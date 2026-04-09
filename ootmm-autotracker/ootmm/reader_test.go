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
