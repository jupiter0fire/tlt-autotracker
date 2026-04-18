package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"ootmm-autotracker/n64"
	"ootmm-autotracker/ootmm"
	"ootmm-autotracker/pj64"
	"ootmm-autotracker/retroarch"
	"ootmm-autotracker/tracker"
	"ootmm-autotracker/ws"
)

const (
	pollInterval = 100 * time.Millisecond
	retryDelay   = 2 * time.Second
)

// emulatorBackend abstracts the lifecycle of RetroArch and PJ64 connections.
type emulatorBackend interface {
	Connect() error
	Close()
}

func main() {
	raHost := flag.String("ra-host", retroarch.DefaultHost, "RetroArch host")
	raPort := flag.Int("ra-port", retroarch.DefaultPort, "RetroArch network command port")
	pj64Mode := flag.Bool("pj64", false, "Use Project64 backend (PJ64 Lua adapter connects to us)")
	pj64Port := flag.Int("pj64-port", pj64.DefaultPort, "PJ64 TCP listen port")
	wsAddr := flag.String("ws-addr", ":17026", "WebSocket listen address")
	flag.Parse()
	consoleCommands := startConsoleCommands()

	fmt.Println("=== OoTMM Autotracker ===")

	// Start WebSocket server
	server := ws.NewServer(*wsAddr)
	server.Start()

	// Create emulator backend
	var (
		backend emulatorBackend
		mem     *n64.Memory
	)

	if *pj64Mode {
		srv := pj64.NewServer(*pj64Port)
		if err := srv.Start(); err != nil {
			log.Fatalf("PJ64 server: %v", err)
		}
		mem = n64.NewMemory(srv)
		mem.SetSwizzle(false)
		mem.SetBaseShift(n64.VirtualBase) // PJ64 always uses virtual addressing
		backend = srv
		fmt.Printf("PJ64:      listening on port %d\n", *pj64Port)
		fmt.Printf("WebSocket: %s\n", *wsAddr)
		fmt.Println("Console: dump [label|path] schreibt einen JSON-Snapshot, help zeigt Befehle")
		fmt.Println()
		fmt.Println("Waiting for PJ64 Lua adapter to connect...")
		fmt.Println("(Load pj64_adapter.lua in Project64's scripting console)")
	} else {
		client := retroarch.NewClient(*raHost, *raPort)
		mem = n64.NewMemory(client)
		backend = client
		fmt.Printf("RetroArch: %s:%d\n", *raHost, *raPort)
		fmt.Printf("WebSocket: %s\n", *wsAddr)
		fmt.Println("Console: dump [label|path] schreibt einen JSON-Snapshot, help zeigt Befehle")
		fmt.Println()
		fmt.Println("Waiting for RetroArch connection...")
	}

	reader := ootmm.NewReader(mem)
	state := tracker.NewState()

	var (
		connected bool
		probed    bool
		lastGame  ootmm.ActiveGame
		lastScene uint16
	)

	for {
		time.Sleep(pollInterval)
		drainConsoleCommands(consoleCommands, connected, probed, mem)

		// Step 1: Ensure connected to emulator
		if !connected {
			if err := backend.Connect(); err != nil {
				continue
			}
			connected = true
			probed = false
			if *pj64Mode {
				log.Println("PJ64 adapter connected")
			} else {
				log.Println("Connected to RetroArch")
			}
		}

		// Step 2: Probe address space (once per connection)
		if !probed {
			if err := mem.Probe(); err != nil {
				// Check if the connection died during probe (e.g. PJ64 Lua crashed)
				if *pj64Mode && !backend.(*pj64.Server).IsConnected() {
					log.Println("PJ64 connection lost during probe")
					connected = false
				}
				// Not an OoTMM session yet, or emulator not running a game
				time.Sleep(retryDelay)
				continue
			}
			probed = true
			log.Println("OoTMM detected!")
		}

		drainConsoleCommands(consoleCommands, connected, probed, mem)

		// Step 3: Read game state
		gs, err := reader.ReadState()
		if err != nil {
			log.Printf("Read error: %v", err)
			// Connection might be lost
			backend.Close()
			connected = false
			time.Sleep(retryDelay)
			continue
		}

		if !gs.Valid {
			// OoTMM context not valid (maybe in title screen)
			continue
		}

		if gs.ActiveGame == ootmm.GameNone {
			// Game transition in progress or data discarded due to mid-read switch
			continue
		}

		// Step 4: Compute deltas
		changedItems, changedChecks, gameChanged := state.Update(gs)
		currentScene := getActiveScene(gs)
		locationChanged := gameChanged || currentScene != lastScene || gs.ActiveGame != lastGame

		// Step 5: Broadcast to trackers
		if server.ClientCount() > 0 {
			if gameChanged {
				log.Printf("Active game: %s", gs.ActiveGame)
			}

			server.BroadcastDelta(gs.ActiveGame, currentScene, locationChanged, changedItems, changedChecks)

			if server.HasPendingFullSync() {
				fullItems, fullChecks := state.FullState(gs)
				server.FlushFullState(fullItems, fullChecks, gs.ActiveGame, currentScene)
			}

			lastScene = currentScene
			lastGame = gs.ActiveGame
		}

		// Step 6: Console output
		if gameChanged || len(changedItems) > 0 || len(changedChecks) > 0 {
			printStatus(gs, changedItems, changedChecks, server.ClientCount())
		}
	}
}

func getActiveScene(gs *ootmm.GameState) uint16 {
	switch gs.ActiveGame {
	case ootmm.GameOot:
		return gs.Oot.SceneID
	case ootmm.GameMm:
		// MM doesn't have a direct sceneId in the same spot;
		// for now return 0 (to be refined)
		return 0
	}
	return 0
}

func printStatus(gs *ootmm.GameState, items []tracker.ItemDiff, checks []tracker.CheckDiff, clients int) {
	fmt.Printf("[%s] Game: %s | Scene: 0x%02X | Δ items: %d | Δ checks: %d | Trackers: %d\n",
		time.Now().Format("15:04:05"),
		gs.ActiveGame,
		getActiveScene(gs),
		len(items),
		len(checks),
		clients,
	)
}
