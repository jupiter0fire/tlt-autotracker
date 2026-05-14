package main

import (
	"strings"
	"testing"
	"time"

	"ootmm-autotracker/n64"
	"ootmm-autotracker/ootmm"
)

func TestResolveSnapshotPathDefault(t *testing.T) {
	path, err := resolveSnapshotPath("", time.Date(2026, time.April, 14, 12, 34, 56, 0, time.UTC))
	if err != nil {
		t.Fatalf("resolveSnapshotPath returned error: %v", err)
	}
	if path != "memory-dumps/snapshot-20260414-123456.json" {
		t.Fatalf("unexpected default path: %q", path)
	}
}

func TestResolveSnapshotPathFromLabel(t *testing.T) {
	path, err := resolveSnapshotPath("Vor Sarias Song", time.Date(2026, time.April, 14, 12, 34, 56, 0, time.UTC))
	if err != nil {
		t.Fatalf("resolveSnapshotPath returned error: %v", err)
	}
	if path != "memory-dumps/vor-sarias-song-20260414-123456.json" {
		t.Fatalf("unexpected label path: %q", path)
	}
}

func TestResolveSnapshotPathKeepsExplicitFile(t *testing.T) {
	path, err := resolveSnapshotPath("memory-dumps/custom.json", time.Date(2026, time.April, 14, 12, 34, 56, 0, time.UTC))
	if err != nil {
		t.Fatalf("resolveSnapshotPath returned error: %v", err)
	}
	if path != "memory-dumps/custom.json" {
		t.Fatalf("unexpected explicit path: %q", path)
	}
}

func TestResolveAutomaticSnapshotPath(t *testing.T) {
	path := resolveAutomaticSnapshotPath(time.Date(2026, time.April, 14, 12, 34, 56, 0, time.UTC))
	if path != "memory-dumps/auto-snapshot-20260414-123456.json" {
		t.Fatalf("unexpected automatic snapshot path: %q", path)
	}
}

type zeroSnapshotCoreReader struct{}

func (zeroSnapshotCoreReader) ReadMemory(addr uint32, size int) ([]byte, error) {
	return make([]byte, size), nil
}

func (zeroSnapshotCoreReader) ReadMemoryLarge(addr uint32, size int) ([]byte, error) {
	return make([]byte, size), nil
}

func TestFilterSnapshotItemsOmitsZeroQty(t *testing.T) {
	items := []ootmm.TrackedItem{
		{ID: "OOT_BOW", Qty: 0},
		{ID: "OOT_HOOKSHOT", Qty: 1},
		{ID: "MM_BOMB", Qty: 3},
		{ID: "MM_ARROW", Qty: 0},
	}

	filtered := filterSnapshotItems(items)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 non-zero items, got %d", len(filtered))
	}
	if filtered[0].ID != "OOT_HOOKSHOT" || filtered[0].Qty != 1 {
		t.Fatalf("unexpected first filtered item: %+v", filtered[0])
	}
	if filtered[1].ID != "MM_BOMB" || filtered[1].Qty != 3 {
		t.Fatalf("unexpected second filtered item: %+v", filtered[1])
	}
}

func TestSnapshotMemoryBlocksOnlyIncludeVariableBlocks(t *testing.T) {
	blocks := snapshotMemoryBlocks()
	if len(blocks) == 0 {
		t.Fatal("expected memory blocks to be present")
	}

	var sharedLive *debugSnapshotMemoryBlock
	for index := range blocks {
		block := &blocks[index]
		if block.BaseKind == "active-save-context" || block.BaseKind == "inactive-foreign-save-copy" || block.BaseKind == "payload-shared-slot-copy" {
			t.Fatalf("unexpected static block kind %q for %s", block.BaseKind, block.LogicalID)
		}
		switch block.LogicalID {
		case "shared.live_near_foreign_mm_payload.xflagsMm":
			sharedLive = block
		}
	}

	if len(blocks) < 2 {
		t.Fatalf("expected variable block metadata for both live shared variants, got %d entries", len(blocks))
	}

	if sharedLive == nil {
		t.Fatal("missing shared live near-foreign MM payload block")
	}
	if sharedLive.BaseKind != "live-shared-near-foreign" {
		t.Fatalf("unexpected live shared base kind: %q", sharedLive.BaseKind)
	}
	if !strings.Contains(sharedLive.RegionOffset, "located foreign OoT save offset") {
		t.Fatalf("unexpected live shared region formula: %q", sharedLive.RegionOffset)
	}
}

func TestSnapshotResolvedAddressesDescribeDynamicCandidates(t *testing.T) {
	resolved := ootmm.DebugResolvedAddresses{
		ForeignMmSaveAddr:  ootmm.AddrOotPayload + 0x6000,
		ForeignOotSaveAddr: ootmm.AddrMmPayload + 0x5000,
		OotSilverDataAddr:  ootmm.AddrOotPayload + 0x1200,
		OotMaxKeysAddr:     ootmm.AddrOotPayload + 0x2200,
		ComboConfigOotAddr: ootmm.AddrOotPayload + 0x3200,
		OotPlayStateAddr:   ootmm.AddrOotPlayStateNtsc10,
	}

	addresses := snapshotResolvedAddresses(resolved)
	if len(addresses) == 0 {
		t.Fatal("expected resolved addresses to be present")
	}

	byID := make(map[string]debugSnapshotResolvedAddress, len(addresses))
	for _, entry := range addresses {
		byID[entry.LogicalID] = entry
	}

	if _, ok := byID["shared.slot_copy_oot_payload"]; ok {
		t.Fatal("fixed shared slot copy address should not be exported")
	}
	if got := byID["mm.foreign_save_in_oot_payload"].Address; got != "0x80406000" {
		t.Fatalf("unexpected foreign MM address: %q", got)
	}
	if got := byID["oot.foreign_save_in_mm_payload"].Address; got != "0x80735000" {
		t.Fatalf("unexpected foreign OoT address: %q", got)
	}
	if got := byID["oot.live.play_state"].Selection; got != "candidate-probe-selected" {
		t.Fatalf("unexpected OoT play state selection: %q", got)
	}
	if _, ok := byID["oot.active.scene_flags"]; ok {
		t.Fatal("fixed save-context blocks should not appear in resolved addresses")
	}
}

func TestSelectSnapshotExportRegionsKeepsCapturedCoreRegions(t *testing.T) {
	coreRegions := []capturedSnapshotRegion{
		{name: "ootSaveContext", address: ootmm.AddrOotSaveCtx, size: 0x100, data: make([]byte, 0x100)},
		{name: "mmSaveContext", address: ootmm.AddrMmSaveCtx, size: 0x100, data: make([]byte, 0x100)},
		{name: "ootPayload", address: ootmm.AddrOotPayload, size: 0x9000, data: make([]byte, 0x9000)},
		{name: "mmPayload", address: ootmm.AddrMmPayload, size: 0x9000, data: make([]byte, 0x9000)},
	}
	resolved := ootmm.DebugResolvedAddresses{
		ForeignMmSaveAddr:  ootmm.AddrOotPayload + 0x4000,
		ForeignOotSaveAddr: ootmm.AddrMmPayload + 0x5000,
		OotSilverDataAddr:  ootmm.AddrOotPayload + 0x1200,
		ComboConfigMmAddr:  ootmm.AddrMmPayload + 0x1800,
	}

	regions := selectSnapshotExportRegions(coreRegions, resolved)
	if len(regions) != len(coreRegions) {
		t.Fatalf("expected %d exported regions, got %d", len(coreRegions), len(regions))
	}

	names := make(map[string]bool, len(regions))
	for _, region := range regions {
		names[region.name] = true
	}

	for _, required := range []string{
		"ootSaveContext",
		"mmSaveContext",
		"ootPayload",
		"mmPayload",
	} {
		if !names[required] {
			t.Fatalf("missing exported region %q", required)
		}
	}
	if len(regions[0].data) == 0 {
		t.Fatal("expected exported region data to be copied")
	}
}

func TestSnapshotCandidateRegionAddressesExposeAddressesWithoutData(t *testing.T) {
	resolved := ootmm.DebugResolvedAddresses{
		ForeignMmSaveAddr:  ootmm.AddrOotPayload + 0x4000,
		ForeignOotSaveAddr: ootmm.AddrMmPayload + 0x5000,
		OotSilverDataAddr:  ootmm.AddrOotPayload + 0x1200,
		OotMaxKeysAddr:     ootmm.AddrOotPayload + 0x2200,
		ComboConfigMmAddr:  ootmm.AddrMmPayload + 0x1800,
	}

	regions := snapshotCandidateRegionAddresses(resolved)
	if len(regions) == 0 {
		t.Fatal("expected candidate region addresses to be present")
	}

	byName := make(map[string]debugSnapshotRegion, len(regions))
	for _, region := range regions {
		byName[region.Name] = region
	}

	if got := byName["mmPayload.liveSharedNearForeignOot"].Address; got != "0x80734790" {
		t.Fatalf("unexpected live shared near-foreign OoT address: %q", got)
	}
	if got := byName["ootPayload.foreignMmSave"].Address; got != "0x80404000" {
		t.Fatalf("unexpected foreign MM address: %q", got)
	}
	if byName["mmPayload.liveSharedNearForeignOot"].Data != "" {
		t.Fatal("candidate region addresses should not include raw data")
	}
	if byName["mmPayload.liveSharedNearForeignOot"].Encoding != "" {
		t.Fatal("candidate region addresses should not include an encoding")
	}
}

func TestCaptureDebugSnapshotExportsCoreRegionsAndRuntimeMarker(t *testing.T) {
	mem := n64.NewMemory(zeroSnapshotCoreReader{})
	mem.SetBaseShift(n64.VirtualBase)
	mem.SetSwizzle(false)

	snapshot, err := captureDebugSnapshot(mem, ootmm.NewReader(mem))
	if err != nil {
		t.Fatalf("captureDebugSnapshot returned error: %v", err)
	}

	names := make(map[string]bool, len(snapshot.Regions))
	var candidate debugSnapshotRegion
	for _, region := range snapshot.Regions {
		names[region.Name] = true
		if region.Name == "mmPayload.liveSharedNearForeignOot" {
			candidate = region
		}
	}

	for _, required := range []string{"comboCtxOot", "comboCtxMm", "ootSaveContext", "mmSaveContext", "ootPayload", "mmPayload", "mmRuntimeMarker"} {
		if !names[required] {
			t.Fatalf("missing snapshot region %q", required)
		}
	}
	if candidate.Data != "" || candidate.Encoding != "" {
		t.Fatal("candidate region metadata should omit raw data")
	}
}
