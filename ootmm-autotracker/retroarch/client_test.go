package retroarch

import (
	"errors"
	"net"
	"testing"
)

func TestConnectFailsWithoutServer(t *testing.T) {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(DefaultHost), Port: 0})
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	port := listener.LocalAddr().(*net.UDPAddr).Port
	listener.Close()

	client := NewClient(DefaultHost, port)
	if err := client.Connect(); err == nil {
		client.Close()
		t.Fatal("expected connect to fail without a RetroArch server")
	}
	if client.conn != nil {
		client.Close()
		t.Fatal("client retained a connection after failed probe")
	}
}

func TestConnectSucceedsWithProbeResponse(t *testing.T) {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(DefaultHost), Port: 0})
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer listener.Close()

	serverErr := make(chan error, 1)
	go func() {
		buf := make([]byte, MaxProbeSize)
		for {
			n, addr, err := listener.ReadFromUDP(buf)
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				serverErr <- err
				return
			}

			if string(buf[:n]) != "VERSION\n" {
				serverErr <- errors.New("unexpected probe command")
				return
			}

			if _, err := listener.WriteToUDP([]byte("1.21.0\n"), addr); err != nil {
				serverErr <- err
				return
			}
		}
	}()

	client := NewClient(DefaultHost, listener.LocalAddr().(*net.UDPAddr).Port)
	defer client.Close()

	if err := client.Connect(); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if !client.IsConnected() {
		t.Fatal("expected IsConnected to succeed after a valid probe response")
	}

	select {
	case err := <-serverErr:
		t.Fatalf("udp server: %v", err)
	default:
	}
}
