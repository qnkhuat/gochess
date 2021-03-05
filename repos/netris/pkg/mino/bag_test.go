package mino

import (
	"testing"
)

func TestBag(t *testing.T) {
	var (
		minos []Mino
		err   error
	)
	for _, d := range minoTestData {
		minos, err = Generate(d.Rank)
		if err != nil {
			t.Errorf("failed to generate minos for rank %d: %s", d.Rank, err)
		}

		if len(minos) != len(d.Minos) {
			t.Errorf("failed to generate minos for rank %d: unexpected number of minos generated", d.Rank)
		}

		b, err := NewBag(0, minos, 10)
		if err != nil {
			t.Errorf("failed to create bag for rank %d: %s", d.Rank, err)

		}
		taken := make(map[string]int)
		for i := 1; i < 4; i++ {
			for i := 0; i < len(d.Minos); i++ {
				mino := b.Take()
				taken[mino.String()]++
			}

			if len(taken) != len(minos) {
				t.Errorf("minos placed in rank %d bag do not match minos taken - placed: %s - taken: %v", d.Rank, minos, taken)
			}

			for _, mino := range minos {
				if taken[mino.String()] != i {
					t.Fatalf("minos placed in rank %d bag do not match minos taken - placed: %s - taken: %v", d.Rank, minos, taken)
				}
			}
		}
	}
}
