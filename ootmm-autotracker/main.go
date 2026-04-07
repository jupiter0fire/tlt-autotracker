package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ootmm-autotracker/n64"
	"github.com/ootmm-autotracker/ootmm"
	"github.com/ootmm-autotracker/retroarch"
	"github.com/ootmm-autotracker/tracker"
	"github.com/ootmm-autotracker/ws"
)

const (
	pollInterval = 100 * time.Millisecond
	retryDelay   = 2 * time.Second
)

func main() {
	raHost := flag.String("ra-host", retroarch.DefaultHost, "RetroArch host")
	raPort := flag.Int("ra-port", retroarch.DefaultPort, "RetroArch network command port")
	wsAddr := flag.String("ws-addr", ":17026", "WebSocket listen address")
	flag.Parse()

	fmt.Println("=== OoTMM Autotracker ===")
	fmt.Printf("RetroArch: %s:%d\n", *raHost, *raPort)
	fmt.Printf("WebSocket: %s\n", *wsAddr)
	fmt.Println()

	// Start WebSocket server
	server := ws.NewServer(*wsAddr)
	server.Start()

	// Create RetroArch client
	client := retroarch.NewClient(*raHost, *raPort)
	mem := n64.NewMemory(client)
	reader := ootmm.NewReader(mem)
	state := tracker.NewState()

	var (
		connected bool
		probed    bool
		lastGame  ootmm.ActiveGame
		lastScene uint16
	)

	fmt.Println("Waiting for RetroArch connection...")

	for {
		time.Sleep(pollInterval)

		// Step 1: Ensure connected to RetroArch
		if !connected {
			if err := client.Connect(); err != nil {
				continue
			}
			connected = true
			probed = false
			log.Println("Connected to RetroArch")
		}

		// Step 2: Probe address space (once per connection)
		if !probed {
			if err := mem.Probe(); err != nil {
				// Not an OoTMM session yet, or emulator not running a game
				time.Sleep(retryDelay)
				continue
			}
			probed = true
			log.Println("OoTMM detected!")
		}

		// Step 3: Read game state
		gs, err := reader.ReadState()
		if err != nil {
			log.Printf("Read error: %v", err)
			// Connection might be lost
			client.Close()
			connected = false
			time.Sleep(retryDelay)
			continue
		}

		if !gs.Valid {
			// OoTMM context not valid (maybe in title screen)
			continue
		}

		if gs.ActiveGame == ootmm.GameNone {
			continue
		}

		// Step 4: Compute deltas
		changedItems, changedChecks, gameChanged := state.Update(gs)
		currentScene := getActiveScene(gs)

		// Step 5: Broadcast to trackers
		if server.ClientCount() > 0 {
			if gameChanged {
				log.Printf("Active game: %s", gs.ActiveGame)
			}

			server.BroadcastItems(changedItems, true)
			server.BroadcastChecks(changedChecks, true)

			// Send location on game or scene change
			if gameChanged || currentScene != lastScene || gs.ActiveGame != lastGame {
				server.BroadcastLocation(gs.ActiveGame, currentScene)
			}

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
