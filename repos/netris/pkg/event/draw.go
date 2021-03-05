package event

type DrawObject int

const (
	DrawAll DrawObject = iota
	DrawMessages
	DrawPlayerMatrix
	DrawMultiplayerMatrixes
)
