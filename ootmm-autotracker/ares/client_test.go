package ares

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"
)

// gdbStub simulates Ares's GDB remote protocol stub.
type gdbStub struct {
	t        *testing.T
	ln       net.Listener
	addr     string
	commands []string
	mu       chan struct{}
}

func startGDBStub(t *testing.T) *gdbStub {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	s := &gdbStub{
		t:    t,
		ln:   ln,
		addr: ln.Addr().String(),
		mu:   make(chan struct{}, 1),
	}
	return s
}

// handle starts a goroutine that accepts one connection and runs handler.
func (s *gdbStub) handle(handler func(conn net.Conn)) {
	go func() {
		conn, err := s.ln.Accept()
		if err != nil {
			s.t.Logf("accept: %v", err)
			return
		}
		handler(conn)
	}()
}

func (s *gdbStub) Close() {
	s.ln.Close()
}

// TestConnectFailsWithoutServer verifies that Connect returns an error
// when no Ares GDB stub is listening.
func TestConnectFailsWithoutServer(t *testing.T) {
	// Use a port that's unlikely to be open
	client := NewClient(DefaultHost, 51999)
	if err := client.Connect(); err == nil {
		client.Close()
		t.Fatal("expected connect to fail without a GDB server")
	}
	if client.conn != nil {
		client.Close()
		t.Fatal("client retained a connection after failed connect")
	}
}

// TestConnectAndHandshake verifies that Connect succeeds when
// a GDB stub is listening and accepts the initial '+' ack.
func TestConnectAndHandshake(t *testing.T) {
	stub := startGDBStub(t)
	defer stub.Close()

	stub.handle(func(conn net.Conn) {
		defer conn.Close()
		// Expect initial '+' ack
		buf := make([]byte, 1)
		if _, err := conn.Read(buf); err != nil {
			t.Logf("read handshake: %v", err)
			return
		}
		if buf[0] != '+' {
			t.Errorf("expected '+', got %q", buf[0])
		}
	})

	client := NewClient("127.0.0.1", portFromAddr(stub.addr))
	defer client.Close()

	if err := client.Connect(); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if !client.IsConnected() {
		t.Fatal("expected IsConnected to return true after successful connect")
	}
}

// TestReadMemory verifies the full memory read flow using GDB 'm' packet.
func TestReadMemory(t *testing.T) {
	stub := startGDBStub(t)
	defer stub.Close()

	done := make(chan struct{})
	stub.handle(func(conn net.Conn) {
		defer conn.Close()
		defer close(done)

		// Expect initial '+' ack
		buf := make([]byte, 1)
		if _, err := conn.Read(buf); err != nil || buf[0] != '+' {
			t.Logf("handshake: %c err=%v", buf[0], err)
			return
		}

		// Read GDB packet: $m<addr>,<len>#<cksum>
		data := readGDBPacket(t, conn)
		expected := "m00000010,4"
		if data != expected {
			t.Errorf("expected command %q, got %q", expected, data)
		}

		// Send response: $deadbeef#<cksum>
		resp := "$deadbeef#"
		cksum := byte(0)
		for i := 0; i < len("deadbeef"); i++ {
			cksum += "deadbeef"[i]
		}
		resp += fmt.Sprintf("%02x", cksum)
		if _, err := conn.Write([]byte(resp)); err != nil {
			t.Logf("write response: %v", err)
			return
		}

		// Expect '+' ack from client
		if _, err := conn.Read(buf); err != nil {
			t.Logf("read ack: %v", err)
		}
	})

	client := NewClient("127.0.0.1", portFromAddr(stub.addr))
	defer client.Close()

	if err := client.Connect(); err != nil {
		t.Fatalf("connect: %v", err)
	}

	data, err := client.ReadMemory(0x10, 4)
	if err != nil {
		t.Fatalf("ReadMemory: %v", err)
	}

	expected := []byte{0xde, 0xad, 0xbe, 0xef}
	if !bytes.Equal(data, expected) {
		t.Fatalf("ReadMemory = % x, want % x", data, expected)
	}

	<-done
}

// TestReadMemoryLarge verifies that large reads are chunked correctly.
func TestReadMemoryLarge(t *testing.T) {
	stub := startGDBStub(t)
	defer stub.Close()

	commandsRead := make(chan string, 2)
	stub.handle(func(conn net.Conn) {
		defer conn.Close()

		// Handshake
		buf := make([]byte, 1)
		if _, err := conn.Read(buf); err != nil || buf[0] != '+' {
			return
		}

		for i := 0; i < 2; i++ {
			data := readGDBPacket(t, conn)
			commandsRead <- data
			var payload string
			if data == "m00000010,800" {
				// 0x800 bytes of zeros (2048 hex bytes)
				payload = ""
				for j := 0; j < 0x800; j++ {
					payload += "00"
				}
			} else if data == "m00000810,10" {
				// 0x10 bytes of 0x11
				payload = ""
				for j := 0; j < 0x10; j++ {
					payload += "11"
				}
			}
			resp := "$" + payload + "#"
			cksum := byte(0)
			for i := 0; i < len(payload); i++ {
				cksum += payload[i]
			}
			resp += fmt.Sprintf("%02x", cksum)
			if _, err := conn.Write([]byte(resp)); err != nil {
				return
			}
			if _, err := conn.Read(buf); err != nil {
				return
			}
		}
	})

	client := NewClient("127.0.0.1", portFromAddr(stub.addr))
	defer client.Close()

	if err := client.Connect(); err != nil {
		t.Fatalf("connect: %v", err)
	}

	data, err := client.ReadMemoryLarge(0x10, 0x810)
	if err != nil {
		t.Fatalf("ReadMemoryLarge: %v", err)
	}
	if len(data) != 0x810 {
		t.Fatalf("ReadMemoryLarge returned %d bytes, want %d", len(data), 0x810)
	}

	close(commandsRead)
}

// TestReadMemoryGDBError verifies that GDB error responses are handled.
func TestReadMemoryGDBError(t *testing.T) {
	stub := startGDBStub(t)
	defer stub.Close()

	done := make(chan struct{})
	stub.handle(func(conn net.Conn) {
		defer conn.Close()
		defer close(done)

		buf := make([]byte, 1)
		if _, err := conn.Read(buf); err != nil || buf[0] != '+' {
			return
		}

		data := readGDBPacket(t, conn)
		_ = data

		// Send error response: $E01#<cksum>
		resp := "$E01#"
		cksum := byte(0)
		for i := 0; i < len("E01"); i++ {
			cksum += "E01"[i]
		}
		resp += fmt.Sprintf("%02x", cksum)
		if _, err := conn.Write([]byte(resp)); err != nil {
			return
		}

		if _, err := conn.Read(buf); err != nil {
			t.Logf("read ack: %v", err)
		}
	})

	client := NewClient("127.0.0.1", portFromAddr(stub.addr))
	defer client.Close()

	if err := client.Connect(); err != nil {
		t.Fatalf("connect: %v", err)
	}

	if _, err := client.ReadMemory(0x10, 4); err == nil {
		t.Fatal("expected error for GDB error response, got nil")
	}
}

// readGDBPacket reads a single GDB protocol packet ($<data>#<cksum>).
func readGDBPacket(t *testing.T, conn net.Conn) string {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	buf := make([]byte, 1)
	// Read until '$'
	for {
		if _, err := conn.Read(buf); err != nil {
			t.Fatalf("read dollar: %v", err)
		}
		if buf[0] == '$' {
			break
		}
	}

	var data []byte
	for {
		if _, err := conn.Read(buf); err != nil {
			t.Fatalf("read data: %v", err)
		}
		if buf[0] == '#' {
			break
		}
		data = append(data, buf[0])
	}

	// Read and ignore checksum
	cksumBuf := make([]byte, 2)
	if _, err := conn.Read(cksumBuf); err != nil {
		t.Fatalf("read checksum: %v", err)
	}

	return string(data)
}

func portFromAddr(addr string) int {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	return port
}
