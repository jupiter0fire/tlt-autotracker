package ootmm

import (
	"fmt"
	"testing"

	"ootmm-autotracker/n64"
)

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

func TestReadRawFrameForStableStateExportsNamedChunks(t *testing.T) {
	specs := []RawChunkSpec{
		{Name: "oot_save_ctx", Address: AddrOotSaveCtx, Length: maxInt(OotSaveSize, ootSaveCtxUsedSize)},
		{Name: "oot_foreign_mm_save", Address: 0x80443970, Length: MmSaveSize},
		{Name: "oot_shared_custom_save", Address: 0x80443100, Length: int(SharedCustomSaveSize)},
		{Name: "oot_runtime_combo_config", Address: 0x804416c8, Length: OotComboConfigSize},
		{Name: "oot_runtime_silver_rupee_data", Address: 0x8042ec10, Length: 72},
		{Name: "oot_runtime_max_keys", Address: 0x80441c78, Length: 21},
		{Name: "oot_playstate_core", Address: 0x801c8544, Length: ootRawPlayStateCoreSize},
		{Name: "oot_playstate_tail", Address: 0x801da160, Length: ootRawPlayStateTailSize},
	}
	regions := []snapshotFixtureRegion{
		{address: specs[0].Address, data: filledBytes(specs[0].Length, 0x30)},
		{address: specs[1].Address, data: filledBytes(specs[1].Length, 0x40)},
		{address: specs[2].Address, data: filledBytes(specs[2].Length, 0x50)},
		{address: specs[3].Address, data: filledBytes(specs[3].Length, 0x60)},
		{address: specs[4].Address, data: filledBytes(specs[4].Length, 0x70)},
		{address: specs[5].Address, data: filledBytes(specs[5].Length, 0x80)},
		{address: specs[6].Address, data: filledBytes(specs[6].Length, 0x90)},
		{address: specs[7].Address, data: filledBytes(specs[7].Length, 0xA0)},
	}

	mem := n64.NewMemory(&snapshotFixtureCoreReader{regions: regions})
	mem.SetBaseShift(n64.VirtualBase)
	mem.SetSwizzle(false)

	reader := NewReader(mem)
	frame, err := reader.ReadRawFrameForStableState(GameOot, 2, specs)
	if err != nil {
		t.Fatalf("ReadRawFrameForStableState returned error: %v", err)
	}
	if !frame.Valid {
		t.Fatal("expected raw frame to be valid")
	}
	if frame.ActiveGame != GameOot {
		t.Fatalf("raw frame active game = %v, want %v", frame.ActiveGame, GameOot)
	}
	if frame.SaveIndex != 2 {
		t.Fatalf("raw frame save index = %d, want 2", frame.SaveIndex)
	}

	if len(frame.Chunks) != len(specs) {
		t.Fatalf("chunk count = %d, want %d", len(frame.Chunks), len(specs))
	}

	for index, spec := range specs {
		chunk := frame.Chunks[index]
		if chunk.Name != spec.Name {
			t.Fatalf("chunk[%d] name = %q, want %q", index, chunk.Name, spec.Name)
		}
		if chunk.Address != spec.Address {
			t.Fatalf("chunk[%d] address = %#x, want %#x", index, chunk.Address, spec.Address)
		}
		if chunk.Length != spec.Length {
			t.Fatalf("chunk[%d] length = %d, want %d", index, chunk.Length, spec.Length)
		}
		if len(chunk.Data) != spec.Length {
			t.Fatalf("chunk[%d] data length = %d, want %d", index, len(chunk.Data), spec.Length)
		}
	}
}

func TestReadRawFrameForStableStateExportsNamedChunksForMm(t *testing.T) {
	specs := []RawChunkSpec{
		{Name: "mm_save_ctx", Address: AddrMmSaveCtx, Length: maxInt(MmSaveSize, mmSaveCtxUsedSize)},
		{Name: "mm_foreign_oot_save", Address: 0x807729f0, Length: OotSaveSize},
		{Name: "mm_shared_custom_save", Address: 0x80772180, Length: int(SharedCustomSaveSize)},
		{Name: "mm_runtime_combo_config", Address: 0x80770b18, Length: OotComboConfigSize},
		{Name: "mm_playstate_core", Address: 0x803e6bc4, Length: mmRawPlayStateCoreSize},
		{Name: "mm_playstate_tail", Address: 0x8041f220, Length: mmRawPlayStateTailSize},
	}
	regions := []snapshotFixtureRegion{
		{address: specs[0].Address, data: filledBytes(specs[0].Length, 0x30)},
		{address: specs[1].Address, data: filledBytes(specs[1].Length, 0x40)},
		{address: specs[2].Address, data: filledBytes(specs[2].Length, 0x50)},
		{address: specs[3].Address, data: filledBytes(specs[3].Length, 0x60)},
		{address: specs[4].Address, data: filledBytes(specs[4].Length, 0x90)},
		{address: specs[5].Address, data: filledBytes(specs[5].Length, 0xA0)},
	}

	mem := n64.NewMemory(&snapshotFixtureCoreReader{regions: regions})
	mem.SetBaseShift(n64.VirtualBase)
	mem.SetSwizzle(false)

	reader := NewReader(mem)
	frame, err := reader.ReadRawFrameForStableState(GameMm, 2, specs)
	if err != nil {
		t.Fatalf("ReadRawFrameForStableState returned error: %v", err)
	}
	if !frame.Valid {
		t.Fatal("expected raw frame to be valid")
	}
	if frame.ActiveGame != GameMm {
		t.Fatalf("raw frame active game = %v, want %v", frame.ActiveGame, GameMm)
	}
	if frame.SaveIndex != 2 {
		t.Fatalf("raw frame save index = %d, want 2", frame.SaveIndex)
	}

	if len(frame.Chunks) != len(specs) {
		t.Fatalf("chunk count = %d, want %d", len(frame.Chunks), len(specs))
	}

	for index, spec := range specs {
		chunk := frame.Chunks[index]
		if chunk.Name != spec.Name {
			t.Fatalf("chunk[%d] name = %q, want %q", index, chunk.Name, spec.Name)
		}
		if chunk.Address != spec.Address {
			t.Fatalf("chunk[%d] address = %#x, want %#x", index, chunk.Address, spec.Address)
		}
		if chunk.Length != spec.Length {
			t.Fatalf("chunk[%d] length = %d, want %d", index, chunk.Length, spec.Length)
		}
		if len(chunk.Data) != spec.Length {
			t.Fatalf("chunk[%d] data length = %d, want %d", index, len(chunk.Data), spec.Length)
		}
	}
}

func filledBytes(length int, value byte) []byte {
	data := make([]byte, length)
	for index := range data {
		data[index] = value
	}
	return data
}
