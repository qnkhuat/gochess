package event

type Event struct {
	Player  int
	Message string
}

type MessageEvent struct {
	Event
	Message string
}

type NicknameEvent struct {
	Event
	Nickname string
}

type GameOverEvent struct {
	Event
}

type ScoreEvent struct {
	Event
	Score int
}

type SendGarbageEvent struct {
	Event
	Lines int
}
