package mino

import (
	"strconv"
	"strings"
)

type Point struct {
	X, Y int
}

func (p Point) Rotate90() Point  { return Point{p.Y, -p.X} }
func (p Point) Rotate180() Point { return Point{-p.X, -p.Y} }
func (p Point) Rotate270() Point { return Point{-p.Y, p.X} }
func (p Point) Reflect() Point   { return Point{-p.X, p.Y} }

func (p Point) String() string {
	var b strings.Builder
	b.WriteRune('(')
	b.WriteString(strconv.Itoa(p.X))
	b.WriteRune(',')
	b.WriteString(strconv.Itoa(p.Y))
	b.WriteRune(')')

	return b.String()
}

// Neighborhood returns the Von Neumann neighborhood of a point
func (p Point) Neighborhood() Mino {
	return Mino{
		{p.X - 1, p.Y},
		{p.X, p.Y - 1},
		{p.X + 1, p.Y},
		{p.X, p.Y + 1}}
}
