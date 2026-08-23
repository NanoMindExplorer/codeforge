package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/codeforge/tui/internal/tui"
	"golang.org/x/term"
)

func getSockPath() string {
	return filepath.Join(os.TempDir(), "codeforge.sock")
}

func startMultiplayerServer(m *tui.Multiplexer) error {
	sock := getSockPath()
	os.Remove(sock)

	l, err := net.Listen("unix", sock)
	if err != nil {
		return err
	}

	go func() {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				continue
			}
			m.AddClient(conn)
		}
	}()
	return nil
}

func runAttach() int {
	sock := getSockPath()
	conn, err := net.Dial("unix", sock)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not attach to server at %s: %v\n", sock, err)
		return 1
	}
	defer conn.Close()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error putting terminal into raw mode: %v\n", err)
		return 1
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	go io.Copy(conn, os.Stdin)
	io.Copy(os.Stdout, conn)
	return 0
}
