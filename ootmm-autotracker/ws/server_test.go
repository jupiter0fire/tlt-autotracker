package ws

import (
	"encoding/base64"
	"encoding/json"
	"errors"
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

	waitForClientCount(t, server, 1)

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

func TestRawHandshakeSelectsRawModeAndPreservesBase64ChunkOrder(t *testing.T) {
	server := NewServer(":0")
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWS))
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(map[string]interface{}{
		"type":     "handshake",
		"features": []string{"raw"},
		"flags":    map[string]interface{}{},
	}); err != nil {
		t.Fatalf("write handshake: %v", err)
	}

	ack := readJSONMessage(t, conn)
	if got := ack["type"]; got != "handshAck" {
		t.Fatalf("ack type = %v, want handshAck", got)
	}
	if got := ack["mode"]; got != "raw" {
		t.Fatalf("ack mode = %v, want raw", got)
	}

	server.BroadcastRawSnapshot(&ootmm.RawFrame{
		Valid:      true,
		ActiveGame: ootmm.GameMm,
		SaveIndex:  2,
		Chunks: []ootmm.RawChunk{
			{Name: "combo_ctx_oot", Address: 0x80006584, Length: 4, Data: []byte{0x00, 0x01, 0x02, 0x03}},
			{Name: "oot_save_ctx", Address: 0x8011A5D0, Length: 3, Data: []byte{0xFA, 0x00, 0xBC}},
		},
	})

	msg := readJSONMessage(t, conn)
	if got := msg["type"]; got != "raw" {
		t.Fatalf("message type = %v, want raw", got)
	}
	if got := msg["schemaVersion"]; got != rawSchemaVersion {
		t.Fatalf("schemaVersion = %v, want %s", got, rawSchemaVersion)
	}
	if got := msg["diff"]; got != false {
		t.Fatalf("diff = %v, want false", got)
	}
	if got := msg["refresh"]; got != true {
		t.Fatalf("refresh = %v, want true", got)
	}
	if got := msg["sequence"]; got != float64(1) {
		t.Fatalf("sequence = %v, want 1", got)
	}
	if got := msg["game"]; got != "MM" {
		t.Fatalf("game = %v, want MM", got)
	}
	if got := msg["saveIndex"]; got != float64(2) {
		t.Fatalf("saveIndex = %v, want 2", got)
	}

	chunks, ok := msg["chunks"].([]interface{})
	if !ok || len(chunks) != 2 {
		t.Fatalf("chunks = %T %#v, want 2 entries", msg["chunks"], msg["chunks"])
	}

	first, ok := chunks[0].(map[string]interface{})
	if !ok {
		t.Fatalf("first chunk = %T, want object", chunks[0])
	}
	if got := first["name"]; got != "combo_ctx_oot" {
		t.Fatalf("first chunk name = %v, want combo_ctx_oot", got)
	}
	if got := first["address"]; got != float64(0x80006584) {
		t.Fatalf("first chunk address = %v, want %#x", got, 0x80006584)
	}
	if got := first["length"]; got != float64(4) {
		t.Fatalf("first chunk length = %v, want 4", got)
	}
	firstData, ok := first["data"].(string)
	if !ok {
		t.Fatalf("first chunk data = %T, want string", first["data"])
	}
	decodedFirst, err := base64.StdEncoding.DecodeString(firstData)
	if err != nil {
		t.Fatalf("decode first chunk data: %v", err)
	}
	if string(decodedFirst) != string([]byte{0x00, 0x01, 0x02, 0x03}) {
		t.Fatalf("decoded first chunk = % x, want 00 01 02 03", decodedFirst)
	}

	second, ok := chunks[1].(map[string]interface{})
	if !ok {
		t.Fatalf("second chunk = %T, want object", chunks[1])
	}
	if got := second["name"]; got != "oot_save_ctx" {
		t.Fatalf("second chunk name = %v, want oot_save_ctx", got)
	}
	secondData, ok := second["data"].(string)
	if !ok {
		t.Fatalf("second chunk data = %T, want string", second["data"])
	}
	decodedSecond, err := base64.StdEncoding.DecodeString(secondData)
	if err != nil {
		t.Fatalf("decode second chunk data: %v", err)
	}
	if string(decodedSecond) != string([]byte{0xFA, 0x00, 0xBC}) {
		t.Fatalf("decoded second chunk = % x, want fa 00 bc", decodedSecond)
	}
}

func TestRawClientReceivesOnlyRequestedMemoryAreasForActiveGame(t *testing.T) {
	server := NewServer(":0")
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWS))
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(map[string]interface{}{
		"type":     "handshake",
		"features": []string{"raw"},
		"flags":    map[string]interface{}{},
		"memoryAreas": map[string]interface{}{
			"oot": []string{"oot_save_ctx", "oot_payload", "oot_playstate_core", "oot_playstate_tail"},
			"mm":  []string{"mm_save_ctx", "mm_playstate_core", "mm_playstate_tail"},
		},
	}); err != nil {
		t.Fatalf("write handshake: %v", err)
	}

	ack := readJSONMessage(t, conn)
	if got := ack["type"]; got != "handshAck" {
		t.Fatalf("ack type = %v, want handshAck", got)
	}

	server.BroadcastRawSnapshot(&ootmm.RawFrame{
		Valid:      true,
		ActiveGame: ootmm.GameOot,
		SaveIndex:  2,
		Chunks: []ootmm.RawChunk{
			{Name: "combo_ctx_oot", Address: 0x80006584, Length: 4, Data: []byte{0x00, 0x01, 0x02, 0x03}},
			{Name: "oot_save_ctx", Address: 0x8011A5D0, Length: 3, Data: []byte{0xFA, 0x00, 0xBC}},
			{Name: "oot_payload", Address: 0x80400000, Length: 2, Data: []byte{0xAA, 0x55}},
			{Name: "oot_playstate_core", Address: 0x801c84a0, Length: 4, Data: []byte{0x33, 0x44, 0x55, 0x66}},
			{Name: "oot_playstate_tail", Address: 0x801c84b0, Length: 4, Data: []byte{0x77, 0x88, 0x99, 0xAA}},
			{Name: "mm_save_ctx", Address: 0x801ef670, Length: 2, Data: []byte{0x10, 0x20}},
		},
	})

	first := readJSONMessage(t, conn)
	firstChunks, ok := first["chunks"].([]interface{})
	if !ok || len(firstChunks) != 4 {
		t.Fatalf("first chunks = %T %#v, want 4 entries", first["chunks"], first["chunks"])
	}
	if got := chunkName(t, firstChunks[0]); got != "oot_save_ctx" {
		t.Fatalf("first OoT chunk = %v, want oot_save_ctx", got)
	}
	if got := chunkName(t, firstChunks[1]); got != "oot_payload" {
		t.Fatalf("second OoT chunk = %v, want oot_payload", got)
	}
	if got := chunkName(t, firstChunks[2]); got != "oot_playstate_core" {
		t.Fatalf("third OoT chunk = %v, want oot_playstate_core", got)
	}
	if got := chunkName(t, firstChunks[3]); got != "oot_playstate_tail" {
		t.Fatalf("fourth OoT chunk = %v, want oot_playstate_tail", got)
	}

	server.BroadcastRawSnapshot(&ootmm.RawFrame{
		Valid:      true,
		ActiveGame: ootmm.GameMm,
		SaveIndex:  2,
		Chunks: []ootmm.RawChunk{
			{Name: "oot_save_ctx", Address: 0x8011A5D0, Length: 3, Data: []byte{0xFA, 0x00, 0xBC}},
			{Name: "mm_save_ctx", Address: 0x801ef670, Length: 2, Data: []byte{0x10, 0x20}},
			{Name: "mm_payload", Address: 0x80400000, Length: 2, Data: []byte{0xAA, 0x55}},
			{Name: "mm_playstate_core", Address: 0x803e6b20, Length: 4, Data: []byte{0x01, 0x02, 0x03, 0x04}},
			{Name: "mm_playstate_tail", Address: 0x803e6b30, Length: 4, Data: []byte{0x05, 0x06, 0x07, 0x08}},
		},
	})

	second := readJSONMessage(t, conn)
	secondChunks, ok := second["chunks"].([]interface{})
	if !ok || len(secondChunks) != 3 {
		t.Fatalf("second chunks = %T %#v, want 3 entries", second["chunks"], second["chunks"])
	}
	if got := chunkName(t, secondChunks[0]); got != "mm_save_ctx" {
		t.Fatalf("MM chunk = %v, want mm_save_ctx", got)
	}
	if got := chunkName(t, secondChunks[1]); got != "mm_playstate_core" {
		t.Fatalf("MM playstate core = %v, want mm_playstate_core", got)
	}
	if got := chunkName(t, secondChunks[2]); got != "mm_playstate_tail" {
		t.Fatalf("MM playstate tail = %v, want mm_playstate_tail", got)
	}
}

func TestRawClientDoesNotReceiveLegacyDelta(t *testing.T) {
	server := NewServer(":0")
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWS))
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(map[string]interface{}{
		"type":     "handshake",
		"features": []string{"raw"},
		"flags":    map[string]interface{}{},
	}); err != nil {
		t.Fatalf("write handshake: %v", err)
	}
	ack := readJSONMessage(t, conn)
	if got := ack["type"]; got != "handshAck" {
		t.Fatalf("ack type = %v, want handshAck", got)
	}

	server.BroadcastDelta(
		ootmm.GameOot,
		0x10,
		true,
		[]tracker.ItemDiff{{ID: "OOT_SWORD", Qty: 1}},
		nil,
	)

	expectNoMessage(t, conn)
}

func TestLegacyClientDoesNotReceiveRawSnapshots(t *testing.T) {
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

	server.BroadcastRawSnapshot(&ootmm.RawFrame{
		Valid:      true,
		ActiveGame: ootmm.GameOot,
		SaveIndex:  1,
		Chunks: []ootmm.RawChunk{{
			Name:    "combo_ctx_oot",
			Address: 0x80006584,
			Length:  2,
			Data:    []byte{0xAA, 0x55},
		}},
	})

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

func chunkName(t *testing.T, value interface{}) string {
	t.Helper()

	chunk, ok := value.(map[string]interface{})
	if !ok {
		t.Fatalf("chunk = %T, want object", value)
	}

	name, ok := chunk["name"].(string)
	if !ok {
		t.Fatalf("chunk name = %T, want string", chunk["name"])
	}

	return name
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
