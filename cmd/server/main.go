package main

import (
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
	pkg.InitLog("/Users/earther/fun/7_chessterm/log", "SERVER: ")
	log.Println("Server started")
	s = pkg.NewServer()

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
