package pkg

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"github.com/notnil/chess/uci"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	ServerIdleTimeout = 30 * time.Minute
	ServerPort        = ":1998"
	MessageQueueSize  = 20
)

var (
	ChesstermBinary string
	LogPath         string
	SshPort         = ":2222"
)

type Server struct {
	*ssh.Server
	Matches map[string]*Match
	Clients []net.Conn
	Engine  *uci.Engine
	In      chan MessageInterface
	Out     chan MessageInterface
}

type ServerConn struct {
	Conn net.Conn
	Name string
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

	cmd := exec.CommandContext(cmdCtx, ChesstermBinary, "-log", LogPath)

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

func NewServer(binary string, sshPort string, logPath string) *Server {
	SshPort = sshPort
	ChesstermBinary = binary // path to chess term to open it
	LogPath = logPath
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

	// for single player mode
	eng, err := uci.New("stockfish")
	if err != nil {
		panic(err)
	}
	in := make(chan MessageInterface, MessageQueueSize)
	out := make(chan MessageInterface, MessageQueueSize)

	matches := make(map[string]*Match)
	clients := make([]net.Conn, 0)
	server := &Server{
		Server:  s,
		Matches: matches,
		Clients: clients,
		Engine:  eng,
		In:      in,
		Out:     out,
	}

	return server
}

func (s *Server) AddConn(conn net.Conn, matchId, name string, duration, increment int) {
	if name == "" {
		name = randomdata.SillyName()
	}
	if m, ok := s.Matches[matchId]; ok {
		m.AddConn(conn, name)
		return
	}
	s.Matches[matchId] = NewMatch(matchId, false, duration, increment)
	s.Matches[matchId].AddConn(conn, name)
}

func (s *Server) HandleConn(sconn ServerConn) {
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
				if _, err := sconn.Conn.Write(b); err != nil {
					log.Printf("Failed to write: %v Error: %v", message, err)
				}
			}
		}
	}()

	scanner := bufio.NewScanner(sconn.Conn)
	var messageTransport MessageTransport
	for scanner.Scan() {
		Decode(scanner.Bytes(), &messageTransport)
		switch messageTransport.MsgType {
		case TypeMessageGameCommand:
			var message MessageGameCommand
			Decode(messageTransport.Data, &message)
			switch message.Command {

			case CommandPractice:
				var level int
				matchId := s.NewMatchName()
				if len(message.Argument) > 0 {
					level, _ = strconv.Atoi(message.Argument[0])
				} else {
					level = 2
				}
				if sconn.Name == "" {
					sconn.Name = randomdata.SillyName()
				}

				s.Matches[matchId] = NewMatch(matchId, true, 30, 0)
				s.Matches[matchId].Engine = s.Engine
				s.Matches[matchId].AddConn(sconn.Conn, sconn.Name)
				s.Matches[matchId].PracticeLevel = level
				return

			case CommandCreate:
				var matchName string
				duration := 10 // default 10 minutes
				increment := 0 // default is 0 second
				if len(message.Argument) > 0 {
					matchName = message.Argument[0]
				} else {
					matchName = s.NewMatchName()
				}

				if len(message.Argument) > 1 {
					duration, _ = strconv.Atoi(message.Argument[1])
				}

				if len(message.Argument) > 2 {
					increment, _ = strconv.Atoi(message.Argument[2])
				}

				matchName = strings.ToLower(strings.TrimSpace(matchName))
				if !s.IsMatchExisted(matchName) {
					s.AddConn(sconn.Conn, matchName, sconn.Name, duration, increment)
					return
				} else {
					matchName = s.NewMatchName()
					out <- MessageGameCommand{Command: CommandMessage, Argument: []string{fmt.Sprintf("Name existed! How about name it: %s?", matchName)}}
				}

			case CommandJoin:
				var matchName string
				matchName = message.Argument[0]

				//if matchName == "" { // join random
				if len(message.Argument) == 0 { // join random
					for matchId, match := range s.Matches {
						if len(match.Players) < 2 && !match.PracticeMode {
							s.AddConn(sconn.Conn, matchId, sconn.Name, -1, 0)
							return
						}
					}
					out <- MessageGameCommand{Command: CommandMessage, Argument: []string{"No match available! Create one and invite your friend ^^!"}}

				} else if s.IsMatchExisted(message.Argument[0]) {
					s.AddConn(sconn.Conn, message.Argument[0], sconn.Name, -1, 0)
					return
				} else {
					out <- MessageGameCommand{Command: CommandMessage, Argument: []string{fmt.Sprintf("Match name %s not existed! type [green]create %s[white] to create one!", matchName, matchName)}}
				}
			case CommandCallme:
				sconn.Name = message.Argument[0]
				out <- MessageGameCommand{Command: CommandMessage, Argument: []string{fmt.Sprintf("[green]%s[white] it is!", strings.Title(sconn.Name))}}

			case CommandLs:
				listMatchString := "Matches list:\n"
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
					listMatchString += fmt.Sprintf("Match: [red]%s[white] (#Player: %d/2, #Viewer: %d)\n", matchName, player_count, viewer_count)
				}
				if len(s.Matches) == 0 {
					listMatchString = "No match found :( Let's create one 🌝"
				}

				out <- MessageGameCommand{Command: CommandMessage, Argument: []string{listMatchString}}

			default:
				log.Println("Unknown command")
			}
		default:
			log.Printf("Unknown message type: %v", messageTransport.MsgType)
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

func (s *Server) CleanIdleMatches() {
	tick := time.NewTicker(1 * time.Minute)
	for {
		connection_count := 0
		select {
		case <-tick.C:
			for key, m := range s.Matches {
				connection_count += len(m.Players)
				if len(m.Players) == 0 {
					delete(s.Matches, key)
					log.Printf("Deleted match: %s", key)
				}
			}
			log.Printf("Connection count: %d", connection_count)
		}
	}
}
