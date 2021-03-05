package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/event"

	"git.sr.ht/~tslocum/netris/pkg/game"
	"git.sr.ht/~tslocum/netris/pkg/mino"
	"github.com/mattn/go-isatty"
)

var (
	done = make(chan bool)

	activeGame *game.Game

	connectAddress string
	serverAddress  string
	debugAddress   string
	startMatrix    string

	nicknameFlag string

	blockSize      = 0
	fixedBlockSize bool

	logDebug   bool
	logVerbose bool

	logMutex             = new(sync.Mutex)
	wroteFirstLogMessage bool
	showLogLines         = 7
)

const (
	LogTimeFormat = "3:04:05"
)

func init() {
	log.SetFlags(0)
}

func fibonacci(value int) int {
	if value == 0 || value == 1 {
		return value
	}
	return fibonacci(value-2) + fibonacci(value-1)
}
func main() {
	defer func() {
		if r := recover(); r != nil {
			closeGUI()
			time.Sleep(time.Second)

			log.Println()
			log.Println()
			debug.PrintStack()

			log.Println()
			log.Println()
			log.Fatalf("panic: %+v", r)
		}
	}()

	flag.IntVar(&blockSize, "scale", 0, "UI scale")
	flag.StringVar(&nicknameFlag, "nick", "", "nickname")
	flag.StringVar(&startMatrix, "matrix", "", "pre-fill matrix with pieces")
	flag.StringVar(&connectAddress, "connect", "", "connect to server address or socket path")
	flag.StringVar(&serverAddress, "server", game.DefaultServer, "server address or socket path")
	flag.StringVar(&debugAddress, "debug-address", "", "address to serve debug info")
	flag.BoolVar(&logDebug, "debug", false, "enable debug logging")
	flag.BoolVar(&logVerbose, "verbose", false, "enable verbose logging")
	flag.Parse()

	tty := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	if !tty {
		log.Fatal("failed to start netris: non-interactive terminals are not supported")
	}

	if blockSize > 0 {
		fixedBlockSize = true

		if blockSize > 3 {
			blockSize = 3
		}
	}

	logLevel := game.LogStandard
	if logVerbose {
		logLevel = game.LogVerbose
	} else if logDebug {
		logLevel = game.LogDebug
	}

	if game.Nickname(nicknameFlag) != "" {
		nickname = game.Nickname(nicknameFlag)
	}

	if debugAddress != "" {
		go func() {
			log.Fatal(http.ListenAndServe(debugAddress, nil))
		}()
	}

	app, err := initGUI(connectAddress != "")
	if err != nil {
		log.Fatalf("failed to initialize GUI: %s", err)
	}

	go func() {
		if err := app.Run(); err != nil {
			log.Fatalf("failed to run application: %s", err)
		}

		done <- true
	}()

	logger := make(chan string, game.LogQueueSize)
	go func() {
		for msg := range logger {
			logMessage(msg)
		}
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM)
	go func() {
		<-sigc

		done <- true
	}()

	// Connect automatically when an address or path is supplied
	if connectAddress != "" {
		selectMode <- event.ModePlayOnline
	}

	var (
		server         *game.Server
		localListenDir string
	)

	go func(server *game.Server) {
		<-done
		if server != nil {
			server.StopListening()
		}
		if localListenDir != "" {
			os.RemoveAll(localListenDir)
		}

		closeGUI()

		os.Exit(0)
	}(server)

	for {
		mode := <-selectMode
		switch mode {
		case event.ModePlayOnline:
			joinedGame = true
			setTitleVisible(false)

			if connectAddress == "" {
				connectAddress = serverAddress
			}

			connectNetwork, _ := game.NetworkAndAddress(connectAddress)

			if connectNetwork != "unix" {
				logMessage(fmt.Sprintf("* Connecting to %s...", connectAddress))
				draw <- event.DrawMessages
			}

			s := game.Connect(connectAddress)

			activeGame, err = s.JoinGame(nickname, 0, logger, draw)
			if err != nil {
				log.Fatalf("failed to connect to %s: %s", connectAddress, err)
			}

			if activeGame == nil {
				log.Fatal("failed to connect to server")
			}

			activeGame.LogLevel = logLevel
		case event.ModePractice:
			joinedGame = true
			setTitleVisible(false)

			server = game.NewServer(nil)

			server.Logger = make(chan string, game.LogQueueSize)
			if logDebug || logVerbose {
				go func() {
					for msg := range server.Logger {
						logMessage("Local server: " + msg)
					}
				}()
			} else {
				go func() {
					for range server.Logger {
					}
				}()
			}

			localListenDir, err = ioutil.TempDir("", "netris")
			if err != nil {
				log.Fatal(err)
			}

			localListenAddress := path.Join(localListenDir, "netris.sock")

			go server.Listen(localListenAddress)

			localServerConn := game.Connect(localListenAddress)

			activeGame, err = localServerConn.JoinGame(nickname, -1, logger, draw)
			if err != nil {
				panic(err)
			}

			activeGame.LogLevel = logLevel

			if startMatrix != "" {
				activeGame.Players[activeGame.LocalPlayer].Matrix.Lock()
				startMatrixSplit := strings.Split(startMatrix, ",")
				startMatrix = ""
				var (
					token int
					x     int
					err   error
				)
				for i := range startMatrixSplit {
					token, err = strconv.Atoi(startMatrixSplit[i])
					if err != nil {
						panic(fmt.Sprintf("failed to parse initial matrix on token #%d", i))
					}
					if i%2 == 1 {
						activeGame.Players[activeGame.LocalPlayer].Matrix.SetBlock(x, token, mino.BlockGarbage, false)
					} else {
						x = token
					}
				}
				activeGame.Players[activeGame.LocalPlayer].Matrix.Unlock()
			}
		}
	}
}
