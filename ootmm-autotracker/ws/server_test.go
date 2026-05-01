package ws

import (
	"errors"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"ootmm-autotracker/ootmm"
	"ootmm-autotracker/tracker"
)

func TestBroadcastDeltaSendsLocationBeforeItems(t *testing.T) {
	server := NewServer(":0")
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWS))
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	server.BroadcastDelta(
		ootmm.GameOot,
		0x10,
		true,
		[]tracker.ItemDiff{{ID: "OOT_SWORD", Qty: 1}},
		nil,
	)

	first := readJSONMessage(t, conn)
	if got := first["type"]; got != "location" {
		t.Fatalf("first message type = %v, want location", got)
	}
	if got := first["sceneId"]; got != float64(0x10) {
		t.Fatalf("first sceneId = %v, want %d", got, 0x10)
	}

	second := readJSONMessage(t, conn)
	if got := second["type"]; got != "item" {
		t.Fatalf("second message type = %v, want item", got)
	}
	if got := second["diff"]; got != true {
		t.Fatalf("second diff = %v, want true", got)
	}

	items, ok := second["items"].([]interface{})
	if !ok || len(items) != 1 {
		t.Fatalf("second items = %T %#v, want single entry", second["items"], second["items"])
	}

	item, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatalf("item entry = %T, want object", items[0])
	}
	if got := item["id"]; got != "OOT_SWORD" {
		t.Fatalf("item id = %v, want OOT_SWORD", got)
	}
	if got := item["qty"]; got != float64(1) {
		t.Fatalf("item qty = %v, want 1", got)
	}
}

func TestPendingFullSyncClientGetsInitialInventoryInsteadOfDelta(t *testing.T) {
	server := NewServer(":0")
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWS))
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	waitForClientCount(t, server, 1)
	markOnlyClientForFullSync(t, server)

	server.BroadcastDelta(
		ootmm.GameOot,
		0x10,
		true,
		[]tracker.ItemDiff{{ID: "OOT_BOW", Qty: 1}},
		[]tracker.CheckDiff{{Name: "Test Check", Checked: true}},
	)
	server.FlushInitialItems([]tracker.ItemDiff{{ID: "OOT_SWORD", Qty: 3}})

	first := readJSONMessage(t, conn)
	if got := first["type"]; got != "item" {
		t.Fatalf("first message type = %v, want item", got)
	}
	if got := first["diff"]; got != false {
		t.Fatalf("first diff = %v, want false", got)
	}
	if got := first["refresh"]; got != false {
		t.Fatalf("first refresh = %v, want false", got)
	}

	items, ok := first["items"].([]interface{})
	if !ok || len(items) != 1 {
		t.Fatalf("first items = %T %#v, want single entry", first["items"], first["items"])
	}

	item, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatalf("item entry = %T, want object", items[0])
	}
	if got := item["id"]; got != "OOT_SWORD" {
		t.Fatalf("item id = %v, want OOT_SWORD", got)
	}
	if got := item["qty"]; got != float64(3) {
		t.Fatalf("item qty = %v, want 3", got)
	}

	second := readJSONMessage(t, conn)
	if got := second["type"]; got != "refresh" {
		t.Fatalf("second message type = %v, want refresh", got)
	}
	if got := second["refresh"]; got != true {
		t.Fatalf("second refresh = %v, want true", got)
	}

	expectNoMessage(t, conn)
}

func readJSONMessage(t *testing.T, conn *websocket.Conn) map[string]interface{} {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read websocket message: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal websocket message: %v", err)
	}

	return payload
}

func expectNoMessage(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	_, _, err := conn.ReadMessage()
	if err == nil {
		t.Fatal("unexpected websocket message")
	}

	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Fatalf("read websocket message: %v", err)
	}

	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		t.Fatalf("reset read deadline: %v", err)
	}
}

func waitForClientCount(t *testing.T, server *Server, want int) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if server.ClientCount() == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("client count = %d, want %d", server.ClientCount(), want)
}

func markOnlyClientForFullSync(t *testing.T, server *Server) {
	t.Helper()

	server.mu.Lock()
	defer server.mu.Unlock()

	if len(server.clients) != 1 {
		t.Fatalf("client count = %d, want 1", len(server.clients))
	}

	for _, client := range server.clients {
		client.wantsFull = true
		return
	}
}
