package mino

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/event"
)

const (
	GarbageDelay  = 1500 * time.Millisecond
	ComboBaseTime = 2.4 // Seconds
)

type MatrixType int

const (
	MatrixStandard MatrixType = iota
	MatrixPreview
	MatrixCustom
)

type Matrix struct {
	W int `json:"-"` // Width
	H int `json:"-"` // Height
	B int `json:"-"` // Buffer height

	M map[int]Block // Matrix
	O map[int]Block `json:"-"` // Overlay

	Bag        *Bag `json:"-"`
	P          *Piece
	PlayerName string

	Type MatrixType

	Event chan<- interface{} `json:"-"`
	Move  chan int           `json:"-"`
	draw  chan event.DrawObject

	Combo              int
	ComboStart         time.Time `json:"-"`
	ComboEnd           time.Time `json:"-"`
	PendingGarbage     int       `json:"-"`
	PendingGarbageTime time.Time `json:"-"`

	LinesCleared    int
	GarbageSent     int
	GarbageReceived int
	Speed           int

	GameOver bool

	lands []time.Time

	sync.Mutex `json:"-"`
}

// Type alias used during marshalling
type LockedMatrix *Matrix

func I(x int, y int, w int) int {
	return (y * w) + x
}

func NewMatrix(w int, h int, b int, players int, event chan<- interface{}, draw chan event.DrawObject, t MatrixType) *Matrix {
	m := Matrix{W: w, H: h, B: b, M: make(map[int]Block), O: make(map[int]Block), Event: event, draw: draw, Type: t}

	m.Move = make(chan int, 10)

	return &m
}

/*
func (m *Matrix) Lock() {
	log.Println("LOCKING ", string(debug.Stack()))
	m.Mutex.Lock()
	log.Println("LOCKED ", string(debug.Stack()))
}

func (m *Matrix) Unlock() {
	log.Println("UNLOCKED ", string(debug.Stack()))
	m.Mutex.Unlock()
}
*/

func (m *Matrix) MarshalJSON() ([]byte, error) {
	m.Lock()
	defer m.Unlock()

	return json.Marshal(*LockedMatrix(m))
}
func (m *Matrix) HandleReceiveGarbage() {
	t := time.NewTicker(500 * time.Millisecond)
	for {
		<-t.C

		m.ReceiveGarbage()
	}
}

func (m *Matrix) AttachBag(bag *Bag) bool {
	m.Lock()
	defer m.Unlock()

	m.Bag = bag

	return true
}

func (m *Matrix) takePiece() bool {
	if m.Type != MatrixStandard {
		return true
	} else if m.GameOver || m.Bag == nil {
		return false
	}

	p := NewPiece(m.Bag.Take(), Point{0, 0})

	pieceStart := m.pieceStart(p)
	if pieceStart.X < 0 || pieceStart.Y < 0 {
		return false
	}

	p.Point = pieceStart

	m.P = p

	return true
}

func (m *Matrix) TakePiece() bool {
	m.Lock()
	defer m.Unlock()

	return m.takePiece()
}

func (m *Matrix) CanAddAt(mn *Piece, loc Point) bool {
	m.Lock()
	defer m.Unlock()

	return m.canAddAt(mn, loc)
}

func (m *Matrix) canAddAt(mn *Piece, loc Point) bool {
	if m.GameOver || loc.Y < 0 {
		return false
	}

	var (
		x, y  int
		index int
	)

	for _, p := range mn.Mino {
		x = p.X + loc.X
		y = p.Y + loc.Y

		if x < 0 || x >= m.W || y < 0 || y >= m.H+m.B {
			return false
		}

		index = I(x, y, m.W)
		if m.M[index] != BlockNone {
			return false
		}
	}

	return true
}

func (m *Matrix) CanAdd(mn *Piece) bool {
	m.Lock()
	defer m.Unlock()

	if m.GameOver {
		return false
	}

	var (
		x, y  int
		index int
	)

	for _, p := range mn.Mino {
		x = p.X + mn.X
		y = p.Y + mn.Y

		if x < 0 || x >= m.W || y < 0 || y >= m.H+m.B {
			return false
		}

		index = I(x, y, m.W)
		if m.M[index] != BlockNone {
			return false
		}
	}

	return true
}

func (m *Matrix) Add(mn *Piece, b Block, loc Point, overlay bool) error {
	m.Lock()
	defer m.Unlock()

	return m.add(mn, b, loc, overlay)
}

func (m *Matrix) add(mn *Piece, b Block, loc Point, overlay bool) error {
	if m.GameOver {
		return nil
	}

	var (
		x, y  int
		index int

		M map[int]Block
	)

	if overlay {
		M = m.O
	} else {
		M = m.M
	}

	for _, p := range mn.Mino {
		x = p.X + loc.X
		y = p.Y + loc.Y

		if x < 0 || x >= m.W || y < 0 || y >= m.H+m.B {
			return fmt.Errorf("failed to add to matrix at %s: point %s out of bounds (%d, %d)", loc, p, x, y)
		}

		index = I(x, y, m.W)
		if !overlay && M[index] != BlockNone {
			return fmt.Errorf("failed to add to matrix at %s: point %s already contains %s", loc, p, M[index])
		}

		M[index] = b
	}

	return nil
}

func (m *Matrix) Empty(loc Point) bool {
	index := I(loc.X, loc.Y, m.W)
	return m.M[index] == BlockNone
}

func (m *Matrix) LineFilled(y int) bool {
	for x := 0; x < m.W; x++ {
		if m.Empty(Point{x, y}) {
			return false
		}
	}

	return true
}

func (m *Matrix) ClearFilled() int {
	m.Lock()
	defer m.Unlock()

	return m.clearFilled()
}

func (m *Matrix) clearFilled() int {
	cleared := 0

	for y := 0; y < m.H+m.B; y++ {
		for {
			if m.LineFilled(y) {
				for my := y + 1; my < m.H+m.B; my++ {
					for mx := 0; mx < m.W; mx++ {
						m.M[I(mx, my-1, m.W)] = m.M[I(mx, my, m.W)]
					}
				}

				cleared++
				continue
			}

			break
		}
	}

	return cleared
}

func (m *Matrix) AddPendingGarbage(lines int) {
	m.Lock()
	defer m.Unlock()

	if m.PendingGarbage == 0 {
		m.PendingGarbageTime = time.Now().Add(GarbageDelay)
	}

	m.PendingGarbage += lines
}

func (m *Matrix) ReceiveGarbage() {
	m.Lock()
	defer m.Unlock()

	if m.PendingGarbage == 0 || m.GameOver {
		return
	} else if time.Since(m.PendingGarbageTime) < 0 {
		return
	}

	m.PendingGarbage--
	if !m.addGarbage(1) {
		m.Event <- &event.GameOverEvent{}
	}
}

func (m *Matrix) addGarbage(lines int) bool {
	for my := m.H + m.B; my >= 0; my-- {
		for mx := 0; mx < m.W; mx++ {
			if my >= m.H+m.B-lines && m.M[I(mx, my, m.W)] != BlockNone {
				return false
			}

			m.M[I(mx, my+lines, m.W)] = m.M[I(mx, my, m.W)]
		}
	}

	for my := 0; my < lines; my++ {
		hole := m.Bag.GarbageHole()
		for mx := 0; mx < m.W; mx++ {
			if mx == hole {
				m.M[I(mx, my, m.W)] = BlockNone
			} else {
				m.M[I(mx, my, m.W)] = BlockGarbage
			}
		}
	}

	y := m.P.Y
	for {
		if y == m.H+m.B {
			return false
		} else if m.canAddAt(m.P, Point{m.P.X, y}) {
			break
		}

		y++
	}

	m.P.Y = y

	m.Draw()

	return true
}

func (m *Matrix) Draw() {
	if m.draw == nil {
		return
	}

	m.draw <- event.DrawPlayerMatrix
}

func (m *Matrix) ClearOverlay() {
	m.Lock()
	defer m.Unlock()

	m.clearOverlay()
}

func (m *Matrix) clearOverlay() {
	for i := range m.O {
		if m.O[i] == BlockNone {
			continue
		}

		m.O[i] = BlockNone
	}
}

func (m *Matrix) Reset() {
	m.Lock()

	m.GameOver = false
	m.P = nil
	m.lands = nil
	m.Speed = 0
	m.PendingGarbage = 0
	m.PendingGarbageTime = time.Time{}
	m.Unlock()

	m.Clear()
	m.ClearOverlay()
}

func (m *Matrix) Clear() {
	m.Lock()
	defer m.Unlock()

	for i := range m.M {
		if m.M[i] == BlockNone {
			continue
		}

		m.M[i] = BlockNone
	}
}

func (m *Matrix) DrawPieces() {
	m.Lock()
	defer m.Unlock()

	m.DrawPiecesL()
}

func (m *Matrix) DrawPiecesL() {
	if m.Type != MatrixStandard {
		return
	}

	m.clearOverlay()

	if m.GameOver {
		return
	}

	p := m.P
	if p == nil {
		return
	}

	// Draw ghost piece
	for y := p.Y; y >= 0; y-- {
		if y == 0 || !m.canAddAt(p, Point{p.X, y - 1}) {
			err := m.add(p, p.Ghost, Point{p.X, y}, true)
			if err != nil {
				log.Fatalf("failed to draw ghost piece: %+v", err)
			}

			break
		}
	}

	// Draw piece
	err := m.add(p, p.Solid, Point{p.X, p.Y}, true)
	if err != nil {
		log.Fatalf("failed to draw active piece: %+v", err)
	}
}

func (m *Matrix) NewM() map[int]Block {
	newM := make(map[int]Block, len(m.M))
	for i, b := range m.M {
		newM[i] = b
	}

	return newM
}

func (m *Matrix) NewO() map[int]Block {
	newO := make(map[int]Block, len(m.O))
	for i, b := range m.O {
		newO[i] = b
	}

	return newO
}

func (m *Matrix) Block(x int, y int) Block {
	index := I(x, y, m.W)

	// Return overlay block first
	if b, ok := m.O[index]; ok && b != BlockNone {
		return b
	}

	return m.M[index]
}

func (m *Matrix) SetGameOver() {
	m.Lock()
	defer m.Unlock()

	if m.GameOver {
		return
	}

	m.GameOver = true
	m.Combo = 0
	m.ComboStart = time.Time{}
	m.ComboEnd = time.Time{}

	for i := range m.M {
		if m.M[i] != BlockNone && m.M[i] != BlockGarbage {
			switch m.M[i] {
			case BlockSolidBlue:
				m.M[i] = BlockGhostBlue
			case BlockSolidCyan:
				m.M[i] = BlockGhostCyan
			case BlockSolidGreen:
				m.M[i] = BlockGhostGreen
			case BlockSolidMagenta:
				m.M[i] = BlockGhostMagenta
			case BlockSolidOrange:
				m.M[i] = BlockGhostOrange
			case BlockSolidRed:
				m.M[i] = BlockGhostRed
			case BlockSolidYellow:
				m.M[i] = BlockGhostYellow
			}
		}
	}
}

func (m *Matrix) SetBlock(x int, y int, block Block, overlay bool) bool {
	if x < 0 || x >= m.W || y < 0 || y >= m.H+m.B {
		return false
	}

	index := I(x, y, m.W)

	if overlay {
		if b, ok := m.O[index]; ok && b != BlockNone {
			return false
		}

		m.O[index] = block
	} else {
		if b, ok := m.M[index]; ok && b != BlockNone {
			return false
		}

		m.M[index] = block
	}

	return true
}

func (m *Matrix) RotatePiece(rotations int, direction int) bool {
	m.Lock()
	defer m.Unlock()

	if m.GameOver || rotations == 0 {
		return false
	}

	p := m.P

	originalMino := make(Mino, len(p.Mino))
	copy(originalMino, p.Mino)

	p.Mino = p.Rotate(rotations, direction)

	for i := range AllOffsets {
		px := p.X + AllOffsets[i].X
		py := p.Y + AllOffsets[i].Y

		if m.canAddAt(p, Point{px, py}) {
			p.ApplyReset()

			if p.X != px || p.Y != py {
				p.SetLocation(px, py)
			}

			p.ApplyRotation(rotations, direction)

			m.Draw()

			return true
		}

	}

	p.Mino = originalMino
	return false
}

func (m *Matrix) pieceStart(p *Piece) Point {
	w, _ := p.Size()
	x := (m.W / 2) - (w / 2)

	for y := m.H; y < m.H+m.B; y++ {
		if m.canAddAt(p, Point{x, y}) {
			return Point{x, y}
		}
	}

	return Point{-1, -1}

}

func (m *Matrix) Render() string {
	m.Lock()
	defer m.Unlock()

	var b strings.Builder

	for y := m.H - 1; y >= 0; y-- {
		for x := 0; x < m.W; x++ {
			b.WriteRune(m.Block(x, y).Rune())
		}

		if y == 0 {
			break
		}

		b.WriteRune('\n')
	}

	return b.String()
}

// LowerPiece lowers the active piece by one line when possible, otherwise the
// piece is landed
func (m *Matrix) LowerPiece() {
	m.Lock()
	defer m.Unlock()

	if m.GameOver {
		return
	} else if m.canAddAt(m.P, Point{m.P.X, m.P.Y - 1}) {
		m.movePiece(0, -1)
	} else {
		m.landPiece()
	}
}

func (m *Matrix) finishLandingPiece() {
	if m.GameOver || m.P.landed {
		return
	}

	m.P.landed = true

	dropped := false
LANDPIECE:
	for y := m.P.Y; y >= 0; y-- {
		if y == 0 || !m.canAddAt(m.P, Point{m.P.X, y - 1}) {
			for dropY := y - 1; dropY < m.H+m.B; dropY++ {
				if !m.canAddAt(m.P, Point{m.P.X, dropY}) {
					continue
				}

				err := m.add(m.P, m.P.Solid, Point{m.P.X, dropY}, false)
				if err != nil {
					log.Fatalf("failed to add piece when landing piece: %+v", err)
				}

				dropped = true
				break LANDPIECE
			}
		}
	}

	if !dropped {
		m.Event <- event.GameOverEvent{}

		m.Draw()
		return
	}

	cleared := m.clearFilled()

	score := 0
	switch cleared {
	case 0:
		// No score
	case 1:
		score = 100
	case 2:
		score = 300
	case 3:
		score = 500
	case 4:
		score = 800
	default:
		score = 1000 + ((cleared - 5) * 200)
	}

	_ = score

	m.Moved()

	for i := range m.lands {
		if time.Since(m.lands[i]) > 2*time.Minute {
			continue
		}

		if i > 0 {
			m.lands = m.lands[i+1:]
		}
		break
	}
	m.lands = append(m.lands, time.Now())

	numlands := len(m.lands)
	if numlands > 1 {
		m.Speed = int(time.Minute / (time.Since(m.lands[0]) / time.Duration(numlands)))
	}

	if cleared > 0 {
		sendGarbage := m.addToCombo(cleared)
		if sendGarbage > 0 {
			remainingGarbage := sendGarbage
			if m.PendingGarbage > 0 {
				m.PendingGarbage -= sendGarbage

				if m.PendingGarbage < 0 {
					remainingGarbage = m.PendingGarbage * -1
					m.PendingGarbage = 0
				} else {
					remainingGarbage = 0
				}
			}

			if remainingGarbage > 0 {
				m.Event <- &event.SendGarbageEvent{Lines: remainingGarbage}
			}
		}
	}

	if !m.takePiece() {
		m.Event <- &event.GameOverEvent{}
	}

	m.Draw()
}

func (m *Matrix) addToCombo(lines int) int {
	if m.GameOver {
		return 0
	}

	baseTime := ComboBaseTime
	bonusTime := baseTime / 2

	if m.Combo == 0 || time.Until(m.ComboEnd) <= 0 {
		m.Combo = 0
		m.ComboStart = time.Now()
		m.ComboEnd = m.ComboStart
	}

	m.Combo++

	if m.Combo > 1 {
		baseTime /= math.Pow(2, float64(m.Combo-1))
		bonusTime /= math.Pow(2, float64(m.Combo-1))
	}

	m.ComboEnd = m.ComboEnd.Add(time.Duration((baseTime * float64(time.Second)) + (bonusTime * float64(lines) * float64(time.Second))))

	baseGarbage := 0
	if lines > 1 {
		baseGarbage = lines - 1
	}

	bonusGarbage := m.CalculateBonusGarbage()

	return baseGarbage + bonusGarbage
}

func (m *Matrix) CalculateBonusGarbage() int {
	bonusGarbage := 0
	if m.Combo == 1 {
		// No bonus garbage
	} else if m.Combo < 4 {
		bonusGarbage = 1
	} else {
		scoreCombo := m.Combo - 3
		if scoreCombo > 7 {
			scoreCombo = 7
		}

		bonusGarbage = fibonacci(scoreCombo)
	}

	return bonusGarbage
}

func (m *Matrix) landPiece() {
	p := m.P
	p.Lock()
	if p.landing || p.landed || m.GameOver {
		p.Unlock()
		return
	}

	p.landing = true
	p.Unlock()

	go func() {
		landStart := time.Now()

		t := time.NewTicker(100 * time.Millisecond)
		for {
			<-t.C

			m.Lock()
			p.Lock()

			if p.landed {
				p.Unlock()
				m.Unlock()
				return
			}

			if p.resets > 0 && time.Since(p.lastReset) < 500*time.Millisecond {
				p.Unlock()
				m.Unlock()
				continue
			} else if time.Since(landStart) < 500*time.Millisecond {
				p.Unlock()
				m.Unlock()
				continue
			}

			t.Stop()
			break
		}

		p.Unlock()

		m.finishLandingPiece()
		m.Unlock()
	}()
}

func (m *Matrix) MovePiece(x int, y int) bool {
	m.Lock()
	defer m.Unlock()

	return m.movePiece(x, y)
}

func (m *Matrix) movePiece(x int, y int) bool {
	if m.GameOver || (x == 0 && y == 0) {
		return false
	}

	px := m.P.X + x
	py := m.P.Y + y

	if !m.canAddAt(m.P, Point{px, py}) {
		return false
	}

	m.P.ApplyReset()
	m.P.SetLocation(px, py)

	if !m.canAddAt(m.P, Point{m.P.X, m.P.Y - 1}) {
		m.landPiece()
	}

	if y < 0 {
		m.Moved()
	}

	m.Draw()

	return true
}

func (m *Matrix) Moved() {
	if m.Move == nil {
		return
	}

	m.Move <- 0
}

func (m *Matrix) HardDropPiece() {
	m.Lock()
	defer m.Unlock()

	m.finishLandingPiece()
}

func (m *Matrix) Replace(newmtx *Matrix) {
	m.Lock()
	defer m.Unlock()

	m.M = newmtx.M
	m.P = newmtx.P

	m.PlayerName = newmtx.PlayerName
	m.GarbageSent = newmtx.GarbageSent
	m.GarbageReceived = newmtx.GarbageReceived
	m.Speed = newmtx.Speed
}

func fibonacci(value int) int {
	if value == 0 || value == 1 {
		return value
	}
	return fibonacci(value-2) + fibonacci(value-1)
}

func NewTestMatrix() (*Matrix, error) {
	minos, err := Generate(4)
	if err != nil {
		return nil, fmt.Errorf("failed to generate minos: %s", err)
	}

	ev := make(chan interface{})
	go func() {
		for range ev {
		}
	}()

	draw := make(chan event.DrawObject)
	go func() {
		for range draw {
		}
	}()

	m := NewMatrix(10, 20, 4, 1, ev, draw, MatrixStandard)

	bag, err := NewBag(1, minos, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate minos: %s", err)
	}

	m.AttachBag(bag)

	m.TakePiece()

	return m, nil
}

func (m *Matrix) AddTestBlocks() {
	var block Block
	for y := 0; y < 7; y++ {
		for x := 0; x < m.W-1; x++ {
			if y > 3 && (x < 2 || x > 7) {
				continue
			}

			if y == 2 || (y > 4 && x%2 > 0) {
				block = BlockSolidMagenta
			} else {
				block = BlockSolidYellow
			}

			m.M[I(x, y, m.W)] = block
		}
	}
}
