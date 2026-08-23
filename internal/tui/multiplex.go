package tui

import (
	"io"
	"net"
	"os"
	"sync"
)

// Multiplexer broadcasts Output to multiple clients and merges Input from them.
type Multiplexer struct {
	mu      sync.Mutex
	writers []io.Writer

	inReader *io.PipeReader
	inWriter *io.PipeWriter
}

func NewMultiplexer() *Multiplexer {
	r, w := io.Pipe()
	m := &Multiplexer{
		writers:  []io.Writer{os.Stdout}, // default local
		inReader: r,
		inWriter: w,
	}

	// Start a goroutine to copy os.Stdin to our multiplexed input
	go func() {
		_, _ = io.Copy(m.inWriter, os.Stdin)
	}()

	return m
}

func (m *Multiplexer) AddClient(conn net.Conn) {
	m.mu.Lock()
	m.writers = append(m.writers, conn)
	m.mu.Unlock()

	// Copy from client to our merged input
	go func() {
		_, _ = io.Copy(m.inWriter, conn)

		// Remove client on disconnect
		m.mu.Lock()
		for i, w := range m.writers {
			if w == conn {
				m.writers = append(m.writers[:i], m.writers[i+1:]...)
				break
			}
		}
		m.mu.Unlock()
	}()
}

// Write broadcasts to all connected writers
func (m *Multiplexer) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, w := range m.writers {
		// Ignore errors from disconnected clients, they are cleaned up in the read loop
		_, _ = w.Write(p)
	}
	return len(p), nil
}

// Read reads from the merged inputs
func (m *Multiplexer) Read(p []byte) (n int, err error) {
	return m.inReader.Read(p)
}
