package pkg

import (
	"bufio"
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
	Out   chan MessageInterface
	Id    int
	Name  string
}

func NewPlayer(conn net.Conn) *Player {
	Out := make(chan MessageInterface, ConnQueueSize)

	p := &Player{
		Conn: conn,
		Out:  Out,
	}
	return p
}

func (p *Player) HandleRead(In chan MessageInterface) {
	// Receive message, add player info, then forward to server
	scanner := bufio.NewScanner(p.Conn)
	var messageTransport MessageTransport
	for scanner.Scan() {
		Decode(scanner.Bytes(), &messageTransport)
		messageTransport.PlayerId = p.Id
		In <- messageTransport // Forward the message to server
	}
}

func (p *Player) HandleWrite() {
	for message := range p.Out {
		messageData := Encode(message)
		messageTransport := &MessageTransport{MsgType: message.Type(), Data: messageData}
		b := Encode(messageTransport)
		if b[len(b)-1] != '\n' { // EOF
			b = append(b, '\n')
		}
		if _, err := p.Conn.Write(b); err != nil {
			log.Printf("Failed to write: %v Error: %v", message, err)
		}
	}
}

func (p *Player) Disconnect() {
	p.Conn.Close()
}
