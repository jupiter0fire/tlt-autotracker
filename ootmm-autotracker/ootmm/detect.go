package ootmm

import (
	"encoding/binary"
	"fmt"

	"github.com/ootmm-autotracker/n64"
)

// ActiveGame indicates which game is currently running.
type ActiveGame int

const (
	GameNone ActiveGame = iota
	GameOot
	GameMm
)

func (g ActiveGame) String() string {
	switch g {
	case GameOot:
		return "OoT"
	case GameMm:
		return "MM"
	default:
		return "None"
	}
}

// Detector detects whether OoTMM is running and which game is active.
type Detector struct {
	mem *n64.Memory
}

func NewDetector(mem *n64.Memory) *Detector {
	return &Detector{mem: mem}
}

// DetectOoTMM checks if the ComboContext magic is valid.
// OoTMM's Context_Init clears the read address after startup, so the magic
// at the fixed addresses is only transiently present during game switches.
// We check both addresses and fall back to gameMode-based detection.
// Returns (saveIndex, valid, error).
func (d *Detector) DetectOoTMM() (uint32, bool, error) {
	// Try OoT ComboCtx address first
	data, err := d.mem.Read(AddrComboCtxOot, ComboCtxSize)
	if err != nil {
		return 0, false, err
	}

	magic := string(data[CtxOffMagic : CtxOffMagic+8])
	if magic == "OoT+MM<3" {
		valid := binary.BigEndian.Uint32(data[CtxOffValid:])
		saveIndex := binary.BigEndian.Uint32(data[CtxOffSaveIndex:])
		return saveIndex, valid != 0, nil
	}

	// Try MM ComboCtx address
	data, err = d.mem.Read(AddrComboCtxMm, ComboCtxSize)
	if err != nil {
		return 0, false, err
	}

	magic = string(data[CtxOffMagic : CtxOffMagic+8])
	if magic == "OoT+MM<3" {
		valid := binary.BigEndian.Uint32(data[CtxOffValid:])
		saveIndex := binary.BigEndian.Uint32(data[CtxOffSaveIndex:])
		return saveIndex, valid != 0, nil
	}

	// Fallback: Context_Init has cleared the magic. Check if a game is
	// in normal mode (gameMode == 0), which indicates active gameplay.
	ootMode, err := d.mem.ReadU32BE(AddrOotSaveCtx + uint32(OotCtxOffGameMode))
	if err != nil {
		return 0, false, err
	}
	if ootMode == GameModeNormal {
		return 0, true, nil
	}

	mmMode, err := d.mem.ReadU32BE(AddrMmSaveCtx + uint32(MmCtxOffGameMode))
	if err != nil {
		return 0, false, err
	}
	if mmMode == GameModeNormal {
		return 0, true, nil
	}

	return 0, false, nil
}

// DetectActiveGame reads both game modes to determine which game is active.
func (d *Detector) DetectActiveGame() (ActiveGame, error) {
	ootMode, err := d.mem.ReadU32BE(AddrOotSaveCtx + uint32(OotCtxOffGameMode))
	if err != nil {
		return GameNone, fmt.Errorf("read OoT gameMode: %w", err)
	}

	mmMode, err := d.mem.ReadU32BE(AddrMmSaveCtx + uint32(MmCtxOffGameMode))
	if err != nil {
		return GameNone, fmt.Errorf("read MM gameMode: %w", err)
	}

	if ootMode == GameModeNormal {
		return GameOot, nil
	}
	if mmMode == GameModeNormal {
		return GameMm, nil
	}

	return GameNone, nil
}
