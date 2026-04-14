package main

import (
	"testing"
	"time"
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
