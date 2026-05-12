package main

import (
	"testing"
	"time"

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
