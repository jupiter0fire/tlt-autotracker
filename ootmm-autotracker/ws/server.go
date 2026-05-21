package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"ootmm-autotracker/ootmm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

const rawSchemaVersion = "1"

// Server manages WebSocket connections to tracker frontends.
type Server struct {
	mu          sync.Mutex
	clients     map[*websocket.Conn]*clientState
	addr        string
	rawSequence uint64
}

type clientState struct {
	conn      *websocket.Conn
	rawMemory rawMemoryAreaSelection
}

type rawMemoryAreasMessage struct {
	Oot []ootmm.RawChunkSpec `json:"oot"`
	Mm  []ootmm.RawChunkSpec `json:"mm"`
}

type rawMemoryAreaSelection struct {
	hasRequest bool
	ootSpecs   []ootmm.RawChunkSpec
	mmSpecs    []ootmm.RawChunkSpec
	ootNames   map[string]struct{}
	mmNames    map[string]struct{}
}

func NewServer(addr string) *Server {
	return &Server{
		clients: make(map[*websocket.Conn]*clientState),
		addr:    addr,
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
	s.clients[conn] = &clientState{conn: conn}
	s.mu.Unlock()

	log.Printf("Tracker connected (%d total)", s.ClientCount())
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
			Type        string                 `json:"type"`
			Features    []string               `json:"features"`
			Flags       map[string]interface{} `json:"flags"`
			MemoryAreas *rawMemoryAreasMessage `json:"memoryAreas"`
		}
		if err := json.Unmarshal(msg, &envelope); err != nil {
			continue
		}

		switch envelope.Type {
		case "handshake":
			if requestsLegacyProtocol(envelope.Features, envelope.Flags) {
				s.sendProtocolError(conn, "legacy protocol removed")
				return
			}
			s.setClientRawMemoryAreas(conn, envelope.MemoryAreas)

			ack := HandshakeAckMessage{
				Type:     "handshAck",
				Version:  "0.1.0",
				Name:     "ootmm-autotracker",
				Refresh:  true,
				Mode:     "raw",
				Features: []string{"raw"},
			}
			data, _ := json.Marshal(ack)
			_ = conn.WriteMessage(websocket.TextMessage, data)

		case "sendFull", "refresh":
			continue
		}
	}
}

func requestsLegacyProtocol(features []string, flags map[string]interface{}) bool {
	if override, ok := stringFlag(flags, "protocol"); ok && strings.EqualFold(override, "legacy") {
		return true
	}
	if override, ok := stringFlag(flags, "mode"); ok && strings.EqualFold(override, "legacy") {
		return true
	}

	for _, feature := range features {
		switch strings.ToLower(strings.TrimSpace(feature)) {
		case "items", "checks", "locations", "legacy":
			return true
		}
	}

	return false
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

func normalizeRawChunkSpecs(specs []ootmm.RawChunkSpec) ([]ootmm.RawChunkSpec, map[string]struct{}) {
	set := make(map[string]struct{}, len(specs))
	normalized := make([]ootmm.RawChunkSpec, 0, len(specs))
	for _, spec := range specs {
		trimmed := strings.TrimSpace(spec.Name)
		if trimmed == "" || spec.Address == 0 || spec.Length <= 0 {
			continue
		}
		if _, exists := set[trimmed]; exists {
			continue
		}
		spec.Name = trimmed
		normalized = append(normalized, spec)
		set[trimmed] = struct{}{}
	}
	if len(normalized) == 0 {
		return nil, nil
	}
	return normalized, set
}

func normalizeRawMemoryAreas(request *rawMemoryAreasMessage) rawMemoryAreaSelection {
	if request == nil {
		return rawMemoryAreaSelection{}
	}
	ootSpecs, ootNames := normalizeRawChunkSpecs(request.Oot)
	mmSpecs, mmNames := normalizeRawChunkSpecs(request.Mm)

	return rawMemoryAreaSelection{
		hasRequest: true,
		ootSpecs:   ootSpecs,
		mmSpecs:    mmSpecs,
		ootNames:   ootNames,
		mmNames:    mmNames,
	}
}

func (selection rawMemoryAreaSelection) requestedNamesForGame(game ootmm.ActiveGame) (map[string]struct{}, bool) {
	if !selection.hasRequest {
		return nil, false
	}

	switch game {
	case ootmm.GameOot:
		return selection.ootNames, true
	case ootmm.GameMm:
		return selection.mmNames, true
	default:
		return nil, false
	}
}

func (selection rawMemoryAreaSelection) requestedSpecsForGame(game ootmm.ActiveGame) ([]ootmm.RawChunkSpec, bool) {
	if !selection.hasRequest {
		return nil, false
	}

	switch game {
	case ootmm.GameOot:
		return selection.ootSpecs, true
	case ootmm.GameMm:
		return selection.mmSpecs, true
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

func (s *Server) setClientRawMemoryAreas(conn *websocket.Conn, request *rawMemoryAreasMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, ok := s.clients[conn]
	if !ok {
		return
	}
	client.rawMemory = normalizeRawMemoryAreas(request)
}

func mergeRequestedRawChunkSpecs(clients map[*websocket.Conn]*clientState, game ootmm.ActiveGame) []ootmm.RawChunkSpec {
	merged := make(map[string]ootmm.RawChunkSpec)
	for _, client := range clients {
		specs, ok := client.rawMemory.requestedSpecsForGame(game)
		if !ok {
			continue
		}
		for _, spec := range specs {
			if _, exists := merged[spec.Name]; exists {
				continue
			}
			merged[spec.Name] = spec
		}
	}
	result := make([]ootmm.RawChunkSpec, 0, len(merged))
	for _, spec := range merged {
		result = append(result, spec)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Address == result[j].Address {
			return result[i].Name < result[j].Name
		}
		return result[i].Address < result[j].Address
	})
	return result
}

func (s *Server) RequestedRawChunkSpecs() ootmm.RawChunkSpecSelection {
	s.mu.Lock()
	defer s.mu.Unlock()

	return ootmm.RawChunkSpecSelection{
		Oot: mergeRequestedRawChunkSpecs(s.clients, ootmm.GameOot),
		Mm:  mergeRequestedRawChunkSpecs(s.clients, ootmm.GameMm),
	}
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

// ClientCount returns the number of connected tracker clients.
func (s *Server) ClientCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.clients)
}

// BroadcastRawSnapshot sends a complete raw snapshot to connected clients.
func (s *Server) BroadcastRawSnapshot(frame *ootmm.RawFrame) {
	if frame == nil || !frame.Valid || frame.ActiveGame == ootmm.GameNone || len(frame.Chunks) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.rawSequence++
	sequence := s.rawSequence

	for conn, client := range s.clients {
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
		}
	}
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
