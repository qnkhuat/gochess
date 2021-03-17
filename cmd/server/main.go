package main

import (
	"flag"
	"github.com/qnkhuat/chessterm/pkg"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var (
	s    *pkg.Server
	done = make(chan bool)
)

func main() {
	logPath := flag.String("log", "~/log", "path to log file")
	binaryPath := flag.String("binary", "../chessterm/chessterm", "path to chessterm binary")
	sshPort := flag.String("ssh", ":2222", "port to ssh")
	flag.Parse()
	pkg.InitLog(*logPath, "SERVER: ")
	log.Println("Server started")
	s = pkg.NewServer(*binaryPath, *sshPort)

	go s.CleanIdleMatches()

	// Create server to listen for data
	listener, err := net.Listen("tcp", pkg.ServerPort)
	log.Printf("Listening at port %s", pkg.ServerPort)
	defer listener.Close()
	if err != nil {
		log.Panic(err)
	}
	for {
		conn, err := listener.Accept()
		sconn := pkg.ServerConn{Conn: conn}
		if err != nil {
			log.Println("Failed to connect %v", err)
			continue
		}
		go s.HandleConn(sconn)
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
