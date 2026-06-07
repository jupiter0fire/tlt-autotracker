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

	"ootmm-autotracker/ares"
	"ootmm-autotracker/n64"
	"ootmm-autotracker/ootmm"
	"ootmm-autotracker/pj64"
	"ootmm-autotracker/retroarch"
	"ootmm-autotracker/ws"
)

const (
	pollInterval            = 100 * time.Millisecond
	retryDelay              = 2 * time.Second
	startupRetryDelay       = 500 * time.Millisecond
	ootmmLostTimeout        = 20 * time.Second
	shortCommitHashLength   = 12
	defaultWSAddr           = "127.0.0.1:17026"
	defaultWSAllowedOrigins = "http://localhost:5173,https://www.thelasttracker.org,https://www.wbsch.de"
)

var commitHash string

// emulatorBackend abstracts the lifecycle of emulator connections (RetroArch, PJ64, Ares).
type emulatorBackend interface {
	Connect() error
	Close()
	IsConnected() bool
}

type backendKind string

const (
	backendRetroArch backendKind = "retroarch"
	backendPJ64      backendKind = "pj64"
	backendAres      backendKind = "ares"
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
	aresMode := flag.Bool("ares", false, "Force Ares backend (skip auto-detection)")
	aresHost := flag.String("ares-host", ares.DefaultHost, "Ares GDB host")
	aresPort := flag.Int("ares-port", ares.DefaultPort, "Ares GDB port")
	wsAddr := flag.String("ws-addr", defaultWSAddr, "WebSocket listen address")
	wsAllowedOrigins := flag.String("ws-allowed-origins", defaultWSAllowedOrigins, "Comma-separated allowed WebSocket origins")
	flag.Parse()

	fmt.Println("=== OoTMM Autotracker ===")

	// Start WebSocket server
	server, err := ws.NewServer(*wsAddr, splitCommaSeparatedList(*wsAllowedOrigins))
	if err != nil {
		log.Fatalf("websocket server config: %v", err)
	}
	server.Start()

	selected, err := selectBackend(*raHost, *raPort, *pj64Port, *aresHost, *aresPort, *wsAddr, *pj64Mode, *aresMode)
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
		invalidSaveSince      time.Time
		unstableGameSince     time.Time
		hasReadValidSave      bool
		hasSeenActiveGame     bool
	)

	restartSession := func() {
		reader = resetTrackingSession(backend, mem)
		connected = false
		probed = false
		lastGame = ootmm.GameNone
		ootmmUnavailableSince = time.Time{}
		invalidSaveSince = time.Time{}
		unstableGameSince = time.Time{}
		hasReadValidSave = false
		hasSeenActiveGame = false
	}

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
			invalidSaveSince = time.Time{}
			unstableGameSince = time.Time{}
			hasReadValidSave = false
			hasSeenActiveGame = false
			switch selected.kind {
			case backendPJ64:
				log.Println("Project64 adapter connected")
			case backendAres:
				log.Println("Connected to Ares (GDB)")
			default:
				log.Println("Connected to RetroArch")
			}
		}

		// Step 2: Probe address space (once per connection)
		if !probed {
			if err := mem.Probe(); err != nil {
				backendConnected := backend.IsConnected()
				if shouldRestartSessionOnReadFailure(hasReadValidSave, backendConnected) &&
					(!backendConnected || n64.IsProbeReadError(err)) {
					if !backendConnected {
						switch selected.kind {
						case backendPJ64:
							log.Println("Project64 connection lost during probe")
						case backendAres:
							log.Println("Ares connection lost during probe")
						default:
							if hasReadValidSave {
								log.Println("RetroArch connection lost during probe")
							}
						}
					} else {
						log.Printf("Probe read error: %v; reconnecting backend for a fresh session", err)
					}
					restartSession()
					time.Sleep(retryDelay)
					continue
				}
				if elapsed := noteOoTMMUnavailableAfterValidSave(hasReadValidSave, &ootmmUnavailableSince, now); elapsed > ootmmLostTimeout {
					log.Printf("OoTMM unavailable for %s during probe; reconnecting backend for a fresh session", elapsed.Round(time.Second))
					restartSession()
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
			backendConnected := backend.IsConnected()
			if shouldRestartSessionOnReadFailure(hasReadValidSave, backendConnected) {
				if !backendConnected {
					switch selected.kind {
					case backendPJ64:
						log.Println("Project64 connection lost during raw read")
					case backendAres:
						log.Println("Ares connection lost during raw read")
					default:
						if hasReadValidSave {
							log.Println("RetroArch connection lost during raw read")
						}
					}
				} else {
					log.Printf("Raw read error: %v", err)
				}
				restartSession()
			}
			time.Sleep(retryDelay)
			continue
		}

		if !rawFrame.Valid {
			if elapsed := noteOoTMMUnavailableAfterValidSave(hasReadValidSave, &invalidSaveSince, now); elapsed > ootmmLostTimeout {
				log.Printf("OoTMM unavailable for %s; reconnecting backend for a fresh session", elapsed.Round(time.Second))
				restartSession()
				time.Sleep(retryDelay)
			}
			continue
		}
		hasReadValidSave = true
		invalidSaveSince = time.Time{}

		if rawFrame.ActiveGame == ootmm.GameNone {
			if elapsed := noteOoTMMUnavailableAfterStableActiveGame(hasSeenActiveGame, &unstableGameSince, now); elapsed > ootmmLostTimeout {
				log.Printf("OoTMM did not reach a stable active game for %s; reconnecting backend for a fresh session", elapsed.Round(time.Second))
				restartSession()
				time.Sleep(retryDelay)
			}
			continue
		}
		hasSeenActiveGame = true
		unstableGameSince = time.Time{}

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

func splitCommaSeparatedList(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func noteOoTMMUnavailable(unavailableSince *time.Time, now time.Time) time.Duration {
	if unavailableSince.IsZero() {
		*unavailableSince = now
		return 0
	}
	return now.Sub(*unavailableSince)
}

func noteOoTMMUnavailableAfterValidSave(hasReadValidSave bool, unavailableSince *time.Time, now time.Time) time.Duration {
	if !hasReadValidSave {
		return 0
	}
	return noteOoTMMUnavailable(unavailableSince, now)
}

func noteOoTMMUnavailableAfterStableActiveGame(hasSeenActiveGame bool, unavailableSince *time.Time, now time.Time) time.Duration {
	if !hasSeenActiveGame {
		return 0
	}
	return noteOoTMMUnavailable(unavailableSince, now)
}

func shouldRestartSessionOnReadFailure(hasReadValidSave bool, backendConnected bool) bool {
	return !backendConnected || hasReadValidSave
}

func resetTrackingSession(backend emulatorBackend, mem *n64.Memory) *ootmm.Reader {
	backend.Close()
	return ootmm.NewReader(mem)
}

func selectBackend(raHost string, raPort int, pj64Port int, aresHost string, aresPort int, wsAddr string, forcePJ64 bool, forceAres bool) (*backendOption, error) {
	if forcePJ64 {
		fmt.Printf("PJ64:      listening on port %d\n", pj64Port)
		fmt.Printf("WebSocket: %s\n", wsAddr)
		fmt.Println()
		fmt.Println("Waiting for Project64 Lua adapter to connect...")
		fmt.Println("(Load pj64_adapter.lua in Project64's scripting console)")
		return newPJ64Option(pj64Port)
	}

	if forceAres {
		fmt.Printf("Ares:      %s:%d\n", aresHost, aresPort)
		fmt.Printf("WebSocket: %s\n", wsAddr)
		fmt.Println()
		fmt.Println("Waiting for Ares GDB stub to become available...")
		return newAresOption(aresHost, aresPort)
	}

	fmt.Printf("RetroArch: %s:%d\n", raHost, raPort)
	fmt.Printf("PJ64:      listening on port %d\n", pj64Port)
	fmt.Printf("Ares:      %s:%d\n", aresHost, aresPort)
	fmt.Printf("WebSocket: %s\n", wsAddr)
	fmt.Println()
	fmt.Println("Checking RetroArch, Project64, and Ares...")

	retroArchOption := newRetroArchOption(raHost, raPort)
	pj64Option, err := newPJ64Option(pj64Port)
	if err != nil {
		log.Printf("Project64 listener unavailable: %v", err)
	}
	aresOption, err2 := newAresOption(aresHost, aresPort)
	if err2 != nil {
		log.Printf("Ares client unavailable: %v", err2)
	}

	waitingPrinted := false
	for {
		available := detectAvailableBackends(retroArchOption, pj64Option, aresOption)
		switch len(available) {
		case 0:
			if !waitingPrinted {
				fmt.Println("Waiting for RetroArch, Project64, or Ares...")
				fmt.Println("(For PJ64, load pj64_adapter.lua in Project64's scripting console)")
				waitingPrinted = true
			}
			time.Sleep(startupRetryDelay)
		case 1:
			chosen := available[0]
			fmt.Printf("Using %s automatically.\n", chosen.name)
			cleanupUnselectedBackends(chosen, retroArchOption, pj64Option, aresOption)
			return chosen, nil
		default:
			chosen, err := promptForBackendChoice(available)
			if err != nil {
				cleanupUnselectedBackends(nil, retroArchOption, pj64Option, aresOption)
				return nil, err
			}
			fmt.Printf("Using %s.\n", chosen.name)
			cleanupUnselectedBackends(chosen, retroArchOption, pj64Option, aresOption)
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

func newAresOption(host string, port int) (*backendOption, error) {
	client := ares.NewClient(host, port)
	mem := n64.NewMemory(client)
	mem.SetSwizzle(false)
	mem.SetBaseShift(n64.VirtualBase)
	return &backendOption{
		kind:     backendAres,
		name:     "Ares",
		backend:  client,
		mem:      mem,
		shutdown: client.Close,
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
		fmt.Println("Multiple emulators are reachable. Select one:")
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

		fmt.Println("Please enter a number or emulator name.")
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
	case backendAres:
		return choice == "ares"
	default:
		return false
	}
}
