package mino

import (
	"fmt"
	"sync"
	"time"
)

const (
	Rotation0 = 0
	RotationR = 1
	Rotation2 = 2
	RotationL = 3

	RotationStates = 4
)

type RotationOffsets []Point

type PieceType int

const (
	PieceI PieceType = iota
	PieceO
	PieceJ
	PieceL
	PieceS
	PieceT
	PieceZ

	PieceJLSTZ
)

var AllRotationPivotsCW = map[PieceType][]Point{
	PieceI: {{1, -2}, {-1, 0}, {1, -1}, {0, 0}},
	PieceO: {{1, 0}, {1, 0}, {1, 0}, {1, 0}},
	PieceJ: {{1, -1}, {0, 0}, {1, 0}, {1, 0}},
	PieceL: {{1, -1}, {0, 0}, {1, 0}, {1, 0}},
	PieceS: {{1, -1}, {0, 0}, {1, 0}, {1, 0}},
	PieceT: {{1, -1}, {0, 0}, {1, 0}, {1, 0}},
	PieceZ: {{1, -1}, {0, 0}, {1, 0}, {1, 0}},
}

// IN PROGRESS
var AllRotationPivotsCCW = map[PieceType][]Point{
	PieceI: {{2, 1}, {-1, 00}, {2, 2}, {1, 3}},
	PieceO: {{0, 1}, {0, 1}, {0, 1}, {0, 1}},
	PieceJ: {{1, 1}, {0, 0}, {1, 2}, {1, 2}},
	PieceL: {{1, 1}, {0, 0}, {1, 2}, {1, 2}},
	PieceS: {{1, 1}, {0, 0}, {1, 2}, {1, 2}},
	PieceT: {{1, 1}, {0, 2}, {1, 2}, {1, 2}},
	PieceZ: {{1, 1}, {0, 0}, {1, 2}, {1, 2}},
}

// AllRotationOffets is a list of all piece offsets.  Each set includes offsets
// for 0, R, L and 2 rotation states.
var AllOffsets = []Point{{0, 0}, {-1, 0}, {1, 0}, {0, -1}, {-1, -1}, {1, -1}, {-2, 0}, {2, 0}}

/*
var AllRotationOffsets = map[PieceType][]RotationOffsets{
	PieceI: {
		{{0, 0}, {-1, 0}, {-1, 1}, {0, 1}},
		{{-1, 0}, {0, 0}, {1, 1}, {0, 1}},
		{{2, 0}, {0, 0}, {-2, 1}, {0, 1}},
		{{-1, 0}, {0, 1}, {1, 0}, {0, -1}},
		{{2, 0}, {0, -2}, {-2, 0}, {0, 2}}},
	PieceO: {{{0, 0}, {0, -1}, {-1, -1}, {-1, 0}}},
	PieceJLSTZ: {
		{{0, 0}, {0, 0}, {0, 0}, {0, 0}},
		{{0, 0}, {1, 0}, {0, 0}, {-1, 0}},
		{{0, 0}, {1, -1}, {0, 0}, {-1, -1}},
		{{0, 0}, {0, 2}, {0, 0}, {0, 2}},
		{{0, 0}, {1, 2}, {0, 0}, {-1, 2}}}}
*/
type Piece struct {
	Point
	Mino
	Ghost    Block
	Solid    Block
	Rotation int

	original  Mino
	pivotsCW  []Point
	pivotsCCW []Point
	resets    int
	lastReset time.Time
	landing   bool
	landed    bool

	sync.Mutex `json:"-"`
}

type LockedPiece *Piece

func (p *Piece) String() string {
	return fmt.Sprintf("%+v", *p)
}

func NewPiece(m Mino, loc Point) *Piece {
	p := &Piece{Mino: m, original: m, Point: loc}

	var pieceType PieceType
	switch m.Canonical().String() {
	case TetrominoI:
		pieceType = PieceI
		p.Solid = BlockSolidCyan
		p.Ghost = BlockGhostCyan
	case TetrominoO:
		pieceType = PieceO
		p.Solid = BlockSolidYellow
		p.Ghost = BlockGhostYellow
	case TetrominoJ:
		pieceType = PieceJ
		p.Solid = BlockSolidBlue
		p.Ghost = BlockGhostBlue
	case TetrominoL:
		pieceType = PieceL
		p.Solid = BlockSolidOrange
		p.Ghost = BlockGhostOrange
	case TetrominoS:
		pieceType = PieceS
		p.Solid = BlockSolidGreen
		p.Ghost = BlockGhostGreen
	case TetrominoT:
		pieceType = PieceT
		p.Solid = BlockSolidMagenta
		p.Ghost = BlockGhostMagenta
	case TetrominoZ:
		pieceType = PieceZ
		p.Solid = BlockSolidRed
		p.Ghost = BlockGhostRed
	default:
		p.Solid = BlockSolidYellow
		p.Ghost = BlockGhostYellow
	}

	p.pivotsCW = AllRotationPivotsCW[pieceType]
	p.pivotsCCW = AllRotationPivotsCCW[pieceType]

	return p
}

/*
func (p *Piece) MarshalJSON() ([]byte, error) {
	log.Println("LOCK PIECE")
	p.Lock()
	defer p.Unlock()
	defer log.Println("UNLOCKED PIECE")

	return json.Marshal(LockedPiece(p))
}*/

// Rotate returns the new mino of a piece when a rotation is applied
func (p *Piece) Rotate(rotations int, direction int) Mino {
	p.Lock()
	defer p.Unlock()

	if rotations == 0 {
		return p.Mino
	}

	newMino := make(Mino, len(p.Mino))
	copy(newMino, p.Mino.Origin())

	var rotationPivot int
	for j := 0; j < rotations; j++ {
		if direction == 0 {
			rotationPivot = p.Rotation + j
		} else {
			rotationPivot = p.Rotation - j
		}

		if rotationPivot < 0 {
			rotationPivot += RotationStates
		}

		if (rotationPivot == 3 && direction == 0) || (rotationPivot == 1 && direction == 1) {
			newMino = p.original
		} else {
			pp := p.pivotsCW[rotationPivot%RotationStates]
			if direction == 1 {
				pp = p.pivotsCCW[rotationPivot%RotationStates]
			}
			px, py := pp.X, pp.Y

			for i := 0; i < len(newMino); i++ {
				x := newMino[i].X
				y := newMino[i].Y

				if direction == 0 {
					newMino[i] = Point{(0 * (x - px)) + (1 * (y - py)), (-1 * (x - px)) + (0 * (y - py))}
				} else {
					newMino[i] = Point{(0 * (x - px)) + (-1 * (y - py)), (1 * (x - px)) + (0 * (y - py))}
				}
			}
		}
	}

	return newMino
}

func (p *Piece) ApplyReset() {
	p.Lock()
	defer p.Unlock()

	if !p.landing || p.resets >= 15 {
		return
	}

	p.resets++
	p.lastReset = time.Now()
}

func (p *Piece) ApplyRotation(rotations int, direction int) {
	p.Lock()
	defer p.Unlock()

	if direction == 1 {
		rotations *= -1
	}

	p.Rotation = p.Rotation + rotations
	if p.Rotation < 0 {
		p.Rotation += RotationStates
	}
	p.Rotation %= RotationStates
}

func (p *Piece) SetLocation(x int, y int) {
	p.Lock()
	defer p.Unlock()

	p.X = x
	p.Y = y
}
