This document covers command-line options available when launching netris. Some
options (such as keybindings) are only available in-game.

# Client

```
Usage of ./netris:
  -connect string
        connect to server address or socket path
  -debug
        enable debug logging
  -debug-address string
        address to serve debug info
  -matrix string
        pre-fill matrix with pieces
  -nick string
        nickname (default "Anonymous")
  -scale int
        UI scale
  -verbose
        enable verbose logging
```

### -scale

Defaults to 0, automatically scaling up to 3x standard size based on the
terminal window's dimensions.

Specify an integer to only use that UI scale.

### -connect

A TCP address in the form of address:port or socket path may be supplied.

# Server

```
Usage of ./netris-server:
  -debug
        enable debug logging
  -debug-address string
        address to serve debug info
  -listen-socket string
        host server on socket path
  -listen-ssh string
        host SSH server on network address
  -listen-tcp string
        host server on network address
  -netris string
        path to netris client
  -verbose
        enable verbose logging
```

### -listen-ssh and -netris

The netris client will be launched to serve incoming SSH connections. Update
client and server binaries together.
