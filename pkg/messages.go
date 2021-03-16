package pkg

import (
	"encoding/json"
	"log"
	"time"
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
	TypeMessageGameChat
	TypeMessageGameAction
	TypeMessageGameStatus
	TypeMessageMatchRemovePlayer
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
	case TypeMessageGameChat:
		return "TypeMessageGameChat"
	case TypeMessageGameAction:
		return "TypeMessageGameAction"
	case TypeMessageGameStatus:
		return "TypeMessageGameStatus"
	case TypeMessageMatchRemovePlayer:
		return "TypeMessageMatchRemovePlayer"
	default:
		return "Unknown MessageType"
	}
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
	if err != nil {
		log.Panic(err)
	}
}

// Message types

// A generic sturct used to transport between server-client
type MessageTransport struct {
	MsgType  MessageType
	Data     json.RawMessage
	PlayerId int
}

func (m MessageTransport) Type() MessageType {
	return TypeMessageTransport
}

// Move from player
type MessageMove struct {
	Move string
	Msg  string
}

func (m MessageMove) Type() MessageType {
	return TypeMessageMove
}

// Game Update
type MessageGame struct {
	Fen    string
	IsTurn bool
}

func (m MessageGame) Type() MessageType {
	return TypeMessageGame
}

// Initialize connection
type MessageConnect struct {
	Role   PlayerRole
	Fen    string
	IsTurn bool
}

func (m MessageConnect) Type() MessageType {
	return TypeMessageConnect
}

//
type MessageGameAction struct {
	Action  Action
	Message string
}

func (m MessageGameAction) Type() MessageType {
	return TypeMessageGameAction
}

// Chatting purpose
type MessageGameChat struct {
	Message string
	Name    string
	Time    time.Time
}

func (m MessageGameChat) Type() MessageType {
	return TypeMessageGameChat
}

//
type MessageGameStatus struct {
	Message string
}

func (m MessageGameStatus) Type() MessageType {
	return TypeMessageGameStatus
}

//
type MessageMatchRemovePlayer struct {
	PlayerId int
}

func (m MessageMatchRemovePlayer) Type() MessageType {
	return TypeMessageMatchRemovePlayer
}
