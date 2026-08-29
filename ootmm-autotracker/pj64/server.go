package pj64

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

const (
	DefaultPort  = 55190
	MaxReadSize  = 2048
	WriteTimeout = 5 * time.Second
	ReadTimeout  = 10 * time.Second

	opReadBulk = 10

	// adapterGreeting is the 4-byte greeting the PJ64 Lua adapter must
	// send immediately after connecting.  The trailing digit is the
	// adapter protocol version: only scripts presenting the current
	// greeting are served, so outdated Lua scripts are rejected instead
	// of feeding the tracker stale data.  Keep in sync with
	// GREETING_MAGIC/PROTOCOL_VERSION in tlt_autotracking.lua.
	adapterGreeting = "OAT2"

	// greetingLogInterval throttles warning logs while an incompatible
	// adapter keeps reconnecting.
	greetingLogInterval = 30 * time.Second
)

// greetingTimeout is how long the server waits for the adapter greeting
// before dropping the connection.  It is a variable so tests can shorten
// it; the new adapter sends the greeting immediately after connecting.
var greetingTimeout = 2 * time.Second

// Server listens for a PJ64 Lua adapter connection and provides
// memory reads over a simple binary protocol.
type Server struct {
	port            int
	listener        net.Listener
	conn            net.Conn
	mu              sync.Mutex
	lastGreetingLog time.Time
}

func NewServer(port int) *Server {
	return &Server{port: port}
}

// Start begins listening on the configured TCP port.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("listen on port %d: %w", s.port, err)
	}
	s.listener = ln
	log.Printf("PJ64 server listening on port %d", s.port)
	return nil
}

// Connect waits for a PJ64 client to connect and verifies the adapter
// greeting (with a short timeout so the main loop can continue polling).
// Implements the same lifecycle interface as retroarch.Client.Connect().
func (s *Server) Connect() error {
	if s.listener == nil {
		return fmt.Errorf("server not started")
	}
	s.listener.(*net.TCPListener).SetDeadline(time.Now().Add(time.Second))
	conn, err := s.listener.Accept()
	if err != nil {
		return err
	}

	// Only serve adapters that identify themselves with the current
	// greeting, so outdated Lua scripts cannot feed the tracker stale
	// data.
	conn.SetReadDeadline(time.Now().Add(greetingTimeout))
	greeting := make([]byte, len(adapterGreeting))
	if _, err := io.ReadFull(conn, greeting); err != nil {
		conn.Close()
		s.logGreetingMismatch(fmt.Sprintf(
			"PJ64 adapter rejected: no greeting received (%v); please load the current tlt_autotracking_v2.lua from the ootmm-autotracker directory", err))
		return fmt.Errorf("adapter greeting: %w", err)
	}
	if string(greeting) != adapterGreeting {
		conn.Close()
		s.logGreetingMismatch(fmt.Sprintf(
			"PJ64 adapter rejected: greeting %q, want %q; please load the current tlt_autotracking_v2.lua from the ootmm-autotracker directory", string(greeting), adapterGreeting))
		return fmt.Errorf("unsupported adapter greeting %q (want %q)", string(greeting), adapterGreeting)
	}

	s.mu.Lock()
	if s.conn != nil {
		s.conn.Close()
	}
	s.conn = conn
	s.mu.Unlock()
	log.Printf("PJ64 client connected from %s", conn.RemoteAddr())
	return nil
}

// logGreetingMismatch logs msg at most once per greetingLogInterval so
// an incompatible adapter that keeps reconnecting does not spam the log.
func (s *Server) logGreetingMismatch(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if time.Since(s.lastGreetingLog) < greetingLogInterval {
		return
	}
	s.lastGreetingLog = time.Now()
	log.Println(msg)
}

// Close closes the current PJ64 connection (keeps the listener open for reconnects).
func (s *Server) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}

// Stop closes both the connection and the listener.
func (s *Server) Stop() {
	s.Close()
	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}
}

// IsConnected returns true if a PJ64 client is connected.
func (s *Server) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn != nil
}

// ReadMemory reads `size` bytes from the PJ64 emulator's N64 memory at `addr`.
// Bytes are returned in N64 native order (big-endian), not mupen64plus-swizzled.
func (s *Server) ReadMemory(addr uint32, size int) ([]byte, error) {
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()

	if conn == nil {
		return nil, fmt.Errorf("not connected")
	}
	if size > MaxReadSize {
		return nil, fmt.Errorf("read size %d exceeds max %d", size, MaxReadSize)
	}

	// Send: op(1) + addr(4) + size(4) = 9 bytes, all little-endian
	cmd := make([]byte, 9)
	cmd[0] = opReadBulk
	binary.LittleEndian.PutUint32(cmd[1:5], addr)
	binary.LittleEndian.PutUint32(cmd[5:9], uint32(size))

	conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
	if _, err := conn.Write(cmd); err != nil {
		s.disconnect()
		return nil, fmt.Errorf("write: %w", err)
	}

	buf := make([]byte, size)
	conn.SetReadDeadline(time.Now().Add(ReadTimeout))
	if _, err := io.ReadFull(conn, buf); err != nil {
		s.disconnect()
		return nil, fmt.Errorf("read response: %w", err)
	}

	return buf, nil
}

// ReadMemoryLarge reads memory in chunks when size exceeds MaxReadSize.
func (s *Server) ReadMemoryLarge(addr uint32, size int) ([]byte, error) {
	if size <= MaxReadSize {
		return s.ReadMemory(addr, size)
	}

	result := make([]byte, 0, size)
	for offset := 0; offset < size; offset += MaxReadSize {
		chunkSize := MaxReadSize
		if offset+chunkSize > size {
			chunkSize = size - offset
		}
		chunk, err := s.ReadMemory(addr+uint32(offset), chunkSize)
		if err != nil {
			return nil, fmt.Errorf("chunk at offset 0x%x: %w", offset, err)
		}
		result = append(result, chunk...)
	}
	return result, nil
}

func (s *Server) disconnect() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}
