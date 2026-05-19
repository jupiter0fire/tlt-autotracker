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
		{address: AddrComboCtxOot, data: filledBytes(ComboCtxSize, 0x10)},
		{address: AddrComboCtxMm, data: filledBytes(ComboCtxSize, 0x20)},
		{address: AddrOotSaveCtx, data: filledBytes(OotSaveCtxSize, 0x30)},
		{address: AddrMmSaveCtx, data: filledBytes(MmSaveCtxSize, 0x40)},
		{address: AddrOotPayload, data: filledBytes(OotPayloadSize, 0x50)},
		{address: AddrMmPayload, data: filledBytes(MmPayloadSize, 0x60)},
	}

	for index, addr := range ootPlayStateCandidateAddrs {
		regions = append(regions,
			snapshotFixtureRegion{
				address: addr + uint32(OotPlayOffSceneID),
				data:    filledBytes(ootRawPlayStateCoreSize, byte(0x70+index)),
			},
			snapshotFixtureRegion{
				address: addr + uint32(OotPlayOffCurrentRoom),
				data:    filledBytes(ootRawPlayStateTailSize, byte(0x80+index)),
			},
		)
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

	specs := rawChunkSpecs()
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
