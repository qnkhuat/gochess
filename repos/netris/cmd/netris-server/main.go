package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/game"
	"git.sr.ht/~tslocum/netris/pkg/game/ssh"
)

var (
	listenAddressTCP    string
	listenAddressSocket string
	listenAddressSSH    string
	netrisBinary        string
	debugAddress        string

	logDebug   bool
	logVerbose bool

	done = make(chan bool)
)

const (
	LogTimeFormat = "2006-01-02 15:04:05"
)

func init() {
	log.SetFlags(0)

	flag.StringVar(&listenAddressTCP, "listen-tcp", "", "host server on network address")
	flag.StringVar(&listenAddressSocket, "listen-socket", "", "host server on socket path")
	flag.StringVar(&listenAddressSSH, "listen-ssh", "", "host SSH server on network address")
	flag.StringVar(&netrisBinary, "netris", "", "path to netris client")
	flag.StringVar(&debugAddress, "debug-address", "", "address to serve debug info")
	flag.BoolVar(&logDebug, "debug", false, "enable debug logging")
	flag.BoolVar(&logVerbose, "verbose", false, "enable verbose logging")
}

func main() {
	flag.Parse()

	if listenAddressTCP == "" && listenAddressSocket == "" {
		log.Fatal("at least one listen path or address is required (--listen-tcp and/or --listen-socket)")
	}

	if debugAddress != "" {
		go func() {
			log.Fatal(http.ListenAndServe(debugAddress, nil))
		}()
	}

	netrisAddress := listenAddressSocket
	if netrisAddress == "" {
		netrisAddress = listenAddressTCP
	}

	sshServer := &ssh.SSHServer{ListenAddress: listenAddressSSH, NetrisBinary: netrisBinary, NetrisAddress: netrisAddress}

	server := game.NewServer([]game.ServerInterface{sshServer})

	logger := make(chan string, game.LogQueueSize)
	go func() {
		for msg := range logger {
			log.Println(time.Now().Format(LogTimeFormat) + " " + msg)
		}
	}()

	server.Logger = logger

	if listenAddressSocket != "" {
		go server.Listen(listenAddressSocket)
	}
	if listenAddressTCP != "" {
		go server.Listen(listenAddressTCP)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM)
	go func() {
		<-sigc

		done <- true
	}()

	<-done

	server.StopListening()

	/*
		i, err := strconv.Atoi(flag.Arg(0))
		if err != nil {
			panic(err)
		}

		minos, err := mino.Generate(i)
		if err != nil {
			panic(err)
		}
		for _, m := range minos {
			log.Println(m.Render())
			log.Println()
			log.Println()
		}*/
}
