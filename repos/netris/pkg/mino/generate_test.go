package mino

import (
	"testing"
)

func TestGenerate(t *testing.T) {
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
			t.Errorf("failed to generate minos for rank %d: expected to generate %d minos, got %d - %s", d.Rank, len(d.Minos), len(minos), minos)
		}

		found := make(map[string]int)

		for _, ex := range d.Minos {
			for _, m := range minos {
				if m.String() == ex {
					found[m.String()]++
				}
			}
		}

		if len(found) != len(d.Minos) {
			t.Errorf("failed to generate minos for rank %d: got unexpected minos %s", d.Rank, minos)
		}

		for _, ms := range d.Minos {
			if found[ms] != 1 {
				t.Errorf("failed to generate minos for rank %d: expected to generate 1 mino %s, got %d", d.Rank, ms, found[ms])
			}
		}
	}
}

func BenchmarkGenerate(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	var (
		minos []Mino
		err   error
	)
	for n := 0; n < b.N; n++ {
		for _, d := range minoTestData {
			minos, err = Generate(d.Rank)
			if err != nil {
				b.Errorf("failed to generate minos: %s", err)
			}

			if len(minos) != len(d.Minos) {
				b.Errorf("failed to generate minos for rank %d: expected to generate %d minos, got %d", d.Rank, len(d.Minos), len(minos))
			}
		}
	}
}
