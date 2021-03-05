package event

type GameAction int

const (
	ActionUnknown GameAction = iota
	ActionRotateCCW
	ActionRotateCW
	ActionMoveLeft
	ActionMoveRight
	ActionSoftDrop
	ActionHardDrop
)
