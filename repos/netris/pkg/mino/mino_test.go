package mino

type MinoTestData struct {
	Rank  int
	Minos []string
}

var minoTestData = []*MinoTestData{
	{1, []string{Monomino}},
	{2, []string{Domino}},
	{3, []string{TrominoI, TrominoL}},
	{4, []string{TetrominoI, TetrominoO, TetrominoT, TetrominoS, TetrominoZ, TetrominoJ, TetrominoL}},
	{5, []string{PentominoF, PentominoE, PentominoJ, PentominoL, PentominoI, PentominoN, PentominoG, PentominoP, PentominoB, PentominoS, PentominoT, PentominoU, PentominoV, PentominoW, PentominoX, PentominoY, PentominoR, PentominoZ}}}
