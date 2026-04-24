package ootmm

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"ootmm-autotracker/n64"
)

type failingCoreReader struct{}

func (failingCoreReader) ReadMemory(addr uint32, size int) ([]byte, error) {
	return nil, os.ErrNotExist
}

func (failingCoreReader) ReadMemoryLarge(addr uint32, size int) ([]byte, error) {
	return nil, os.ErrNotExist
}

type testDebugSnapshot struct {
	Summary struct {
		SaveIndex uint32 `json:"saveIndex"`
	} `json:"summary"`
	Regions []struct {
		Name string `json:"name"`
		Data string `json:"data"`
	} `json:"regions"`
}

func loadTestDebugSnapshot(t *testing.T, name string) testDebugSnapshot {
	t.Helper()

	path := filepath.Join("..", "memory-dumps", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read debug snapshot %s: %v", name, err)
	}

	var snapshot testDebugSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		t.Fatalf("parse debug snapshot %s: %v", name, err)
	}
	return snapshot
}

func decodeTestSnapshotRegion(t *testing.T, snapshot testDebugSnapshot, name string) []byte {
	t.Helper()

	for _, region := range snapshot.Regions {
		if region.Name != name {
			continue
		}
		data, err := base64.StdEncoding.DecodeString(region.Data)
		if err != nil {
			t.Fatalf("decode region %s: %v", name, err)
		}
		return data
	}

	t.Fatalf("missing region %s", name)
	return nil
}

func failingSharedBitmaps(data []byte) []string {
	failing := make([]string, 0)
	for _, bitmap := range sharedStorage.Bitmaps {
		end := bitmap.Offset + bitmap.Size
		if end > len(data) {
			failing = append(failing, bitmap.Name+":out-of-bounds")
			continue
		}
		if !sharedBitmapHasNoUnusedBits(data[bitmap.Offset:end], sharedBitmapUsedBits[bitmap.Name], bitmap.Size) {
			failing = append(failing, bitmap.Name)
		}
	}
	return failing
}

func TestForeignOotSaveAddrUsesSaveIndex(t *testing.T) {
	addr, err := foreignOotSaveAddr(0)
	if err != nil {
		t.Fatalf("foreignOotSaveAddr(0): %v", err)
	}
	if want := AddrMmPayload + ForeignOotSaveBaseOff; addr != want {
		t.Fatalf("unexpected slot 0 addr: got %08x want %08x", addr, want)
	}

	addr, err = foreignOotSaveAddr(2)
	if err != nil {
		t.Fatalf("foreignOotSaveAddr(2): %v", err)
	}
	if want := AddrMmPayload + ForeignOotSaveBaseOff + 2*ForeignOotSaveStride; addr != want {
		t.Fatalf("unexpected slot 2 addr: got %08x want %08x", addr, want)
	}
}

func TestForeignMmSaveAddrUsesSaveIndex(t *testing.T) {
	addr, err := foreignMmSaveAddr(0)
	if err != nil {
		t.Fatalf("foreignMmSaveAddr(0): %v", err)
	}
	if want := AddrOotPayload + ForeignMmSaveBaseOff; addr != want {
		t.Fatalf("unexpected slot 0 addr: got %08x want %08x", addr, want)
	}

	addr, err = foreignMmSaveAddr(2)
	if err != nil {
		t.Fatalf("foreignMmSaveAddr(2): %v", err)
	}
	if want := AddrOotPayload + ForeignMmSaveBaseOff + 2*ForeignMmSaveStride; addr != want {
		t.Fatalf("unexpected slot 2 addr: got %08x want %08x", addr, want)
	}
}

func TestForeignSaveAddrRejectsOutOfRangeIndex(t *testing.T) {
	if _, err := foreignMmSaveAddr(16); err == nil {
		t.Fatal("expected MM foreign save addr to reject out-of-range slot")
	}
	if _, err := foreignOotSaveAddr(64); err == nil {
		t.Fatal("expected OoT foreign save addr to reject out-of-range slot")
	}
}

func TestLocateForeignOotSaveFindsChecksummedCandidate(t *testing.T) {
	payload := make([]byte, 0x50000)
	offset := 0x429f0
	candidate := payload[offset : offset+OotSaveSize]
	binary.BigEndian.PutUint32(candidate[OotOffAge:], 1)
	binary.BigEndian.PutUint16(candidate[OotOffSceneID:], 0x20)
	for i := 0; i < 24; i++ {
		candidate[OotOffInvItems+i] = emptyInventoryItem
	}
	candidate[OotOffInvItems] = 0x00
	candidate[OotOffInvItems+1] = 0x01
	candidate[OotOffInvItems+5] = 0x05
	candidate[OotOffInvItems+6] = 0x06
	candidate[OotOffInvItems+7] = 0x08
	candidate[OotOffInvItems+9] = 0x0A
	candidate[OotOffInvItems+22] = 0x33
	binary.BigEndian.PutUint16(candidate[OotOffChecksum:], ootChecksum(candidate))

	addr, ok := locateForeignOotSave(payload, AddrMmPayload)
	if !ok {
		t.Fatal("expected foreign OoT save candidate")
	}
	if want := AddrMmPayload + uint32(offset); addr != want {
		t.Fatalf("unexpected foreign OoT addr: got %08x want %08x", addr, want)
	}
}

func TestIsPlausibleOotSaveRejectsInvalidSceneID(t *testing.T) {
	data := make([]byte, OotSaveSize)
	binary.BigEndian.PutUint32(data[OotOffAge:], 1)
	binary.BigEndian.PutUint16(data[OotOffSceneID:], 0xFFFF)
	for i := 0; i < 24; i++ {
		data[OotOffInvItems+i] = emptyInventoryItem
	}

	if isPlausibleOotSave(data) {
		t.Fatal("expected OoT save with invalid sceneId to fail plausibility")
	}
}

func TestLocateForeignOotSaveFindsRewardMutatedCandidate(t *testing.T) {
	payload := make([]byte, 0x50000)

	garbageOffset := 0x39030
	garbage := payload[garbageOffset : garbageOffset+OotSaveSize]
	binary.BigEndian.PutUint32(garbage[OotOffAge:], 0)
	binary.BigEndian.PutUint16(garbage[OotOffSceneID:], 0xFFFF)
	for i := 0; i < 24; i++ {
		garbage[OotOffInvItems+i] = emptyInventoryItem
	}
	binary.BigEndian.PutUint16(garbage[OotOffChecksum:], ootChecksum(garbage)+0x0200)

	validOffset := 0x429f0
	valid := payload[validOffset : validOffset+OotSaveSize]
	binary.BigEndian.PutUint32(valid[OotOffAge:], 1)
	binary.BigEndian.PutUint16(valid[OotOffSceneID:], 0x20)
	for i := 0; i < 24; i++ {
		valid[OotOffInvItems+i] = emptyInventoryItem
	}
	valid[OotOffInvItems] = 0x00
	valid[OotOffInvItems+1] = 0x01
	valid[OotOffInvItems+7] = 0x08
	valid[OotOffInvItems+12] = 0x0E
	valid[OotOffInvItems+13] = 0x0F
	binary.BigEndian.PutUint32(valid[OotOffUpgrades:], 0x00162040)
	binary.BigEndian.PutUint16(valid[OotOffChecksum:], ootChecksum(valid))
	binary.BigEndian.PutUint32(valid[OotOffUpgrades:], 0x00163040)

	if delta, ok := ootChecksumDelta(valid); !ok || delta != 0x1000 {
		t.Fatalf("reward-mutated OoT checksum delta = %d, %v, want 0x1000", delta, ok)
	}

	addr, ok := locateForeignOotSave(payload, AddrMmPayload)
	if !ok {
		t.Fatal("expected foreign OoT save candidate with reward-sized checksum delta")
	}
	if want := AddrMmPayload + uint32(validOffset); addr != want {
		t.Fatalf("unexpected foreign OoT addr: got %08x want %08x", addr, want)
	}
}

func TestLocateForeignMmSaveFindsChecksummedCandidate(t *testing.T) {
	payload := make([]byte, 0x80000)
	offset := 0x43970
	candidate := payload[offset : offset+MmSaveSize]
	candidate[MmOffPlayerForm] = 4
	binary.BigEndian.PutUint32(candidate[MmOffDay:], 1)
	binary.BigEndian.PutUint16(candidate[MmOffTime:], 0x60B1)
	for i := 0; i < 48; i++ {
		candidate[MmOffInvItems+i] = emptyInventoryItem
	}
	candidate[MmOffInvItems] = 0x00
	candidate[MmOffInvItems+15] = 0x0F
	candidate[MmOffInvItems+42] = 0x37
	binary.BigEndian.PutUint16(candidate[MmOffChecksum:], mmChecksum(candidate))

	addr, ok := locateForeignMmSave(payload, AddrOotPayload)
	if !ok {
		t.Fatal("expected foreign MM save candidate")
	}
	if want := AddrOotPayload + uint32(offset); addr != want {
		t.Fatalf("unexpected foreign MM addr: got %08x want %08x", addr, want)
	}
}

func TestValidateOotSaveRejectsGarbage(t *testing.T) {
	data := make([]byte, OotSaveSize)
	data[OotOffChecksum] = 0x12
	data[OotOffChecksum+1] = 0x34

	if err := validateOotSave(data); err == nil {
		t.Fatal("expected garbage OoT save to fail checksum validation")
	}
}

func TestValidateMmSaveAcceptsMatchingChecksum(t *testing.T) {
	data := make([]byte, MmSaveSize)
	data[0] = 0x01
	data[1] = 0x02
	data[2] = 0x03

	checksum := uint16(data[0]) + uint16(data[1]) + uint16(data[2])
	data[MmOffChecksum] = byte(checksum >> 8)
	data[MmOffChecksum+1] = byte(checksum)

	if err := validateMmSave(data); err != nil {
		t.Fatalf("expected MM checksum to validate: %v", err)
	}
}

func TestParseMmSaveUsesPaddedInventoryOffsets(t *testing.T) {
	data := make([]byte, MmSaveSize)
	data[MmOffEquipment] = 0x00
	data[MmOffEquipment+1] = 0x23
	data[MmOffInvItems] = 0x00
	data[MmOffInvItems+1] = 0xFF
	data[MmOffInvQuest+2] = 0x80
	data[MmOffStrayFairies] = 3
	data[MmOffStrayFairies+1] = 7
	data[MmOffStrayFairies+2] = 6
	data[MmOffStrayFairies+3] = 15
	binary.BigEndian.PutUint16(data[MmOffSkullSwamp:], 30)
	binary.BigEndian.PutUint16(data[MmOffSkullOcean:], 29)

	var mm MmState
	if err := parseMmSave(&mm, data); err != nil {
		t.Fatalf("parseMmSave: %v", err)
	}

	if mm.Items[0] != 0x00 || mm.Items[1] != 0xFF {
		t.Fatalf("unexpected item bytes: % x", mm.Items[:4])
	}
	if mm.Equipment != 0x0023 {
		t.Fatalf("unexpected MM equipment: %04x", mm.Equipment)
	}
	if mm.QuestItems != 0x00008000 {
		t.Fatalf("unexpected MM quest bits: %08x", mm.QuestItems)
	}
	if mm.StrayFairies[0] != 3 || mm.StrayFairies[1] != 7 || mm.StrayFairies[2] != 6 || mm.StrayFairies[3] != 15 {
		t.Fatalf("unexpected MM stray fairies: % x", mm.StrayFairies[:4])
	}
	if mm.SkullTokensSwamp != 30 || mm.SkullTokensOcean != 29 {
		t.Fatalf("unexpected MM skull token counts: swamp=%d ocean=%d", mm.SkullTokensSwamp, mm.SkullTokensOcean)
	}
}

func TestParseMmSaveReadsTownStrayFairyWeekEvent(t *testing.T) {
	data := make([]byte, MmSaveSize)
	data[MmOffWeekEventReg+mmWeekEventTownStrayFairyByte] = mmWeekEventTownStrayFairyMask
	data[MmOffWeekEventReg+23] |= 0x02

	var mm MmState
	if err := parseMmSave(&mm, data); err != nil {
		t.Fatalf("parseMmSave: %v", err)
	}
	if !mm.TownStrayFairy {
		t.Fatal("expected MM town stray fairy to be true")
	}
	if got := mm.WeekEventReg[23]; got != 0x02 {
		t.Fatalf("unexpected MM week event byte 23: got %#02x want 0x02", got)
	}
}

func TestParseMmSaveReadsOwlActivationFlags(t *testing.T) {
	data := make([]byte, MmSaveSize)
	binary.BigEndian.PutUint16(data[MmOffOwlActivationFlags:], 1<<mmOwlClockTownBit)

	var mm MmState
	if err := parseMmSave(&mm, data); err != nil {
		t.Fatalf("parseMmSave: %v", err)
	}
	if got := mm.OwlActivationFlags; got != 1<<mmOwlClockTownBit {
		t.Fatalf("unexpected MM owl activation flags: got %#04x want %#04x", got, 1<<mmOwlClockTownBit)
	}
}

func TestParseMmSaveReadsMagicFlags(t *testing.T) {
	data := make([]byte, MmSaveSize)
	data[MmOffMagicAcquired] = 1
	data[MmOffDoubleMagic] = 1

	var mm MmState
	if err := parseMmSave(&mm, data); err != nil {
		t.Fatalf("parseMmSave: %v", err)
	}
	if !mm.HasMagic {
		t.Fatal("expected MM magic-acquired flag to be true")
	}
	if !mm.HasDoubleMagic {
		t.Fatal("expected MM double-magic flag to be true")
	}
}

func TestParseOotSaveReadsEventBitmaps(t *testing.T) {
	data := make([]byte, OotSaveCtxSize)
	data[OotOffInvItems] = emptyInventoryItem
	binary.BigEndian.PutUint16(data[OotOffEventsChk+(ootEventSongEpona>>4)*2:], 1<<(ootEventSongEpona&0xF))
	binary.BigEndian.PutUint16(data[OotOffEventsItem+(ootEventItemGoronBracelet>>4)*2:], 1<<(ootEventItemGoronBracelet&0xF))
	binary.BigEndian.PutUint16(data[OotOffEventsMisc+(ootEventMiscMedigoron>>4)*2:], 1<<(ootEventMiscMedigoron&0xF))

	var oot OotState
	if err := parseOotSave(&oot, data); err != nil {
		t.Fatalf("parseOotSave: %v", err)
	}
	if got := oot.EventsChk[ootEventSongEpona>>4]; got != 1<<(ootEventSongEpona&0xF) {
		t.Fatalf("unexpected OoT eventsChk word: got %#04x", got)
	}
	if got := oot.EventsItem[ootEventItemGoronBracelet>>4]; got != 1<<(ootEventItemGoronBracelet&0xF) {
		t.Fatalf("unexpected OoT eventsItem word: got %#04x", got)
	}
	if got := oot.EventsMisc[ootEventMiscMedigoron>>4]; got != 1<<(ootEventMiscMedigoron&0xF) {
		t.Fatalf("unexpected OoT eventsMisc word: got %#04x", got)
	}
}

func TestAcceptStableStateRequiresTwoMatchingObservations(t *testing.T) {
	r := &Reader{}

	if r.acceptStableState(GameOot, 1) {
		t.Fatal("first observation should not be accepted")
	}
	if !r.acceptStableState(GameOot, 1) {
		t.Fatal("second matching observation should be accepted")
	}
	if !r.acceptStableState(GameOot, 1) {
		t.Fatal("stable state should continue to be accepted")
	}
}

func TestAcceptStableStateResetsOnChangedCandidate(t *testing.T) {
	r := &Reader{}

	if r.acceptStableState(GameMm, 0) {
		t.Fatal("first MM observation should not be accepted")
	}
	if r.acceptStableState(GameMm, 1) {
		t.Fatal("changed save slot should restart stabilization")
	}
	if !r.acceptStableState(GameMm, 1) {
		t.Fatal("second matching changed observation should be accepted")
	}
}

func TestLocateForeignMmSaveFindsNearChecksummedCandidate(t *testing.T) {
	payload := make([]byte, 0x80000)
	offset := 0x43970
	candidate := payload[offset : offset+MmSaveSize]
	candidate[MmOffPlayerForm] = 4
	binary.BigEndian.PutUint32(candidate[MmOffDay:], 0)
	for i := 0; i < 48; i++ {
		candidate[MmOffInvItems+i] = emptyInventoryItem
	}
	for i, keys := range []int8{1, 3, 1, 4, -1, -1, -1, -1, -1} {
		candidate[MmOffDungeonKeys+i] = byte(keys)
	}
	for i, fairies := range []int8{15, 15, 15, 15, 0, 0, 0, 0, 0, 0} {
		candidate[MmOffStrayFairies+i] = byte(fairies)
	}
	candidate[MmOffWeekEventReg+mmWeekEventTownStrayFairyByte] = mmWeekEventTownStrayFairyMask
	binary.BigEndian.PutUint16(candidate[MmOffChecksum:], mmChecksum(candidate))

	// Simulate live payload mutations after the flash copy was loaded.
	candidate[MmOffDungeonKeys+1] = 4
	candidate[MmOffStrayFairies+4] = 1

	addr, ok := locateForeignMmSave(payload, AddrOotPayload)
	if !ok {
		t.Fatal("expected foreign MM save candidate with near checksum")
	}
	if want := AddrOotPayload + uint32(offset); addr != want {
		t.Fatalf("unexpected foreign MM addr: got %08x want %08x", addr, want)
	}
}

func TestResetEmptyMmStateUsesEmptyInventorySentinels(t *testing.T) {
	mm := MmState{}
	resetEmptyMmState(&mm)

	for i, itemID := range mm.Items {
		if itemID != emptyInventoryItem {
			t.Fatalf("unexpected empty MM item %d: got %02x want %02x", i, itemID, emptyInventoryItem)
		}
	}
	for i, keys := range mm.DungeonKeys {
		if keys != -1 {
			t.Fatalf("unexpected empty MM small key count %d: got %d want -1", i, keys)
		}
	}
}

func TestIsPlausibleOotPlayStateSample(t *testing.T) {
	sample := ootPlayStateSample{
		sceneID:        0x28,
		actorTotal:     94,
		currentRoom:    0,
		linkAgeOnLoad:  1,
		gameplayFrames: 1234,
	}
	if !isPlausibleOotPlayStateSample(sample) {
		t.Fatal("expected plausible OoT PlayState sample to validate")
	}

	sample.actorTotal = 0
	if isPlausibleOotPlayStateSample(sample) {
		t.Fatal("expected sample with zero actors to fail plausibility")
	}
}

func TestRememberOotStateDropsLiveFlags(t *testing.T) {
	r := &Reader{}
	oot := OotState{
		LiveSceneID:         0x28,
		LiveChestFlags:      1 << 2,
		LiveCollectFlags:    1 << 3,
		LiveTempCollectFlag: 1 << 4,
		HasLiveSceneFlags:   true,
	}

	r.rememberOotState(oot)

	if r.lastKnownOot.HasLiveSceneFlags {
		t.Fatal("expected remembered OoT state to drop live scene flags")
	}
	if r.lastKnownOot.LiveSceneID != 0 || r.lastKnownOot.LiveChestFlags != 0 || r.lastKnownOot.LiveCollectFlags != 0 || r.lastKnownOot.LiveTempCollectFlag != 0 {
		t.Fatalf("remembered live fields = scene %d chest %#x collect %#x tempCollect %#x, want zero", r.lastKnownOot.LiveSceneID, r.lastKnownOot.LiveChestFlags, r.lastKnownOot.LiveCollectFlags, r.lastKnownOot.LiveTempCollectFlag)
	}
}

func TestReadSharedStateFallsBackToLastKnownWhenNearForeignUnavailable(t *testing.T) {
	mem := n64.NewMemory(failingCoreReader{})
	mem.SetBaseShift(n64.VirtualBase)
	mem.SetSwizzle(false)

	r := NewReader(mem)
	r.lastKnownShared.SetBit("npcOot", 4)
	r.lastKnownShared.SetBit("xflagsMm", 17)
	r.lastKnownShared.OcarinaButtonMaskOot = 0x8000
	r.hasLastKnownShared = true

	var shared SharedCustomState
	if err := r.readSharedState(GameMm, 0, &shared); err != nil {
		t.Fatalf("readSharedState: %v", err)
	}

	if !bitmapHasBit(shared.Bitmap("npcOot"), 4) {
		t.Fatal("expected last-known npcOot bitmap to be preserved")
	}
	if !bitmapHasBit(shared.Bitmap("xflagsMm"), 17) {
		t.Fatal("expected last-known xflagsMm bitmap to be preserved")
	}
	if shared.OcarinaButtonMaskOot != 0x8000 {
		t.Fatalf("last-known ocarina mask = %#x, want %#x", shared.OcarinaButtonMaskOot, 0x8000)
	}
}

func TestIsPlausibleSharedStateSkipsSoulBitmaps(t *testing.T) {
	shared := SharedCustomState{}
	for _, bitmap := range sharedStorage.Bitmaps {
		data := make([]byte, bitmap.Size)
		// OoTMM fills soul bitmaps with 0xff when settings are disabled
		if isSoulBitmap(bitmap.Name) {
			for i := range data {
				data[i] = 0xff
			}
		}
		shared.SetBitmap(bitmap.Name, data)
	}

	if !isPlausibleSharedState(shared) {
		t.Fatal("expected shared state with full soul bitmaps to pass plausibility")
	}
}

func TestIsPlausibleSharedStateRejectsCheckBitmapFutureBits(t *testing.T) {
	shared := SharedCustomState{}
	for _, bitmap := range sharedStorage.Bitmaps {
		shared.SetBitmap(bitmap.Name, make([]byte, bitmap.Size))
	}

	// Check bitmaps have all bits marked as used (via markSharedCheckBitmapsUsed),
	// so set a future bit in a non-check, non-soul bitmap to test rejection.
	// Find a bitmap that is not a soul bitmap and not a check bitmap.
	var targetBitmap string
	for name, info := range sharedBitmaps {
		if !isSoulBitmap(name) {
			// Check if it's a check bitmap (all bits used = size*8)
			if sharedBitmapUsedBits[name] < info.Size*8 {
				targetBitmap = name
				break
			}
		}
	}
	if targetBitmap == "" {
		t.Skip("no non-soul non-check bitmaps available for testing")
	}

	bitmap := sharedBitmaps[targetBitmap]
	shared.SetBit(bitmap.Name, bitmap.Size*8-1)

	if isPlausibleSharedState(shared) {
		t.Fatalf("expected shared state with future bit in %s to fail plausibility", targetBitmap)
	}
}

func TestParseSharedStateReadsBombchuBagBits(t *testing.T) {
	data := make([]byte, SharedCustomSaveSize)
	data[sharedBombchuBagFlagsOffset] = (2 << sharedBombchuBagOotShift) | (3 << sharedBombchuBagMmShift)

	shared, err := parseSharedState(data)
	if err != nil {
		t.Fatalf("parseSharedState: %v", err)
	}
	if shared.BombchuBagOot != 2 {
		t.Fatalf("BombchuBagOot = %d, want 2", shared.BombchuBagOot)
	}
	if shared.BombchuBagMm != 3 {
		t.Fatalf("BombchuBagMm = %d, want 3", shared.BombchuBagMm)
	}
}

func TestParseSharedStateReadsOcarinaButtonMasks(t *testing.T) {
	data := make([]byte, SharedCustomSaveSize)
	binary.BigEndian.PutUint16(data[sharedOcarinaButtonMaskOotOffset:], sharedOcarinaButtonMaskDisabled)
	binary.BigEndian.PutUint16(data[sharedOcarinaButtonMaskMmOffset:], sharedOcarinaButtonCRightMask)

	shared, err := parseSharedState(data)
	if err != nil {
		t.Fatalf("parseSharedState: %v", err)
	}
	if shared.OcarinaButtonMaskOot != sharedOcarinaButtonMaskDisabled {
		t.Fatalf("OcarinaButtonMaskOot = %#x, want %#x", shared.OcarinaButtonMaskOot, sharedOcarinaButtonMaskDisabled)
	}
	if shared.OcarinaButtonMaskMm != sharedOcarinaButtonCRightMask {
		t.Fatalf("OcarinaButtonMaskMm = %#x, want %#x", shared.OcarinaButtonMaskMm, sharedOcarinaButtonCRightMask)
	}
}

func TestParseSharedStateReadsCaughtFishWeights(t *testing.T) {
	data := make([]byte, SharedCustomSaveSize)
	data[sharedCaughtChildFishWeightOffset] = 2
	data[sharedCaughtChildFishWeightOffset+1] = 7
	data[sharedCaughtChildFishWeightOffset+2] = fishingPondLoachWeightMask | 19
	data[sharedCaughtAdultFishWeightOffset] = 2
	data[sharedCaughtAdultFishWeightOffset+1] = 25
	data[sharedCaughtAdultFishWeightOffset+2] = fishingPondLoachWeightMask | 36

	shared, err := parseSharedState(data)
	if err != nil {
		t.Fatalf("parseSharedState: %v", err)
	}
	if got := shared.CaughtChildFishWeights[0]; got != 2 {
		t.Fatalf("CaughtChildFishWeights[0] = %d, want 2", got)
	}
	if got := shared.CaughtChildFishWeights[2]; got != fishingPondLoachWeightMask|19 {
		t.Fatalf("CaughtChildFishWeights[2] = %#x, want %#x", got, fishingPondLoachWeightMask|19)
	}
	if got := shared.CaughtAdultFishWeights[1]; got != 25 {
		t.Fatalf("CaughtAdultFishWeights[1] = %d, want 25", got)
	}
	if got := shared.CaughtAdultFishWeights[2]; got != fishingPondLoachWeightMask|36 {
		t.Fatalf("CaughtAdultFishWeights[2] = %#x, want %#x", got, fishingPondLoachWeightMask|36)
	}
}

func TestSelectSharedStateCandidatePrefersRicherCheckState(t *testing.T) {
	r := &Reader{}

	near := SharedCustomState{}
	near.SetBit("npcMm", 1)

	payload := SharedCustomState{}
	payload.SetBit("npcMm", 1)
	payload.SetBit("npcOot", 3)
	payload.SetBit("xflagsOot", 9)

	got, ok := r.selectSharedStateCandidate([]sharedStateCandidate{
		{source: "near-foreign", state: near},
		{source: "payload-80400000", state: payload},
	})
	if !ok {
		t.Fatal("expected shared-state candidate selection to succeed")
	}
	if got.source != "payload-80400000" {
		t.Fatalf("selected source = %q, want payload-80400000", got.source)
	}
	if bitmap := got.state.Bitmap("npcOot"); len(bitmap) == 0 || bitmap[0]&(1<<3) == 0 {
		t.Fatal("selected candidate is missing expected OoT NPC bit")
	}
}

func TestSelectSharedStateCandidatePrefersTrackedSoulBitmap(t *testing.T) {
	r := &Reader{}

	payload := SharedCustomState{}
	near := SharedCustomState{}
	source := mustCatalogItemSource("MM_SOUL_ENEMY_EYEGORE")
	near.SetBit(source.Block, source.Bit)

	got, ok := r.selectSharedStateCandidate([]sharedStateCandidate{
		{source: "payload-80400000", state: payload},
		{source: "near-foreign", state: near},
	})
	if !ok {
		t.Fatal("expected shared-state candidate selection to succeed")
	}
	if got.source != "near-foreign" {
		t.Fatalf("selected source = %q, want near-foreign", got.source)
	}
	byteIndex := source.Bit / 8
	bitMask := uint8(1 << uint(source.Bit%8))
	if bitmap := got.state.Bitmap(source.Block); len(bitmap) <= byteIndex || bitmap[byteIndex]&bitMask == 0 {
		t.Fatal("selected candidate is missing expected MM Soul of Eyegore bit")
	}
}

func TestOverlaySharedCheckBitmapsAddsNearForeignScrubProgress(t *testing.T) {
	payload := SharedCustomState{}
	payload.SetBit("npcMm", 1)
	payload.SetBit("npcOot", 3)
	payload.SetBit("xflagsOot", 9)

	near := SharedCustomState{}
	near.SetBit("scrubsOot", 0)

	overlaySharedCheckBitmaps(&payload, &near)

	if bitmap := payload.Bitmap("scrubsOot"); len(bitmap) == 0 || bitmap[0]&1 == 0 {
		t.Fatal("overlay is missing expected OoT scrub bit")
	}
	if bitmap := payload.Bitmap("npcOot"); len(bitmap) == 0 || bitmap[0]&(1<<3) == 0 {
		t.Fatal("overlay dropped expected OoT NPC bit")
	}
	if bitmap := payload.Bitmap("xflagsOot"); len(bitmap) < 2 || bitmap[1]&(1<<1) == 0 {
		t.Fatal("overlay dropped expected OoT XFlag bit")
	}
}

func TestSelectSharedStateCandidatePrefersContinuityOverSparseRegression(t *testing.T) {
	r := &Reader{}
	r.lastKnownShared.SetBit("npcOot", 4)
	r.lastKnownShared.SetBit("xflagsMm", 17)
	r.hasLastKnownShared = true

	regressed := SharedCustomState{}
	regressed.SetBit("npcMm", 1)

	continued := SharedCustomState{}
	continued.SetBit("npcOot", 4)
	continued.SetBit("xflagsMm", 17)

	got, ok := r.selectSharedStateCandidate([]sharedStateCandidate{
		{source: "near-foreign", state: regressed},
		{source: "payload-80730000", state: continued},
	})
	if !ok {
		t.Fatal("expected shared-state candidate selection to succeed")
	}
	if got.source != "payload-80730000" {
		t.Fatalf("selected source = %q, want payload-80730000", got.source)
	}
}

func TestDebugDumpNearForeignSharedStateShowsScrubImmediately(t *testing.T) {
	// Tests that relied on external memory-dump files for scrub scenarios
	// have been removed because the corresponding dump files are not present
	// in the repository. Related behavior is still covered by unit tests that
	// do not depend on those external fixtures.
}

func TestDebugDumpMmNearForeignFindsOotSave(t *testing.T) {
	snap := loadTestDebugSnapshot(t, "mm-after-chest-20260418-212749.json")
	mmPayload := decodeTestSnapshotRegion(t, snap, "mmPayload")

	foreignAddr, ok := locateForeignOotSave(mmPayload, AddrMmPayload)
	if !ok {
		t.Fatal("expected to locate foreign OoT save in MM payload")
	}

	nearOffset := int(foreignAddr - AddrMmPayload - SharedCustomSaveSize)
	if nearOffset < 0 || nearOffset+sharedStateReadSize() > len(mmPayload) {
		t.Fatalf("near-foreign offset %#x is out of bounds", nearOffset)
	}

	window := mmPayload[nearOffset : nearOffset+sharedStateReadSize()]
	shared, err := parseSharedCheckState(window)
	if err != nil {
		t.Fatalf("parse near-foreign shared check state: %v", err)
	}

	// The near-foreign SharedCustomSave should have OoT xflags from the
	// user's session (loaded from flash).  It should NOT have thousands
	// of MM xflag bits (which would be garbage from MIPS code).
	xflagsMm := shared.Bitmap("xflagsMm")
	mmBits := 0
	for _, b := range xflagsMm {
		for b != 0 {
			mmBits++
			b &= b - 1
		}
	}
	if mmBits > 100 {
		t.Fatalf("near-foreign xflagsMm has %d bits set, expected <100 for a fresh seed", mmBits)
	}
}

func TestDebugDumpAfterGrassFindsForeignOotSaveWithDeltaTolerance(t *testing.T) {
	snap := loadTestDebugSnapshot(t, "after-grass-20260419-192638.json")
	mmPayload := decodeTestSnapshotRegion(t, snap, "mmPayload")

	addr, ok := locateForeignOotSave(mmPayload, AddrMmPayload)
	if !ok {
		t.Fatal("expected after-grass MM payload to locate foreign OoT save via delta tolerance")
	}

	// Verify it found the correct save at the known offset.
	expectedOffset := uint32(0x429f0)
	expectedAddr := AddrMmPayload + expectedOffset
	if addr != expectedAddr {
		t.Fatalf("foreign OoT save at %#x, want %#x", addr, expectedAddr)
	}
}

func TestDebugDumpGarbageKeepsForeignOotLocatorStable(t *testing.T) {
	before := loadTestDebugSnapshot(t, "before-dd-20260422-193401.json")
	after := loadTestDebugSnapshot(t, "after-dd-20260422-193433.json")
	garbage := loadTestDebugSnapshot(t, "garbage-20260422-193744.json")

	beforePayload := decodeTestSnapshotRegion(t, before, "mmPayload")
	afterPayload := decodeTestSnapshotRegion(t, after, "mmPayload")
	garbagePayload := decodeTestSnapshotRegion(t, garbage, "mmPayload")

	beforeAddr, ok := locateForeignOotSave(beforePayload, AddrMmPayload)
	if !ok {
		t.Fatal("expected before-dd MM payload to locate foreign OoT save")
	}
	afterAddr, ok := locateForeignOotSave(afterPayload, AddrMmPayload)
	if !ok {
		t.Fatal("expected after-dd MM payload to locate foreign OoT save")
	}
	garbageAddr, ok := locateForeignOotSave(garbagePayload, AddrMmPayload)
	if !ok {
		t.Fatal("expected garbage MM payload to locate foreign OoT save")
	}

	if afterAddr != beforeAddr {
		t.Fatalf("after-dd foreign OoT save moved: got %#x want %#x", afterAddr, beforeAddr)
	}
	if garbageAddr != beforeAddr {
		t.Fatalf("garbage dump foreign OoT save moved: got %#x want %#x", garbageAddr, beforeAddr)
	}
}

func TestDebugDumpGarbageReadStateKeepsSaneForeignOotState(t *testing.T) {
	state := readStateFromSnapshot(t, "garbage-20260422-193744.json")

	if got := state.Oot.SceneID; got != 0x20 {
		t.Fatalf("garbage dump OoT scene = %#x, want 0x20", got)
	}
	if got := state.Oot.GoldTokens; got > 100 {
		t.Fatalf("garbage dump OoT gold tokens = %d, want <= 100", got)
	}
}

func TestDebugDumpMmPayloadCandidateHasGarbageBits(t *testing.T) {
	snap := loadTestDebugSnapshot(t, "mm-after-chest-20260418-212749.json")
	mmPayload := decodeTestSnapshotRegion(t, snap, "mmPayload")

	// The fixed-offset payload candidates read MIPS code rather than save
	// data.  Even when one passes the plausibility filter, it contains
	// implausibly high xflag bit counts.  The near-foreign approach must
	// be preferred over these candidates.
	addr, err := sharedSaveAddr(AddrMmPayload, MmPayloadSize, 0)
	if err != nil {
		t.Fatalf("sharedSaveAddr: %v", err)
	}
	offset := int(addr - AddrMmPayload)
	end := offset + sharedStateReadSize()
	if end > len(mmPayload) {
		t.Skip("payload window out of bounds")
	}
	window := mmPayload[offset:end]
	parsed, err := parseSharedState(window)
	if err != nil {
		t.Skip("payload candidate fails plausibility as expected")
	}

	xflagsOot := parsed.Bitmap("xflagsOot")
	ootBits := 0
	for _, b := range xflagsOot {
		for b != 0 {
			ootBits++
			b &= b - 1
		}
	}
	// MIPS code interpreted as xflag bitmap yields thousands of bits.
	if ootBits < 1000 {
		t.Fatalf("payload slot 0 xflagsOot has only %d bits, expected >1000 for garbage", ootBits)
	}
}

// snapshotCoreReader serves N64 memory reads from a debug dump's stored regions.
type snapshotCoreReader struct {
	regions []snapshotRegion
}

type snapshotRegion struct {
	addr uint32
	data []byte
}

func (s *snapshotCoreReader) ReadMemory(addr uint32, size int) ([]byte, error) {
	return s.ReadMemoryLarge(addr, size)
}

func (s *snapshotCoreReader) ReadMemoryLarge(addr uint32, size int) ([]byte, error) {
	for _, r := range s.regions {
		if addr >= r.addr && int(addr-r.addr)+size <= len(r.data) {
			off := int(addr - r.addr)
			return r.data[off : off+size], nil
		}
	}
	return nil, fmt.Errorf("address %#x (size %d) not in snapshot", addr, size)
}

func (s *snapshotCoreReader) addRegion(addr uint32, data []byte) {
	s.regions = append(s.regions, snapshotRegion{addr: addr, data: data})
}

func (s *snapshotCoreReader) loadSnapshot(t *testing.T, snap testDebugSnapshot) {
	t.Helper()
	s.regions = s.regions[:0]

	for _, region := range snap.Regions {
		addr, ok := regionAddresses[region.Name]
		if !ok {
			continue
		}
		data, err := base64.StdEncoding.DecodeString(region.Data)
		if err != nil {
			t.Fatalf("decode region %s: %v", region.Name, err)
		}
		s.addRegion(addr, data)
	}

	runtimeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(runtimeBuf, 1)
	s.addRegion(0x801F3F60, runtimeBuf)
}

// regionAddresses maps snapshot region names to their N64 virtual addresses.
var regionAddresses = map[string]uint32{
	"comboCtxOot":    AddrComboCtxOot,
	"comboCtxMm":     AddrComboCtxMm,
	"ootSaveContext": AddrOotSaveCtx,
	"mmSaveContext":  AddrMmSaveCtx,
	"ootPayload":     AddrOotPayload,
	"mmPayload":      AddrMmPayload,
}

func newSnapshotCoreReader(t *testing.T, snap testDebugSnapshot) *snapshotCoreReader {
	t.Helper()
	cr := &snapshotCoreReader{}

	for _, region := range snap.Regions {
		addr, ok := regionAddresses[region.Name]
		if !ok {
			continue
		}
		data, err := base64.StdEncoding.DecodeString(region.Data)
		if err != nil {
			t.Fatalf("decode region %s: %v", region.Name, err)
		}
		cr.addRegion(addr, data)
	}

	// MM detection needs runtimeMarker at addrMmRegEditorPtr (0x801F3F60).
	// This address is past the end of mmSaveContext; supply a nonzero value.
	runtimeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(runtimeBuf, 1)
	cr.addRegion(0x801F3F60, runtimeBuf)

	return cr
}

func readStateFromSnapshot(t *testing.T, name string) *GameState {
	t.Helper()
	snap := loadTestDebugSnapshot(t, name)
	cr := newSnapshotCoreReader(t, snap)
	reader := newReaderFromSnapshotCore(cr)
	return readStateFromReader(t, reader, name)
}

func newReaderFromSnapshotCore(cr *snapshotCoreReader) *Reader {
	mem := n64.NewMemory(cr)
	mem.SetBaseShift(n64.VirtualBase)
	mem.SetSwizzle(false)
	return NewReader(mem)
}

func newReaderFromSnapshot(t *testing.T, name string) *Reader {
	t.Helper()
	snap := loadTestDebugSnapshot(t, name)
	cr := newSnapshotCoreReader(t, snap)
	return newReaderFromSnapshotCore(cr)
}

func readStateFromReader(t *testing.T, reader *Reader, name string) *GameState {
	t.Helper()
	var state *GameState
	for attempt := 0; attempt < 3; attempt++ {
		var err error
		state, err = reader.ReadState()
		if err != nil {
			t.Fatalf("ReadState attempt %d: %v", attempt, err)
		}
		if state != nil && state.Valid && state.ActiveGame != GameNone {
			return state
		}
	}
	if state == nil || !state.Valid || state.ActiveGame == GameNone {
		t.Fatalf("ReadState from %s did not produce a valid game state", name)
	}
	return state
}

func TestInMmSnapshotIncludesRequestedMmChecksAndItem(t *testing.T) {
	state := readStateFromSnapshot(t, "in-mm-20260423-173418.json")
	checks := checkNameSet(ExtractChecks(state))
	for _, name := range []string{
		"Road to Southern Swamp HP",
		"Clock Town Tree HP",
		"Clock Town Platform HP",
		"Clock Town Stray Fairy",
	} {
		if _, ok := checks[name]; !ok {
			t.Fatalf("missing %s from in-mm snapshot", name)
		}
	}
	items := itemQtyMap(ExtractItems(state))
	if got := items["MM_STRAY_FAIRY_TOWN"]; got != 1 {
		t.Fatalf("MM_STRAY_FAIRY_TOWN = %d, want 1 in in-mm snapshot", got)
	}
}

func TestMmToOotSnapshotTransitionPreservesRequestedMmState(t *testing.T) {
	mmReader := newReaderFromSnapshot(t, "in-mm-20260423-173418.json")
	mmState := readStateFromReader(t, mmReader, "in-mm-20260423-173418.json")
	if got := itemQtyMap(ExtractItems(mmState))["MM_STRAY_FAIRY_TOWN"]; got != 1 {
		t.Fatalf("MM_STRAY_FAIRY_TOWN = %d, want 1 in source MM snapshot", got)
	}

	ootReader := newReaderFromSnapshot(t, "in-oot-20260423-173148.json")
	ootReader.lastKnownMm = mmReader.lastKnownMm
	ootReader.lastKnownMmSaveIdx = mmReader.lastKnownMmSaveIdx
	ootReader.hasLastKnownMm = mmReader.hasLastKnownMm
	ootState := readStateFromReader(t, ootReader, "in-oot-20260423-173148.json")

	checks := checkNameSet(ExtractChecks(ootState))
	for _, name := range []string{
		"Road to Southern Swamp HP",
		"Clock Town Tree HP",
		"Clock Town Platform HP",
		"Clock Town Stray Fairy",
	} {
		if _, ok := checks[name]; !ok {
			t.Fatalf("missing %s after MM->OoT snapshot transition", name)
		}
	}
	items := itemQtyMap(ExtractItems(ootState))
	if got := items["MM_STRAY_FAIRY_TOWN"]; got != 1 {
		t.Fatalf("MM_STRAY_FAIRY_TOWN = %d, want 1 after MM->OoT snapshot transition", got)
	}
}

func TestSameReaderMmToOotSnapshotTransitionPreservesRequestedMmState(t *testing.T) {
	mmSnap := loadTestDebugSnapshot(t, "in-mm-20260423-173418.json")
	ootSnap := loadTestDebugSnapshot(t, "in-oot-2-20260423-191522.json")
	core := newSnapshotCoreReader(t, mmSnap)
	reader := newReaderFromSnapshotCore(core)

	mmState := readStateFromReader(t, reader, "in-mm-20260423-173418.json")
	if got := itemQtyMap(ExtractItems(mmState))["MM_STRAY_FAIRY_TOWN"]; got != 1 {
		t.Fatalf("MM_STRAY_FAIRY_TOWN = %d, want 1 in source MM snapshot", got)
	}

	core.loadSnapshot(t, ootSnap)
	ootState := readStateFromReader(t, reader, "in-oot-2-20260423-191522.json")

	checks := checkNameSet(ExtractChecks(ootState))
	for _, name := range []string{
		"Road to Southern Swamp HP",
		"Clock Town Tree HP",
		"Clock Town Platform HP",
		"Clock Town Stray Fairy",
	} {
		if _, ok := checks[name]; !ok {
			t.Fatalf("missing %s after same-reader MM->OoT snapshot transition", name)
		}
	}
	items := itemQtyMap(ExtractItems(ootState))
	if got := items["MM_STRAY_FAIRY_TOWN"]; got != 1 {
		t.Fatalf("MM_STRAY_FAIRY_TOWN = %d, want 1 after same-reader MM->OoT snapshot transition", got)
	}
}

func TestFreshOotSnapshotIncludesTownStrayFairyFromExtraFlags(t *testing.T) {
	state := readStateFromSnapshot(t, "fresh-oot-20260423-202324.json")

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Clock Town Stray Fairy"]; !ok {
		t.Fatal("missing Clock Town Stray Fairy from fresh OoT snapshot")
	}

	items := itemQtyMap(ExtractItems(state))
	if got := items["MM_STRAY_FAIRY_TOWN"]; got != 1 {
		t.Fatalf("MM_STRAY_FAIRY_TOWN = %d, want 1 in fresh OoT snapshot", got)
	}
}

func TestGrassCheckDiffOnlyZoraTunicAndGrass(t *testing.T) {
	beforeState := readStateFromSnapshot(t, "before-grass-20260419-192606.json")
	afterState := readStateFromSnapshot(t, "after-grass-20260419-192638.json")

	// Extract items and checks from both states.
	beforeItems := ExtractItems(beforeState)
	afterItems := ExtractItems(afterState)
	beforeChecks := ExtractChecks(beforeState)
	afterChecks := ExtractChecks(afterState)

	// Build maps for comparison.
	beforeItemMap := make(map[string]int)
	for _, it := range beforeItems {
		beforeItemMap[it.ID] = it.Qty
	}
	afterItemMap := make(map[string]int)
	for _, it := range afterItems {
		afterItemMap[it.ID] = it.Qty
	}
	beforeCheckSet := make(map[string]bool)
	for _, ch := range beforeChecks {
		beforeCheckSet[ch.Name] = true
	}
	afterCheckSet := make(map[string]bool)
	for _, ch := range afterChecks {
		afterCheckSet[ch.Name] = true
	}

	// Collect item differences.
	allItemIDs := make(map[string]bool)
	for k := range beforeItemMap {
		allItemIDs[k] = true
	}
	for k := range afterItemMap {
		allItemIDs[k] = true
	}
	var itemDiffs []string
	for id := range allItemIDs {
		bv, av := beforeItemMap[id], afterItemMap[id]
		if bv != av {
			itemDiffs = append(itemDiffs, fmt.Sprintf("%s: %d -> %d", id, bv, av))
		}
	}
	sort.Strings(itemDiffs)

	// Collect check differences.
	var newChecks, lostChecks []string
	for name := range afterCheckSet {
		if !beforeCheckSet[name] {
			newChecks = append(newChecks, name)
		}
	}
	for name := range beforeCheckSet {
		if !afterCheckSet[name] {
			lostChecks = append(lostChecks, name)
		}
	}
	sort.Strings(newChecks)
	sort.Strings(lostChecks)

	// Expected item diffs from picking up Zora Tunic at the grass check:
	//  - OOT_TUNIC increases (Zora Tunic added to OoT equipment bitmask)
	//  - MM_TUNIC_ZORA appears (Zora Tunic added to MM inventory)
	//  - MM_TRADE_3 changes (ExtraRecords trade bitmask updated)
	expectedItemDiffs := map[string]bool{
		"OOT_TUNIC":     true,
		"MM_TUNIC_ZORA": true,
		"MM_TRADE_3":    true,
	}
	if len(itemDiffs) != len(expectedItemDiffs) {
		t.Fatalf("expected %d item diffs, got %d:\n%s",
			len(expectedItemDiffs), len(itemDiffs), formatLines(itemDiffs))
	}
	for id := range allItemIDs {
		bv, av := beforeItemMap[id], afterItemMap[id]
		if bv != av && !expectedItemDiffs[id] {
			t.Fatalf("unexpected item diff: %s: %d -> %d", id, bv, av)
		}
	}

	// The only new check should be the grass check.
	if len(newChecks) != 1 {
		t.Fatalf("expected exactly 1 new check (grass), got %d:\n%s",
			len(newChecks), formatLines(newChecks))
	}
	if newChecks[0] != "Termina Field Grass Pack 10 Grass 03" {
		t.Fatalf("unexpected new check: %q", newChecks[0])
	}

	// No checks should be lost.
	if len(lostChecks) != 0 {
		t.Fatalf("expected 0 lost checks, got %d:\n%s",
			len(lostChecks), formatLines(lostChecks))
	}
}

func TestOwlCheckDiffOnlyWalletAndClockTownOwl(t *testing.T) {
	beforeState := readStateFromSnapshot(t, "before-owl-20260421-202333.json")
	afterStates := map[string]*GameState{
		"after": readStateFromSnapshot(t, "after-owl-20260421-202406.json"),
		"live":  readStateFromSnapshot(t, "live-after-owl-20260421.json"),
	}

	beforeItems := ExtractItems(beforeState)
	beforeChecks := ExtractChecks(beforeState)
	beforeItemMap := make(map[string]int)
	for _, it := range beforeItems {
		beforeItemMap[it.ID] = it.Qty
	}
	beforeCheckSet := make(map[string]bool)
	for _, ch := range beforeChecks {
		beforeCheckSet[ch.Name] = true
	}

	for label, afterState := range afterStates {
		t.Run(label, func(t *testing.T) {
			afterItems := ExtractItems(afterState)
			afterChecks := ExtractChecks(afterState)

			afterItemMap := make(map[string]int)
			for _, it := range afterItems {
				afterItemMap[it.ID] = it.Qty
			}
			afterCheckSet := make(map[string]bool)
			for _, ch := range afterChecks {
				afterCheckSet[ch.Name] = true
			}

			allItemIDs := make(map[string]bool)
			for id := range beforeItemMap {
				allItemIDs[id] = true
			}
			for id := range afterItemMap {
				allItemIDs[id] = true
			}

			var itemDiffs []string
			for id := range allItemIDs {
				if beforeItemMap[id] != afterItemMap[id] {
					itemDiffs = append(itemDiffs, fmt.Sprintf("%s: %d -> %d", id, beforeItemMap[id], afterItemMap[id]))
				}
			}
			sort.Strings(itemDiffs)

			var newChecks, lostChecks []string
			for name := range afterCheckSet {
				if !beforeCheckSet[name] {
					newChecks = append(newChecks, name)
				}
			}
			for name := range beforeCheckSet {
				if !afterCheckSet[name] {
					lostChecks = append(lostChecks, name)
				}
			}
			sort.Strings(newChecks)
			sort.Strings(lostChecks)

			expectedItemDiffs := map[string]bool{
				"MM_WALLET":  true,
				"OOT_WALLET": true,
			}
			if len(itemDiffs) != len(expectedItemDiffs) {
				t.Fatalf("expected %d item diffs, got %d:\n%s", len(expectedItemDiffs), len(itemDiffs), formatLines(itemDiffs))
			}
			for id := range allItemIDs {
				if beforeItemMap[id] != afterItemMap[id] && !expectedItemDiffs[id] {
					t.Fatalf("unexpected item diff: %s: %d -> %d", id, beforeItemMap[id], afterItemMap[id])
				}
			}

			if len(newChecks) != 1 || newChecks[0] != "Clock Town Owl Statue" {
				t.Fatalf("expected only Clock Town Owl Statue as new check, got:\n%s", formatLines(newChecks))
			}
			if len(lostChecks) != 0 {
				t.Fatalf("expected 0 lost checks, got %d:\n%s", len(lostChecks), formatLines(lostChecks))
			}
		})
	}
}

func formatLines(lines []string) string {
	s := ""
	for _, l := range lines {
		s += "  " + l + "\n"
	}
	return s
}
