# GoChess
Play chess with your friends on terminal!
# Play
`ssh gochess.club`

To play in singple player mode ( against stockfish bot ), just type `practice`

To play with your friend:
- Create a room with `create [roomname]`
- Tell your friend to join with command `create [roomname]`


# Screenshots
### Menu
![](./statics/menu.png)
### Game play
![](./statics/gochess.png)

# Libaries
The following libraries are used to build gochess:
- [notnil/chess](https://github.com/notnil/chess) - Chess engine
- [rivo/tview](https://github.com/rivo/tview) - UI
- [creack/pty](https://github.com/creack/pty) - Pseudo-terminal interfacej
- [gliderlabs/ssh](https://github.com/gliderlabs/ssh) - SSH server
# TODO
- [x] Single player mode
- [ ] Add mouse control
- [ ] Add timer
- [ ] Hint moves 

# Disclaimer
I'm building this project while learning Go. So any Comments on code quality/logic will be much appricated!

Please add your comment in this [issue](https://github.com/qnkhuat/chessterm/issues/1) if you have any!
