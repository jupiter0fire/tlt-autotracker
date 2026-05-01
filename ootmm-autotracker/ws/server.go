package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"ootmm-autotracker/ootmm"
	"ootmm-autotracker/tracker"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Server manages WebSocket connections to tracker frontends.
type Server struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]*clientState
	addr    string
}

type clientState struct {
	conn      *websocket.Conn
	wantsFull bool
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
			Type string `json:"type"`
		}
		if err := json.Unmarshal(msg, &envelope); err != nil {
			continue
		}

		switch envelope.Type {
		case "handshake":
			s.markClientForFullSync(conn)

			ack := map[string]interface{}{
				"type":    "handshAck",
				"version": "0.1.0",
				"name":    "ootmm-autotracker",
				"refresh": true,
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
	s.broadcastReadyClients(msg)
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
	s.broadcastReadyClients(msg)
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
	s.broadcastReadyClients(msg)
}

func (s *Server) broadcastReadyClients(msg interface{}) {
	s.broadcast(msg, func(client *clientState) bool {
		return !client.wantsFull
	})
}

func (s *Server) broadcast(msg interface{}, include func(*clientState) bool) {
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
		}
	}
}

// HasPendingFullSync reports whether any client requested a full state.
func (s *Server) HasPendingFullSync() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, client := range s.clients {
		if client.wantsFull {
			return true
		}
	}

	return false
}

// FlushFullState sends a full snapshot to clients that requested one.
func (s *Server) FlushFullState(items []tracker.ItemDiff, checks []tracker.CheckDiff, game ootmm.ActiveGame, sceneID uint16) {
	conns := s.consumeFullSyncRequests()
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
	conns := s.consumeFullSyncRequests()
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

func (s *Server) consumeFullSyncRequests() []*websocket.Conn {
	s.mu.Lock()
	defer s.mu.Unlock()

	conns := make([]*websocket.Conn, 0, len(s.clients))
	for _, client := range s.clients {
		if !client.wantsFull {
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
