// Package ares implements a GDB remote protocol client for the Ares emulator.
//
// Ares exposes a GDB remote serial protocol stub over TCP (default port 9123).
// Memory is read using the 'm' packet and written using the 'M' packet.
// The address space uses N64 virtual addressing (0x80xxxxxx), matching the
// CPU's view of memory — same as Project64.
package ares

import (
	"encoding/hex"
	"fmt"
	"net"
	"time"
)

const (
	DefaultHost = "127.0.0.1"
	DefaultPort = 9123

	ConnectTimeout = 2 * time.Second
	SendTimeout    = 5 * time.Second
	RecvTimeout    = 10 * time.Second
	MaxReadSize    = 2048
)

// Client communicates with Ares over its GDB remote protocol stub.
type Client struct {
	host string
	port int
	conn net.Conn
}

// NewClient creates a new Ares GDB client.
func NewClient(host string, port int) *Client {
	return &Client{
		host: host,
		port: port,
	}
}

// Connect dials the Ares GDB stub and performs the initial handshake.
func (c *Client) Connect() error {
	addr := net.JoinHostPort(c.host, fmt.Sprintf("%d", c.port))
	conn, err := net.DialTimeout("tcp", addr, ConnectTimeout)
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}

	// GDB remote protocol: send '+' to acknowledge/start.
	// Ares requires this initial ack before accepting commands.
	conn.SetWriteDeadline(time.Now().Add(SendTimeout))
	if _, err := conn.Write([]byte("+")); err != nil {
		conn.Close()
		return fmt.Errorf("initial handshake write: %w", err)
	}

	c.conn = conn
	return nil
}

// Close closes the connection to Ares.
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

// IsConnected returns true if the client is currently connected.
func (c *Client) IsConnected() bool {
	if c.conn == nil {
		return false
	}
	return true
}

// ReadMemory reads `size` bytes from the emulated N64 memory at `addr`
// using the GDB remote 'm' packet.
func (c *Client) ReadMemory(addr uint32, size int) ([]byte, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}
	if size > MaxReadSize {
		return nil, fmt.Errorf("read size %d exceeds max %d", size, MaxReadSize)
	}

	// Build GDB 'm' packet: $m<addr>,<len>#<cksum>
	cmd := buildPacket(fmt.Sprintf("m%08x,%x", addr, size))

	c.conn.SetWriteDeadline(time.Now().Add(SendTimeout))
	if _, err := c.conn.Write([]byte(cmd)); err != nil {
		c.Close()
		return nil, fmt.Errorf("write: %w", err)
	}

	// Read response: $<hex>#<cksum>
	resp, err := c.readPacket()
	if err != nil {
		return nil, err
	}

	// Check for GDB error response (starts with 'E')
	if len(resp) > 0 && resp[0] == 'E' {
		return nil, fmt.Errorf("gdb error: %s", resp)
	}

	data, err := hex.DecodeString(resp)
	if err != nil {
		return nil, fmt.Errorf("decode hex response: %w", err)
	}

	if len(data) != size {
		return nil, fmt.Errorf("expected %d bytes, got %d", size, len(data))
	}

	return data, nil
}

// ReadMemoryLarge reads memory in chunks when size exceeds MaxReadSize.
func (c *Client) ReadMemoryLarge(addr uint32, size int) ([]byte, error) {
	if size <= MaxReadSize {
		return c.ReadMemory(addr, size)
	}

	result := make([]byte, 0, size)
	for offset := 0; offset < size; offset += MaxReadSize {
		chunkSize := MaxReadSize
		if offset+chunkSize > size {
			chunkSize = size - offset
		}
		chunk, err := c.ReadMemory(addr+uint32(offset), chunkSize)
		if err != nil {
			return nil, fmt.Errorf("chunk at offset 0x%x: %w", offset, err)
		}
		result = append(result, chunk...)
	}
	return result, nil
}

// buildPacket creates a GDB remote protocol packet:
// $<data>#<checksum>
// where checksum is the modulo-256 sum of all bytes in data.
func buildPacket(data string) string {
	cksum := byte(0)
	for i := 0; i < len(data); i++ {
		cksum += data[i]
	}
	return fmt.Sprintf("$%s#%02x", data, cksum)
}

// readPacket reads a single GDB remote protocol response packet.
// The format is: $<data>#<cksum>
// After receiving a valid packet we send '+' to acknowledge.
func (c *Client) readPacket() (string, error) {
	// Read until we find '$' and then until '#'
	buf := make([]byte, 1)
	// Find start of packet
	for {
		c.conn.SetReadDeadline(time.Now().Add(RecvTimeout))
		if _, err := c.conn.Read(buf); err != nil {
			c.Close()
			return "", fmt.Errorf("read start marker: %w", err)
		}
		if buf[0] == '$' {
			break
		}
		// Ignore any non-packet data (e.g. '+' acks, notifications)
	}

	// Read data until '#'
	var data []byte
	for {
		c.conn.SetReadDeadline(time.Now().Add(RecvTimeout))
		if _, err := c.conn.Read(buf); err != nil {
			c.Close()
			return "", fmt.Errorf("read data: %w", err)
		}
		if buf[0] == '#' {
			break
		}
		data = append(data, buf[0])
	}

	// Read 2-byte checksum
	cksumBuf := make([]byte, 2)
	c.conn.SetReadDeadline(time.Now().Add(RecvTimeout))
	if _, err := c.conn.Read(cksumBuf); err != nil {
		c.Close()
		return "", fmt.Errorf("read checksum: %w", err)
	}

	// Verify checksum
	expected := checksum(data)
	got := string(cksumBuf)
	if got != fmt.Sprintf("%02x", expected) {
		c.Close()
		return "", fmt.Errorf("checksum mismatch: expected %02x, got %s", expected, got)
	}

	// Send acknowledgment
	c.conn.SetWriteDeadline(time.Now().Add(SendTimeout))
	if _, err := c.conn.Write([]byte("+")); err != nil {
		// Don't close on ack failure — the read succeeded
		return string(data), fmt.Errorf("ack write: %w", err)
	}

	return string(data), nil
}

func checksum(data []byte) byte {
	var cksum byte
	for _, b := range data {
		cksum += b
	}
	return cksum
}
