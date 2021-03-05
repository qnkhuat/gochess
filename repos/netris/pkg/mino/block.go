package mino

type Block int

func (b Block) String() string {
	return string(b.Rune())
}

func (b Block) Rune() rune {
	switch b {
	case BlockNone:
		return ' '
	case BlockGhostBlue, BlockGhostCyan, BlockGhostRed, BlockGhostYellow, BlockGhostMagenta, BlockGhostGreen, BlockGhostOrange:
		return '▓'
	case BlockGarbage, BlockSolidBlue, BlockSolidCyan, BlockSolidRed, BlockSolidYellow, BlockSolidMagenta, BlockSolidGreen, BlockSolidOrange:
		return '█'
	default:
		return '?'
	}
}

const (
	BlockNone Block = iota
	BlockGarbage
	BlockGhostBlue
	BlockGhostCyan
	BlockGhostRed
	BlockGhostYellow
	BlockGhostMagenta
	BlockGhostGreen
	BlockGhostOrange
	BlockSolidBlue
	BlockSolidCyan
	BlockSolidRed
	BlockSolidYellow
	BlockSolidMagenta
	BlockSolidGreen
	BlockSolidOrange
)
