package ws

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"ootmm-autotracker/buildinfo"
	"ootmm-autotracker/ootmm"
)

const (
	allowedDevOrigin  = "http://localhost:5173"
	allowedProdOrigin = "https://www.thelasttracker.org"
)

func TestRawHandshakeSelectsRawModeAndPreservesBase64ChunkOrder(t *testing.T) {
	server := newTestServer(t)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWS))
	defer httpServer.Close()

	conn, _, err := dialTestWebSocket(httpServer.URL, allowedDevOrigin)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(map[string]interface{}{
		"type":     "handshake",
		"features": []string{"raw"},
		"flags":    map[string]interface{}{},
		"memoryAreas": map[string]interface{}{
			"oot": []map[string]interface{}{{
				"name":    "oot_save_ctx",
				"address": 0x8011A5D0,
				"length":  3,
			}},
			"mm": []map[string]interface{}{{
				"name":    "combo_ctx_oot",
				"address": 0x80006584,
				"length":  4,
			}, {
				"name":    "oot_save_ctx",
				"address": 0x8011A5D0,
				"length":  3,
			}},
		},
	}); err != nil {
		t.Fatalf("write handshake: %v", err)
	}

	ack := readJSONMessage(t, conn)
	if got := ack["type"]; got != "handshAck" {
		t.Fatalf("ack type = %v, want handshAck", got)
	}
	if got := ack["version"]; got != buildinfo.Version {
		t.Fatalf("ack version = %v, want %s", got, buildinfo.Version)
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
	server := newTestServer(t)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWS))
	defer httpServer.Close()

	conn, _, err := dialTestWebSocket(httpServer.URL, allowedDevOrigin)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(map[string]interface{}{
		"type":     "handshake",
		"features": []string{"raw"},
		"flags":    map[string]interface{}{},
		"memoryAreas": map[string]interface{}{
			"oot": []map[string]interface{}{
				{"name": "oot_save_ctx", "address": 0x8011A5D0, "length": 3},
				{"name": "oot_foreign_mm_save", "address": 0x80443970, "length": 2},
				{"name": "oot_shared_custom_save", "address": 0x80443100, "length": 2},
				{"name": "oot_playstate_core", "address": 0x801c84a0, "length": 4},
				{"name": "oot_playstate_tail", "address": 0x801c84b0, "length": 4},
			},
			"mm": []map[string]interface{}{
				{"name": "mm_save_ctx", "address": 0x801ef670, "length": 2},
				{"name": "mm_foreign_oot_save", "address": 0x807729f0, "length": 2},
				{"name": "mm_shared_custom_save", "address": 0x80772180, "length": 2},
				{"name": "mm_playstate_core", "address": 0x803e6b20, "length": 4},
				{"name": "mm_playstate_tail", "address": 0x803e6b30, "length": 4},
			},
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
			{Name: "oot_foreign_mm_save", Address: 0x80443970, Length: 2, Data: []byte{0xAA, 0x55}},
			{Name: "oot_shared_custom_save", Address: 0x80443100, Length: 2, Data: []byte{0x11, 0x22}},
			{Name: "oot_playstate_core", Address: 0x801c84a0, Length: 4, Data: []byte{0x33, 0x44, 0x55, 0x66}},
			{Name: "oot_playstate_tail", Address: 0x801c84b0, Length: 4, Data: []byte{0x77, 0x88, 0x99, 0xAA}},
			{Name: "mm_save_ctx", Address: 0x801ef670, Length: 2, Data: []byte{0x10, 0x20}},
		},
	})

	first := readJSONMessage(t, conn)
	firstChunks, ok := first["chunks"].([]interface{})
	if !ok || len(firstChunks) != 5 {
		t.Fatalf("first chunks = %T %#v, want 5 entries", first["chunks"], first["chunks"])
	}
	if got := chunkName(t, firstChunks[0]); got != "oot_save_ctx" {
		t.Fatalf("first OoT chunk = %v, want oot_save_ctx", got)
	}
	if got := chunkName(t, firstChunks[1]); got != "oot_foreign_mm_save" {
		t.Fatalf("second OoT chunk = %v, want oot_foreign_mm_save", got)
	}
	if got := chunkName(t, firstChunks[2]); got != "oot_shared_custom_save" {
		t.Fatalf("third OoT chunk = %v, want oot_shared_custom_save", got)
	}
	if got := chunkName(t, firstChunks[3]); got != "oot_playstate_core" {
		t.Fatalf("fourth OoT chunk = %v, want oot_playstate_core", got)
	}
	if got := chunkName(t, firstChunks[4]); got != "oot_playstate_tail" {
		t.Fatalf("fifth OoT chunk = %v, want oot_playstate_tail", got)
	}

	server.BroadcastRawSnapshot(&ootmm.RawFrame{
		Valid:      true,
		ActiveGame: ootmm.GameMm,
		SaveIndex:  2,
		Chunks: []ootmm.RawChunk{
			{Name: "oot_save_ctx", Address: 0x8011A5D0, Length: 3, Data: []byte{0xFA, 0x00, 0xBC}},
			{Name: "mm_save_ctx", Address: 0x801ef670, Length: 2, Data: []byte{0x10, 0x20}},
			{Name: "mm_foreign_oot_save", Address: 0x807729f0, Length: 2, Data: []byte{0xAA, 0x55}},
			{Name: "mm_shared_custom_save", Address: 0x80772180, Length: 2, Data: []byte{0x11, 0x22}},
			{Name: "mm_playstate_core", Address: 0x803e6b20, Length: 4, Data: []byte{0x01, 0x02, 0x03, 0x04}},
			{Name: "mm_playstate_tail", Address: 0x803e6b30, Length: 4, Data: []byte{0x05, 0x06, 0x07, 0x08}},
		},
	})

	second := readJSONMessage(t, conn)
	secondChunks, ok := second["chunks"].([]interface{})
	if !ok || len(secondChunks) != 5 {
		t.Fatalf("second chunks = %T %#v, want 5 entries", second["chunks"], second["chunks"])
	}
	if got := chunkName(t, secondChunks[0]); got != "mm_save_ctx" {
		t.Fatalf("MM chunk = %v, want mm_save_ctx", got)
	}
	if got := chunkName(t, secondChunks[1]); got != "mm_foreign_oot_save" {
		t.Fatalf("MM foreign OoT save = %v, want mm_foreign_oot_save", got)
	}
	if got := chunkName(t, secondChunks[2]); got != "mm_shared_custom_save" {
		t.Fatalf("MM shared custom save = %v, want mm_shared_custom_save", got)
	}
	if got := chunkName(t, secondChunks[3]); got != "mm_playstate_core" {
		t.Fatalf("MM playstate core = %v, want mm_playstate_core", got)
	}
	if got := chunkName(t, secondChunks[4]); got != "mm_playstate_tail" {
		t.Fatalf("MM playstate tail = %v, want mm_playstate_tail", got)
	}
}

func TestRawClientDoesNotReceiveUpdateWhenOnlyUnwatchedChunkChanges(t *testing.T) {
	server := newTestServer(t)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWS))
	defer httpServer.Close()

	conn, _, err := dialTestWebSocket(httpServer.URL, allowedDevOrigin)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(map[string]interface{}{
		"type":     "handshake",
		"features": []string{"raw"},
		"flags":    map[string]interface{}{},
		"memoryAreas": map[string]interface{}{
			"oot": []map[string]interface{}{{
				"name":    "oot_save_ctx",
				"address": 0x8011A5D0,
				"length":  3,
			}},
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
			{Name: "oot_save_ctx", Address: 0x8011A5D0, Length: 3, Data: []byte{0xFA, 0x00, 0xBC}},
			{Name: "oot_foreign_mm_save", Address: 0x80443970, Length: 2, Data: []byte{0xAA, 0x55}},
		},
	})

	readJSONMessage(t, conn)

	server.BroadcastRawSnapshot(&ootmm.RawFrame{
		Valid:      true,
		ActiveGame: ootmm.GameOot,
		SaveIndex:  2,
		Chunks: []ootmm.RawChunk{
			{Name: "oot_save_ctx", Address: 0x8011A5D0, Length: 3, Data: []byte{0xFA, 0x00, 0xBC}},
			{Name: "oot_foreign_mm_save", Address: 0x80443970, Length: 2, Data: []byte{0xDE, 0xAD}},
		},
	})

	expectNoJSONMessage(t, conn)
}

func TestRawClientReceivesUpdateWhenWatchedChunkChanges(t *testing.T) {
	server := newTestServer(t)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWS))
	defer httpServer.Close()

	conn, _, err := dialTestWebSocket(httpServer.URL, allowedDevOrigin)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(map[string]interface{}{
		"type":     "handshake",
		"features": []string{"raw"},
		"flags":    map[string]interface{}{},
		"memoryAreas": map[string]interface{}{
			"oot": []map[string]interface{}{{
				"name":    "oot_save_ctx",
				"address": 0x8011A5D0,
				"length":  3,
			}},
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
		Chunks: []ootmm.RawChunk{{
			Name:    "oot_save_ctx",
			Address: 0x8011A5D0,
			Length:  3,
			Data:    []byte{0xFA, 0x00, 0xBC},
		}},
	})

	readJSONMessage(t, conn)

	server.BroadcastRawSnapshot(&ootmm.RawFrame{
		Valid:      true,
		ActiveGame: ootmm.GameOot,
		SaveIndex:  2,
		Chunks: []ootmm.RawChunk{{
			Name:    "oot_save_ctx",
			Address: 0x8011A5D0,
			Length:  3,
			Data:    []byte{0xFA, 0x01, 0xBC},
		}},
	})

	msg := readJSONMessage(t, conn)
	chunks, ok := msg["chunks"].([]interface{})
	if !ok || len(chunks) != 1 {
		t.Fatalf("chunks = %T %#v, want 1 entry", msg["chunks"], msg["chunks"])
	}

	chunk, ok := chunks[0].(map[string]interface{})
	if !ok {
		t.Fatalf("chunk = %T, want object", chunks[0])
	}
	data, ok := chunk["data"].(string)
	if !ok {
		t.Fatalf("chunk data = %T, want string", chunk["data"])
	}
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		t.Fatalf("decode chunk data: %v", err)
	}
	if string(decoded) != string([]byte{0xFA, 0x01, 0xBC}) {
		t.Fatalf("decoded chunk = % x, want fa 01 bc", decoded)
	}
}

func TestLegacyHandshakeIsRejected(t *testing.T) {
	server := newTestServer(t)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWS))
	defer httpServer.Close()

	conn, _, err := dialTestWebSocket(httpServer.URL, allowedDevOrigin)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(map[string]interface{}{
		"type":     "handshake",
		"features": []string{"items", "checks"},
		"flags": map[string]interface{}{
			"protocol": "legacy",
		},
	}); err != nil {
		t.Fatalf("write legacy handshake: %v", err)
	}

	msg := readJSONMessage(t, conn)
	if got := msg["type"]; got != "error" {
		t.Fatalf("message type = %v, want error", got)
	}
	if got := msg["message"]; got != "legacy protocol removed" {
		t.Fatalf("message = %v, want legacy protocol removed", got)
	}
}

func TestWebSocketUpgradeAllowsConfiguredProductionOrigin(t *testing.T) {
	server := newTestServer(t)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWS))
	defer httpServer.Close()

	conn, _, err := dialTestWebSocket(httpServer.URL, allowedProdOrigin)
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
}

func TestWebSocketUpgradeRejectsDisallowedOrigin(t *testing.T) {
	server := newTestServer(t)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWS))
	defer httpServer.Close()

	_, response, err := dialTestWebSocket(httpServer.URL, "http://evil.example")
	if err == nil {
		t.Fatal("dial websocket unexpectedly succeeded")
	}
	if response == nil {
		t.Fatalf("dial websocket returned nil response for error: %v", err)
	}
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("status code = %d, want %d", response.StatusCode, http.StatusForbidden)
	}
}

func TestWebSocketUpgradeLogsRepeatedDisallowedOriginOnce(t *testing.T) {
	server := newTestServer(t)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWS))
	defer httpServer.Close()

	var logOutput bytes.Buffer
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	log.SetOutput(&logOutput)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
	})

	for attempt := 0; attempt < 2; attempt++ {
		_, response, err := dialTestWebSocket(httpServer.URL, "http://evil.example")
		if err == nil {
			t.Fatal("dial websocket unexpectedly succeeded")
		}
		if response == nil {
			t.Fatalf("dial websocket returned nil response for error: %v", err)
		}
		if response.StatusCode != http.StatusForbidden {
			t.Fatalf("status code = %d, want %d", response.StatusCode, http.StatusForbidden)
		}
	}

	output := logOutput.String()
	if count := strings.Count(output, "Rejected WebSocket upgrade"); count != 1 {
		t.Fatalf("rejection log count = %d, want 1; output=%q", count, output)
	}
	if strings.Contains(output, "WebSocket upgrade error:") {
		t.Fatalf("unexpected duplicate upgrade error log: %q", output)
	}
}

func TestWebSocketUpgradeRejectsMissingOrigin(t *testing.T) {
	server := newTestServer(t)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWS))
	defer httpServer.Close()

	_, response, err := dialTestWebSocket(httpServer.URL, "")
	if err == nil {
		t.Fatal("dial websocket unexpectedly succeeded")
	}
	if response == nil {
		t.Fatalf("dial websocket returned nil response for error: %v", err)
	}
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("status code = %d, want %d", response.StatusCode, http.StatusForbidden)
	}
}

func TestWebSocketUpgradeRejectsNullOrigin(t *testing.T) {
	server := newTestServer(t)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWS))
	defer httpServer.Close()

	_, response, err := dialTestWebSocket(httpServer.URL, "null")
	if err == nil {
		t.Fatal("dial websocket unexpectedly succeeded")
	}
	if response == nil {
		t.Fatalf("dial websocket returned nil response for error: %v", err)
	}
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("status code = %d, want %d", response.StatusCode, http.StatusForbidden)
	}
}

func newTestServer(t *testing.T) *Server {
	t.Helper()

	server, err := NewServer(":0", []string{allowedDevOrigin, allowedProdOrigin})
	if err != nil {
		t.Fatalf("NewServer returned error: %v", err)
	}
	return server
}

func dialTestWebSocket(httpURL string, origin string) (*websocket.Conn, *http.Response, error) {
	wsURL := "ws" + strings.TrimPrefix(httpURL, "http")
	headers := make(http.Header)
	if origin != "" {
		headers.Set("Origin", origin)
	}
	return websocket.DefaultDialer.Dial(wsURL, headers)
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

func expectNoJSONMessage(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	_, _, err := conn.ReadMessage()
	if err == nil {
		t.Fatal("read websocket message: got unexpected message")
	}

	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Fatalf("read websocket message: %v", err)
	}
}
