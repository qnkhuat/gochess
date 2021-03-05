package mino

import (
	"math/rand"
	"sync"
)

type Bag struct {
	Minos    []Mino
	Original []Mino

	minoRandomizer    *rand.Rand
	garbageRandomizer *rand.Rand

	i     int
	width int
	*sync.Mutex
}

func NewBag(seed int64, minos []Mino, width int) (*Bag, error) {
	minoSource := rand.NewSource(seed)
	garbageSource := rand.NewSource(seed)
	b := &Bag{Original: minos, minoRandomizer: rand.New(minoSource), garbageRandomizer: rand.New(garbageSource), width: width, Mutex: new(sync.Mutex)}

	b.shuffle()

	return b, nil
}

func (b *Bag) Take() Mino {
	b.Lock()
	defer b.Unlock()

	mino := b.Minos[b.i]
	if b.i == len(b.Minos)-1 {
		b.shuffle()

		b.i = 0
	} else {
		b.i++
	}

	return mino
}

func (b *Bag) Next() Mino {
	b.Lock()
	defer b.Unlock()

	return b.Minos[b.i]
}

func (b *Bag) shuffle() {
	if b.Minos == nil {
		b.Minos = make([]Mino, len(b.Original))
	}
	copy(b.Minos, b.Original)

	b.minoRandomizer.Shuffle(len(b.Minos), func(i, j int) { b.Minos[i], b.Minos[j] = b.Minos[j], b.Minos[i] })
}

func (b *Bag) GarbageHole() int {
	b.Lock()
	defer b.Unlock()

	return b.garbageRandomizer.Intn(b.width)
}
