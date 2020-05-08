// Copyright (c) 2020 BlockDev AG
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package transport

import (
	"fmt"
	"log"
	"net"
	"os"
)

const sock = "/var/run/myst.sock"

// Start starts a listener on a unix domain socket.
// Conversation is handled by the handlerFunc.
func Start(handle handlerFunc) error {
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
			handle(conn)
			if err := conn.Close(); err != nil {
				log.Println("Error closing connection for:", peer, err)
			}
			log.Println("Client disconnected:", peer)
		}()
	}
}
