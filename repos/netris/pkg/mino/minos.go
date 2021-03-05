package mino

type Minos []Mino

func (ms Minos) Has(m Mino) bool {
	for _, msm := range ms {
		if msm.Equal(m) {
			return true
		}
	}

	return false
}
