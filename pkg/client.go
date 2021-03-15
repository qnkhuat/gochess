package pkg

import (
	"bufio"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/notnil/chess"
	"github.com/rivo/tview"
	"log"
	"net"
	"strings"
	"time"
)

type Client struct {
	Game           *chess.Game
	App            *tview.Application
	Board          *tview.Table
	Layout         *tview.Grid
	Conn           net.Conn
	In             chan MessageInterface
	Out            chan MessageInterface
	selecting      bool
	lastSelection  chess.Square
	highlights     map[chess.Square]bool
	Color          PlayerColor
	ChatTextView   *tview.TextView
	StatusTextView *tview.TextView
}

const (
	numrows             = 8
	numcols             = 8
	numOfSquaresInBoard = 8 * 8
	ConnQueueSize       = 10
)

func NewClient() *Client {
	app := tview.NewApplication()

	In := make(chan MessageInterface, ConnQueueSize)
	Out := make(chan MessageInterface, ConnQueueSize)
	highlights := make(map[chess.Square]bool)
	cl := &Client{
		App:        app,
		Game:       chess.NewGame(chess.UseNotation(chess.UCINotation{})),
		In:         In,
		Out:        Out,
		highlights: highlights,
	}
	cl.InitGUI()

	return cl
}

func (cl *Client) Disconnect() {
	cl.App.Stop()
	cl.Conn.Close()
}

func (cl *Client) InitGUI() {
	// Game options
	drawBtn := tview.NewButton("Draw").SetSelectedFunc(func() {
		cl.Disconnect()
		// Send draw offer
	})

	resignBtn := tview.NewButton("Resign").SetSelectedFunc(func() {
		cl.Disconnect()
		// Send resign
	})

	statusText := tview.NewTextView().
		SetText("This is where message is gonna be")

	gameOptions := tview.NewGrid().
		SetColumns(3, 10, 1, 10, 3).
		SetRows(3, 1, 3, -1).
		AddItem(statusText, 0, 0, 1, 5, 0, 0, false).
		AddItem(drawBtn, 2, 1, 1, 1, 0, 0, false).
		AddItem(resignBtn, 2, 3, 1, 1, 0, 0, false)

	messageInput := tview.NewInputField()
	messageInput.SetLabel("[red]>[red] ").
		SetDoneFunc(func(key tcell.Key) {
			cl.Out <- MessageGameChat{Message: messageInput.GetText(), Time: time.Now()}
			messageInput.SetText("")
		})

	chatTextView := tview.NewTextView().
		SetText("a Message from user\nNgoc dep trai vl").
		SetScrollable(true)

	chatGrid := tview.NewGrid().
		SetColumns(60).
		SetRows(9, 1, 1).
		AddItem(chatTextView, 0, 0, 1, 1, 0, 0, false).
		AddItem(messageInput, 2, 0, 1, 1, 0, 0, false)

	board := tview.NewTable()

	layout := tview.NewGrid().
		SetRows(-1, 10, 11, -1).
		SetColumns(-1, 30, 30, -1).
		AddItem(board, 1, 1, 1, 1, 0, 0, true).
		AddItem(gameOptions, 1, 2, 1, 1, 0, 0, false).
		AddItem(chatGrid, 2, 1, 1, 2, 0, 0, false)

	cl.StatusTextView = statusText
	cl.ChatTextView = chatTextView
	cl.Layout = layout
	cl.Board = board

	cl.init_table()
}

func (cl *Client) init_table() {
	cl.RenderTable()
	cl.Board.SetSelectable(true, true)
	cl.Board.Select(0, 0).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			cl.Disconnect()
		}
		if key == tcell.KeyEnter {
			cl.Board.SetSelectable(true, true)
		}
	}).SetSelectedFunc(func(row, col int) {
		// TODO handle when promoting
		//sq := posToSquare(row, col)
		sq := cl.posToSquare(row, col)
		if cl.selecting {
			if sq == cl.lastSelection { // chose the last move to deactivate
				cl.selecting = false
				cl.lastSelection = 0
				cl.Board.GetCell(row, col).SetBackgroundColor(squareToColor(sq, cl.highlights))
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
		cl.RenderTable() // Not need to if the we have a seperated routine to highlights
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
				cl.Board.SetCell(r, f, cell)
				continue
			}

			if r == numrows && f > 0 { // Draw files square
				file := chess.File(f - 1)
				cell := tview.NewTableCell(fmt.Sprintf(" %s", file.String())).
					SetAlign(tview.AlignCenter).
					SetSelectable(false)
				cl.Board.SetCell(r, f, cell)
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
			cl.Board.SetCell(r, f, cell)
		}
	}
	cl.Board.GetCell(numrows, 0).SetSelectable(false) // The bottom left tile is not used
	go cl.App.Draw()

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
		commandData := Encode(command)
		commandTransport := MessageTransport{MsgType: command.Type(), Data: commandData}
		b := Encode(commandTransport)
		if b[len(b)-1] != '\n' { // EOF
			b = append(b, '\n')
		}
		if _, err := cl.Conn.Write(b); err != nil {
			log.Fatal(err)
		}
		log.Printf("Send a msg type :%s", command.Type())
	}
}

func (cl *Client) HandleRead() {
	scanner := bufio.NewScanner(cl.Conn)
	var messageTransport MessageTransport
	for scanner.Scan() {
		Decode(scanner.Bytes(), &messageTransport)
		log.Printf("Received a message type: %s", messageTransport.MsgType)
		switch messageTransport.MsgType {
		case TypeMessageGame:
			var message MessageGame
			Decode(messageTransport.Data, &message)
			cl.Game = GameFromFEN(message.Fen)
			if message.IsTurn {
				cl.StatusTextView.SetText("Your turn!")
			} else {
				cl.StatusTextView.SetText("Opponent turn!")
			}
			cl.RenderTable()

		case TypeMessageConnect:
			var message MessageConnect
			Decode(messageTransport.Data, &message)
			cl.Game = GameFromFEN(message.Fen)
			cl.Color = message.Color
			if message.IsTurn {
				cl.StatusTextView.SetText("Your turn!")
			} else {
				cl.StatusTextView.SetText("Opponent turn!")
			}

			cl.RenderTable()

		case TypeMessageGameChat:
			var message MessageGameChat
			Decode(messageTransport.Data, &message)
			currentText := cl.ChatTextView.GetText(true)
			displayText := fmt.Sprintf("%s: %s", message.Name, message.Message)
			cl.ChatTextView.SetText(fmt.Sprintf("%s\n%s", currentText, displayText))
			go cl.App.Draw()

		default:
			log.Printf("Received Unknown message")
		}
	}
}
func (cl *Client) posToSquare(row, col int) chess.Square {
	// A1 is square 0
	if cl.Color != Black { // decending order if is white
		row = numrows - row - 1
	}
	col = col - 1 // 1 column for the rank
	return chess.Square(row*8 + col)
}
