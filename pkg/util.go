package pkg

import (
	"github.com/gdamore/tcell/v2"
	"github.com/notnil/chess"
	"log"
	"os"
)

func getSquare(f chess.File, r chess.Rank) chess.Square {
	return chess.Square((int(r) * 8) + int(f))
}

func posToSquare(row, col int, flip bool) chess.Square {
	//row = 7 - row
	return chess.Square((numrows-row-1)*8 + col - 1)
}

func squareToColor(sq chess.Square, highlights map[chess.Square]bool) tcell.Color {
	if hl, ok := highlights[sq]; ok && hl {
		return tcell.ColorRed
	} else if (int(sq.File())+int(sq.Rank()))%2 == 0 {
		return tcell.ColorBlue
	} else {
		return tcell.ColorGreen
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
