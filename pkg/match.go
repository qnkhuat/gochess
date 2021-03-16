package pkg

import (
	"fmt"
	"github.com/notnil/chess"
	"log"
	"net"
	"strconv"
)

type Match struct {
	//Players [2]*Player
	Players map[int]*Player
	Game    *chess.Game
	Server  Server
	Turn    PlayerRole
	In      chan MessageInterface
	Out     chan MessageInterface
}

func NewGame() *chess.Game {
	return chess.NewGame(chess.UseNotation(chess.UCINotation{}))
}

func NewMatch() *Match {
	game := NewGame()
	in := make(chan MessageInterface, MessageQueueSize)
	out := make(chan MessageInterface, MessageQueueSize)
	players := make(map[int]*Player)

	match := &Match{
		In:      in,
		Out:     out,
		Players: players,
		Game:    game,
		Turn:    White, // White move first

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
func (m *Match) AddConn(conn net.Conn) {
	p := NewPlayer(conn)

	role := m.availableRole()
	p.Role = role
	// Id of white,black player is unique, Viewer instead can have as many as we want
	if role == Black || role == White {
		p.Id = int(role)
		go p.HandleRead(m.In)
	} else {
		p.Id = int(Viewer) + len(m.Players)
	}
	m.Players[p.Id] = p

	go p.HandleWrite()
	p.Out <- MessageConnect{
		Fen:    m.GameFEN(),
		IsTurn: m.Turn == p.Role,
		Role:   p.Role,
	}

	message := MessageGameChat{
		Message: fmt.Sprintf("[grey]%s has joined[white]", p.Role),
		Name:    "Server",
	}
	m.In <- MessageTransport{MsgType: message.Type(), Data: Encode(message)}

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
				message := MessageGame{Fen: m.GameFEN()}
				for _, p := range m.Players { // Broadcast the game to all users
					message.IsTurn = p.Role == m.Turn
					p.Out <- message
				}

				if m.Game.Outcome() != chess.NoOutcome {
					for _, p := range m.Players { // Broadcast the game to all users
						if (p.Role == White && m.Game.Outcome() == chess.WhiteWon) || (p.Role == Black && m.Game.Outcome() == chess.BlackWon) {
							p.Out <- MessageGameAction{Action: ActionWin}
						} else {
							p.Out <- MessageGameAction{Action: ActionLose}
						}
					}
				}
			}
		case TypeMessageGameChat:
			var message MessageGameChat
			Decode(messageTransport.Data, &message)

			var senderName string
			if message.Name == "Server" { // Broadcast message from server
				senderName = message.Name
			} else if m.Players[messageTransport.PlayerId].Name != "" {
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
						p.Out <- MessageGameAction{Action: ActionLose, Message: "by resignation"}
					} else {
						p.Out <- MessageGameAction{Action: ActionWin, Message: "by resigination"}
					}
					// TODO handle case for viewer
				}

			// Draw
			case ActionDrawOffer:
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
				for _, p := range m.Players {
					if p.Id != messageTransport.PlayerId {
						p.Out <- MessageGameAction{Action: ActionNewGameOffer}
						p.Out <- MessageGameStatus{Message: "New Game?"}
					}
				}

			case ActionNewGameAccept:
				m.ReMatch()
				message := MessageGame{Fen: m.GameFEN()}
				for _, p := range m.Players {
					// TODO switch color
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
