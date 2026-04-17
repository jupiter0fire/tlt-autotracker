package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ootmm-autotracker/n64"
	"ootmm-autotracker/ootmm"
)

type consoleCommand struct {
	name string
	args string
}

type debugSnapshot struct {
	SchemaVersion int                   `json:"schemaVersion"`
	CreatedAt     string                `json:"createdAt"`
	Summary       debugSnapshotSummary  `json:"summary"`
	Regions       []debugSnapshotRegion `json:"regions"`
	ReadError     string                `json:"readError,omitempty"`
}

type debugSnapshotSummary struct {
	Valid        bool                 `json:"valid"`
	ActiveGame   string               `json:"activeGame"`
	SaveIndex    uint32               `json:"saveIndex"`
	OotSceneID   uint16               `json:"ootSceneId"`
	MmDay        uint32               `json:"mmDay"`
	MmPlayerForm uint8                `json:"mmPlayerForm"`
	Items        []ootmm.TrackedItem  `json:"items,omitempty"`
	Checks       []ootmm.TrackedCheck `json:"checks,omitempty"`
}

type debugSnapshotRegion struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Size     int    `json:"size"`
	Encoding string `json:"encoding"`
	Data     string `json:"data"`
}

type memoryRegionSpec struct {
	name    string
	address uint32
	size    int
}

func startConsoleCommands() <-chan consoleCommand {
	commands := make(chan consoleCommand, 4)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			command := consoleCommand{name: strings.ToLower(parts[0])}
			if len(parts) > 1 {
				command.args = strings.TrimSpace(line[len(parts[0]):])
			}
			select {
			case commands <- command:
			default:
				log.Printf("Konsole: Befehlswarteschlange voll, verwerfe %q", line)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Konsole: Eingabefehler: %v", err)
		}
	}()
	return commands
}

func drainConsoleCommands(commands <-chan consoleCommand, connected, probed bool, mem *n64.Memory) {
	for {
		select {
		case command := <-commands:
			handleConsoleCommand(command, connected, probed, mem)
		default:
			return
		}
	}
}

func handleConsoleCommand(command consoleCommand, connected, probed bool, mem *n64.Memory) {
	switch command.name {
	case "help":
		printConsoleHelp()
	case "dump", "snapshot":
		if !connected || !probed {
			log.Printf("Snapshot nicht moeglich: RetroArch/OoTMM ist aktuell nicht verbunden")
			return
		}
		path, err := resolveSnapshotPath(command.args, time.Now())
		if err != nil {
			log.Printf("Snapshot-Pfad ungueltig: %v", err)
			return
		}
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			log.Printf("Snapshot-Pfad konnte nicht aufgeloest werden: %v", err)
			return
		}
		log.Printf("Schreibe Snapshot nach %s", absolutePath)
		if err := writeDebugSnapshot(absolutePath, mem); err != nil {
			log.Printf("Snapshot fehlgeschlagen: %v", err)
			return
		}
		log.Printf("Snapshot gespeichert: %s", absolutePath)
	default:
		log.Printf("Unbekannter Konsolenbefehl %q", command.name)
		printConsoleHelp()
	}
}

func printConsoleHelp() {
	log.Printf("Konsole: help | dump [label|pfad] | snapshot [label|pfad]")
	log.Printf("Beispiel: dump vor-sarias-song")
	log.Printf("Beispiel: dump memory-dumps/nach-check.json")
}

func resolveSnapshotPath(raw string, now time.Time) (string, error) {
	timestamp := now.Format("20060102-150405")
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return filepath.Join("memory-dumps", fmt.Sprintf("snapshot-%s.json", timestamp)), nil
	}

	if strings.Contains(trimmed, "..") {
		return "", fmt.Errorf("Pfad mit '..' wird nicht unterstuetzt")
	}

	if strings.ContainsRune(trimmed, os.PathSeparator) || filepath.Ext(trimmed) != "" {
		return trimmed, nil
	}

	label := sanitizeSnapshotLabel(trimmed)
	if label == "" {
		return "", fmt.Errorf("leerer Snapshot-Name")
	}
	return filepath.Join("memory-dumps", fmt.Sprintf("%s-%s.json", label, timestamp)), nil
}

func sanitizeSnapshotLabel(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return ""
	}

	var builder strings.Builder
	lastDash := false
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		case r == '-', r == '_', r == ' ', r == '.':
			if !lastDash && builder.Len() > 0 {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}

	return strings.Trim(builder.String(), "-")
}

func writeDebugSnapshot(path string, mem *n64.Memory) error {
	snapshot, err := captureDebugSnapshot(mem)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("snapshot JSON erstellen: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("Snapshot-Verzeichnis anlegen: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("Snapshot schreiben: %w", err)
	}
	return nil
}

func captureDebugSnapshot(mem *n64.Memory) (*debugSnapshot, error) {
	snapshot := &debugSnapshot{
		SchemaVersion: 1,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}

	reader := ootmm.NewReader(mem)
	var (
		state   *ootmm.GameState
		readErr error
	)
	for attempt := 0; attempt < 3; attempt++ {
		state, readErr = reader.ReadState()
		if readErr != nil {
			snapshot.ReadError = readErr.Error()
			break
		}
		if state == nil {
			continue
		}
		if state.Valid && state.ActiveGame != ootmm.GameNone {
			break
		}
	}

	if state != nil {
		snapshot.Summary = debugSnapshotSummary{
			Valid:        state.Valid,
			ActiveGame:   state.ActiveGame.String(),
			SaveIndex:    state.SaveIndex,
			OotSceneID:   state.Oot.SceneID,
			MmDay:        state.Mm.Day,
			MmPlayerForm: state.Mm.PlayerForm,
		}
		if state.Valid && state.ActiveGame != ootmm.GameNone {
			snapshot.Summary.Items = ootmm.ExtractItems(state)
			snapshot.Summary.Checks = ootmm.ExtractChecks(state)
		}
	}

	regions, err := readSnapshotRegions(mem)
	if err != nil {
		return nil, err
	}
	snapshot.Regions = regions
	return snapshot, nil
}

func readSnapshotRegions(mem *n64.Memory) ([]debugSnapshotRegion, error) {
	specs := []memoryRegionSpec{
		{name: "comboCtxOot", address: ootmm.AddrComboCtxOot, size: ootmm.ComboCtxSize},
		{name: "comboCtxMm", address: ootmm.AddrComboCtxMm, size: ootmm.ComboCtxSize},
		{name: "ootSaveContext", address: ootmm.AddrOotSaveCtx, size: ootmm.OotSaveCtxSize},
		{name: "mmSaveContext", address: ootmm.AddrMmSaveCtx, size: ootmm.MmSaveCtxSize},
		{name: "ootPayload", address: ootmm.AddrOotPayload, size: ootmm.OotPayloadSize},
		{name: "mmPayload", address: ootmm.AddrMmPayload, size: ootmm.MmPayloadSize},
	}

	regions := make([]debugSnapshotRegion, 0, len(specs))
	for _, spec := range specs {
		data, err := mem.Read(spec.address, spec.size)
		if err != nil {
			return nil, fmt.Errorf("Region %s lesen: %w", spec.name, err)
		}
		regions = append(regions, debugSnapshotRegion{
			Name:     spec.name,
			Address:  fmt.Sprintf("0x%08x", spec.address),
			Size:     spec.size,
			Encoding: "base64",
			Data:     base64.StdEncoding.EncodeToString(data),
		})
	}

	return regions, nil
}
