package ootmm

import (
	"encoding/base64"
	"encoding/binary"
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
	return loadSnapshotFixtureStateOptions(t, name, false)
}

func loadSnapshotFixtureStateAllowGameNone(t *testing.T, name string) *GameState {
	return loadSnapshotFixtureStateOptions(t, name, true)
}

func loadSnapshotFixtureStateOptions(t *testing.T, name string, allowGameNone bool) *GameState {
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
	if allowGameNone {
		fallback, err := loadSnapshotFixtureOfflineState(mem, state)
		if err != nil {
			t.Fatalf("read offline snapshot fixture state %s: %v", name, err)
		}
		return fallback
	}
	if state == nil {
		t.Fatalf("snapshot fixture %s produced no state", name)
	}
	t.Fatalf("snapshot fixture %s produced incomplete state: valid=%v activeGame=%v", name, state.Valid, state.ActiveGame)
	return nil
}

func loadSnapshotFixtureOfflineState(mem *n64.Memory, base *GameState) (*GameState, error) {
	state := &GameState{}
	if base != nil {
		*state = *base
	}

	ootData, err := mem.Read(AddrOotSaveCtx, OotSaveCtxSize)
	if err != nil {
		return nil, fmt.Errorf("read OoT save context: %w", err)
	}
	if err := parseOotSave(&state.Oot, ootData); err != nil {
		return nil, fmt.Errorf("parse OoT save context: %w", err)
	}

	if mmData, err := mem.Read(AddrMmSaveCtx, MmSaveCtxSize); err == nil {
		if err := parseMmSave(&state.Mm, mmData); err != nil {
			return nil, fmt.Errorf("parse MM save context: %w", err)
		}
		state.Mm.ExtraFlags2 = state.Oot.ExtraRecords[ExtraIdxMmFlags2]
	}

	state.Valid = true
	return state, nil
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
}

func TestSnapshotFixtureMemoryGameTransitionIncludesCheck(t *testing.T) {
	state := loadSnapshotFixtureStateAllowGameNone(t, "test-20260501-110855.json")

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Lost Woods Memory Game"]; !ok {
		t.Fatal("missing Lost Woods Memory Game check from transition snapshot fixture")
	}
	if _, ok := checks["Shooting Gallery Child"]; !ok {
		t.Fatal("missing Shooting Gallery Child check from transition snapshot fixture")
	}
}

func TestSnapshotFixtureMemoryGameAndChildShootingIncludesChecks(t *testing.T) {
	state := loadSnapshotFixtureState(t, "test-20260501-125454.json")

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Lost Woods Memory Game"]; !ok {
		t.Fatal("missing Lost Woods Memory Game check from combined snapshot fixture")
	}
	if _, ok := checks["Shooting Gallery Child"]; !ok {
		t.Fatal("missing Shooting Gallery Child check from combined snapshot fixture")
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

func TestSnapshotFixtureInitialSongOfHealingExtraFlag(t *testing.T) {
	withoutState := loadSnapshotFixtureState(t, "mm-without-initial-song-of-healing-20260501-143756.json")
	withState := loadSnapshotFixtureState(t, "mm-with-initial-song-of-healing-20260501-143938.json")

	withoutChecks := checkNameSet(ExtractChecks(withoutState))
	if _, ok := withoutChecks["Initial Song of Healing"]; ok {
		t.Fatal("unexpected Initial Song of Healing check in without snapshot fixture")
	}

	withChecks := checkNameSet(ExtractChecks(withState))
	if _, ok := withChecks["Initial Song of Healing"]; !ok {
		t.Fatal("missing Initial Song of Healing check from snapshot fixture")
	}
}

func TestSnapshotFixtureTingleMapWeekEventFallback(t *testing.T) {
	beforeState := loadSnapshotFixtureState(t, "before-tingle-20260501-170052.json")
	afterState := loadSnapshotFixtureState(t, "after-tingle-20260501-170137.json")

	beforeChecks := checkNameSet(ExtractChecks(beforeState))
	afterChecks := checkNameSet(ExtractChecks(afterState))

	for _, name := range []string{"Tingle Map Clock Town", "Tingle Map Woodfall"} {
		if _, ok := beforeChecks[name]; ok {
			t.Fatalf("unexpected %s check in before snapshot fixture", name)
		}
		if _, ok := afterChecks[name]; !ok {
			t.Fatalf("missing %s check in after snapshot fixture", name)
		}
	}

	for _, name := range []string{"Tingle Map Snowhead", "Tingle Map Ranch", "Tingle Map Great Bay", "Tingle Map Ikana"} {
		if _, ok := afterChecks[name]; ok {
			t.Fatalf("unexpected %s check in after snapshot fixture", name)
		}
	}

	newChecks := 0
	for name := range afterChecks {
		if _, ok := beforeChecks[name]; ok {
			continue
		}
		newChecks++
		if name != "Tingle Map Clock Town" && name != "Tingle Map Woodfall" {
			t.Fatalf("unexpected new check in after snapshot fixture: %s", name)
		}
	}
	if newChecks != 2 {
		t.Fatalf("unexpected new check count across Tingle snapshot fixtures: got %d, want 2", newChecks)
	}
}

func TestSnapshotFixtureMadameAromaMaskKafeiFallback(t *testing.T) {
	beforeState := loadSnapshotFixtureState(t, "before-madame-aroma-20260501-170327.json")
	afterState := loadSnapshotFixtureState(t, "after-madame-aroma-20260501-170357.json")

	beforeChecks := checkNameSet(ExtractChecks(beforeState))
	afterChecks := checkNameSet(ExtractChecks(afterState))

	if _, ok := beforeChecks["Mayor's Office Kafei's Mask"]; ok {
		t.Fatal("unexpected Mayor's Office Kafei's Mask check in before snapshot fixture")
	}
	if _, ok := afterChecks["Mayor's Office Kafei's Mask"]; !ok {
		t.Fatal("missing Mayor's Office Kafei's Mask check in after snapshot fixture")
	}

	newChecks := 0
	for name := range afterChecks {
		if _, ok := beforeChecks[name]; ok {
			continue
		}
		newChecks++
		if name != "Mayor's Office Kafei's Mask" {
			t.Fatalf("unexpected new check in after Madame Aroma snapshot fixture: %s", name)
		}
	}
	if newChecks != 1 {
		t.Fatalf("unexpected new check count across Madame Aroma snapshot fixtures: got %d, want 1", newChecks)
	}

	for name := range beforeChecks {
		if _, ok := afterChecks[name]; !ok {
			t.Fatalf("unexpected removed check in after Madame Aroma snapshot fixture: %s", name)
		}
	}
}
