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
	regions := []snapshotFixtureRegion{
		{address: AddrOotSaveCtx, data: filledBytes(maxInt(OotSaveSize, ootSaveCtxUsedSize), 0x30)},
		{address: AddrOotForeignMmSaveLive, data: filledBytes(MmSaveSize, 0x40)},
		{address: AddrOotSharedCustomSaveLive, data: filledBytes(sharedStateReadSize(), 0x50)},
		{address: AddrOotRuntimeOotComboConfigLive, data: filledBytes(OotComboConfigSize, 0x60)},
		{address: AddrOotRuntimeSilverRupeeDataLive, data: filledBytes(OotSilverRupeeDataSize, 0x70)},
		{address: AddrOotRuntimeMaxKeysLive, data: filledBytes(OotMaxKeysBlockSize, 0x80)},
	}
	for index, addr := range ootPlayStateCandidateAddrs {
		regions = append(regions,
			snapshotFixtureRegion{
				address: addr + uint32(OotPlayOffSceneID),
				data:    filledBytes(ootRawPlayStateCoreSize, byte(0x90+index)),
			},
			snapshotFixtureRegion{
				address: addr + uint32(OotPlayOffCurrentRoom),
				data:    filledBytes(ootRawPlayStateTailSize, byte(0xA0+index)),
			},
		)
	}

	mem := n64.NewMemory(&snapshotFixtureCoreReader{regions: regions})
	mem.SetBaseShift(n64.VirtualBase)
	mem.SetSwizzle(false)

	reader := NewReader(mem)
	frame, err := reader.ReadRawFrameForStableState(GameOot, 2)
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

	specs := append(rawChunkSpecs(GameOot), reader.selectedPlayStateChunkSpecs(GameOot)...)
	if len(frame.Chunks) != len(specs) {
		t.Fatalf("chunk count = %d, want %d", len(frame.Chunks), len(specs))
	}

	for index, spec := range specs {
		chunk := frame.Chunks[index]
		if chunk.Name != spec.name {
			t.Fatalf("chunk[%d] name = %q, want %q", index, chunk.Name, spec.name)
		}
		if chunk.Address != spec.address {
			t.Fatalf("chunk[%d] address = %#x, want %#x", index, chunk.Address, spec.address)
		}
		if chunk.Length != spec.length {
			t.Fatalf("chunk[%d] length = %d, want %d", index, chunk.Length, spec.length)
		}
		if len(chunk.Data) != spec.length {
			t.Fatalf("chunk[%d] data length = %d, want %d", index, len(chunk.Data), spec.length)
		}
	}
}

func TestReadRawFrameForStableStateExportsNamedChunksForMm(t *testing.T) {
	regions := []snapshotFixtureRegion{
		{address: AddrMmSaveCtx, data: filledBytes(maxInt(MmSaveSize, mmSaveCtxUsedSize), 0x30)},
		{address: AddrMmForeignOotSaveLive, data: filledBytes(OotSaveSize, 0x40)},
		{address: AddrMmSharedCustomSaveLive, data: filledBytes(sharedStateReadSize(), 0x50)},
		{address: AddrMmRuntimeOotComboConfigLive, data: filledBytes(OotComboConfigSize, 0x60)},
	}
	for index, addr := range mmPlayStateCandidateAddrs {
		regions = append(regions,
			snapshotFixtureRegion{
				address: addr + uint32(MmPlayOffSceneID),
				data:    filledBytes(mmRawPlayStateCoreSize, byte(0x90+index)),
			},
			snapshotFixtureRegion{
				address: addr + uint32(MmPlayOffCurrentRoom),
				data:    filledBytes(mmRawPlayStateTailSize, byte(0xA0+index)),
			},
		)
	}

	mem := n64.NewMemory(&snapshotFixtureCoreReader{regions: regions})
	mem.SetBaseShift(n64.VirtualBase)
	mem.SetSwizzle(false)

	reader := NewReader(mem)
	frame, err := reader.ReadRawFrameForStableState(GameMm, 2)
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

	specs := append(rawChunkSpecs(GameMm), reader.selectedPlayStateChunkSpecs(GameMm)...)
	if len(frame.Chunks) != len(specs) {
		t.Fatalf("chunk count = %d, want %d", len(frame.Chunks), len(specs))
	}

	for index, spec := range specs {
		chunk := frame.Chunks[index]
		if chunk.Name != spec.name {
			t.Fatalf("chunk[%d] name = %q, want %q", index, chunk.Name, spec.name)
		}
		if chunk.Address != spec.address {
			t.Fatalf("chunk[%d] address = %#x, want %#x", index, chunk.Address, spec.address)
		}
		if chunk.Length != spec.length {
			t.Fatalf("chunk[%d] length = %d, want %d", index, chunk.Length, spec.length)
		}
		if len(chunk.Data) != spec.length {
			t.Fatalf("chunk[%d] data length = %d, want %d", index, len(chunk.Data), spec.length)
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
