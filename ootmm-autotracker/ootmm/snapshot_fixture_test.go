package ootmm

import (
	"encoding/binary"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"ootmm-autotracker/n64"
)

type debugSnapshotFixture struct {
	Regions []debugSnapshotFixtureRegion `json:"regions"`
}

type debugSnapshotFixtureRegion struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Size     int    `json:"size"`
	Encoding string `json:"encoding"`
	Data     string `json:"data"`
}

type snapshotFixtureRegion struct {
	address uint32
	data    []byte
}

type snapshotFixtureCoreReader struct {
	regions []snapshotFixtureRegion
}

func (s *snapshotFixtureCoreReader) ReadMemory(addr uint32, size int) ([]byte, error) {
	return s.ReadMemoryLarge(addr, size)
}

func (s *snapshotFixtureCoreReader) ReadMemoryLarge(addr uint32, size int) ([]byte, error) {
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

func loadSnapshotFixtureState(t *testing.T, name string) *GameState {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("testdata", "dumps", name))
	if err != nil {
		t.Fatalf("read snapshot fixture %s: %v", name, err)
	}

	var snapshot debugSnapshotFixture
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatalf("decode snapshot fixture %s: %v", name, err)
	}

	regions := make([]snapshotFixtureRegion, 0, len(snapshot.Regions))
	for _, region := range snapshot.Regions {
		if region.Encoding != "base64" {
			t.Fatalf("snapshot fixture region %s uses unsupported encoding %q", region.Name, region.Encoding)
		}
		addr, err := strconv.ParseUint(region.Address, 0, 32)
		if err != nil {
			t.Fatalf("parse snapshot fixture address %q: %v", region.Address, err)
		}
		data, err := base64.StdEncoding.DecodeString(region.Data)
		if err != nil {
			t.Fatalf("decode snapshot fixture region %s: %v", region.Name, err)
		}
		if len(data) != region.Size {
			t.Fatalf("snapshot fixture region %s size = %d, want %d", region.Name, len(data), region.Size)
		}
		regions = append(regions, snapshotFixtureRegion{address: uint32(addr), data: data})
	}
	runtimeMarker := make([]byte, 4)
	binary.BigEndian.PutUint32(runtimeMarker, 1)
	regions = append(regions, snapshotFixtureRegion{address: addrMmRegEditorPtr, data: runtimeMarker})

	mem := n64.NewMemory(&snapshotFixtureCoreReader{regions: regions})
	mem.SetBaseShift(n64.VirtualBase)
	mem.SetSwizzle(false)

	reader := NewReader(mem)
	var state *GameState
	for attempt := 0; attempt < 3; attempt++ {
		state, err = reader.ReadState()
		if err != nil {
			t.Fatalf("read snapshot fixture state %s: %v", name, err)
		}
		if state != nil && state.Valid && state.ActiveGame != GameNone {
			return state
		}
	}
	if state == nil {
		t.Fatalf("snapshot fixture %s produced no state", name)
	}
	t.Fatalf("snapshot fixture %s produced incomplete state: valid=%v activeGame=%v", name, state.Valid, state.ActiveGame)
	return nil
}

func TestSnapshotFixtureGerudoCardIncludesCheck(t *testing.T) {
	state := loadSnapshotFixtureState(t, "gerudo-card-20260429-201847.json")

	items := itemQtyMap(ExtractItems(state))
	if got := items["OOT_GERUDO_CARD"]; got != 1 {
		t.Fatalf("OOT_GERUDO_CARD = %d, want 1", got)
	}

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Gerudo Member Card"]; !ok {
		t.Fatal("missing Gerudo Member Card check from snapshot fixture")
	}
}

func TestSnapshotFixtureMarketDogLadyIncludesCheck(t *testing.T) {
	state := loadSnapshotFixtureState(t, "market-dog-lady-20260429-205725.json")

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Market Dog Lady HP"]; !ok {
		t.Fatal("missing Market Dog Lady HP check from snapshot fixture")
	}
}

func TestSnapshotFixtureChildShootingGalleryIncludesCheck(t *testing.T) {
	state := loadSnapshotFixtureState(t, "child-shooting-gallery-20260429-210339.json")

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Shooting Gallery Child"]; !ok {
		t.Fatal("missing Shooting Gallery Child check from snapshot fixture")
	}
}

func TestSnapshotFixtureOcarinaGameIncludesCheck(t *testing.T) {
	state := loadSnapshotFixtureState(t, "ocarina-game-20260429-213033.json")

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Lost Woods Memory Game"]; !ok {
		t.Fatal("missing Lost Woods Memory Game check from snapshot fixture")
	}
	if _, ok := checks["Shooting Gallery Child"]; ok {
		t.Fatal("unexpected Shooting Gallery Child check from ocarina-game snapshot fixture")
	}
}

func TestSnapshotFixtureLonLonRanchTalonBottleIncludesCheck(t *testing.T) {
	state := loadSnapshotFixtureState(t, "lon-lon-ranch-talon-bottle-20260429-210824.json")

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Lon Lon Ranch Talon Bottle"]; !ok {
		t.Fatal("missing Lon Lon Ranch Talon Bottle check from snapshot fixture")
	}
}

func TestSnapshotFixtureHoneyDarlingFalsePositive(t *testing.T) {
	state := loadSnapshotFixtureState(t, "honey-darling-false-20260429-204043.json")

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Honey & Darling Reward Any Day"]; ok {
		t.Fatal("unexpected Honey & Darling Reward Any Day check from snapshot fixture")
	}
	if _, ok := checks["Honey & Darling Reward All Days"]; ok {
		t.Fatal("unexpected Honey & Darling Reward All Days check from snapshot fixture")
	}
}
