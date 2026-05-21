package n64

import (
	"bytes"
	"errors"
	"testing"
)

type probeTestCoreReader struct {
	err error
}

func (p *probeTestCoreReader) ReadMemory(addr uint32, size int) ([]byte, error) {
	return p.ReadMemoryLarge(addr, size)
}

func (p *probeTestCoreReader) ReadMemoryLarge(addr uint32, size int) ([]byte, error) {
	if p.err != nil {
		return nil, p.err
	}
	return make([]byte, size), nil
}

func TestAlignedWordRead(t *testing.T) {
	start, size := alignedWordRead(0x74, 24)
	if start != 0x74 {
		t.Fatalf("start = 0x%x, want 0x74", start)
	}
	if size != 24 {
		t.Fatalf("size = %d, want 24", size)
	}

	start, size = alignedWordRead(0x75, 1)
	if start != 0x74 {
		t.Fatalf("unaligned start = 0x%x, want 0x74", start)
	}
	if size != 4 {
		t.Fatalf("unaligned size = %d, want 4", size)
	}
}

func TestUnswizzleRangeAligned(t *testing.T) {
	raw := []byte{0xff, 0xff, 0x01, 0x00, 0x08, 0xff, 0xff, 0xff}
	data, err := unswizzleRange(raw, 0x74, 0x74, 8)
	if err != nil {
		t.Fatalf("unswizzleRange returned error: %v", err)
	}

	want := []byte{0x00, 0x01, 0xff, 0xff, 0xff, 0xff, 0xff, 0x08}
	if !bytes.Equal(data, want) {
		t.Fatalf("aligned unswizzle = % x, want % x", data, want)
	}
}

func TestUnswizzleRangeUnaligned(t *testing.T) {
	raw := []byte{0xff, 0xff, 0x01, 0x00}
	data, err := unswizzleRange(raw, 0x74, 0x75, 1)
	if err != nil {
		t.Fatalf("unswizzleRange returned error: %v", err)
	}

	want := []byte{0x01}
	if !bytes.Equal(data, want) {
		t.Fatalf("unaligned unswizzle = % x, want % x", data, want)
	}
}

func TestProbeReturnsReadErrorWhenAllReadsFail(t *testing.T) {
	backendErr := errors.New("backend read failed")
	mem := NewMemory(&probeTestCoreReader{err: backendErr})
	mem.SetSwizzle(false)

	err := mem.Probe()
	if err == nil {
		t.Fatal("Probe() returned nil, want error")
	}
	if !IsProbeReadError(err) {
		t.Fatalf("Probe() error classification = false, want true: %v", err)
	}
	if !errors.Is(err, backendErr) {
		t.Fatalf("Probe() error = %v, want wrapped backend error %v", err, backendErr)
	}
}

func TestProbeFixedModeReturnsReadErrorWhenAllReadsFail(t *testing.T) {
	backendErr := errors.New("backend read failed")
	mem := NewMemory(&probeTestCoreReader{err: backendErr})
	mem.SetSwizzle(false)
	mem.SetBaseShift(VirtualBase)

	err := mem.Probe()
	if err == nil {
		t.Fatal("Probe() returned nil, want error")
	}
	if !IsProbeReadError(err) {
		t.Fatalf("Probe() error classification = false, want true: %v", err)
	}
	if !errors.Is(err, backendErr) {
		t.Fatalf("Probe() error = %v, want wrapped backend error %v", err, backendErr)
	}
}

func TestProbeReturnsNotDetectedWhenReadsSucceedWithoutOoTMM(t *testing.T) {
	mem := NewMemory(&probeTestCoreReader{})
	mem.SetSwizzle(false)

	err := mem.Probe()
	if err == nil {
		t.Fatal("Probe() returned nil, want error")
	}
	if IsProbeReadError(err) {
		t.Fatalf("Probe() error classification = true, want false: %v", err)
	}
}
