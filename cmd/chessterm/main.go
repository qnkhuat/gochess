package main

import (
	"github.com/qnkhuat/chessterm/pkg"
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
	pkg.InitLog("/Users/earther/fun/7_chessterm/log", "CLIENT: ")

	cl := pkg.NewClient(ServerPort)
	cl.RenderTable()

	if err := cl.App.SetRoot(cl.Table, true).EnableMouse(true).Run(); err != nil {
		panic(err)
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
