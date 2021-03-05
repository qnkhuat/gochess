//+build windows

package ssh

import "git.sr.ht/~tslocum/netris/pkg/game"

// SSH server is unsupported on Windows

type SSHServer struct {
	ListenAddress string
	NetrisBinary  string
	NetrisAddress string
}

func (s *SSHServer) Host(newPlayers chan<- *game.IncomingPlayer) {
}

func (s *SSHServer) Shutdown(reason string) {
}
