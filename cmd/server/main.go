package main

import (
	"bufio"
	"fmt"
	"strconv"
	"io"
	"log"
	"os"
	"time"
	"path"
	"os/exec"
	"os/signal"
	"syscall"
	"unsafe"
	"context"
	"net"

	"github.com/gliderlabs/ssh"
	"github.com/creack/pty"
)

var (
	ServerIdleTimeout = 1 * time.Minute
	SshPort = ":2222"
	done = make(chan bool)
	count = 0
)

type Server struct {
	*ssh.Server
	Logger chan string

}


func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
	uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func sshHandle(s ssh.Session){
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

func (s *Server) Log(msg string){
	s.Logger <- msg
}

func NewServer() *Server{
	s:= &ssh.Server{
		Addr: SshPort,
		IdleTimeout: ServerIdleTimeout,
		Handler: sshHandle,
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
	server := &Server{s, nil}

	return server
}

func sendMsg(msg string, conn net.Conn){
	msg = msg + "\n"
	if _, err := io.WriteString(conn, msg); err != nil {
		log.Fatal(err)
	}
}


func handleConn(conn net.Conn){
	// Handle requests
	sendMsg(strconv.Itoa(count), conn)
	defer conn.Close()
	input := bufio.NewScanner(conn)
	for input.Scan() {
		switch input.Text() {
		case "inc":
			count++
			go sendMsg(strconv.Itoa(count), conn)
		case "dec":
			count--
			go sendMsg(strconv.Itoa(count), conn)
		default:
			fmt.Printf("Invalid command\n")
		}
		fmt.Printf("Count now: %d\n", count)
	}
}


func main() {
	s := NewServer()

	// Setup Logger
	logger := make(chan string, 10)
	go func() {
		for msg := range logger {
			log.Println(time.Now().Format("3:04:05") + " " + msg)
		}
	}()

	s.Logger = logger
	s.Log("Server started")

	// Create server to listen for data
	listener, err := net.Listen("tcp", ":1998")
	defer listener.Close()
	if err != nil {
		log.Panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to connect %v", err)
			continue
		}
		log.Println("New connection on port :1998")
		go handleConn(conn)
	}

	// Keep the server run
	sigc := make(chan os.Signal, 1)
	// Wait for teminate signal
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM)
	go func() {
		<-sigc

		done <- true
	}()

	<-done
	
}
