package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"github.com/notnil/chess"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"
	"unsafe"
)

const (
	ServerIdleTimeout = 1 * time.Minute
	SshPort           = ":2222"
	ServerPort        = ":1998"
	MessageQueueSize  = 20
)

type Server struct {
	*ssh.Server
	Game    *chess.Game
	Players []*Player
	Turn    PlayerColor
	// TODO : find out a better way to generic this
	In  chan MessageTransport
	Out chan MessageGame
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

	In := make(chan MessageTransport, MessageQueueSize)
	Out := make(chan MessageGame, MessageQueueSize)
	game := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	server := &Server{
		Server: s,
		Game:   game,
		Turn:   White, // White move first
		In:     In,
		Out:    Out,
	}
	go server.HandleWrite()
	go server.HandleRead()

	return server
}

func (s *Server) AddPlayer(p *Player) {
	var color PlayerColor
	num_players := len(s.Players)
	if num_players == 0 {
		color = White
	} else if num_players == 1 {
		color = Black
	} else {
		color = Viewer
	}
	p.Color = color
	p.Id = len(s.Players)
	log.Printf("Added a Player: %s", p.Color)
	s.Players = append(s.Players, p)
}

func (s *Server) GameFEN() string {
	return s.Game.Position().String()
}

func (s *Server) HandleWrite() {
	for messageTransport := range s.In {
		switch messageTransport.MsgType {
		case TypeMessageMove:
			var message MessageMove
			err := json.Unmarshal(messageTransport.Data, &message)
			if err != nil {
				log.Panic(err)
			}
			// Validate if the sender is the one who allowed to move
			if s.Players[messageTransport.PlayerId].Color == s.Turn {
				s.Game.MoveStr(message.Move)
				// Switch turn
				if s.Turn == White {
					s.Turn = Black
				} else {
					s.Turn = White
				}
				s.Out <- MessageGame{Fen: s.GameFEN()}
			} else {
				log.Println("Not your turn bro")
			}

		default:
			log.Printf("Received Unknown message")
		}
	}
}

func (s *Server) HandleRead() {
	for message := range s.Out {
		switch message.Type() {
		case TypeMessageGame:
			log.Println("pay attention")
			for _, p := range s.Players { // Broadcast the game to all users
				message.IsTurn = p.Color == s.Turn
				p.Out <- message
			}
		default:
			log.Println("Received Unknown message")
		}
	}
}