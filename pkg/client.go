package pkg

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/notnil/chess"
	"github.com/rivo/tview"
	"log"
	"net"
	"strconv"
)

type Client struct {
	g             *chess.Game
	selecting     bool
	lastSelection chess.Square
	App           *tview.Application
	Table         *tview.Table
	highlights    map[chess.Square]bool
	Conn          net.Conn
}

const (
	numrows             = 8
	numcols             = 8
	numOfSquaresInBoard = 8 * 8
)

func NewClient() *Client {
	app := tview.NewApplication()
	table := tview.NewTable()
	cl := &Client{
		App:   app,
		g:     chess.NewGame(chess.UseNotation(chess.UCINotation{})),
		Table: table,
	}
	cl.highlights = make(map[chess.Square]bool)
	cl.init_table()
	return cl
}

func (cl *Client) init_table() {
	cl.RenderTable()
	cl.Table.SetSelectable(true, true)
	cl.Table.Select(0, 0).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			cl.App.Stop()
		}
		if key == tcell.KeyEnter {
			cl.Table.SetSelectable(true, true)
		}
	}).SetSelectedFunc(func(row int, col int) {
		// TODO handle when promoting
		sq := posToSquare(row, col)
		log.Printf("Selecting %s, %d, %d", sq.String(), row, col)
		if cl.selecting {
			if sq == cl.lastSelection { // chose the last move to deactivate
				cl.selecting = false
				cl.lastSelection = 0
				cl.Table.GetCell(row, col).SetBackgroundColor(squareToColor(sq, cl.highlights))
				delete(cl.highlights, sq)
			} else { // Chosing destination
				move := fmt.Sprintf("%s%s", cl.lastSelection.String(), sq.String())
				err := cl.g.MoveStr(move)
				if err != nil {
					log.Printf("invalid moves %s: %v", move, err)
					delete(cl.highlights, sq)
					delete(cl.highlights, cl.lastSelection)
					cl.selecting = false
					cl.lastSelection = 0
				} else { // success
					log.Printf("Move: %s", move)
					delete(cl.highlights, cl.lastSelection)
					cl.lastSelection = 0
					cl.selecting = false
				}
			}
		} else {
			cl.highlights[sq] = true
			cl.selecting = true
			cl.lastSelection = sq
		}
		cl.RenderTable()
	})
}

func (cl *Client) RenderTable() {
	board := cl.g.Position().Board()
	var r, f int
	var color tcell.Color
	// Step through the ranks starting with the top row
	for r = 0; r <= numrows; r++ {
		if r != numrows { // Draw numbers square
			cell := tview.NewTableCell(strconv.Itoa(numrows - r)).
				SetAlign(tview.AlignCenter).
				SetSelectable(false)
			cl.Table.SetCell(r, 0, cell)
		}

		// Walk the board
		for f = 1; f <= numcols; f++ {
			file := chess.File(f - 1)
			if r == numrows { // Draw files square
				cell := tview.NewTableCell(fmt.Sprintf(" %s", file.String())).
					SetAlign(tview.AlignCenter).
					SetSelectable(false)
				cl.Table.SetCell(r, f, cell)
				continue
			}
			sq := posToSquare(r, f)
			p := board.Piece(sq)
			ps := fmt.Sprintf(" %s", p.String())
			color = squareToColor(sq, cl.highlights)
			cell := tview.NewTableCell(ps).
				SetAlign(tview.AlignCenter).
				SetBackgroundColor(color)

			cl.Table.SetCell(r, f, cell)
		}
	}
	cl.Table.GetCell(numrows, 0).SetSelectable(false) // The bottom left tile is not used
}

func (cl *Client) Connect(port string) {
	conn, err := net.Dial("tcp", port)
	if err != nil {
		log.Panic(err)
	}
	cl.Conn = conn
}
