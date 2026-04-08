package ootmm

import "testing"

func TestLocateForeignSaveFindsAlignedCandidate(t *testing.T) {
	payload := make([]byte, 0x400)
	offset := 0x120
	copy(payload[offset+0x24:], []byte("ZELDA3"))

	addr, ok := locateForeignSave(payload, 0x80400000, 0x200, 0x24)
	if !ok {
		t.Fatal("expected foreign save candidate")
	}
	if want := uint32(0x80400120); addr != want {
		t.Fatalf("unexpected candidate addr: got %08x want %08x", addr, want)
	}
}

func TestLocateForeignSaveRejectsUnalignedCandidate(t *testing.T) {
	payload := make([]byte, 0x400)
	copy(payload[0x121+0x24:], []byte("ZELDA3"))

	if _, ok := locateForeignSave(payload, 0x80400000, 0x200, 0x24); ok {
		t.Fatal("expected unaligned candidate to be rejected")
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