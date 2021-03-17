package pkg

import (
	"github.com/gdamore/tcell/v2"
	"github.com/notnil/chess"
	"github.com/rivo/tview"
	"log"
	"os"
)

func remove(s []int, i int) []int {
	s[i] = s[len(s)-1]
	// We do not need to put s[i] at the end, as it will be discarded anyway
	return s[:len(s)-1]
}

func getSquare(f chess.File, r chess.Rank) chess.Square {
	return chess.Square((int(r) * 8) + int(f))
}

func posToSquare(row, col int, flip bool) chess.Square {
	//row = 7 - row
	return chess.Square((numrows-row-1)*8 + col - 1)
}

func squareToColor(sq chess.Square) tcell.Color {
	if (int(sq.File())+int(sq.Rank()))%2 == 0 {
		return tcell.GetColor("#ffffdf") // light
	} else {
		return tcell.GetColor("#dfdfdf") // Dark
	}
}

func GameFromFEN(gamefen string) *chess.Game {
	fen, err := chess.FEN(gamefen)
	if err != nil {
		log.Panic(err)
	}
	game := chess.NewGame(fen, chess.UseNotation(chess.UCINotation{}))
	return game
}

func InitLog(dest, prefix string) {
	f, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	log.SetPrefix(prefix)
}

// Center returns a new primitive which shows the provided primitive in its
// center, given the provided primitive's size.
func Center(width, height int, p tview.Primitive) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}
