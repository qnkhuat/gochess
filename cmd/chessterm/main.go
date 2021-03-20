package main

import (
	"flag"
	"github.com/qnkhuat/gochess/pkg"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	numrows             = 8
	numcols             = 8
	numOfSquaresInBoard = 8 * 8
	ServerPort          = ":1998"
)

var (
	done = make(chan bool)
)

func main() {
	logPath := flag.String("log", "./log", "path to log file")
	flag.Parse()
	pkg.InitLog(*logPath, "CLIENT: ")

	log.Println("New Client")
	cl := pkg.NewClient()
	cl.Connect(ServerPort)
	go cl.HandleRead()
	go cl.HandleWrite()
	if err := cl.App.SetRoot(cl.MenuLayout, true).EnableMouse(true).Run(); err != nil {
		done <- true
	}

	// Keep the client up
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM)
	go func() { // Down when receive killed signal
		<-sigc

		done <- true
	}()

	go func() { // Down when self-killed
		<-done
		cl.Disconnect()
		os.Exit(0)
	}()
}
