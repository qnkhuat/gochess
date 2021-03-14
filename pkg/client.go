package pkg

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/notnil/chess"
	"github.com/rivo/tview"
	"log"
	"net"
	"os"
	"strings"
)

type Client struct {
	Game          *chess.Game
	App           *tview.Application
	Table         *tview.Table
	Conn          net.Conn
	In            chan MessageInterface
	Out           chan MessageInterface
	selecting     bool
	lastSelection chess.Square
	highlights    map[chess.Square]bool
	Color         PlayerColor
}

const (
	numrows             = 8
	numcols             = 8
	numOfSquaresInBoard = 8 * 8
	ConnQueueSize       = 10
)

func NewClient() *Client {
	app := tview.NewApplication()
	table := tview.NewTable()
	In := make(chan MessageInterface, ConnQueueSize)
	Out := make(chan MessageInterface, ConnQueueSize)
	cl := &Client{
		App:   app,
		Game:  chess.NewGame(chess.UseNotation(chess.UCINotation{})),
		Table: table,
		In:    In,
		Out:   Out,
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
			cl.Conn.Close()
			os.Exit(0)
		}
		if key == tcell.KeyEnter {
			cl.Table.SetSelectable(true, true)
		}
	}).SetSelectedFunc(func(row, col int) {
		// TODO handle when promoting
		//sq := posToSquare(row, col)
		sq := cl.posToSquare(row, col)
		if cl.selecting {
			if sq == cl.lastSelection { // chose the last move to deactivate
				cl.selecting = false
				cl.lastSelection = 0
				cl.Table.GetCell(row, col).SetBackgroundColor(squareToColor(sq, cl.highlights))
				delete(cl.highlights, sq)
			} else { // Chosing destination
				move := fmt.Sprintf("%s%s", cl.lastSelection.String(), sq.String())
				validMoves := cl.Game.ValidMoves()
				isValid := false
				for _, validMove := range validMoves {
					if strings.Compare(validMove.String(), move) == 0 {
						isValid = true
					}
				}
				if !isValid {
					log.Printf("invalid moves %s", move)
					delete(cl.highlights, sq)
					delete(cl.highlights, cl.lastSelection)
					cl.selecting = false
					cl.lastSelection = 0
				} else { // success
					log.Printf("Move: %s", move)
					cl.Out <- MessageMove{Move: move, Msg: "Hi"}
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
		cl.RenderTable() // Not need to
	})
}

func (cl *Client) RenderTable() {
	board := cl.Game.Position().Board()
	var (
		r, f  int
		color tcell.Color
	)
	// Step through the ranks starting with the top row
	for r = 0; r <= numrows; r++ {
		// Each column
		for f = 0; f <= numcols; f++ {
			if f == 0 && r != numrows { // draw rank square
				var rank chess.Rank
				if cl.Color == White {
					rank = chess.Rank(numrows - r - 1)
				} else {
					rank = chess.Rank(r)
				}
				cell := tview.NewTableCell(rank.String()).
					SetAlign(tview.AlignCenter).
					SetSelectable(false)
				cl.Table.SetCell(r, f, cell)
				continue
			}

			if r == numrows && f > 0 { // Draw files square
				file := chess.File(f - 1)
				cell := tview.NewTableCell(fmt.Sprintf(" %s", file.String())).
					SetAlign(tview.AlignCenter).
					SetSelectable(false)
				cl.Table.SetCell(r, f, cell)
				continue
			}

			if r == numrows && f == 0 {
				continue
			}

			// Draw the pieces

			sq := cl.posToSquare(r, f)
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
	go func() {
		cl.App.Draw()
	}()

}

func (cl *Client) Connect(port string) {
	log.Printf("Connecting to port: %s", port)
	conn, err := net.Dial("tcp", port)
	if err != nil {
		log.Panic(err)
	}
	cl.Conn = conn
}

func (cl *Client) HandleWrite() {
	for command := range cl.Out {
		commandData := command.Encode()
		commandTransport := MessageTransport{MsgType: command.Type(), Data: commandData}
		b := commandTransport.Encode()
		if b[len(b)-1] != '\n' { // EOF
			b = append(b, '\n')
		}
		if _, err := cl.Conn.Write(b); err != nil {
			log.Fatal(err)
		}
		log.Printf("Send a msg: %v with type :%s", b, command.Type())
	}
}

func (cl *Client) HandleRead() {
	scanner := bufio.NewScanner(cl.Conn)
	var messageTransport MessageTransport
	for scanner.Scan() {
		err := json.Unmarshal(scanner.Bytes(), &messageTransport)
		if err != nil {
			log.Panic(err)
		}
		switch messageTransport.MsgType {
		case TypeMessageGame:
			var message MessageGame
			err := json.Unmarshal(messageTransport.Data, &message)
			if err != nil {
				log.Panic(err)
			}
			cl.Game = GameFromFEN(message.Fen)
			cl.RenderTable()

		case TypeMessageConnect:
			var message MessageConnect
			err := json.Unmarshal(messageTransport.Data, &message)
			if err != nil {
				log.Panic(err)
			}
			cl.Game = GameFromFEN(message.Fen)
			cl.Color = message.Color
			cl.RenderTable()

		default:
			log.Printf("Received Unknown message")
		}
	}
}
func (cl *Client) posToSquare(row, col int) chess.Square {
	// A1 is square 0
	if cl.Color == White { // decending order if is white
		row = numrows - row - 1
	}
	col = col - 1 // 1 column for the rank
	return chess.Square(row*8 + col)
}
