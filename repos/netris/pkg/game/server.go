package game

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/event"
)

type Server struct {
	I []ServerInterface

	In  chan GameCommandInterface
	Out chan GameCommandInterface

	NewPlayers chan *IncomingPlayer

	Games map[int]*Game

	Logger chan string

	listeners []net.Listener
	sync.RWMutex
}

type IncomingPlayer struct {
	Name string
	Conn *ServerConn
}

type ServerInterface interface {
	// Load config
	Host(newPlayers chan<- *IncomingPlayer)
	Shutdown(reason string)
}

func NewServer(si []ServerInterface) *Server {
	in := make(chan GameCommandInterface, CommandQueueSize)
	out := make(chan GameCommandInterface, CommandQueueSize)

	s := &Server{I: si, In: in, Out: out, Games: make(map[int]*Game)}

	s.NewPlayers = make(chan *IncomingPlayer, CommandQueueSize)

	go s.accept()

	for _, serverInterface := range si {
		serverInterface.Host(s.NewPlayers)
	}

	return s
}

func (s *Server) NewGame() (*Game, error) {
	gameID := 1
	for {
		if _, ok := s.Games[gameID]; !ok {
			break
		}

		gameID++
	}

	draw := make(chan event.DrawObject)
	go func() {
		for range draw {
		}
	}()

	logger := make(chan string, LogQueueSize)
	go func() {
		for msg := range logger {
			s.Log(fmt.Sprintf("Game %d: %s", gameID, msg))
		}
	}()

	g, err := NewGame(4, nil, logger, draw)
	if err != nil {
		return nil, err
	}

	s.Games[gameID] = g

	return g, nil
}

func (s *Server) FindGame(p *Player, gameID int) *Game {
	s.Lock()
	defer s.Unlock()

	var (
		g   *Game
		err error
	)

	if gm, ok := s.Games[gameID]; ok {
		g = gm
	}

	if g == nil {
		for gameID, g = range s.Games {
			if g != nil {
				if g.Terminated {
					delete(s.Games, gameID)
					g = nil

					s.Log("Cleaned up game ", gameID)
					continue
				}

				break
			}
		}
	}

	if g == nil {
		g, err = s.NewGame()
		if err != nil {
			panic(err)
		}

		if gameID == -1 {
			g.Local = true
		}
	}

	g.Lock()

	g.AddPlayerL(p)

	if gameID == -1 {
		go g.Start(0)
	} else if len(g.Players) > 1 {
		go s.initiateAutoStart(g)
	} else if !g.Started {
		p.Write(&GameCommandMessage{Message: "Waiting for at least two players to join..."})
	}

	g.Unlock()

	return g
}

func (s *Server) accept() {
	for {
		np := <-s.NewPlayers

		p := NewPlayer(np.Name, np.Conn)

		s.Log("Incoming connection from ", np.Name)

		go s.handleJoinGame(p)
	}
}

func (s *Server) handleJoinGame(pl *Player) {
	for e := range pl.In {
		if e.Command() == CommandJoinGame {
			if p, ok := e.(*GameCommandJoinGame); ok {
				pl.Name = Nickname(p.Name)

				g := s.FindGame(pl, p.GameID)

				s.Logf("Adding %s to game %d", pl.Name, p.GameID)

				go s.handleGameCommands(pl, g)
				return
			}
		}
	}
}

func (s *Server) initiateAutoStart(g *Game) {
	g.Lock()
	defer g.Unlock()

	if g.Starting || g.Started {
		return
	}

	g.Starting = true

	go func() {
		g.WriteMessage("Starting game...")
		time.Sleep(2 * time.Second)
		g.Start(0)
	}()
}

func (s *Server) handleGameCommands(pl *Player, g *Game) {
	for e := range pl.In {
		c := e.Command()
		if (c != CommandPing && c != CommandPong && c != CommandUpdateMatrix) || g.LogLevel >= LogVerbose {
			s.Log("REMOTE handle game command ", e.Command(), " from ", e.Source(), e)
		}

		g.Lock()

		switch c {
		case CommandMessage:
			if p, ok := e.(*GameCommandMessage); ok {
				if player, ok := g.Players[p.SourcePlayer]; ok {
					s.Log("<" + player.Name + "> " + p.Message)

					msg := strings.ReplaceAll(strings.TrimSpace(p.Message), "\n", "")
					if msg != "" {
						g.WriteAllL(&GameCommandMessage{Player: p.SourcePlayer, Message: msg})
					}
				}
			}
		case CommandNickname:
			if p, ok := e.(*GameCommandNickname); ok {
				if player, ok := g.Players[p.SourcePlayer]; ok {
					newNick := Nickname(p.Nickname)
					if newNick != "" && newNick != player.Name {
						oldNick := player.Name
						player.Name = newNick

						g.Logf(LogStandard, "* %s is now known as %s", oldNick, newNick)
						g.WriteAllL(&GameCommandNickname{Player: p.SourcePlayer, Nickname: newNick})
					}
				}
			}
		case CommandUpdateMatrix:
			if p, ok := e.(*GameCommandUpdateMatrix); ok {
				for _, m := range p.Matrixes {
					g.Players[p.SourcePlayer].Matrix.Replace(m)
				}
			}
		case CommandGameOver:
			if p, ok := e.(*GameCommandGameOver); ok {
				g.Players[p.SourcePlayer].Matrix.SetGameOver()

				g.WriteMessage(fmt.Sprintf("%s was knocked out", g.Players[p.SourcePlayer].Name))
				g.WriteAllL(&GameCommandGameOver{Player: p.SourcePlayer})
			}
		case CommandSendGarbage:
			if p, ok := e.(*GameCommandSendGarbage); ok {
				leastGarbagePlayer := -1
				leastGarbage := -1
				for playerID, player := range g.Players {
					if playerID == p.SourcePlayer || player.Matrix.GameOver {
						continue
					}

					if leastGarbage == -1 || player.totalGarbageReceived < leastGarbage {
						leastGarbagePlayer = playerID
						leastGarbage = player.totalGarbageReceived
					}
				}

				if leastGarbagePlayer != -1 {
					g.Players[leastGarbagePlayer].totalGarbageReceived += p.Lines
					g.Players[leastGarbagePlayer].pendingGarbage += p.Lines

					g.Players[p.SourcePlayer].totalGarbageSent += p.Lines
				}
			}
		}

		g.Unlock()
	}
}

func (s *Server) Listen(address string) {
	var network string
	network, address = NetworkAndAddress(address)

	listener, err := net.Listen(network, address)
	if err != nil {
		log.Fatalf("failed to listen on %s: %s", address, err)
	}

	s.listeners = append(s.listeners, listener)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		s.NewPlayers <- &IncomingPlayer{Name: "Anonymous", Conn: NewServerConn(conn, nil)}
	}
}

func (s *Server) StopListening() {
	for i := range s.listeners {
		s.listeners[i].Close()
	}
}

func (s *Server) Log(a ...interface{}) {
	if s.Logger == nil {
		return
	}

	s.Logger <- fmt.Sprint(a...)
}

func (s *Server) Logf(format string, a ...interface{}) {
	if s.Logger == nil {
		return
	}

	s.Logger <- fmt.Sprintf(format, a...)
}

func NetworkAndAddress(address string) (string, string) {
	var network string
	if strings.ContainsAny(address, `\/`) {
		network = "unix"
	} else {
		network = "tcp"

		if !strings.Contains(address, `:`) {
			address = fmt.Sprintf("%s:%d", address, DefaultPort)
		}
	}

	return network, address
}
