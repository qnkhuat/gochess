go build -gcflags '-N -l'
if [ $? -eq 0 ]
then
	./server -log /Users/earther/fun/7_chessterm/log -ssh :2022
fi
