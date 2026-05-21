package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"ootmm-autotracker/n64"
	"ootmm-autotracker/ootmm"
	"ootmm-autotracker/pj64"
	"ootmm-autotracker/retroarch"
	"ootmm-autotracker/ws"
)

const (
	pollInterval          = 100 * time.Millisecond
	retryDelay            = 2 * time.Second
	startupRetryDelay     = 500 * time.Millisecond
	ootmmLostTimeout      = 20 * time.Second
	shortCommitHashLength = 12
)

var commitHash string

// emulatorBackend abstracts the lifecycle of RetroArch and PJ64 connections.
type emulatorBackend interface {
	Connect() error
	Close()
	IsConnected() bool
}

type backendKind string

const (
	backendRetroArch backendKind = "retroarch"
	backendPJ64      backendKind = "pj64"
)

type backendOption struct {
	kind      backendKind
	name      string
	backend   emulatorBackend
	mem       *n64.Memory
	connected bool
	shutdown  func()
}

func main() {
	launched, err := relaunchInTerminalIfNeeded()
	if err != nil {
		log.Printf("automatic terminal relaunch unavailable: %v", err)
	}
	if launched {
		return
	}

	fmt.Println(startupCommitHash())

	raHost := flag.String("ra-host", retroarch.DefaultHost, "RetroArch host")
	raPort := flag.Int("ra-port", retroarch.DefaultPort, "RetroArch network command port")
	pj64Mode := flag.Bool("pj64", false, "Force Project64 backend (skip auto-detection)")
	pj64Port := flag.Int("pj64-port", pj64.DefaultPort, "PJ64 TCP listen port")
	wsAddr := flag.String("ws-addr", ":17026", "WebSocket listen address")
	flag.Parse()

	fmt.Println("=== OoTMM Autotracker ===")

	// Start WebSocket server
	server := ws.NewServer(*wsAddr)
	server.Start()

	selected, err := selectBackend(*raHost, *raPort, *pj64Port, *wsAddr, *pj64Mode)
	if err != nil {
		log.Fatalf("backend startup: %v", err)
	}
	defer selected.shutdown()

	backend := selected.backend
	mem := selected.mem

	reader := ootmm.NewReader(mem)

	var (
		connected             = selected.connected
		probed                bool
		lastGame              ootmm.ActiveGame
		ootmmUnavailableSince time.Time
	)

	for {
		time.Sleep(pollInterval)
		now := time.Now()

		// Step 1: Ensure connected to emulator
		if !connected {
			if err := backend.Connect(); err != nil {
				continue
			}
			connected = true
			probed = false
			ootmmUnavailableSince = time.Time{}
			if selected.kind == backendPJ64 {
				log.Println("Project64 adapter connected")
			} else {
				log.Println("Connected to RetroArch")
			}
		}

		// Step 2: Probe address space (once per connection)
		if !probed {
			if err := mem.Probe(); err != nil {
				// Check if the connection died during probe (e.g. PJ64 Lua crashed)
				if selected.kind == backendPJ64 && !backend.IsConnected() {
					log.Println("Project64 connection lost during probe")
					connected = false
				} else if elapsed := noteOoTMMUnavailable(&ootmmUnavailableSince, now); elapsed > ootmmLostTimeout {
					log.Printf("OoTMM unavailable for %s during probe; reconnecting backend for a fresh session", elapsed.Round(time.Second))
					reader = resetTrackingSession(backend, mem)
					connected = false
					probed = false
					lastGame = ootmm.GameNone
					ootmmUnavailableSince = time.Time{}
				}
				// Not an OoTMM session yet, or emulator not running a game
				time.Sleep(retryDelay)
				continue
			}
			probed = true
			ootmmUnavailableSince = time.Time{}
			log.Println("OoTMM detected!")
		}

		// Step 3: Read game state
		rawFrame, err := reader.ReadRawFrameWithSelection(server.RequestedRawChunkSpecs())
		if err != nil {
			log.Printf("Raw read error: %v", err)
			reader = resetTrackingSession(backend, mem)
			connected = false
			probed = false
			lastGame = ootmm.GameNone
			ootmmUnavailableSince = time.Time{}
			time.Sleep(retryDelay)
			continue
		}

		if !rawFrame.Valid {
			if elapsed := noteOoTMMUnavailable(&ootmmUnavailableSince, now); elapsed > ootmmLostTimeout {
				log.Printf("OoTMM unavailable for %s; reconnecting backend for a fresh session", elapsed.Round(time.Second))
				reader = resetTrackingSession(backend, mem)
				connected = false
				probed = false
				lastGame = ootmm.GameNone
				ootmmUnavailableSince = time.Time{}
				time.Sleep(retryDelay)
			}
			continue
		}

		if rawFrame.ActiveGame == ootmm.GameNone {
			if elapsed := noteOoTMMUnavailable(&ootmmUnavailableSince, now); elapsed > ootmmLostTimeout {
				log.Printf("OoTMM did not reach a stable active game for %s; reconnecting backend for a fresh session", elapsed.Round(time.Second))
				reader = resetTrackingSession(backend, mem)
				connected = false
				probed = false
				lastGame = ootmm.GameNone
				ootmmUnavailableSince = time.Time{}
				time.Sleep(retryDelay)
			}
			continue
		}
		ootmmUnavailableSince = time.Time{}

		if rawFrame.ActiveGame != lastGame {
			log.Printf("Active game: %s", rawFrame.ActiveGame)
			lastGame = rawFrame.ActiveGame
		}

		server.BroadcastRawSnapshot(rawFrame)
	}
}

func startupCommitHash() string {
	if hash := shortCommitHash(commitHash); hash != "" {
		return hash
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	for _, setting := range buildInfo.Settings {
		if setting.Key == "vcs.revision" {
			if hash := shortCommitHash(setting.Value); hash != "" {
				return hash
			}
			break
		}
	}

	return "unknown"
}

func shortCommitHash(hash string) string {
	hash = strings.TrimSpace(hash)
	if hash == "" || hash == "(devel)" {
		return ""
	}
	if len(hash) <= shortCommitHashLength {
		return hash
	}
	return hash[:shortCommitHashLength]
}

func noteOoTMMUnavailable(unavailableSince *time.Time, now time.Time) time.Duration {
	if unavailableSince.IsZero() {
		*unavailableSince = now
		return 0
	}
	return now.Sub(*unavailableSince)
}

func resetTrackingSession(backend emulatorBackend, mem *n64.Memory) *ootmm.Reader {
	backend.Close()
	return ootmm.NewReader(mem)
}

func selectBackend(raHost string, raPort int, pj64Port int, wsAddr string, forcePJ64 bool) (*backendOption, error) {
	if forcePJ64 {
		fmt.Printf("PJ64:      listening on port %d\n", pj64Port)
		fmt.Printf("WebSocket: %s\n", wsAddr)
		fmt.Println()
		fmt.Println("Waiting for Project64 Lua adapter to connect...")
		fmt.Println("(Load pj64_adapter.lua in Project64's scripting console)")
		return newPJ64Option(pj64Port)
	}

	fmt.Printf("RetroArch: %s:%d\n", raHost, raPort)
	fmt.Printf("PJ64:      listening on port %d\n", pj64Port)
	fmt.Printf("WebSocket: %s\n", wsAddr)
	fmt.Println()
	fmt.Println("Checking RetroArch and Project64...")

	retroArchOption := newRetroArchOption(raHost, raPort)
	pj64Option, err := newPJ64Option(pj64Port)
	if err != nil {
		log.Printf("Project64 listener unavailable: %v", err)
	}

	waitingPrinted := false
	for {
		available := detectAvailableBackends(retroArchOption, pj64Option)
		switch len(available) {
		case 0:
			if !waitingPrinted {
				fmt.Println("Waiting for RetroArch or Project64...")
				fmt.Println("(Load pj64_adapter.lua in Project64's scripting console if needed)")
				waitingPrinted = true
			}
			time.Sleep(startupRetryDelay)
		case 1:
			chosen := available[0]
			fmt.Printf("Using %s automatically.\n", chosen.name)
			cleanupUnselectedBackends(chosen, retroArchOption, pj64Option)
			return chosen, nil
		default:
			chosen, err := promptForBackendChoice(available)
			if err != nil {
				cleanupUnselectedBackends(nil, retroArchOption, pj64Option)
				return nil, err
			}
			fmt.Printf("Using %s.\n", chosen.name)
			cleanupUnselectedBackends(chosen, retroArchOption, pj64Option)
			return chosen, nil
		}
	}
}

func newRetroArchOption(host string, port int) *backendOption {
	client := retroarch.NewClient(host, port)
	mem := n64.NewMemory(client)
	return &backendOption{
		kind:     backendRetroArch,
		name:     "RetroArch",
		backend:  client,
		mem:      mem,
		shutdown: client.Close,
	}
}

func newPJ64Option(port int) (*backendOption, error) {
	srv := pj64.NewServer(port)
	if err := srv.Start(); err != nil {
		return nil, err
	}
	mem := n64.NewMemory(srv)
	mem.SetSwizzle(false)
	mem.SetBaseShift(n64.VirtualBase)
	return &backendOption{
		kind:     backendPJ64,
		name:     "Project64",
		backend:  srv,
		mem:      mem,
		shutdown: srv.Stop,
	}, nil
}

func detectAvailableBackends(options ...*backendOption) []*backendOption {
	available := make([]*backendOption, 0, len(options))
	for _, option := range options {
		if option == nil {
			continue
		}
		if option.connected && option.backend.IsConnected() {
			available = append(available, option)
			continue
		}
		option.connected = false
		if err := option.backend.Connect(); err != nil {
			continue
		}
		option.connected = true
		available = append(available, option)
	}
	return available
}

func cleanupUnselectedBackends(chosen *backendOption, options ...*backendOption) {
	for _, option := range options {
		if option == nil || option == chosen {
			continue
		}
		option.shutdown()
	}
}

func promptForBackendChoice(options []*backendOption) (*backendOption, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Both RetroArch and Project64 are reachable. Select one:")
		for index, option := range options {
			fmt.Printf("  %d) %s\n", index+1, option.name)
		}
		fmt.Print("> ")

		input, err := reader.ReadString('\n')
		if err != nil && len(strings.TrimSpace(input)) == 0 {
			return nil, fmt.Errorf("read backend selection: %w", err)
		}

		if chosen := parseBackendChoice(input, options); chosen != nil {
			return chosen, nil
		}

		fmt.Println("Please enter 1/2 or the emulator name.")
	}
}

func parseBackendChoice(input string, options []*backendOption) *backendOption {
	choice := strings.ToLower(strings.TrimSpace(input))
	for index, option := range options {
		if choice == fmt.Sprintf("%d", index+1) || backendChoiceAlias(option.kind, choice) {
			return option
		}
	}
	return nil
}

func backendChoiceAlias(kind backendKind, choice string) bool {
	switch kind {
	case backendRetroArch:
		return choice == "retroarch" || choice == "ra"
	case backendPJ64:
		return choice == "project64" || choice == "project" || choice == "pj64"
	default:
		return false
	}
}
