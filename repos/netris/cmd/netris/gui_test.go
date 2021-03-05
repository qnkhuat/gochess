package main

import (
	"testing"

	"git.sr.ht/~tslocum/netris/pkg/mino"
)

func TestRenderMatrix(t *testing.T) {
	renderLock.Lock()
	defer renderLock.Unlock()

	blockSize = 1

	m, err := mino.NewTestMatrix()
	if err != nil {
		t.Error(err)
	}

	m.AddTestBlocks()

	renderMatrix(m)
}

func BenchmarkRenderStandardMatrix(b *testing.B) {
	renderLock.Lock()
	defer renderLock.Unlock()

	blockSize = 1

	m, err := mino.NewTestMatrix()
	if err != nil {
		b.Error(err)
	}

	m.AddTestBlocks()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		renderMatrix(m)
	}
}

func BenchmarkRenderLargeMatrix(b *testing.B) {
	renderLock.Lock()
	defer renderLock.Unlock()

	blockSize = 2

	m, err := mino.NewTestMatrix()
	if err != nil {
		b.Error(err)
	}

	m.AddTestBlocks()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		renderMatrix(m)
	}

	blockSize = 1
}
