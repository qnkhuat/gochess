package pkg

//type Match struct {
//	Players [2]*Player
//	Viewers []*Player
//	Game    *chess.Game
//	Server  Server
//	Turn    PlayerColor
//}
//
//func NewGame() *chess.Game {
//	return chess.NewGame(chess.UseNotation(chess.UCINotation{}))
//}
//
//func NewMatch() *Match {
//	game := NewGame()
//
//	match := &Match{}
//	return match
//
//}
//
//func (m *Match) AddPlayer(p *Player) {
//	var color PlayerColor
//	num_players := len(m.Players)
//	if num_players == 0 {
//		color = White
//	} else if num_players == 1 {
//		color = Black
//	} else {
//		color = Viewer
//	}
//	p.Color = color
//	p.Id = len(m.Players)
//	m.Players = append(m.Players, p)
//
//	go p.HandleWrite()
//	go p.HandleRead(m.In)
//	p.Out <- MessageConnect{
//		Fen:    m.GameFEN(),
//		IsTurn: m.Turn == p.Color,
//		Color:  p.Color,
//	}
//
//	m := MessageGameChat{
//		Message: fmt.Sprintf("[grey]Player %s has joined[white]", p.Color),
//		Name:    "Server",
//	}
//	messageData := Encode(m)
//	m.In <- MessageTransport{MsgType: m.Type(), Data: messageData}
//
//	log.Printf("Added a Player: %s", p.Color)
//}
