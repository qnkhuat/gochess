package pkg

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	ServerIdleTimeout = 5 * time.Minute
	SshPort           = ":2222"
	ServerPort        = ":1998"
	MessageQueueSize  = 20
)

type Server struct {
	*ssh.Server
	Matches map[string]*Match
	Clients []net.Conn
}

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func sshHandle(s ssh.Session) {
	ptyReq, winCh, isPty := s.Pty()
	if !isPty {
		io.WriteString(s, "non-interactive terminals are not supported\n")

		s.Exit(1)
		return
	}

	cmdCtx, cancelCmd := context.WithCancel(s.Context())
	defer cancelCmd()

	cmd := exec.CommandContext(cmdCtx, "/Users/earther/fun/7_chessterm/cmd/chessterm/chessterm")

	cmd.Env = append(s.Environ(), fmt.Sprintf("TERM=%s", ptyReq.Term))

	f, err := pty.Start(cmd)
	if err != nil {
		io.WriteString(s, fmt.Sprintf("failed to initialize pseudo-terminal: %s\n", err))
		s.Exit(1)
		return
	}
	defer f.Close()

	go func() {
		for win := range winCh {
			setWinsize(f, win.Width, win.Height)
		}
	}()

	go func() {
		io.Copy(f, s)
	}()
	io.Copy(s, f)

	f.Close()
	cmd.Wait()

}

func NewServer() *Server {
	s := &ssh.Server{
		Addr:        SshPort,
		IdleTimeout: ServerIdleTimeout,
		Handler:     sshHandle,
	}

	// TODO: understand what does it do?
	homeDir, err := os.UserHomeDir()
	err = s.SetOption(ssh.HostKeyFile(path.Join(homeDir, ".ssh", "id_rsa")))

	if err != nil {
		log.Panic(err)
	}
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()

	matches := make(map[string]*Match)
	clients := make([]net.Conn, 0)
	server := &Server{
		Server:  s,
		Matches: matches,
		Clients: clients,
	}

	return server
}

func (s *Server) AddConn(conn net.Conn, matchId string) {
	if m, ok := s.Matches[matchId]; ok {
		m.AddConn(conn)
		return
	}
	s.Matches[matchId] = NewMatch()
	s.Matches[matchId].AddConn(conn)
}

func (s *Server) HandleConn(conn net.Conn) {

	out := make(chan MessageInterface)
	go func() {
		for {
			for message := range out {
				messageData := Encode(message)
				messageTransport := &MessageTransport{MsgType: message.Type(), Data: messageData}
				b := Encode(messageTransport)
				if b[len(b)-1] != '\n' { // EOF
					b = append(b, '\n')
				}
				if _, err := conn.Write(b); err != nil {
					log.Printf("Failed to write: %v Error: %v", message, err)
				}
			}
		}
	}()

	scanner := bufio.NewScanner(conn)
	var messageTransport MessageTransport
	for scanner.Scan() {
		Decode(scanner.Bytes(), &messageTransport)
		switch messageTransport.MsgType {
		case TypeMessageGameCommand:
			var message MessageGameCommand
			Decode(messageTransport.Data, &message)
			switch message.Command {

			case CommandCreate:
				var matchName string
				if message.Argument == "" {
					matchName = s.NewMatchName()
				} else {
					matchName = message.Argument
				}
				matchName = strings.ToLower(strings.TrimSpace(matchName))
				if !s.IsMatchExisted(matchName) {
					s.AddConn(conn, matchName)
					return
				} else {
					matchName = s.NewMatchName()
					out <- MessageGameCommand{Command: CommandMessage, Argument: fmt.Sprintf("Name existed! How about name it: %s?", matchName)}
				}

			case CommandJoin:
				var matchName string
				matchName = message.Argument

				if s.IsMatchExisted(matchName) {
					s.AddConn(conn, matchName)
					return
				} else {
					out <- MessageGameCommand{Command: CommandMessage, Argument: fmt.Sprintf("Match name %s not existed! type [green]create %s[white] to create one!", matchName, matchName)}
				}

			case CommandLs:
				listMatchString := ""
				for matchName, match := range s.Matches {
					player_count := 0
					viewer_count := 0
					for _, p := range match.Players {
						if p.Role == White || p.Role == Black {
							player_count++
						} else {
							viewer_count++
						}
					}
					listMatchString += fmt.Sprintf("Match: %s (#Player: %d/2, #Viewer: %d)\n", matchName, player_count, viewer_count)
				}
				out <- MessageGameCommand{Command: CommandMessage, Argument: listMatchString}

			default:
				log.Println("Unknown command")
			}
		default:
			log.Println("Unknown message type")
		}
	}
}

func (s *Server) IsMatchExisted(name string) bool {
	_, ok := s.Matches[name]
	return ok
}

func (s *Server) NewMatchName() string {
	// TODO there might be a case when we ran out of countries name, but I'm afraid so lol
	for {
		matchName := randomdata.Country(randomdata.FullCountry)
		if !s.IsMatchExisted(matchName) {
			return matchName
		}
	}
}
