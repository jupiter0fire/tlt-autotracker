package n64

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/ootmm-autotracker/retroarch"
)

const (
	// Mupen64Plus-Next uses physical addressing (virtual 0x80xxxxxx → physical 0x00xxxxxx).
	VirtualBase  uint32 = 0x80000000
	PhysicalBase uint32 = 0x00000000
)

// Memory provides N64-specific memory access through a RetroArch client.
// It probes the address space at startup and then translates virtual addresses.
type Memory struct {
	client    *retroarch.Client
	baseShift uint32 // 0 for physical, 0x80000000 for virtual
}

func NewMemory(client *retroarch.Client) *Memory {
	return &Memory{
		client:    client,
		baseShift: 0, // default to physical (Mupen64Plus-Next)
	}
}

// Probe detects whether the N64 core uses physical or virtual addressing
// by looking for the OoTMM ComboContext magic string "OoT+MM<3".
// OoTMM's Context_Init clears the read address after startup, so we check
// both ComboCtx addresses (OoT and MM) and fall back to checking the
// payload area at 0x80400000 which is only populated by OoTMM.
func (m *Memory) Probe() error {
	magic := []byte("OoT+MM<3")

	// ComboCtx addresses: OoT side and MM side
	ctxAddrs := []uint32{0x00006584, 0x00098280}

	// Try physical first (most common for Mupen64Plus-Next)
	for _, addr := range ctxAddrs {
		data, err := m.client.ReadMemory(addr, 8)
		if err == nil && string(data) == string(magic) {
			m.baseShift = 0
			log.Println("N64 core uses physical addressing")
			return nil
		}
	}

	// Try virtual
	for _, addr := range ctxAddrs {
		data, err := m.client.ReadMemory(addr|VirtualBase, 8)
		if err == nil && string(data) == string(magic) {
			m.baseShift = VirtualBase
			log.Println("N64 core uses virtual addressing")
			return nil
		}
	}

	// Fallback: OoTMM loads a payload into the Expansion Pack area (0x80400000).
	// Regular OoT doesn't use this region, so non-zero data here indicates OoTMM.
	payloadAddr := uint32(0x00400000)
	data, err := m.client.ReadMemory(payloadAddr, 8)
	if err == nil && !isAllZero(data) {
		m.baseShift = 0
		log.Println("N64 core uses physical addressing (detected via payload)")
		return nil
	}

	data, err = m.client.ReadMemory(payloadAddr|VirtualBase, 8)
	if err == nil && !isAllZero(data) {
		m.baseShift = VirtualBase
		log.Println("N64 core uses virtual addressing (detected via payload)")
		return nil
	}

	return fmt.Errorf("unable to detect N64 address space (OoTMM not detected)")
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
	return m.client.ReadMemoryLarge(coreAddr, size)
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
