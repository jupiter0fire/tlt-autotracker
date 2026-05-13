package main

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
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
	SchemaVersion     int                            `json:"schemaVersion"`
	CreatedAt         string                         `json:"createdAt"`
	Summary           debugSnapshotSummary           `json:"summary"`
	MemoryBlocks      []debugSnapshotMemoryBlock     `json:"memoryBlocks"`
	ResolvedAddresses []debugSnapshotResolvedAddress `json:"resolvedAddresses,omitempty"`
	Regions           []debugSnapshotRegion          `json:"regions"`
	ReadError         string                         `json:"readError,omitempty"`
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

type debugSnapshotResolvedAddress struct {
	LogicalID string `json:"logical_id"`
	Address   string `json:"address,omitempty"`
	SizeBytes int    `json:"size_bytes,omitempty"`
	Selection string `json:"selection"`
	Meaning   string `json:"meaning"`
}

type debugSnapshotMemoryBlock struct {
	LogicalID                string `json:"logical_id"`
	DumpRegion               string `json:"dump_region"`
	BaseKind                 string `json:"base_kind"`
	RegionOffset             string `json:"region_offset"`
	AbsoluteAddressOrFormula string `json:"absolute_address_or_formula"`
	SizeBytes                int    `json:"size_bytes"`
	Layout                   string `json:"layout"`
	AuthoritativeWhen        string `json:"authoritative_when"`
	Meaning                  string `json:"meaning"`
}

type snapshotMemoryBlockSpec struct {
	LogicalID string
	Offset    int
	SizeBytes int
	Layout    string
	Meaning   string
}

const debugSnapshotSchemaVersion = 4

type memoryRegionSpec struct {
	name    string
	address uint32
	size    int
}

type capturedSnapshotRegion struct {
	name    string
	address uint32
	size    int
	data    []byte
}

type snapshotExportRegionSpec struct {
	name    string
	address uint32
	size    int
}

type snapshotCoreReader struct {
	regions []capturedSnapshotRegion
}

func (s *snapshotCoreReader) ReadMemory(addr uint32, size int) ([]byte, error) {
	return s.ReadMemoryLarge(addr, size)
}

func (s *snapshotCoreReader) ReadMemoryLarge(addr uint32, size int) ([]byte, error) {
	for _, region := range s.regions {
		if addr < region.address {
			continue
		}
		offset := int(addr - region.address)
		if offset < 0 || offset+size > len(region.data) {
			continue
		}
		return region.data[offset : offset+size], nil
	}
	return nil, fmt.Errorf("addr %#x size %d not found", addr, size)
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
				log.Printf("Console: command queue full, dropping %q", line)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Console: input error: %v", err)
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
			log.Printf("Snapshot not possible: RetroArch/OoTMM is not currently connected")
			return
		}
		path, err := resolveSnapshotPath(command.args, time.Now())
		if err != nil {
			log.Printf("Snapshot path invalid: %v", err)
			return
		}
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			log.Printf("Snapshot path could not be resolved: %v", err)
			return
		}
		log.Printf("Writing snapshot to %s", absolutePath)
		if err := writeDebugSnapshot(absolutePath, mem); err != nil {
			log.Printf("Snapshot failed: %v", err)
			return
		}
		log.Printf("Snapshot saved: %s", absolutePath)
	default:
		log.Printf("Unknown console command %q", command.name)
		printConsoleHelp()
	}
}

func printConsoleHelp() {
	log.Printf("Console: help | dump [label|path] | snapshot [label|path]")
	log.Printf("Example: dump vor-sarias-song")
	log.Printf("Example: dump memory-dumps/nach-check.json")
}

func resolveSnapshotPath(raw string, now time.Time) (string, error) {
	timestamp := now.Format("20060102-150405")
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return filepath.Join("memory-dumps", fmt.Sprintf("snapshot-%s.json", timestamp)), nil
	}

	if strings.Contains(trimmed, "..") {
		return "", fmt.Errorf("path containing '..' is not supported")
	}

	if strings.ContainsRune(trimmed, os.PathSeparator) || filepath.Ext(trimmed) != "" {
		return trimmed, nil
	}

	label := sanitizeSnapshotLabel(trimmed)
	if label == "" {
		return "", fmt.Errorf("empty snapshot name")
	}
	return filepath.Join("memory-dumps", fmt.Sprintf("%s-%s.json", label, timestamp)), nil
}

func resolveAutomaticSnapshotPath(now time.Time) string {
	timestamp := now.Format("20060102-150405")
	return filepath.Join("memory-dumps", fmt.Sprintf("auto-snapshot-%s.json", timestamp))
}

func writeAutomaticSnapshot(mem *n64.Memory, now time.Time) error {
	absolutePath, err := filepath.Abs(resolveAutomaticSnapshotPath(now))
	if err != nil {
		return fmt.Errorf("automatic snapshot path could not be resolved: %w", err)
	}

	log.Printf("Writing automatic snapshot to %s", absolutePath)
	if err := writeDebugSnapshot(absolutePath, mem); err != nil {
		return err
	}
	log.Printf("Automatic snapshot saved: %s", absolutePath)
	return nil
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
		return fmt.Errorf("creating snapshot JSON: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating snapshot directory: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("writing snapshot: %w", err)
	}
	return nil
}

func captureDebugSnapshot(mem *n64.Memory) (*debugSnapshot, error) {
	snapshot := &debugSnapshot{
		SchemaVersion: debugSnapshotSchemaVersion,
		CreatedAt:     time.Now().Format(time.RFC3339),
		MemoryBlocks:  snapshotMemoryBlocks(),
	}

	coreRegions, err := captureSnapshotCoreRegions(mem)
	if err != nil {
		return nil, err
	}

	readerMem := n64.NewMemory(&snapshotCoreReader{regions: append([]capturedSnapshotRegion(nil), coreRegions...)})
	readerMem.SetBaseShift(n64.VirtualBase)
	readerMem.SetSwizzle(false)
	reader := ootmm.NewReader(readerMem)
	var (
		state    *ootmm.GameState
		readErr  error
		resolved ootmm.DebugResolvedAddresses
	)
	for attempt := 0; attempt < 3; attempt++ {
		state, readErr = reader.ReadState()
		resolved = reader.DebugResolvedAddresses()
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
		reader.StabilizeSnapshotState(state)
		snapshot.Summary = debugSnapshotSummary{
			Valid:        state.Valid,
			ActiveGame:   state.ActiveGame.String(),
			SaveIndex:    state.SaveIndex,
			OotSceneID:   state.Oot.SceneID,
			MmDay:        state.Mm.Day,
			MmPlayerForm: state.Mm.PlayerForm,
		}
		if state.Valid && state.ActiveGame != ootmm.GameNone {
			snapshot.Summary.Items = filterSnapshotItems(ootmm.ExtractItems(state))
			snapshot.Summary.Checks = ootmm.ExtractChecks(state)
		}
	}
	snapshot.ResolvedAddresses = snapshotResolvedAddresses(resolved)
	snapshot.Regions = encodeSnapshotRegions(selectSnapshotExportRegions(coreRegions, resolved))
	return snapshot, nil
}

func filterSnapshotItems(items []ootmm.TrackedItem) []ootmm.TrackedItem {
	filtered := make([]ootmm.TrackedItem, 0, len(items))
	for _, item := range items {
		if item.Qty == 0 {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func snapshotCoreRegionSpecs() []memoryRegionSpec {
	return []memoryRegionSpec{
		{name: "comboCtxOot", address: ootmm.AddrComboCtxOot, size: ootmm.ComboCtxSize},
		{name: "comboCtxMm", address: ootmm.AddrComboCtxMm, size: ootmm.ComboCtxSize},
		{name: "ootSaveContext", address: ootmm.AddrOotSaveCtx, size: ootmm.OotSaveCtxSize},
		{name: "mmSaveContext", address: ootmm.AddrMmSaveCtx, size: ootmm.MmSaveCtxSize},
		{name: "ootPayload", address: ootmm.AddrOotPayload, size: ootmm.OotPayloadSize},
		{name: "mmPayload", address: ootmm.AddrMmPayload, size: ootmm.MmPayloadSize},
	}
}

func captureSnapshotCoreRegions(mem *n64.Memory) ([]capturedSnapshotRegion, error) {
	regions := make([]capturedSnapshotRegion, 0, len(snapshotCoreRegionSpecs())+1)
	for _, spec := range snapshotCoreRegionSpecs() {
		data, err := mem.Read(spec.address, spec.size)
		if err != nil {
			return nil, fmt.Errorf("reading region %s: %w", spec.name, err)
		}
		regions = append(regions, capturedSnapshotRegion{
			name:    spec.name,
			address: spec.address,
			size:    spec.size,
			data:    data,
		})
	}

	// MM detection reads this runtime marker directly, so freeze it alongside the
	// captured regions to keep the derived summary consistent with the snapshot.
	runtimeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(runtimeBuf, 1)
	regions = append(regions, capturedSnapshotRegion{
		name:    "mmRuntimeMarker",
		address: 0x801F3F60,
		size:    len(runtimeBuf),
		data:    runtimeBuf,
	})

	return regions, nil
}

func encodeSnapshotRegions(captured []capturedSnapshotRegion) []debugSnapshotRegion {
	regions := make([]debugSnapshotRegion, 0, len(captured))
	for _, region := range captured {
		if region.name == "mmRuntimeMarker" {
			continue
		}
		regions = append(regions, debugSnapshotRegion{
			Name:     region.name,
			Address:  fmt.Sprintf("0x%08x", region.address),
			Size:     region.size,
			Encoding: "base64",
			Data:     base64.StdEncoding.EncodeToString(region.data),
		})
	}
	return regions
}

func snapshotMemoryBlocks() []debugSnapshotMemoryBlock {
	blocks := make([]debugSnapshotMemoryBlock, 0, 32)

	sharedLayout := ootmm.SharedStorageSnapshotLayout()
	blocks = appendSharedSnapshotBlocks(
		blocks,
		"shared.live_near_foreign_oot_payload.",
		"ootPayload",
		"live-shared-near-foreign",
		"(<located foreign MM save offset in ootPayload>) - "+hexUint32(ootmm.SharedCustomSaveSize),
		"(<located foreign MM save absolute address>) - "+hexUint32(ootmm.SharedCustomSaveSize),
		"Authoritative live SharedCustomSave when summary.activeGame = \"OoT\". Locate the validated foreign MM save in ootPayload, then step back 0x870 bytes.",
		sharedLayout,
	)
	blocks = appendSharedSnapshotBlocks(
		blocks,
		"shared.live_near_foreign_mm_payload.",
		"mmPayload",
		"live-shared-near-foreign",
		"(<located foreign OoT save offset in mmPayload>) - "+hexUint32(ootmm.SharedCustomSaveSize),
		"(<located foreign OoT save absolute address>) - "+hexUint32(ootmm.SharedCustomSaveSize),
		"Authoritative live SharedCustomSave when summary.activeGame = \"MM\". Locate the validated foreign OoT save in mmPayload, then step back 0x870 bytes.",
		sharedLayout,
	)

	return blocks
}

func appendFormulaSnapshotBlocks(blocks []debugSnapshotMemoryBlock, dumpRegion string, baseKind string, regionBase string, absoluteBase string, authoritativeWhen string, specs []snapshotMemoryBlockSpec) []debugSnapshotMemoryBlock {
	for _, spec := range specs {
		blocks = append(blocks, debugSnapshotMemoryBlock{
			LogicalID:                spec.LogicalID,
			DumpRegion:               dumpRegion,
			BaseKind:                 baseKind,
			RegionOffset:             formulaWithOffset(regionBase, spec.Offset),
			AbsoluteAddressOrFormula: formulaWithOffset(absoluteBase, spec.Offset),
			SizeBytes:                spec.SizeBytes,
			Layout:                   spec.Layout,
			AuthoritativeWhen:        authoritativeWhen,
			Meaning:                  spec.Meaning,
		})
	}
	return blocks
}

func appendSharedSnapshotBlocks(blocks []debugSnapshotMemoryBlock, logicalPrefix string, dumpRegion string, baseKind string, regionBase string, absoluteBase string, authoritativeWhen string, layout ootmm.SnapshotSharedStorageLayout) []debugSnapshotMemoryBlock {
	sharedSpecs := make([]snapshotMemoryBlockSpec, 0, len(layout.Bitmaps)+5)
	for _, bitmap := range layout.Bitmaps {
		sharedSpecs = append(sharedSpecs, snapshotMemoryBlockSpec{
			LogicalID: logicalPrefix + bitmap.Name,
			Offset:    bitmap.Offset,
			SizeBytes: bitmap.Size,
			Layout:    fmt.Sprintf("bitmap, %d bytes (%d bits)", bitmap.Size, bitmap.Size*8),
			Meaning:   sharedBitmapMeaning(bitmap.Name),
		})
	}
	sharedSpecs = append(sharedSpecs,
		snapshotMemoryBlockSpec{LogicalID: logicalPrefix + "ocarinaButtonMaskOot", Offset: ootmm.SharedOcarinaButtonMaskOotOffset, SizeBytes: 2, Layout: "u16 bitmask", Meaning: "Shared OoT ocarina button ownership mask."},
		snapshotMemoryBlockSpec{LogicalID: logicalPrefix + "ocarinaButtonMaskMm", Offset: ootmm.SharedOcarinaButtonMaskMmOffset, SizeBytes: 2, Layout: "u16 bitmask", Meaning: "Shared MM ocarina button ownership mask."},
		snapshotMemoryBlockSpec{LogicalID: logicalPrefix + "caughtChildFishWeights", Offset: ootmm.SharedCaughtChildFishWeightOffset, SizeBytes: ootmm.SharedCaughtFishWeightCount, Layout: "20 x u8 stack bytes", Meaning: "Child fishing record stack; byte 0 is count, later bytes are stored fish weights."},
		snapshotMemoryBlockSpec{LogicalID: logicalPrefix + "caughtAdultFishWeights", Offset: ootmm.SharedCaughtAdultFishWeightOffset, SizeBytes: ootmm.SharedCaughtFishWeightCount, Layout: "20 x u8 stack bytes", Meaning: "Adult fishing record stack; byte 0 is count, later bytes are stored fish weights."},
		snapshotMemoryBlockSpec{LogicalID: logicalPrefix + "bombchuBagFlags", Offset: ootmm.SharedBombchuBagFlagsOffset, SizeBytes: 1, Layout: "packed flag byte", Meaning: "Shared Bombchu Bag progression byte containing both OoT and MM 2-bit bag levels."},
	)

	return appendFormulaSnapshotBlocks(blocks, dumpRegion, baseKind, regionBase, absoluteBase, authoritativeWhen, sharedSpecs)
}

func sharedBitmapMeaning(name string) string {
	switch name {
	case "xflagsOot":
		return "OoT randomized XFlag bitmap for custom check locations."
	case "npcOot":
		return "OoT shared NPC and reward bitmap."
	case "shopsOot":
		return "OoT shop purchase bitmap."
	case "scrubsOot":
		return "OoT Deku Scrub reward bitmap."
	case "srOot":
		return "OoT silver-rupee reward bitmap."
	case "xflagsMm":
		return "MM randomized XFlag bitmap for custom check locations."
	case "npcMm":
		return "MM shared NPC and reward bitmap."
	case "shopsMm":
		return "MM shop purchase bitmap."
	case "soulsEnemyOot":
		return "OoT enemy soul ownership bitmap."
	case "soulsEnemyMm":
		return "MM enemy soul ownership bitmap."
	case "soulsBossOot":
		return "OoT boss soul ownership bitmap."
	case "soulsBossMm":
		return "MM boss soul ownership bitmap."
	case "soulsNpcOot":
		return "OoT NPC soul ownership bitmap."
	case "soulsNpcMm":
		return "MM NPC soul ownership bitmap."
	case "soulsAnimalOot":
		return "OoT animal soul ownership bitmap."
	case "soulsAnimalMm":
		return "MM animal soul ownership bitmap."
	case "soulsMiscOot":
		return "OoT miscellaneous soul ownership bitmap."
	case "soulsMiscMm":
		return "MM miscellaneous soul ownership bitmap."
	default:
		return "Shared bitmap stored in SharedCustomSave."
	}
}

func formulaWithOffset(base string, offset int) string {
	if offset == 0 {
		return base
	}
	return base + " + " + hexInt(offset)
}

func hexUint32(value uint32) string {
	return fmt.Sprintf("0x%X", value)
}

func hexInt(value int) string {
	return fmt.Sprintf("0x%X", value)
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
			return nil, fmt.Errorf("reading region %s: %w", spec.name, err)
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

func snapshotSharedStateSize() int {
	size := int(ootmm.SharedCustomSaveSize)
	layout := ootmm.SharedStorageSnapshotLayout()
	if size < layout.TrackedSize {
		size = layout.TrackedSize
	}
	if size < ootmm.SharedBombchuBagFlagsOffset+1 {
		size = ootmm.SharedBombchuBagFlagsOffset + 1
	}
	return size
}

func snapshotResolvedAddresses(resolved ootmm.DebugResolvedAddresses) []debugSnapshotResolvedAddress {
	addresses := make([]debugSnapshotResolvedAddress, 0, 12)
	sharedSize := snapshotSharedStateSize()
	appendAddress := func(logicalID string, address uint32, size int, selection string, meaning string) {
		if address == 0 {
			return
		}
		entry := debugSnapshotResolvedAddress{
			LogicalID: logicalID,
			Address:   fmt.Sprintf("0x%08x", address),
			Selection: selection,
			Meaning:   meaning,
		}
		if size > 0 {
			entry.SizeBytes = size
		}
		addresses = append(addresses, entry)
	}

	if resolved.ForeignMmSaveAddr >= ootmm.SharedCustomSaveSize {
		appendAddress("shared.live_near_foreign_oot_payload", resolved.ForeignMmSaveAddr-ootmm.SharedCustomSaveSize, sharedSize, "adjacent-to-selected-foreign-save", "Live SharedCustomSave window immediately before the selected foreign MM save inside the OoT payload.")
	}
	appendAddress("mm.foreign_save_in_oot_payload", resolved.ForeignMmSaveAddr, ootmm.MmSaveSize, "payload-scan-selected", "Selected foreign MM save candidate inside the OoT payload.")

	if resolved.ForeignOotSaveAddr >= ootmm.SharedCustomSaveSize {
		appendAddress("shared.live_near_foreign_mm_payload", resolved.ForeignOotSaveAddr-ootmm.SharedCustomSaveSize, sharedSize, "adjacent-to-selected-foreign-save", "Live SharedCustomSave window immediately before the selected foreign OoT save inside the MM payload.")
	}
	appendAddress("oot.foreign_save_in_mm_payload", resolved.ForeignOotSaveAddr, ootmm.OotSaveSize, "payload-scan-selected", "Selected foreign OoT save candidate inside the MM payload.")

	appendAddress("oot.runtime.silver_rupee_data", resolved.OotSilverDataAddr, ootmm.OotSilverRupeeDataSize, "payload-scan-selected", "Located OoT runtime silver-rupee metadata block.")
	appendAddress("oot.runtime.max_keys", resolved.OotMaxKeysAddr, ootmm.OotMaxKeysBlockSize, "payload-scan-selected", "Located OoT runtime max-small-keys block.")
	appendAddress("oot.runtime.combo_config", resolved.ComboConfigOotAddr, ootmm.OotComboConfigSize, "payload-scan-selected", "Located OoT runtime combo config block in the active payload.")
	appendAddress("mm.runtime.combo_config", resolved.ComboConfigMmAddr, ootmm.OotComboConfigSize, "payload-scan-selected", "Located OoT runtime combo config block while MM is active.")
	appendAddress("oot.live.play_state", resolved.OotPlayStateAddr, 0, "candidate-probe-selected", "Selected OoT PlayState base; the reader samples non-contiguous live fields from this struct.")
	appendAddress("mm.live.play_state", resolved.MmPlayStateAddr, 0, "candidate-probe-selected", "Selected MM PlayState base; the reader samples non-contiguous live fields from this struct.")

	return addresses
}

func selectSnapshotExportRegions(coreRegions []capturedSnapshotRegion, resolved ootmm.DebugResolvedAddresses) []capturedSnapshotRegion {
	specs := make([]snapshotExportRegionSpec, 0, 8)
	sharedSize := snapshotSharedStateSize()

	if resolved.ForeignMmSaveAddr >= ootmm.SharedCustomSaveSize {
		specs = append(specs, snapshotExportRegionSpec{
			name:    "ootPayload.liveSharedNearForeignMm",
			address: resolved.ForeignMmSaveAddr - ootmm.SharedCustomSaveSize,
			size:    sharedSize,
		})
	}
	if resolved.ForeignMmSaveAddr != 0 {
		specs = append(specs, snapshotExportRegionSpec{
			name:    "ootPayload.foreignMmSave",
			address: resolved.ForeignMmSaveAddr,
			size:    ootmm.MmSaveSize,
		})
	}
	if resolved.ForeignOotSaveAddr >= ootmm.SharedCustomSaveSize {
		specs = append(specs, snapshotExportRegionSpec{
			name:    "mmPayload.liveSharedNearForeignOot",
			address: resolved.ForeignOotSaveAddr - ootmm.SharedCustomSaveSize,
			size:    sharedSize,
		})
	}
	if resolved.ForeignOotSaveAddr != 0 {
		specs = append(specs, snapshotExportRegionSpec{
			name:    "mmPayload.foreignOotSave",
			address: resolved.ForeignOotSaveAddr,
			size:    ootmm.OotSaveSize,
		})
	}
	if resolved.OotSilverDataAddr != 0 {
		specs = append(specs, snapshotExportRegionSpec{
			name:    "ootPayload.runtimeSilverRupeeData",
			address: resolved.OotSilverDataAddr,
			size:    ootmm.OotSilverRupeeDataSize,
		})
	}
	if resolved.OotMaxKeysAddr != 0 {
		specs = append(specs, snapshotExportRegionSpec{
			name:    "ootPayload.runtimeMaxKeys",
			address: resolved.OotMaxKeysAddr,
			size:    ootmm.OotMaxKeysBlockSize,
		})
	}
	if resolved.ComboConfigOotAddr != 0 {
		specs = append(specs, snapshotExportRegionSpec{
			name:    "ootPayload.runtimeOotComboConfig",
			address: resolved.ComboConfigOotAddr,
			size:    ootmm.OotComboConfigSize,
		})
	}
	if resolved.ComboConfigMmAddr != 0 {
		specs = append(specs, snapshotExportRegionSpec{
			name:    "mmPayload.runtimeOotComboConfig",
			address: resolved.ComboConfigMmAddr,
			size:    ootmm.OotComboConfigSize,
		})
	}

	exports := make([]capturedSnapshotRegion, 0, len(specs))
	for _, spec := range specs {
		if region, ok := sliceCapturedSnapshotRegion(coreRegions, spec.name, spec.address, spec.size); ok {
			exports = append(exports, region)
		}
	}
	return exports
}

func sliceCapturedSnapshotRegion(regions []capturedSnapshotRegion, name string, address uint32, size int) (capturedSnapshotRegion, bool) {
	if size <= 0 {
		return capturedSnapshotRegion{}, false
	}
	for _, region := range regions {
		if address < region.address {
			continue
		}
		offset := int(address - region.address)
		if offset < 0 || offset+size > len(region.data) {
			continue
		}
		data := append([]byte(nil), region.data[offset:offset+size]...)
		return capturedSnapshotRegion{
			name:    name,
			address: address,
			size:    size,
			data:    data,
		}, true
	}
	return capturedSnapshotRegion{}, false
}
