package pkg

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
)

type PlayerColor int

const (
	White PlayerColor = iota
	Black
	Viewer
	Unknown
)

func (pc PlayerColor) String() string {
	switch pc {
	case White:
		return "White"
	case Black:
		return "Black"
	case Viewer:
		return "Viewer"
	default:
		return "Unknown"
	}
}

type Player struct {
	Conn  net.Conn
	Color PlayerColor
	In    chan MessageInterface
	Out   chan MessageInterface
	Id    int
}

func NewPlayer(conn net.Conn) *Player {
	In := make(chan MessageInterface, ConnQueueSize)
	Out := make(chan MessageInterface, ConnQueueSize)

	p := &Player{
		Conn: conn,
		In:   In,
		Out:  Out,
	}

	go p.HandleWrite()
	return p
}

func (p *Player) HandleRead(In chan MessageTransport) {
	// Receive message, add player info, then forward to server
	scanner := bufio.NewScanner(p.Conn)
	var messageTransport MessageTransport
	for scanner.Scan() {
		err := json.Unmarshal(scanner.Bytes(), &messageTransport)
		if err != nil {
			log.Panic(err)
		}
		messageTransport.PlayerId = p.Id
		In <- messageTransport // Forward the message to server
	}
}

func (p *Player) HandleWrite() {
	for message := range p.Out {
		messageData := message.Encode()
		messageTransport := &MessageTransport{MsgType: message.Type(), Data: messageData}
		b := messageTransport.Encode()
		if b[len(b)-1] != '\n' { // EOF
			b = append(b, '\n')
		}
		if _, err := p.Conn.Write(b); err != nil {
			log.Fatal(err)
		}
		log.Printf("Send a msg: %v with type :%s", b, message.Type())
	}
}
