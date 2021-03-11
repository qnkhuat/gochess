package main

import (
	"github.com/qnkhuat/chessterm/pkg"
	"log"
)

const (
	numrows             = 8
	numcols             = 8
	numOfSquaresInBoard = 8 * 8
	ServerPort          = ":1998"
)

func main() {
	pkg.InitLog("/Users/earther/fun/7_chessterm/cmd/chessterm/log")

	cl := pkg.NewClient()
	cl.RenderTable()
	log.Println("Connecting")
	cl.Connect(ServerPort)
	log.Println("Connected")

	if err := cl.App.SetRoot(cl.Table, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
