package pkg

import (
	"fmt"
	"github.com/notnil/chess"
	"github.com/notnil/chess/uci"
	"log"
	"math"
	"net"
	"strconv"
	"strings"
	"time"
)

type Match struct {
	//Players [2]*Player
	Players       map[int]*Player
	Game          *chess.Game
	Server        Server
	Turn          PlayerRole
	In            chan MessageInterface
	Out           chan MessageInterface
	Name          string
	PracticeMode  bool
	Engine        *uci.Engine
	PracticeLevel int
	Duration      time.Duration
	Increment     time.Duration
}

func NewGame() *chess.Game {
	return chess.NewGame(chess.UseNotation(chess.UCINotation{}))
}

func NewMatch(name string, practiceMode bool, duration, increment int) *Match {
	game := NewGame()
	in := make(chan MessageInterface, MessageQueueSize)
	out := make(chan MessageInterface, MessageQueueSize)
	players := make(map[int]*Player)

	match := &Match{
		Name:          name,
		In:            in,
		Out:           out,
		Players:       players,
		Game:          game,
		Turn:          White, // White move first
		PracticeMode:  practiceMode,
		PracticeLevel: 2, // Default level for hardress in single player mode
		Duration:      time.Duration(duration) * time.Minute,
		Increment:     time.Duration(increment) * time.Second,
	}

	go match.HandleRead()
	return match
}

func (m *Match) ReMatch() {
	m.Game = NewGame()
	m.Turn = White
}

func (m *Match) GameFEN() string {
	return m.Game.Position().String()
}

func (m *Match) GameMoves() []string {
	var moves []string
	for _, move := range m.Game.Moves() {
		moves = append(moves, move.String())
	}
	return moves
}

func (m *Match) isPlayer() bool {
	if len(m.Players) <= 2 {
		return true
	}
	return false
}

func (m *Match) availableRole() PlayerRole {
	if _, ok := m.Players[int(White)]; !ok {
		return White
	} else if _, ok := m.Players[int(Black)]; !ok {
		return Black
	}
	return Viewer
}

func (m *Match) AddConn(conn net.Conn, name string) {
	p := NewPlayer(conn, name)

	role := m.availableRole()
	p.Role = role
	// Id of white, black player is unique, Viewer instead can have as many as we want
	if role == Black || role == White {
		p.Id = int(role)
		go p.HandleRead(m.In)
	} else {
		p.Id = int(Viewer) + len(m.Players)
	}
	m.Players[p.Id] = p

	go p.HandleWrite()

	// Connect player to the game
	p.Out <- MessageConnect{
		Fen:       m.GameFEN(),
		IsTurn:    m.Turn == p.Role,
		Role:      p.Role,
		Duration:  m.Duration,
		Increment: m.Increment,
	}

	// Broadcast new player for all player in the game
	for id, pl := range m.Players {
		if id == p.Id {
			continue
		}
		pl.Out <- MessageGameChat{
			Message: fmt.Sprintf("[gray]Player [green]%s[gray] has joined[white]", strings.Title(p.Name)),
		}
	}

	p.Out <- MessageGameChat{
		Message: fmt.Sprintf(`[gray]You have joined room [red]%s[gray] as [red]%s[gray] player with name [green]%s[white].
To move piece: [green]click[white] on piece to select and [green]click[white] again on destination
Also, you might want to zoom in to see the pieces clearer! Have fun :)
`, m.Name, p.Role, strings.Title(p.Name)),
	}

	log.Printf("Added a Player: %s", p.Role)
}

func (m *Match) HandleRead() {
	for inMessage := range m.In {
		messageTransport := inMessage.(MessageTransport)
		switch messageTransport.MsgType {

		case TypeMessageMatchRemovePlayer:
			var message MessageMatchRemovePlayer
			Decode(messageTransport.Data, &message)
			m.Players[message.PlayerId].Disconnect()
			delete(m.Players, message.PlayerId)

		case TypeMessageMove:
			var message MessageMove
			Decode(messageTransport.Data, &message)
			// Validate if the sender is the one who allowed to move
			if m.Players[messageTransport.PlayerId].Role == m.Turn {
				m.Game.MoveStr(message.Move)
				// Switch turn
				if m.Turn == White {
					m.Turn = Black
				} else {
					m.Turn = White
				}
				log.Println("Out moves:")
				message := MessageGame{Fen: m.GameFEN(), Moves: m.GameMoves()}
				for _, p := range m.Players { // Broadcast the game to all users
					message.IsTurn = p.Role == m.Turn
					p.Out <- message
				}

				// Practice mode will move immediately after client move
				if m.PracticeMode {
					time.Sleep(time.Second / 2) // Fake processing time
					m.Turn = White              // Player is always white
					m.Game.MoveStr(m.NextMove())
					message := MessageGame{Fen: m.GameFEN(), Moves: m.GameMoves()}
					for _, p := range m.Players { // Broadcast the game to all users
						message.IsTurn = p.Role == m.Turn
						p.Out <- message
					}
					log.Println(m.Game.Moves())
				}

				if m.Game.Outcome() != chess.NoOutcome {
					for _, p := range m.Players { // Broadcast the game to all users
						if (p.Role == White && m.Game.Outcome() == chess.WhiteWon) || (p.Role == Black && m.Game.Outcome() == chess.BlackWon) {
							p.Out <- MessageGameAction{Action: ActionWin, Message: m.Game.Method().String()}
						} else {
							p.Out <- MessageGameAction{Action: ActionLose, Message: m.Game.Method().String()}
						}
					}
				}
			}
		case TypeMessageGameChat:
			var message MessageGameChat
			Decode(messageTransport.Data, &message)

			var senderName string
			if m.Players[messageTransport.PlayerId].Name != "" {
				senderName = m.Players[messageTransport.PlayerId].Name
			} else {
				senderName = fmt.Sprintf("ID[%v]", strconv.Itoa(messageTransport.PlayerId))
			}
			message.Name = senderName
			for _, p := range m.Players { // Broadcast the game to all users
				p.Out <- message
			}
		case TypeMessageGameAction:
			var message MessageGameAction
			Decode(messageTransport.Data, &message)
			switch message.Action {
			case ActionResignYes:
				for _, p := range m.Players {
					if p.Id == messageTransport.PlayerId {
						p.Out <- MessageGameAction{Action: ActionLose, Message: "by Resignation"}
					} else {
						p.Out <- MessageGameAction{Action: ActionWin, Message: "by Resigination"}
					}
				}

			case ActionTimeOut:
				log.Println("Got the time out")
				for _, p := range m.Players {
					if p.Id == messageTransport.PlayerId {
						p.Out <- MessageGameAction{Action: ActionLose, Message: "by Time Out"}
					} else {
						p.Out <- MessageGameAction{Action: ActionWin, Message: "by Time Out"}
					}
				}

			case ActionDrawOffer:
				if m.PracticeMode {
					for _, p := range m.Players {
						p.Out <- MessageGameStatus{Message: "Rejected draw offer"}
					}
				}
				for _, p := range m.Players {
					if p.Id != messageTransport.PlayerId {
						p.Out <- MessageGameAction{Action: ActionDrawOffer}
						p.Out <- MessageGameStatus{Message: "Opponent offer draw!"}
					}
				}

			case ActionDrawAccept:
				for _, p := range m.Players {
					p.Out <- MessageGameAction{Action: ActionDraw}
				}

			case ActionDrawReject:
				for _, p := range m.Players {
					if p.Id != messageTransport.PlayerId {
						p.Out <- MessageGameStatus{Message: "Rejected draw offer"}
					}
				}

			// New Game
			case ActionNewGameOffer:
				if m.PracticeMode {
					m.ReMatch()
					message := MessageGame{Fen: m.GameFEN()}
					for _, p := range m.Players {
						message.IsTurn = p.Role == m.Turn
						p.Out <- message
					}

				} else {
					for _, p := range m.Players {
						if p.Id != messageTransport.PlayerId {
							p.Out <- MessageGameAction{Action: ActionNewGameOffer}
							p.Out <- MessageGameStatus{Message: "New Game?"}
						}
					}
				}

			case ActionNewGameAccept:
				m.ReMatch()
				message := MessageGame{Fen: m.GameFEN()}
				for _, p := range m.Players {
					// TODO: switch color
					message.IsTurn = p.Role == m.Turn
					p.Out <- message
				}

			case ActionNewGameReject:
				for _, p := range m.Players {
					if p.Id != messageTransport.PlayerId {
						p.Out <- MessageGameStatus{Message: "Rejected New Game offer"}
					}
				}

			// Exit
			case ActionExit:
				for _, p := range m.Players {
					if p.Id != messageTransport.PlayerId {
						p.Out <- MessageGameStatus{Message: "Opponent exited!"}
					}
				}

			default:
				log.Printf("Received Unknown message")
			}
		}
	}
}

func (m *Match) NextMove() string { // used for singple player mode
	cmdPos := uci.CmdPosition{Position: m.Game.Position()}
	cmdGo := uci.CmdGo{MoveTime: time.Second / time.Duration(200/math.Pow(float64(m.PracticeLevel), 2.))} // the higher the level the longer the compute
	if err := m.Engine.Run(cmdPos, cmdGo); err != nil {
		panic(err)
	}
	move := m.Engine.SearchResults().BestMove
	return move.String()
}
