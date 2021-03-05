package mino

import (
	"errors"
)

// Generate
func Generate(rank int) ([]Mino, error) {
	switch {
	case rank < 0:
		return nil, errors.New("invalid rank")
	case rank == 0:
		return []Mino{}, nil
	case rank == 1:
		return []Mino{monomino()}, nil
	default:
		r, err := Generate(rank - 1)
		if err != nil {
			return nil, err
		}

		var minos []Mino
		found := make(map[string]bool)
		for _, mino := range r {
			for _, newMino := range mino.newMinos() {
				if s := newMino.Canonical().String(); !found[s] {
					minos = append(minos, newMino.Canonical())
					found[s] = true
				}
			}
		}

		return minos, nil
	}
}

func monomino() Mino {
	return Mino{{0, 0}}
}
