package ootmm

import (
	"encoding/binary"
	"testing"
)

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

func ootChecksum(data []byte) uint16 {
	checksum := uint16(0)
	for i := 0; i < OotSaveSize; i += 2 {
		if i == OotOffChecksum {
			continue
		}
		checksum += binary.BigEndian.Uint16(data[i:])
	}
	return checksum
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

	var mm MmState
	if err := parseMmSave(&mm, data); err != nil {
		t.Fatalf("parseMmSave: %v", err)
	}
	if !mm.TownStrayFairy {
		t.Fatal("expected MM town stray fairy to be true")
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

func TestIsPlausibleSharedStateRejectsUnknownFutureBits(t *testing.T) {
	shared := SharedCustomState{}
	for _, bitmap := range sharedStorage.Bitmaps {
		shared.SetBitmap(bitmap.Name, make([]byte, bitmap.Size))
	}

	bitmap := sharedBitmaps["soulsEnemyOot"]
	shared.SetBit(bitmap.Name, bitmap.Size*8-1)

	if isPlausibleSharedState(shared) {
		t.Fatal("expected shared state with future bits to fail plausibility")
	}
}

func TestValidateSilverRupeeDataAcceptsLiveLikeMqConfig(t *testing.T) {
	data := []byte{
		0x00, 0x01, 0x25, 0x05,
		0x00, 0x08, 0x1f, 0x05,
		0x00, 0x06, 0x05, 0x05,
		0x00, 0x06, 0x0a, 0x05,
		0x00, 0x06, 0x02, 0x05,
		0x00, 0x07, 0x01, 0x05,
		0x00, 0x07, 0x00, 0x00,
		0x00, 0x07, 0x09, 0x05,
		0x00, 0x07, 0x08, 0x05,
		0x00, 0x09, 0x08, 0x05,
		0x00, 0x09, 0x09, 0x05,
		0x00, 0x0b, 0x1c, 0x05,
		0x00, 0x0b, 0x0c, 0x06,
		0x00, 0x0b, 0x1b, 0x03,
		0x00, 0x0d, 0x0b, 0x05,
		0x00, 0x0d, 0x02, 0x05,
		0x00, 0x0d, 0x01, 0x05,
		0x00, 0x0d, 0x00, 0x00,
	}
	if !validateSilverRupeeData(data) {
		t.Fatal("expected live-like silver rupee metadata to validate")
	}
}

func TestLocateSilverRupeeDataFindsStructuredCandidate(t *testing.T) {
	payload := make([]byte, OotPayloadSize)
	off := 0x2ec10
	copy(payload[off:], []byte{
		0x00, 0x01, 0x25, 0x05,
		0x00, 0x08, 0x1f, 0x05,
		0x00, 0x06, 0x05, 0x05,
		0x00, 0x06, 0x0a, 0x05,
		0x00, 0x06, 0x02, 0x05,
		0x00, 0x07, 0x01, 0x05,
		0x00, 0x07, 0x00, 0x00,
		0x00, 0x07, 0x09, 0x05,
		0x00, 0x07, 0x08, 0x05,
		0x00, 0x09, 0x08, 0x05,
		0x00, 0x09, 0x09, 0x05,
		0x00, 0x0b, 0x1c, 0x05,
		0x00, 0x0b, 0x0c, 0x06,
		0x00, 0x0b, 0x1b, 0x03,
		0x00, 0x0d, 0x0b, 0x05,
		0x00, 0x0d, 0x02, 0x05,
		0x00, 0x0d, 0x01, 0x05,
		0x00, 0x0d, 0x00, 0x00,
	})

	got, ok := locateSilverRupeeData(payload)
	if !ok {
		t.Fatal("expected silver rupee metadata candidate")
	}
	if got != off {
		t.Fatalf("unexpected silver rupee offset: got %#x want %#x", got, off)
	}
}

func TestValidateOotMaxKeyBlockAcceptsLiveLikeConfig(t *testing.T) {
	data := []byte{0, 0, 0, 5, 5, 5, 5, 5, 3, 0, 0, 3, 4, 3, 0, 0, 6, 1, 3, 1, 4}
	if !validateOotMaxKeyBlock(data) {
		t.Fatal("expected live-like OoT max-key block to validate")
	}
}

func TestLocateOotMaxKeysFindsStructuredCandidate(t *testing.T) {
	payload := make([]byte, OotPayloadSize)
	off := 0x41c78
	copy(payload[off-8:], []byte{0x00, 0x00, 0x08, 0x01, 0x00, 0x01, 0x0b, 0x14})
	copy(payload[off:], []byte{0, 0, 0, 5, 5, 5, 5, 5, 3, 0, 0, 3, 4, 3, 0, 0, 6, 1, 3, 1, 4})
	copy(payload[off+OotMaxKeysBlockSize:], []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0x00, 0x00})

	got, ok := locateOotMaxKeys(payload)
	if !ok {
		t.Fatal("expected OoT max-key candidate")
	}
	if got != off {
		t.Fatalf("unexpected OoT max-key offset: got %#x want %#x", got, off)
	}
}
