package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"ootmm-autotracker/ootmm"
	"ootmm-autotracker/tracker"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type ProtocolMode string

const (
	ProtocolModeLegacy ProtocolMode = "legacy"
	ProtocolModeRaw    ProtocolMode = "raw"
	rawSchemaVersion   string       = "1"
)

type Options struct {
	EnableLegacyInterpretation bool
	DefaultProtocolMode        ProtocolMode
}

func ParseProtocolMode(value string) (ProtocolMode, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", string(ProtocolModeLegacy):
		return ProtocolModeLegacy, nil
	case string(ProtocolModeRaw):
		return ProtocolModeRaw, nil
	default:
		return "", fmt.Errorf("unsupported protocol mode %q", value)
	}
}

// Server manages WebSocket connections to tracker frontends.
type Server struct {
	mu                         sync.Mutex
	clients                    map[*websocket.Conn]*clientState
	addr                       string
	enableLegacyInterpretation bool
	defaultProtocolMode        ProtocolMode
	rawSequence                uint64
}

type clientState struct {
	conn      *websocket.Conn
	wantsFull bool
	mode      ProtocolMode
	rawMemory rawMemoryAreaSelection
}

type rawMemoryAreasMessage struct {
	Oot []string `json:"oot"`
	Mm  []string `json:"mm"`
}

type rawMemoryAreaSelection struct {
	hasRequest bool
	oot        map[string]struct{}
	mm         map[string]struct{}
}

func NewServer(addr string, options ...Options) *Server {
	config := Options{
		EnableLegacyInterpretation: true,
		DefaultProtocolMode:        ProtocolModeLegacy,
	}
	if len(options) > 0 {
		config = options[0]
		if config.DefaultProtocolMode == "" {
			config.DefaultProtocolMode = ProtocolModeLegacy
		}
	}
	if !config.EnableLegacyInterpretation && config.DefaultProtocolMode == ProtocolModeLegacy {
		config.DefaultProtocolMode = ProtocolModeRaw
	}

	return &Server{
		clients:                    make(map[*websocket.Conn]*clientState),
		addr:                       addr,
		enableLegacyInterpretation: config.EnableLegacyInterpretation,
		defaultProtocolMode:        config.DefaultProtocolMode,
	}
}

// Start begins listening for WebSocket connections in a background goroutine.
func (s *Server) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleWS)

	go func() {
		log.Printf("WebSocket server listening on %s", s.addr)
		if err := http.ListenAndServe(s.addr, mux); err != nil {
			log.Fatalf("WebSocket server error: %v", err)
		}
	}()
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	s.mu.Lock()
	s.clients[conn] = &clientState{conn: conn, mode: s.defaultProtocolMode}
	s.mu.Unlock()

	log.Printf("Tracker connected (%d total)", s.ClientCount())

	// Read messages in a goroutine (handle handshake, refresh requests)
	go s.readLoop(conn)
}

func (s *Server) readLoop(conn *websocket.Conn) {
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		conn.Close()
		log.Printf("Tracker disconnected (%d remaining)", s.ClientCount())
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var envelope struct {
			Type        string                  `json:"type"`
			Features    []string                `json:"features"`
			Flags       map[string]interface{}  `json:"flags"`
			MemoryAreas *rawMemoryAreasMessage `json:"memoryAreas"`
		}
		if err := json.Unmarshal(msg, &envelope); err != nil {
			continue
		}

		switch envelope.Type {
		case "handshake":
			mode, err := s.resolveProtocolMode(envelope.Features, envelope.Flags)
			if err != nil {
				s.sendProtocolError(conn, err.Error())
				return
			}
			if mode == ProtocolModeLegacy && !s.enableLegacyInterpretation {
				s.sendProtocolError(conn, "legacy protocol disabled")
				return
			}
			s.setClientMode(conn, mode)
			s.setClientRawMemoryAreas(conn, envelope.MemoryAreas)
			s.markClientForFullSync(conn)

			ack := HandshakeAckMessage{
				Type:     "handshAck",
				Version:  "0.1.0",
				Name:     "ootmm-autotracker",
				Refresh:  true,
				Mode:     string(mode),
				Features: supportedFeatures(mode),
			}
			data, _ := json.Marshal(ack)
			conn.WriteMessage(websocket.TextMessage, data)

		case "sendFull":
			s.markClientForFullSync(conn)

		case "refresh":
			s.markClientForFullSync(conn)
			log.Println("Refresh requested by tracker")
		}
	}
}

func (s *Server) resolveProtocolMode(features []string, flags map[string]interface{}) (ProtocolMode, error) {
	if override, ok := stringFlag(flags, "protocol"); ok {
		return ParseProtocolMode(override)
	}
	if override, ok := stringFlag(flags, "mode"); ok {
		return ParseProtocolMode(override)
	}

	for _, feature := range features {
		switch strings.ToLower(strings.TrimSpace(feature)) {
		case string(ProtocolModeRaw):
			return ProtocolModeRaw, nil
		case "items", "checks", "locations":
			return ProtocolModeLegacy, nil
		}
	}

	return s.defaultProtocolMode, nil
}

func stringFlag(flags map[string]interface{}, key string) (string, bool) {
	if len(flags) == 0 {
		return "", false
	}
	value, ok := flags[key]
	if !ok {
		return "", false
	}
	text, ok := value.(string)
	if !ok {
		return "", false
	}
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", false
	}
	return trimmed, true
}

func normalizeRawMemoryAreaNames(names []string) map[string]struct{} {
	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		set[trimmed] = struct{}{}
	}
	if len(set) == 0 {
		return nil
	}
	return set
}

func normalizeRawMemoryAreas(request *rawMemoryAreasMessage) rawMemoryAreaSelection {
	if request == nil {
		return rawMemoryAreaSelection{}
	}

	return rawMemoryAreaSelection{
		hasRequest: true,
		oot:        normalizeRawMemoryAreaNames(request.Oot),
		mm:         normalizeRawMemoryAreaNames(request.Mm),
	}
}

func (selection rawMemoryAreaSelection) requestedNamesForGame(game ootmm.ActiveGame) (map[string]struct{}, bool) {
	if !selection.hasRequest {
		return nil, false
	}

	switch game {
	case ootmm.GameOot:
		return selection.oot, true
	case ootmm.GameMm:
		return selection.mm, true
	default:
		return nil, false
	}
}

func filterRawChunksForGame(
	chunks []ootmm.RawChunk,
	selection rawMemoryAreaSelection,
	game ootmm.ActiveGame,
) []ootmm.RawChunk {
	requestedNames, filtered := selection.requestedNamesForGame(game)
	if !filtered {
		return chunks
	}
	if len(requestedNames) == 0 {
		return nil
	}

	result := make([]ootmm.RawChunk, 0, len(chunks))
	for _, chunk := range chunks {
		if _, ok := requestedNames[chunk.Name]; ok {
			result = append(result, chunk)
		}
	}
	return result
}

func supportedFeatures(mode ProtocolMode) []string {
	if mode == ProtocolModeRaw {
		return []string{"raw"}
	}
	return []string{"items", "checks", "location"}
}

func (s *Server) setClientMode(conn *websocket.Conn, mode ProtocolMode) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, ok := s.clients[conn]
	if !ok {
		return
	}
	client.mode = mode
}

func (s *Server) setClientRawMemoryAreas(conn *websocket.Conn, request *rawMemoryAreasMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, ok := s.clients[conn]
	if !ok {
		return
	}
	client.rawMemory = normalizeRawMemoryAreas(request)
}

func (s *Server) sendProtocolError(conn *websocket.Conn, message string) {
	data, err := json.Marshal(map[string]interface{}{
		"type":    "error",
		"message": message,
	})
	if err == nil {
		_ = conn.WriteMessage(websocket.TextMessage, data)
	}
}

func (s *Server) markClientForFullSync(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, ok := s.clients[conn]
	if !ok {
		return
	}

	client.wantsFull = true
}

// RequestFullSyncAll marks every connected client to receive a fresh snapshot.
func (s *Server) RequestFullSyncAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, client := range s.clients {
		client.wantsFull = true
	}
}

// ClientCount returns the number of connected tracker clients.
func (s *Server) ClientCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.clients)
}

// LegacyClientCount returns the number of clients subscribed to the legacy
// interpreted item/check/location stream.
func (s *Server) LegacyClientCount() int {
	return s.clientCountByMode(ProtocolModeLegacy)
}

// RawClientCount returns the number of clients subscribed to the raw stream.
func (s *Server) RawClientCount() int {
	return s.clientCountByMode(ProtocolModeRaw)
}

func (s *Server) clientCountByMode(mode ProtocolMode) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for _, client := range s.clients {
		if client.mode == mode {
			count++
		}
	}
	return count
}

// BroadcastItems sends item updates to all connected clients.
func (s *Server) BroadcastItems(items []tracker.ItemDiff, diff bool) {
	if len(items) == 0 {
		return
	}
	msg := ItemMessage{
		Type:    "item",
		Diff:    diff,
		Refresh: true,
		Items:   items,
	}
	s.broadcastReadyLegacyClients(msg)
}

// BroadcastChecks sends check updates to all connected clients.
func (s *Server) BroadcastChecks(checks []tracker.CheckDiff, diff bool) {
	if len(checks) == 0 {
		return
	}
	msg := CheckMessage{
		Type:    "check",
		Diff:    diff,
		Refresh: true,
		Checks:  checks,
	}
	s.broadcastReadyLegacyClients(msg)
}

// BroadcastDelta sends location updates before item/check diffs when they changed in the same poll.
func (s *Server) BroadcastDelta(game ootmm.ActiveGame, sceneID uint16, locationChanged bool, items []tracker.ItemDiff, checks []tracker.CheckDiff) {
	if locationChanged {
		s.BroadcastLocation(game, sceneID)
	}

	s.BroadcastItems(items, true)
	s.BroadcastChecks(checks, true)
}

// BroadcastLocation sends the current location to all clients.
func (s *Server) BroadcastLocation(game ootmm.ActiveGame, sceneID uint16) {
	msg := LocationMessage{
		Type:    "location",
		Refresh: true,
		Game:    game.String(),
		SceneID: sceneID,
	}
	s.broadcastReadyLegacyClients(msg)
}

// BroadcastRawSnapshot sends a complete raw snapshot to raw-mode clients.
func (s *Server) BroadcastRawSnapshot(frame *ootmm.RawFrame) {
	if frame == nil || !frame.Valid || frame.ActiveGame == ootmm.GameNone || len(frame.Chunks) == 0 {
		return
	}

	sequence := s.nextRawSequence()

	s.mu.Lock()
	defer s.mu.Unlock()

	for conn, client := range s.clients {
		if client.mode != ProtocolModeRaw {
			continue
		}

		chunks := filterRawChunksForGame(frame.Chunks, client.rawMemory, frame.ActiveGame)
		if len(chunks) == 0 {
			continue
		}

		msg := RawMessage{
			Type:          "raw",
			SchemaVersion: rawSchemaVersion,
			Diff:          false,
			Refresh:       true,
			Sequence:      sequence,
			Game:          frame.ActiveGame.String(),
			SaveIndex:     frame.SaveIndex,
			Chunks:        chunks,
		}

		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("JSON marshal error: %v", err)
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			conn.Close()
			delete(s.clients, conn)
			continue
		}
		client.wantsFull = false
	}
}

func (s *Server) nextRawSequence() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rawSequence++
	return s.rawSequence
}

func (s *Server) broadcastReadyLegacyClients(msg interface{}) {
	s.broadcast(msg, func(client *clientState) bool {
		return client.mode == ProtocolModeLegacy && !client.wantsFull
	})
}

func (s *Server) broadcast(msg interface{}, include func(*clientState) bool, postSend ...func(*clientState)) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for conn, client := range s.clients {
		if include != nil && !include(client) {
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			conn.Close()
			delete(s.clients, conn)
			continue
		}
		if len(postSend) > 0 && postSend[0] != nil {
			postSend[0](client)
		}
	}
}

// HasPendingFullSync reports whether any client requested a full state.
func (s *Server) HasPendingFullSync() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, client := range s.clients {
		if client.mode == ProtocolModeLegacy && client.wantsFull {
			return true
		}
	}

	return false
}

// FlushFullState sends a full snapshot to clients that requested one.
func (s *Server) FlushFullState(items []tracker.ItemDiff, checks []tracker.CheckDiff, game ootmm.ActiveGame, sceneID uint16) {
	conns := s.consumeLegacyFullSyncRequests()
	if len(conns) == 0 {
		return
	}

	s.sendToClients(conns, LocationMessage{
		Type:    "location",
		Refresh: false,
		Game:    game.String(),
		SceneID: sceneID,
	})

	if len(items) > 0 {
		s.sendToClients(conns, ItemMessage{
			Type:    "item",
			Diff:    false,
			Refresh: false,
			Items:   items,
		})
	}

	s.sendToClients(conns, CheckMessage{
		Type:    "check",
		Diff:    false,
		Refresh: false,
		Checks:  checks,
	})

	s.sendToClients(conns, map[string]interface{}{
		"type":    "refresh",
		"refresh": true,
	})
}

// FlushInitialItems sends only the current item inventory to clients that
// connected before the tracker had an established baseline.
func (s *Server) FlushInitialItems(items []tracker.ItemDiff) {
	conns := s.consumeLegacyFullSyncRequests()
	if len(conns) == 0 {
		return
	}

	if len(items) > 0 {
		s.sendToClients(conns, ItemMessage{
			Type:    "item",
			Diff:    false,
			Refresh: false,
			Items:   items,
		})
	}

	s.sendToClients(conns, map[string]interface{}{
		"type":    "refresh",
		"refresh": true,
	})
}

func (s *Server) consumeLegacyFullSyncRequests() []*websocket.Conn {
	s.mu.Lock()
	defer s.mu.Unlock()

	conns := make([]*websocket.Conn, 0, len(s.clients))
	for _, client := range s.clients {
		if client.mode != ProtocolModeLegacy || !client.wantsFull {
			continue
		}

		client.wantsFull = false
		conns = append(conns, client.conn)
	}

	return conns
}

func (s *Server) sendToClients(conns []*websocket.Conn, msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return
	}

	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			s.mu.Lock()
			conn.Close()
			delete(s.clients, conn)
			s.mu.Unlock()
		}
	}
}

type ItemMessage struct {
	Type    string             `json:"type"`
	Diff    bool               `json:"diff"`
	Refresh bool               `json:"refresh"`
	Items   []tracker.ItemDiff `json:"items"`
}

type CheckMessage struct {
	Type    string              `json:"type"`
	Diff    bool                `json:"diff"`
	Refresh bool                `json:"refresh"`
	Checks  []tracker.CheckDiff `json:"checks"`
}

type LocationMessage struct {
	Type    string `json:"type"`
	Refresh bool   `json:"refresh"`
	Game    string `json:"game"`
	SceneID uint16 `json:"sceneId"`
}

type HandshakeAckMessage struct {
	Type     string   `json:"type"`
	Version  string   `json:"version"`
	Name     string   `json:"name"`
	Refresh  bool     `json:"refresh"`
	Mode     string   `json:"mode,omitempty"`
	Features []string `json:"features,omitempty"`
}

type RawMessage struct {
	Type          string           `json:"type"`
	SchemaVersion string           `json:"schemaVersion"`
	Diff          bool             `json:"diff"`
	Refresh       bool             `json:"refresh"`
	Sequence      uint64           `json:"sequence"`
	Game          string           `json:"game"`
	SaveIndex     uint32           `json:"saveIndex"`
	Chunks        []ootmm.RawChunk `json:"chunks"`
}
