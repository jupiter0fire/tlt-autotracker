package ootmm

import (
	"encoding/binary"
	"os"
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

func TestIsPlausibleOotSaveAllowsRicherInventory(t *testing.T) {
	data := make([]byte, OotSaveSize)
	binary.BigEndian.PutUint32(data[OotOffAge:], 1)
	binary.BigEndian.PutUint16(data[OotOffSceneID:], 0x20)
	for i := 0; i < 24; i++ {
		data[OotOffInvItems+i] = emptyInventoryItem
	}
	for i := 0; i < 17; i++ {
		data[OotOffInvItems+i] = byte(i)
	}

	if !isPlausibleOotSave(data) {
		t.Fatal("expected OoT save with 17 filled inventory slots to remain plausible")
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

func TestStabilizeSnapshotStateUsesSlotAnchoredForeignOotSave(t *testing.T) {
	foreignAddr, err := foreignOotSaveAddr(0)
	if err != nil {
		t.Fatalf("foreignOotSaveAddr(0): %v", err)
	}
	sharedAddr, err := sharedSaveAddr(AddrMmPayload, MmPayloadSize, 0)
	if err != nil {
		t.Fatalf("sharedSaveAddr(MM, 0): %v", err)
	}

	ootData := make([]byte, OotSaveSize)
	for i := 0; i < len(ootData[OotOffInvItems:OotOffInvItems+24]); i++ {
		ootData[OotOffInvItems+i] = emptyInventoryItem
	}
	for i := 0; i < 19; i++ {
		ootData[OotOffDungeonKeys+i] = 0xff
	}
	binary.BigEndian.PutUint32(ootData[OotOffAge:], 1)
	binary.BigEndian.PutUint16(ootData[OotOffSceneID:], 0x20)

	sharedData := make([]byte, sharedStateReadSize())

	mem := n64.NewMemory(&snapshotFixtureCoreReader{regions: []snapshotFixtureRegion{
		{address: foreignAddr, data: ootData},
		{address: sharedAddr, data: sharedData},
	}})
	mem.SetBaseShift(n64.VirtualBase)
	mem.SetSwizzle(false)

	r := NewReader(mem)
	state := &GameState{
		Valid:      true,
		ActiveGame: GameMm,
		SaveIndex:  0,
	}
	resetEmptyOotState(&state.Oot)
	state.Oot.SceneID = 0x52
	state.Oot.Items[mustOotInventorySlotIndex("OOT_BOOMERANG")] = 0x06
	state.Oot.Items[mustOotInventorySlotIndex("OOT_BOTTLE_4")] = 0x16
	state.Oot.ExtraRecords[ExtraIdxMmFlags2] = 0xfeedbeef
	state.Mm.ExtraFlags2 = 0xfeedbeef
	state.Shared.BombchuBagOot = 3

	r.StabilizeSnapshotState(state)

	items := itemQtyMap(ExtractItems(state))
	if got := items["OOT_BOOMERANG"]; got != 0 {
		t.Fatalf("OOT_BOOMERANG = %d, want 0", got)
	}
	if got := items["OOT_BOTTLE_4"]; got != 0 {
		t.Fatalf("OOT_BOTTLE_4 = %d, want 0", got)
	}
	if state.Mm.ExtraFlags2 != 0 {
		t.Fatalf("MM extra flags 2 = %#x, want 0", state.Mm.ExtraFlags2)
	}
	if state.Shared.BombchuBagOot != 0 {
		t.Fatalf("shared OoT bombchu bag = %d, want 0", state.Shared.BombchuBagOot)
	}
	for index, itemID := range state.Oot.Items {
		if itemID != emptyInventoryItem {
			t.Fatalf("unexpected stabilized OoT item %d: got %#02x want %#02x", index, itemID, emptyInventoryItem)
		}
	}
	if state.Oot.SceneID != 0x20 {
		t.Fatalf("stabilized OoT scene = %#x, want 0x20", state.Oot.SceneID)
	}
	if state.Oot.Age != 1 {
		t.Fatalf("stabilized OoT age = %d, want 1", state.Oot.Age)
	}
	if state.Oot.HasRuntimeMqBits {
		t.Fatal("expected snapshot-stabilized OoT state to keep runtime MQ bits disabled without combo config data")
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

func TestReadForeignOotSavePrefersChecksummedCandidateOverWeakCachedAddress(t *testing.T) {
	payload := make([]byte, MmPayloadSize)
	weakOffset := 0x408a0
	weak := payload[weakOffset : weakOffset+OotSaveSize]
	binary.BigEndian.PutUint32(weak[OotOffAge:], 0)
	binary.BigEndian.PutUint16(weak[OotOffSceneID:], 0x00)
	for i := 0; i < 24; i++ {
		weak[OotOffInvItems+i] = emptyInventoryItem
	}
	for i := 0; i < 16; i++ {
		weak[OotOffInvItems+i] = byte(i)
	}

	exactOffset := 0x429f0
	exact := payload[exactOffset : exactOffset+OotSaveSize]
	binary.BigEndian.PutUint32(exact[OotOffAge:], 1)
	binary.BigEndian.PutUint16(exact[OotOffSceneID:], 0x20)
	binary.BigEndian.PutUint16(exact[OotOffGoldTokens:], 31)
	binary.BigEndian.PutUint32(exact[OotOffUpgrades:], 0x00120089)
	for i := 0; i < 24; i++ {
		exact[OotOffInvItems+i] = emptyInventoryItem
	}
	for i := 0; i < 17; i++ {
		exact[OotOffInvItems+i] = byte(i)
	}
	binary.BigEndian.PutUint16(exact[OotOffChecksum:], ootChecksum(exact))

	mem := n64.NewMemory(&snapshotFixtureCoreReader{regions: []snapshotFixtureRegion{{address: AddrMmPayload, data: payload}}})
	mem.SetBaseShift(n64.VirtualBase)
	mem.SetSwizzle(false)

	r := NewReader(mem)
	r.foreignOotSaveAddr = AddrMmPayload + uint32(weakOffset)

	if err := r.validateForeignOotSaveAt(r.foreignOotSaveAddr, weak); err != nil {
		t.Fatalf("expected weak cached candidate to remain prefix-valid: %v", err)
	}

	var oot OotState
	if err := r.readForeignOotSave(&oot); err != nil {
		t.Fatalf("readForeignOotSave: %v", err)
	}

	if want := AddrMmPayload + uint32(exactOffset); r.foreignOotSaveAddr != want {
		t.Fatalf("foreignOotSaveAddr = %08x, want %08x", r.foreignOotSaveAddr, want)
	}
	if oot.SceneID != 0x20 {
		t.Fatalf("OoT scene = %#x, want 0x20", oot.SceneID)
	}
	if got := oot.Items[16]; got != 0x10 {
		t.Fatalf("OoT item 16 = %#02x, want 0x10", got)
	}
	if oot.GoldTokens != 31 {
		t.Fatalf("OoT gold tokens = %d, want 31", oot.GoldTokens)
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
