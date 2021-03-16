package pkg

import (
	"context"
	"fmt"
	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
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
	server := &Server{
		Server:  s,
		Matches: matches,
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
	return
}
