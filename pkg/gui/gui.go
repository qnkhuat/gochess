package main

import (
	"fmt"
	"strings"
	"strconv"

	//"github.com/gdamore/tcell/v2"
	//"github.com/rivo/tview"
	"github.com/notnil/chess"
)

func main() {
	//app := tview.NewApplication()
	game := chess.NewGame()
	//fmt.Println(game.String())
	//gameStr := game.Position().Board().Draw()
	gameStr := strings.Split(game.Position().Board().Draw(), "\n")
	
	for _, s := range gameStr{
		for _, c := range s {
			fmt.Printf("%s", strconv.Atoi(c))
		}
		//fmt.Println(s)
		fmt.Printf("\n")
	}
		




	//table := tview.NewTable().
	//	SetBorders(true).
	//	SetSelectable(true, true)
	//lorem := strings.Split("Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.", " ")
	//cols, rows := 10, 40
	//word := 0
	//for r := 0; r < rows; r++ {
	//	for c := 0; c < cols; c++ {
	//		color := tcell.ColorWhite
	//		if c < 1 || r < 1 {
	//			color = tcell.ColorYellow
	//		}
	//		table.SetCell(r, c,
	//			tview.NewTableCell(lorem[word]).
	//			SetTextColor(color).
	//			SetAlign(tview.AlignCenter))
	//		word = (word + 1) % len(lorem)
	//	}
	//}
	//table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
	//	if key == tcell.KeyEscape {
	//		app.Stop()
	//	}
	//	if key == tcell.KeyEnter {
	//		table.SetSelectable(true, true)
	//	}
	//}).SetSelectedFunc(func(row int, column int) {
	//	table.GetCell(row, column).SetTextColor(tcell.ColorRed)
	//	table.SetSelectable(false, false)
	//})
	//if err := app.SetRoot(table, true).SetFocus(table).Run(); err != nil {
	//	panic(err)
	//}
}

