package pkg

import (
	"bufio"
	"log"
	"net"
)

type PlayerRole int

const (
	White PlayerRole = iota
	Black
	Viewer
)

func (pc PlayerRole) String() string {
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
	Conn net.Conn
	Role PlayerRole
	Out  chan MessageInterface
	Id   int
	Name string
	// Time -- User time here
}

func NewPlayer(conn net.Conn, name string) *Player {
	out := make(chan MessageInterface, ConnQueueSize)
	p := &Player{
		Conn: conn,
		Out:  out,
		Name: name,
	}
	return p
}

func (p *Player) HandleRead(In chan MessageInterface) {
	// Receive message, add player info, then forward to server
	defer p.Disconnect()
	scanner := bufio.NewScanner(p.Conn)
	var messageTransport MessageTransport
	for scanner.Scan() {
		Decode(scanner.Bytes(), &messageTransport)
		messageTransport.PlayerId = p.Id
		In <- messageTransport // Forward the message to server
	}

	message := MessageMatchRemovePlayer{
		PlayerId: p.Id,
	}
	In <- MessageTransport{
		MsgType:  message.Type(),
		Data:     Encode(message),
		PlayerId: p.Id,
	}
	log.Println("Player Disconnected")
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
