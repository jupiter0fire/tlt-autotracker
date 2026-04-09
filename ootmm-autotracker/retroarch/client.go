package retroarch

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	DefaultHost  = "127.0.0.1"
	DefaultPort  = 55355
	SendTimeout  = 1 * time.Second
	ProbeTimeout = 500 * time.Millisecond
	RecvTimeout  = 10 * time.Second
	MaxReadSize  = 2048
	MaxProbeSize = 256
)

type Client struct {
	host string
	port int
	conn net.Conn
}

func NewClient(host string, port int) *Client {
	return &Client{
		host: host,
		port: port,
	}
}

func (c *Client) Connect() error {
	addr := net.JoinHostPort(c.host, fmt.Sprintf("%d", c.port))
	conn, err := net.DialTimeout("udp", addr, SendTimeout)
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}
	if err := c.pingConn(conn); err != nil {
		conn.Close()
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

func (c *Client) IsConnected() bool {
	if c.conn == nil {
		return false
	}
	return c.pingConn(c.conn) == nil
}

func (c *Client) pingConn(conn net.Conn) error {
	conn.SetWriteDeadline(time.Now().Add(SendTimeout))
	if _, err := conn.Write([]byte("VERSION\n")); err != nil {
		return fmt.Errorf("probe write: %w", err)
	}

	buf := make([]byte, MaxProbeSize)
	conn.SetReadDeadline(time.Now().Add(ProbeTimeout))
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("probe read: %w", err)
	}
	if strings.TrimSpace(string(buf[:n])) == "" {
		return fmt.Errorf("probe read: empty response")
	}
	return nil
}

// ReadMemory reads `size` bytes from the emulated core's memory at `addr`.
// The address is in the core's native address space (physical for Mupen64Plus-Next).
func (c *Client) ReadMemory(addr uint32, size int) ([]byte, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}
	if size > MaxReadSize {
		return nil, fmt.Errorf("read size %d exceeds max %d", size, MaxReadSize)
	}

	cmd := fmt.Sprintf("READ_CORE_MEMORY %x %d\n", addr, size)

	c.conn.SetWriteDeadline(time.Now().Add(SendTimeout))
	_, err := c.conn.Write([]byte(cmd))
	if err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	// Response: "READ_CORE_MEMORY <addr> <hexdata>\n"
	// or "READ_CORE_MEMORY -1\n" on error
	buf := make([]byte, len(cmd)+size*3+64)
	c.conn.SetReadDeadline(time.Now().Add(RecvTimeout))
	n, err := c.conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	resp := strings.TrimSpace(string(buf[:n]))
	parts := strings.SplitN(resp, " ", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected response: %q", resp)
	}

	if parts[1] == "-1" {
		return nil, fmt.Errorf("core returned error for address 0x%x", addr)
	}

	// RetroArch returns hex bytes space-separated (e.g. "44 80 02 3C").
	hexStr := strings.ReplaceAll(parts[2], " ", "")
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("decode hex: %w", err)
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
