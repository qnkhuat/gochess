package main

import (
	pkg "github.com/qnkhuat/chessterm/pkg"
)

const (
	numrows             = 8
	numcols             = 8
	numOfSquaresInBoard = 8 * 8
)

func main() {
	pkg.Init_log("log")

	cl := pkg.NewClient()
	cl.RenderTable()

	if err := cl.App().SetRoot(cl.Table(), true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}
