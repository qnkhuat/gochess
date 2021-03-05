package main

import (
	"log"
	"math/rand"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/event"
	"git.sr.ht/~tslocum/netris/pkg/game"
	"git.sr.ht/~tslocum/netris/pkg/mino"
	"github.com/tslocum/tview"
)

var (
	titleVisible               bool
	titleScreen                int
	titleSelectedButton        int
	gameSettingsSelectedButton int
	drawTitle                  = make(chan struct{}, game.CommandQueueSize)

	titleGrid          *tview.Grid
	titleContainerGrid *tview.Grid

	playerSettingsForm          *tview.Form
	playerSettingsGrid          *tview.Grid
	playerSettingsContainerGrid *tview.Grid

	gameSettingsGrid          *tview.Grid
	gameSettingsContainerGrid *tview.Grid
	gameGrid                  *tview.Grid

	titleName *tview.TextView
	titleL    *tview.TextView
	titleR    *tview.TextView

	titleMatrixL = newTitleMatrixSide()
	titleMatrix  = newTitleMatrixName()
	titleMatrixR = newTitleMatrixSide()
	titlePiecesL []*mino.Piece
	titlePiecesR []*mino.Piece

	buttonA *tview.Button
	buttonB *tview.Button
	buttonC *tview.Button

	buttonLabelA *tview.TextView
	buttonLabelB *tview.TextView
	buttonLabelC *tview.TextView
)

func previousTitleButton() {
	if titleSelectedButton == 0 {
		return
	}

	titleSelectedButton--
}

func nextTitleButton() {
	if titleSelectedButton == 2 {
		return
	}

	titleSelectedButton++
}

func updateGameSettings() {
	switch gameSettingsSelectedButton {
	case 0:
		app.SetFocus(buttonKeybindRotateCCW)
	case 1:
		app.SetFocus(buttonKeybindRotateCW)
	case 2:
		app.SetFocus(buttonKeybindMoveLeft)
	case 3:
		app.SetFocus(buttonKeybindMoveRight)
	case 4:
		app.SetFocus(buttonKeybindSoftDrop)
	case 5:
		app.SetFocus(buttonKeybindHardDrop)
	case 6:
		app.SetFocus(buttonKeybindCancel)
	case 7:
		app.SetFocus(buttonKeybindSave)
	}
}

func setTitleVisible(visible bool) {
	if titleVisible == visible {
		return
	}

	titleVisible = visible

	if !titleVisible {
		app.SetRoot(gameGrid, true)

		app.SetFocus(nil)
	} else {
		titleScreen = 0
		titleSelectedButton = 0

		drawTitle <- struct{}{}

		app.SetRoot(titleContainerGrid, true)

		updateTitle()
	}
}

func updateTitle() {
	if titleScreen == 1 {
		buttonA.SetLabel("Player Settings")
		buttonLabelA.SetText("\nChange name")

		buttonB.SetLabel("Game Settings")
		buttonLabelB.SetText("\nChange keybindings")

		buttonC.SetLabel("Return")
		buttonLabelC.SetText("\nReturn to the last screen")
	} else {
		if joinedGame {
			buttonA.SetLabel("Resume")
			buttonLabelA.SetText("\nResume game in progress")

			buttonB.SetLabel("Settings")
			buttonLabelB.SetText("\nChange player name, keybindings, etc")

			buttonC.SetLabel("Quit")
			buttonLabelC.SetText("\nQuit game")
		} else {
			buttonA.SetLabel("Play")
			buttonLabelA.SetText("\nPlay with others online")

			buttonB.SetLabel("Practice")
			buttonLabelB.SetText("\nPlay by yourself")

			buttonC.SetLabel("Settings")
			buttonLabelC.SetText("\nPlayer name, keybindings, etc.")
		}
	}

	if titleScreen > 1 {
		return
	}

	switch titleSelectedButton {
	case 1:
		app.SetFocus(buttonB)
	case 2:
		app.SetFocus(buttonC)
	default:
		app.SetFocus(buttonA)
	}
}

func handleTitle() {
	var t *time.Ticker
	for {
		if t == nil {
			t = time.NewTicker(850 * time.Millisecond)
		} else {
			select {
			case <-t.C:
			case <-drawTitle:
				if t != nil {
					t.Stop()
				}

				t = time.NewTicker(850 * time.Millisecond)
			}
		}

		if !titleVisible {
			continue
		}

		titleMatrixL.ClearOverlay()

		for _, p := range titlePiecesL {
			p.Y -= 1
			if p.Y < -3 {
				p.Y = titleMatrixL.H + 2
			}
			if rand.Intn(4) == 0 {
				p.Mino = p.Rotate(1, 0)
				p.ApplyRotation(1, 0)
			}

			for _, m := range p.Mino {
				titleMatrixL.SetBlock(p.X+m.X, p.Y+m.Y, p.Solid, true)
			}
		}

		titleMatrixR.ClearOverlay()

		for _, p := range titlePiecesR {
			p.Y -= 1
			if p.Y < -3 {
				p.Y = titleMatrixL.H + 2
			}
			if rand.Intn(4) == 0 {
				p.Mino = p.Rotate(1, 0)
				p.ApplyRotation(1, 0)
			}

			for _, m := range p.Mino {
				if titleMatrixR.Block(p.X+m.X, p.Y+m.Y) != mino.BlockNone {
					continue
				}

				titleMatrixR.SetBlock(p.X+m.X, p.Y+m.Y, p.Solid, true)
			}
		}

		app.QueueUpdateDraw(renderTitle)
	}
}

func renderTitle() {
	var newBlock mino.Block
	for i, b := range titleMatrix.M {
		switch b {
		case mino.BlockSolidRed:
			newBlock = mino.BlockSolidMagenta
		case mino.BlockSolidYellow:
			newBlock = mino.BlockSolidRed
		case mino.BlockSolidGreen:
			newBlock = mino.BlockSolidYellow
		case mino.BlockSolidCyan:
			newBlock = mino.BlockSolidGreen
		case mino.BlockSolidBlue:
			newBlock = mino.BlockSolidCyan
		case mino.BlockSolidMagenta:
			newBlock = mino.BlockSolidBlue
		default:
			continue
		}

		titleMatrix.M[i] = newBlock
	}

	renderLock.Lock()

	renderMatrix(titleMatrix)
	titleName.Clear()
	titleName.Write(renderBuffer.Bytes())

	renderMatrix(titleMatrixL)
	titleL.Clear()
	titleL.Write(renderBuffer.Bytes())

	renderMatrix(titleMatrixR)
	titleR.Clear()
	titleR.Write(renderBuffer.Bytes())

	renderLock.Unlock()
}

func newTitleMatrixSide() *mino.Matrix {
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

	m := mino.NewMatrix(21, 24, 0, 1, ev, draw, mino.MatrixCustom)

	return m
}

func newTitleMatrixName() *mino.Matrix {
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

	m := mino.NewMatrix(36, 7, 0, 1, ev, draw, mino.MatrixCustom)

	baseStart := 1
	centerStart := (m.W / 2) - 17

	var titleBlocks = []struct {
		mino.Point
		mino.Block
	}{
		// N
		{mino.Point{0, 0}, mino.BlockSolidRed},
		{mino.Point{0, 1}, mino.BlockSolidRed},
		{mino.Point{0, 2}, mino.BlockSolidRed},
		{mino.Point{0, 3}, mino.BlockSolidRed},
		{mino.Point{0, 4}, mino.BlockSolidRed},
		{mino.Point{1, 3}, mino.BlockSolidRed},
		{mino.Point{2, 2}, mino.BlockSolidRed},
		{mino.Point{3, 1}, mino.BlockSolidRed},
		{mino.Point{4, 0}, mino.BlockSolidRed},
		{mino.Point{4, 1}, mino.BlockSolidRed},
		{mino.Point{4, 2}, mino.BlockSolidRed},
		{mino.Point{4, 3}, mino.BlockSolidRed},
		{mino.Point{4, 4}, mino.BlockSolidRed},

		// E
		{mino.Point{7, 0}, mino.BlockSolidYellow},
		{mino.Point{7, 1}, mino.BlockSolidYellow},
		{mino.Point{7, 2}, mino.BlockSolidYellow},
		{mino.Point{7, 3}, mino.BlockSolidYellow},
		{mino.Point{7, 4}, mino.BlockSolidYellow},
		{mino.Point{8, 0}, mino.BlockSolidYellow},
		{mino.Point{9, 0}, mino.BlockSolidYellow},
		{mino.Point{8, 2}, mino.BlockSolidYellow},
		{mino.Point{9, 2}, mino.BlockSolidYellow},
		{mino.Point{8, 4}, mino.BlockSolidYellow},
		{mino.Point{9, 4}, mino.BlockSolidYellow},

		// T
		{mino.Point{12, 4}, mino.BlockSolidGreen},
		{mino.Point{13, 4}, mino.BlockSolidGreen},
		{mino.Point{14, 0}, mino.BlockSolidGreen},
		{mino.Point{14, 1}, mino.BlockSolidGreen},
		{mino.Point{14, 2}, mino.BlockSolidGreen},
		{mino.Point{14, 3}, mino.BlockSolidGreen},
		{mino.Point{14, 4}, mino.BlockSolidGreen},
		{mino.Point{15, 4}, mino.BlockSolidGreen},
		{mino.Point{16, 4}, mino.BlockSolidGreen},

		// R
		{mino.Point{19, 0}, mino.BlockSolidCyan},
		{mino.Point{19, 1}, mino.BlockSolidCyan},
		{mino.Point{19, 2}, mino.BlockSolidCyan},
		{mino.Point{19, 3}, mino.BlockSolidCyan},
		{mino.Point{19, 4}, mino.BlockSolidCyan},
		{mino.Point{20, 2}, mino.BlockSolidCyan},
		{mino.Point{20, 4}, mino.BlockSolidCyan},
		{mino.Point{21, 2}, mino.BlockSolidCyan},
		{mino.Point{21, 4}, mino.BlockSolidCyan},
		{mino.Point{22, 0}, mino.BlockSolidCyan},
		{mino.Point{22, 1}, mino.BlockSolidCyan},
		{mino.Point{22, 3}, mino.BlockSolidCyan},

		// I
		{mino.Point{25, 0}, mino.BlockSolidBlue},
		{mino.Point{25, 1}, mino.BlockSolidBlue},
		{mino.Point{25, 2}, mino.BlockSolidBlue},
		{mino.Point{25, 3}, mino.BlockSolidBlue},
		{mino.Point{25, 4}, mino.BlockSolidBlue},

		// S
		{mino.Point{28, 0}, mino.BlockSolidMagenta},
		{mino.Point{29, 0}, mino.BlockSolidMagenta},
		{mino.Point{30, 0}, mino.BlockSolidMagenta},
		{mino.Point{31, 1}, mino.BlockSolidMagenta},
		{mino.Point{29, 2}, mino.BlockSolidMagenta},
		{mino.Point{30, 2}, mino.BlockSolidMagenta},
		{mino.Point{28, 3}, mino.BlockSolidMagenta},
		{mino.Point{29, 4}, mino.BlockSolidMagenta},
		{mino.Point{30, 4}, mino.BlockSolidMagenta},
		{mino.Point{31, 4}, mino.BlockSolidMagenta},
	}

	for _, titleBlock := range titleBlocks {
		if !m.SetBlock(centerStart+titleBlock.X, baseStart+titleBlock.Y, titleBlock.Block, false) {
			log.Fatalf("failed to set title block %s", titleBlock.Point)
		}
	}

	return m
}
