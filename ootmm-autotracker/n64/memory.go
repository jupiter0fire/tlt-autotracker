package n64

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
)

const (
	// Mupen64Plus-Next uses physical addressing (virtual 0x80xxxxxx → physical 0x00xxxxxx).
	VirtualBase  uint32 = 0x80000000
	PhysicalBase uint32 = 0x00000000
)

// CoreReader is the interface for reading emulator core memory.
// Both retroarch.Client and pj64.Server implement this.
type CoreReader interface {
	ReadMemory(addr uint32, size int) ([]byte, error)
	ReadMemoryLarge(addr uint32, size int) ([]byte, error)
}

// Memory provides N64-specific memory access through an emulator backend.
// It probes the address space at startup and then translates virtual addresses.
type Memory struct {
	client         CoreReader
	baseShift      uint32 // 0 for physical, 0x80000000 for virtual
	swizzle        bool   // true for RetroArch/mupen64plus word-swizzle, false for PJ64
	fixedBaseShift bool   // true when SetBaseShift() was called; Probe() won't override
}

type probeReadError struct {
	err error
}

func (e *probeReadError) Error() string {
	return fmt.Sprintf("probe read: %v", e.err)
}

func (e *probeReadError) Unwrap() error {
	return e.err
}

func IsProbeReadError(err error) bool {
	var target *probeReadError
	return errors.As(err, &target)
}

func NewMemory(client CoreReader) *Memory {
	return &Memory{
		client:    client,
		baseShift: 0, // default to physical (Mupen64Plus-Next)
		swizzle:   true,
	}
}

// SetSwizzle controls whether byte-level word unswizzling is applied.
// RetroArch/mupen64plus stores RDRAM in host byte order (needs swizzle=true).
// PJ64 returns bytes in N64 native order (needs swizzle=false).
func (m *Memory) SetSwizzle(s bool) {
	m.swizzle = s
}

// SetBaseShift forces a specific address translation mode.
// Use 0 for physical addressing (Mupen64Plus-Next), VirtualBase for virtual (PJ64).
// When set, Probe() will not override the base shift.
func (m *Memory) SetBaseShift(shift uint32) {
	m.baseShift = shift
	m.fixedBaseShift = true
}

// Probe detects whether the N64 core uses physical or virtual addressing
// by looking for the OoTMM ComboContext magic string "OoT+MM<3".
// OoTMM's Context_Init clears the read address after startup, so we check
// both ComboCtx addresses (OoT and MM) and fall back to checking the
// payload area at 0x80400000 which is only populated by OoTMM.
func (m *Memory) Probe() error {
	var (
		hadSuccessfulRead bool
		lastReadErr       error
	)

	noteReadResult := func(err error) {
		if err == nil {
			hadSuccessfulRead = true
			return
		}
		lastReadErr = err
	}

	// When the address mode has been forced (e.g. PJ64), skip detection
	// and only verify that OoTMM is actually running.
	if m.fixedBaseShift {
		return m.probeFixedMode()
	}

	magic := []byte("OoT+MM<3")

	// ComboCtx addresses: OoT side and MM side
	ctxAddrs := []uint32{0x00006584, 0x00098280}

	// Try physical first (most common for Mupen64Plus-Next)
	for _, addr := range ctxAddrs {
		data, err := m.readCoreLogical(addr, 8)
		noteReadResult(err)
		if err == nil && string(data) == string(magic) {
			m.baseShift = 0
			log.Println("N64 core uses physical addressing")
			return nil
		}
	}

	// Try virtual
	for _, addr := range ctxAddrs {
		data, err := m.readCoreLogical(addr|VirtualBase, 8)
		noteReadResult(err)
		if err == nil && string(data) == string(magic) {
			m.baseShift = VirtualBase
			log.Println("N64 core uses virtual addressing")
			return nil
		}
	}

	// Fallback: OoTMM loads a payload into the Expansion Pack area (0x80400000).
	// Regular OoT doesn't use this region, so non-zero data here indicates OoTMM.
	payloadAddr := uint32(0x00400000)
	data, err := m.readCoreLogical(payloadAddr, 8)
	noteReadResult(err)
	if err == nil && !isAllZero(data) {
		m.baseShift = 0
		log.Println("N64 core uses physical addressing (detected via payload)")
		return nil
	}

	data, err = m.readCoreLogical(payloadAddr|VirtualBase, 8)
	noteReadResult(err)
	if err == nil && !isAllZero(data) {
		m.baseShift = VirtualBase
		log.Println("N64 core uses virtual addressing (detected via payload)")
		return nil
	}

	if !hadSuccessfulRead && lastReadErr != nil {
		return &probeReadError{err: lastReadErr}
	}

	return fmt.Errorf("unable to detect N64 address space (OoTMM not detected)")
}

// probeFixedMode checks for OoTMM presence when the addressing mode is already known.
func (m *Memory) probeFixedMode() error {
	var (
		hadSuccessfulRead bool
		lastReadErr       error
	)

	noteReadResult := func(err error) {
		if err == nil {
			hadSuccessfulRead = true
			return
		}
		lastReadErr = err
	}

	magic := []byte("OoT+MM<3")
	mode := "virtual"
	if m.baseShift == 0 {
		mode = "physical"
	}

	// Check ComboCtx addresses using the pre-set base shift
	for _, virtAddr := range []uint32{0x80006584, 0x80098280} {
		data, err := m.Read(virtAddr, 8)
		noteReadResult(err)
		if err == nil && string(data) == string(magic) {
			log.Printf("N64 core uses %s addressing (fixed)", mode)
			return nil
		}
	}

	// Fallback: check payload area
	data, err := m.Read(0x80400000, 8)
	noteReadResult(err)
	if err == nil && !isAllZero(data) {
		log.Printf("N64 core uses %s addressing (fixed, detected via payload)", mode)
		return nil
	}

	if !hadSuccessfulRead && lastReadErr != nil {
		return &probeReadError{err: lastReadErr}
	}

	return fmt.Errorf("unable to detect OoTMM (address mode: %s)", mode)
}

func isAllZero(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// translate converts a virtual N64 address (0x80xxxxxx) to the core's address space.
func (m *Memory) translate(virtualAddr uint32) uint32 {
	// Strip virtual prefix if present, then add baseShift
	physical := virtualAddr &^ VirtualBase
	return physical | m.baseShift
}

// Read reads `size` bytes from a virtual N64 address.
func (m *Memory) Read(virtualAddr uint32, size int) ([]byte, error) {
	coreAddr := m.translate(virtualAddr)
	return m.readCoreLogical(coreAddr, size)
}

// ReadU8 reads a single byte.
func (m *Memory) ReadU8(virtualAddr uint32) (uint8, error) {
	data, err := m.Read(virtualAddr, 1)
	if err != nil {
		return 0, err
	}
	return data[0], nil
}

// ReadU16BE reads a big-endian uint16 (N64 is big-endian).
func (m *Memory) ReadU16BE(virtualAddr uint32) (uint16, error) {
	data, err := m.Read(virtualAddr, 2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(data), nil
}

// ReadU32BE reads a big-endian uint32.
func (m *Memory) ReadU32BE(virtualAddr uint32) (uint32, error) {
	data, err := m.Read(virtualAddr, 4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(data), nil
}

func (m *Memory) readCoreLogical(coreAddr uint32, size int) ([]byte, error) {
	if !m.swizzle {
		// PJ64 returns bytes in N64 native order; no alignment or unswizzle needed.
		return m.client.ReadMemoryLarge(coreAddr, size)
	}
	alignedStart, alignedSize := alignedWordRead(coreAddr, size)
	raw, err := m.client.ReadMemoryLarge(alignedStart, alignedSize)
	if err != nil {
		return nil, err
	}
	return unswizzleRange(raw, alignedStart, coreAddr, size)
}

func alignedWordRead(coreAddr uint32, size int) (uint32, int) {
	alignedStart := coreAddr &^ 3
	end := coreAddr + uint32(size)
	alignedEnd := (end + 3) &^ 3
	return alignedStart, int(alignedEnd - alignedStart)
}

func unswizzleRange(raw []byte, rawBase uint32, logicalBase uint32, size int) ([]byte, error) {
	data := make([]byte, size)
	for index := 0; index < size; index++ {
		rawAddr := (logicalBase + uint32(index)) ^ 3
		rawIndex := int(rawAddr - rawBase)
		if rawIndex < 0 || rawIndex >= len(raw) {
			return nil, fmt.Errorf("unswizzle out of range: raw index %d for base 0x%x size %d", rawIndex, rawBase, len(raw))
		}
		data[index] = raw[rawIndex]
	}
	return data, nil
}
