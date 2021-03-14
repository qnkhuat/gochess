package pkg

import (
	"encoding/json"
	"log"
)

type MessageInterface interface {
	Type() MessageType
}
type MessageType int

const (
	TypeMessageGame MessageType = iota
	TypeMessageMove
	TypeMessageTransport
	TypeMessageConnect
)

func (m MessageType) String() string {
	switch m {
	case TypeMessageGame:
		return "TypeMessageGame"
	case TypeMessageMove:
		return "TypeMessageMove"
	case TypeMessageTransport:
		return "TypeMessageTransport"
	case TypeMessageConnect:
		return "TypeMessageConnect"
	default:
		return "Unknown MessageType"
	}
}

// Message types

//
type MessageTransport struct {
	MsgType  MessageType
	Data     json.RawMessage
	PlayerId int
}

func (m MessageTransport) Type() MessageType {
	return TypeMessageTransport
}

//
type MessageMove struct {
	Move string
	Msg  string
}

func (m MessageMove) Type() MessageType {
	return TypeMessageMove
}

//
type MessageGame struct {
	Fen    string
	IsTurn bool
}

func (m MessageGame) Type() MessageType {
	return TypeMessageGame
}

//
type MessageConnect struct {
	Color  PlayerColor
	Fen    string
	IsTurn bool
}

func (m MessageConnect) Type() MessageType {
	return TypeMessageConnect
}

func Encode(o interface{}) json.RawMessage {
	data, err := json.Marshal(o)
	if err != nil {
		log.Panic(err)
	}
	return data
}

func Decode(data []byte, o interface{}) {
	err := json.Unmarshal(data, o)
	log.Printf("Decode %v", o)
	if err != nil {
		log.Panic(err)
	}
}
