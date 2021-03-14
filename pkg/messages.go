// TODO : Duplicated code, Use Interface extensively
package pkg

import (
	"encoding/json"
	"log"
)

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

type MessageInterface interface {
	Type() MessageType
	Encode() json.RawMessage
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

func (m MessageTransport) Encode() json.RawMessage {
	data, err := json.Marshal(m)
	if err != nil {
		log.Panic(err)
	}
	return data
}

//
type MessageMove struct {
	Move string
	Msg  string
}

func (m MessageMove) Type() MessageType {
	return TypeMessageMove
}

func (m MessageMove) Encode() json.RawMessage {
	data, err := json.Marshal(m)
	if err != nil {
		log.Panic(err)
	}
	return data
}

//
type MessageGame struct {
	Fen    string
	IsTurn bool
}

func (m MessageGame) Type() MessageType {
	return TypeMessageGame
}

func (m MessageGame) Encode() json.RawMessage {
	data, err := json.Marshal(m)
	if err != nil {
		log.Panic(err)
	}

	return data
}

type MessageConnect struct {
	Color  PlayerColor
	Fen    string
	IsTurn bool
}

func (m MessageConnect) Type() MessageType {
	return TypeMessageConnect
}

func (m MessageConnect) Encode() json.RawMessage {
	data, err := json.Marshal(m)
	if err != nil {
		log.Panic(err)
	}

	return data
}
