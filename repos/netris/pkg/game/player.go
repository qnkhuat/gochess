package game

import (
	"regexp"
	"strconv"

	"git.sr.ht/~tslocum/netris/pkg/mino"
)

const (
	CommandQueueSize = 10
	LogQueueSize     = 10
	PlayerHost       = -1
	PlayerUnknown    = 0
)

type Player struct {
	Name string

	*ServerConn

	Score   int
	Preview *mino.Matrix
	Matrix  *mino.Matrix

	pendingGarbage       int
	totalGarbageSent     int
	totalGarbageReceived int
}

type ConnectingPlayer struct {
	Client ClientInterface

	Name string
}

func NewPlayer(name string, conn *ServerConn) *Player {
	/*in := make(chan *GameCommand, CommandQueueSize)
	out := make(chan *GameCommand, CommandQueueSize)

	p := &Player{Conn: conn, In: in, Out: out}

	go p.handleRead()
	go p.handleWrite()*/

	if conn == nil {
		conn = &ServerConn{}
	}

	p := &Player{Name: Nickname(name), ServerConn: conn}

	return p
}

/*
func (p *Player) handleRead() {
	if p.Conn == nil {
		return
	}

	scanner := bufio.NewScanner(p.Conn)
	for scanner.Scan() {
		log.Println("unmarshal [" + scanner.Text() + "]")

		var gameCommand GameCommand
		err := json.Unmarshal(scanner.Bytes(), &gameCommand)
		if err != nil {
			panic(err)
		}

		p.In <- &gameCommand

		log.Println("read player ")
	}
}

func (p *Player) handleWrite() {
	if p.Conn == nil {
		for range p.Out {
		}
		return
	}

	var (
		j   []byte
		err error
	)
	for e := range p.Out {
		j, err = json.Marshal(e)
		if err != nil {
			log.Printf("attempting to marshal %+v", e)
			panic(err)
		}
		j = append(j, '\n')
		_, err = p.Conn.Write(j)
		if err != nil {
			p.Conn.Close()
		}
	}
}*/

type ClientInterface interface {
	Attach(in chan<- GameCommandInterface, out <-chan GameCommandInterface)
	Detach(reason string)
}

type Command int

func (c Command) String() string {
	switch c {
	case CommandUnknown:
		return "Unknown"
	case CommandDisconnect:
		return "Disconnect"
	case CommandNickname:
		return "Nickname"
	case CommandMessage:
		return "Message"
	case CommandNewGame:
		return "NewGame"
	case CommandJoinGame:
		return "JoinGame"
	case CommandQuitGame:
		return "QuitGame"
	case CommandUpdateGame:
		return "UpdateGame"
	case CommandStartGame:
		return "StartGame"
	case CommandGameOver:
		return "GameOver"
	case CommandUpdateMatrix:
		return "UpdateMatrix"
	case CommandSendGarbage:
		return "Garbage-OUT"
	case CommandReceiveGarbage:
		return "Garbage-IN"
	default:
		return strconv.Itoa(int(c))
	}
}

// The order of these constants must be preserved
const (
	CommandUnknown Command = iota
	CommandDisconnect
	CommandPing
	CommandPong
	CommandNickname
	CommandMessage
	CommandNewGame
	CommandJoinGame
	CommandQuitGame
	CommandUpdateGame
	CommandStartGame
	CommandGameOver
	CommandUpdateMatrix
	CommandSendGarbage
	CommandReceiveGarbage
)

type GameCommand struct {
	SourcePlayer int
}

func (gc *GameCommand) Source() int {
	if gc == nil {
		return 0
	}

	return gc.SourcePlayer
}

func (gc *GameCommand) SetSource(source int) {
	if gc == nil {
		return
	}

	gc.SourcePlayer = source
}

type GameCommandInterface interface {
	Command() Command
	Source() int
	SetSource(int)
}

type GameCommandPing struct {
	GameCommand
	Message string
}

func (gc GameCommandPing) Command() Command {
	return CommandPing
}

type GameCommandPong struct {
	GameCommand
	Message string
}

func (gc GameCommandPong) Command() Command {
	return CommandPong
}

type GameCommandMessage struct {
	GameCommand
	Player  int
	Message string
}

func (gc GameCommandMessage) Command() Command {
	return CommandMessage
}

type GameCommandJoinGame struct {
	GameCommand
	Name     string
	GameID   int
	PlayerID int
}

func (gc GameCommandJoinGame) Command() Command {
	return CommandJoinGame
}

type GameCommandNickname struct {
	GameCommand

	Player   int
	Nickname string
}

func (gc GameCommandNickname) Command() Command {
	return CommandNickname
}

type GameCommandQuitGame struct {
	GameCommand
	Player int
}

func (gc GameCommandQuitGame) Command() Command {
	return CommandQuitGame
}

type GameCommandUpdateGame struct {
	GameCommand
	Players map[int]string
}

func (gc GameCommandUpdateGame) Command() Command {
	return CommandUpdateGame
}

type GameCommandStartGame struct {
	GameCommand
	Seed    int64
	Started bool
}

func (gc GameCommandStartGame) Command() Command {
	return CommandStartGame
}

type GameCommandUpdateMatrix struct {
	GameCommand
	Matrixes map[int]*mino.Matrix
}

func (gc GameCommandUpdateMatrix) Command() Command {
	return CommandUpdateMatrix
}

type GameCommandGameOver struct {
	GameCommand
	Player int
	Winner string
}

func (gc GameCommandGameOver) Command() Command {
	return CommandGameOver
}

type GameCommandSendGarbage struct {
	GameCommand
	Lines int
}

func (gc GameCommandSendGarbage) Command() Command {
	return CommandSendGarbage
}

type GameCommandReceiveGarbage struct {
	GameCommand
	Lines int
}

func (gc GameCommandReceiveGarbage) Command() Command {
	return CommandReceiveGarbage
}

var nickRegexp = regexp.MustCompile(`[^a-zA-Z0-9_\-!@#$%^&*+=,./]+`)

func Nickname(nick string) string {
	nick = nickRegexp.ReplaceAllString(nick, "")
	if len(nick) > 10 {
		nick = nick[:10]
	} else if nick == "" {
		nick = "Anonymous"
	}

	return nick
}
