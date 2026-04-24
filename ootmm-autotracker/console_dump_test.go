package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"ootmm-autotracker/n64"
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

type limitedSnapshotCore struct {
	regions        []capturedSnapshotRegion
	readsRemaining int
	readCount      int
}

func (c *limitedSnapshotCore) ReadMemory(addr uint32, size int) ([]byte, error) {
	return c.ReadMemoryLarge(addr, size)
}

func (c *limitedSnapshotCore) ReadMemoryLarge(addr uint32, size int) ([]byte, error) {
	if c.readsRemaining <= 0 {
		return nil, fmt.Errorf("unexpected live-core read at %#x size %d", addr, size)
	}
	c.readsRemaining--
	c.readCount++
	for _, region := range c.regions {
		if addr < region.address {
			continue
		}
		offset := int(addr - region.address)
		if offset < 0 || offset+size > len(region.data) {
			continue
		}
		return region.data[offset : offset+size], nil
	}
	return nil, fmt.Errorf("missing region for %#x size %d", addr, size)
}

func loadCapturedSnapshotRegions(t *testing.T, name string) []capturedSnapshotRegion {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("memory-dumps", name))
	if err != nil {
		t.Fatalf("read snapshot fixture %s: %v", name, err)
	}

	var snapshot debugSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		t.Fatalf("parse snapshot fixture %s: %v", name, err)
	}

	regions := make([]capturedSnapshotRegion, 0, len(snapshot.Regions))
	for _, region := range snapshot.Regions {
		decoded, err := base64.StdEncoding.DecodeString(region.Data)
		if err != nil {
			t.Fatalf("decode region %s: %v", region.Name, err)
		}
		addr, err := strconv.ParseUint(strings.TrimPrefix(region.Address, "0x"), 16, 32)
		if err != nil {
			t.Fatalf("parse region address %s: %v", region.Address, err)
		}
		regions = append(regions, capturedSnapshotRegion{
			name:    region.Name,
			address: uint32(addr),
			size:    region.Size,
			data:    decoded,
		})
	}

	return regions
}

func hasSnapshotCheck(snapshot *debugSnapshot, name string) bool {
	for _, check := range snapshot.Summary.Checks {
		if check.Name == name && check.Checked {
			return true
		}
	}
	return false
}

func snapshotItemQty(snapshot *debugSnapshot, itemID string) int {
	for _, item := range snapshot.Summary.Items {
		if item.ID == itemID {
			return item.Qty
		}
	}
	return -1
}

func TestCaptureDebugSnapshotBuildsSummaryFromFrozenRegions(t *testing.T) {
	regions := loadCapturedSnapshotRegions(t, "in-mm-20260423-173418.json")
	core := &limitedSnapshotCore{
		regions:        regions,
		readsRemaining: len(snapshotRegionSpecs()),
	}
	mem := n64.NewMemory(core)
	mem.SetBaseShift(n64.VirtualBase)
	mem.SetSwizzle(false)

	snapshot, err := captureDebugSnapshot(mem)
	if err != nil {
		t.Fatalf("captureDebugSnapshot: %v", err)
	}
	if core.readCount != len(snapshotRegionSpecs()) {
		t.Fatalf("live core read count = %d, want %d", core.readCount, len(snapshotRegionSpecs()))
	}
	if snapshot.ReadError != "" {
		t.Fatalf("unexpected snapshot read error: %s", snapshot.ReadError)
	}
	if snapshot.Summary.ActiveGame != "MM" {
		t.Fatalf("summary active game = %q, want MM", snapshot.Summary.ActiveGame)
	}
	if snapshot.Summary.MmDay != 1 {
		t.Fatalf("summary mmDay = %d, want 1", snapshot.Summary.MmDay)
	}
	if snapshot.Summary.MmPlayerForm != 4 {
		t.Fatalf("summary mmPlayerForm = %d, want 4", snapshot.Summary.MmPlayerForm)
	}
	if got := snapshotItemQty(snapshot, "MM_STRAY_FAIRY_TOWN"); got != 1 {
		t.Fatalf("MM_STRAY_FAIRY_TOWN = %d, want 1", got)
	}
	for _, check := range []string{
		"Road to Southern Swamp HP",
		"Clock Town Tree HP",
		"Clock Town Platform HP",
	} {
		if !hasSnapshotCheck(snapshot, check) {
			t.Fatalf("missing snapshot check %q", check)
		}
	}
}
