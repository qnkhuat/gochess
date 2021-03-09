//// Demo code for a timer based update
//package main
//
//import (
//	"fmt"
//	"time"
//	"github.com/rivo/tview"
//	"bufio"
//	"io"
//	"net"
//	"log"
//	"strconv"
//)
//
//const refreshInterval = 500 * time.Millisecond
//
//var (
//	view *tview.Modal
//	app  *tview.Application
//	count int
//)
//
//func sendMsg(msg string, conn net.Conn){
//	if _, err := io.WriteString(conn, msg); err != nil {
//		log.Fatal(err)
//	}
//}
//
//func recvMsg(conn net.Conn) string{
//	input := bufio.NewScanner(conn)
//	input.Scan()
//	return input.Text()
//}
//
//func update(conn net.Conn){
//	count, _ := strconv.Atoi(recvMsg(conn))
//	app.QueueUpdateDraw(func() {
//		view.SetText(fmt.Sprintf("Count: %d", count))
//	})
//}
//
//func main() {
//	app = tview.NewApplication()
//	conn, err := net.Dial("tcp", ":1998")
//	defer conn.Close()
//	if err != nil {
//		log.Panic(err)
//	}
//
//	view = tview.NewModal().
//		SetText(fmt.Sprintf("Count: %d", count)).
//		AddButtons([]string{"Inc", "Dec", "Quit"}).
//		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
//			if buttonLabel == "Quit" {
//				app.Stop()
//			} else if buttonLabel == "Inc"{
//				count++
//				go sendMsg("inc\n", conn)
//				go update(conn)
//				
//			} else if buttonLabel == "Dec"{
//				count--
//				go sendMsg("dec\n", conn)
//				go update(conn)
//			}
//		})
//
//		count, err = strconv.Atoi(recvMsg(conn))
//		view.SetText(fmt.Sprintf("Count: %d", count))
//	if err := app.SetRoot(view, false).Run(); err != nil {
//		panic(err)
//	}
//}
package main
