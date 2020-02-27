package daemon

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/mysteriumnetwork/node-supervisor/config"
)

const sock = "/var/run/myst.sock"

// Daemon - supervisor process.
type Daemon struct {
	cfg *config.Config
}

// New creates a new daemon.
func New(cfg *config.Config) Daemon {
	return Daemon{cfg: cfg}
}

// Start supervisor daemon. Blocks.
func (d Daemon) Start() error {
	if err := os.RemoveAll(sock); err != nil {
		return fmt.Errorf("could not remove sock: %w", err)
	}
	l, err := net.Listen("unix", sock)
	if err != nil {
		return fmt.Errorf("error listening: %w", err)
	}
	if err := os.Chmod(sock, 0777); err != nil {
		return fmt.Errorf("failed to chmod the sock: %w", err)
	}
	defer func() {
		if err := l.Close(); err != nil {
			log.Println("Error closing listener:", err)
		}
	}()
	for {
		log.Println("Waiting for connections...")
		conn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("accept error: %w", err)
		}
		go func() {
			peer := conn.RemoteAddr().Network()
			log.Println("Client connected:", peer)
			d.serve(conn)
			if err := conn.Close(); err != nil {
				log.Println("Error closing connection for:", peer, err)
			}
			log.Println("Client disconnected:", peer)
		}()
	}
}

// serve talks to the client via established connection.
func (d Daemon) serve(c net.Conn) {
	scan := bufio.NewScanner(c)
	for scan.Scan() {
		line := scan.Bytes()
		cmd := strings.Split(string(line), " ")
		op := cmd[0]
		switch op {
		case "BYE":
			message(c, "BYE")
			return
		case "PING":
			message(c, "PONG")
			return
		case "RUN":
			go func() {
				err := d.RunMyst()
				if err != nil {
					log.Println("Error running myst:", err)
				}
			}()
		case "KILL":
		}
	}
}

func message(w io.Writer, msg string) {
	if _, err := fmt.Fprintln(w, msg); err != nil {
		log.Println("Could not send message:", msg)
	}
}
