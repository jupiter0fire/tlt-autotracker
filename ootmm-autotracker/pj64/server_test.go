package pj64

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

// startTestServer starts a PJ64 server on an ephemeral port.
func startTestServer(t *testing.T) *Server {
	t.Helper()
	srv := NewServer(0)
	if err := srv.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(srv.Stop)
	return srv
}

// dialTestAdapter connects a raw TCP client simulating the Lua adapter.
func dialTestAdapter(t *testing.T, s *Server) net.Conn {
	t.Helper()
	conn, err := net.Dial("tcp", s.listener.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

// TestConnectAcceptsMatchingGreeting verifies that a client sending the
// current greeting is accepted and can serve memory reads.
func TestConnectAcceptsMatchingGreeting(t *testing.T) {
	srv := startTestServer(t)
	adapter := dialTestAdapter(t, srv)

	if _, err := adapter.Write([]byte(adapterGreeting)); err != nil {
		t.Fatalf("write greeting: %v", err)
	}

	if err := srv.Connect(); err != nil {
		t.Fatalf("Connect with matching greeting: %v", err)
	}
	defer srv.Close()

	if !srv.IsConnected() {
		t.Fatal("expected server to report connected after matching greeting")
	}

	// Round-trip a bulk read: server sends op(1)+addr(4)+size(4)
	// little-endian, adapter replies with size payload bytes.
	errCh := make(chan error, 1)
	go func() {
		cmd := make([]byte, 9)
		if _, err := io.ReadFull(adapter, cmd); err != nil {
			errCh <- fmt.Errorf("read command: %w", err)
			return
		}
		if cmd[0] != opReadBulk {
			errCh <- fmt.Errorf("op = %d, want %d", cmd[0], opReadBulk)
			return
		}
		addr := binary.LittleEndian.Uint32(cmd[1:5])
		size := binary.LittleEndian.Uint32(cmd[5:9])
		payload := make([]byte, size)
		for i := range payload {
			payload[i] = byte(addr + uint32(i))
		}
		if _, err := adapter.Write(payload); err != nil {
			errCh <- fmt.Errorf("write payload: %w", err)
		}
		errCh <- nil
	}()

	data, err := srv.ReadMemory(0x80000000, 16)
	if err != nil {
		t.Fatalf("ReadMemory: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}

	want := make([]byte, 16)
	for i := range want {
		want[i] = byte(0x80000000 + uint32(i))
	}
	for i := range want {
		if data[i] != want[i] {
			t.Fatalf("data[%d] = 0x%02x, want 0x%02x", i, data[i], want[i])
		}
	}
}

// TestConnectRejectsWrongGreeting verifies that a client sending an
// outdated greeting is rejected and its connection closed.
func TestConnectRejectsWrongGreeting(t *testing.T) {
	srv := startTestServer(t)
	adapter := dialTestAdapter(t, srv)

	if _, err := adapter.Write([]byte("OAT1")); err != nil {
		t.Fatalf("write greeting: %v", err)
	}

	if err := srv.Connect(); err == nil {
		t.Fatal("expected Connect to fail for wrong greeting")
	}
	if srv.IsConnected() {
		t.Fatal("expected server not to be connected after rejected greeting")
	}

	// The server must have closed the adapter's connection.
	adapter.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := adapter.Read(make([]byte, 1)); err == nil {
		t.Fatal("expected adapter connection to be closed after rejected greeting")
	}
}

// TestConnectRejectsMissingGreeting verifies that a client sending no
// greeting at all (an old Lua script) is rejected.
func TestConnectRejectsMissingGreeting(t *testing.T) {
	orig := greetingTimeout
	greetingTimeout = 50 * time.Millisecond
	defer func() { greetingTimeout = orig }()

	srv := startTestServer(t)
	adapter := dialTestAdapter(t, srv) // never sends a greeting

	start := time.Now()
	if err := srv.Connect(); err == nil {
		t.Fatal("expected Connect to fail when no greeting is sent")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("rejection took too long: %v", elapsed)
	}
	if srv.IsConnected() {
		t.Fatal("expected server not to be connected after missing greeting")
	}

	adapter.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := adapter.Read(make([]byte, 1)); err == nil {
		t.Fatal("expected adapter connection to be closed after missing greeting")
	}
}
