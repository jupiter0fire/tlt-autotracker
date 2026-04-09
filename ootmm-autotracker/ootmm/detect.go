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

const playableSaveSlotCount uint32 = 3
const addrMmRegEditorPtr uint32 = 0x801F3F60

func NewDetector(mem *n64.Memory) *Detector {
	return &Detector{mem: mem}
}

func isReadyGameState(game ActiveGame, gameMode, saveIndex, runtimeMarker uint32) bool {
	if gameMode != GameModeNormal || saveIndex >= playableSaveSlotCount {
		return false
	}

	if game == GameMm && runtimeMarker == 0 {
		return false
	}

	return true
}

func (d *Detector) detectReadyGame(game ActiveGame) (uint32, bool, error) {
	var modeAddr uint32
	var saveAddr uint32
	runtimeMarker := uint32(1)

	switch game {
	case GameOot:
		modeAddr = AddrOotSaveCtx + uint32(OotCtxOffGameMode)
		saveAddr = AddrOotSaveCtx + uint32(OotCtxOffFileNum)
	case GameMm:
		modeAddr = AddrMmSaveCtx + uint32(MmCtxOffGameMode)
		saveAddr = AddrMmSaveCtx + uint32(MmCtxOffFileNum)
		ptr, err := d.mem.ReadU32BE(addrMmRegEditorPtr)
		if err != nil {
			return 0, false, err
		}
		runtimeMarker = ptr
	default:
		return 0, false, nil
	}

	gameMode, err := d.mem.ReadU32BE(modeAddr)
	if err != nil {
		return 0, false, err
	}

	saveIndex, err := d.mem.ReadU32BE(saveAddr)
	if err != nil {
		return 0, false, err
	}

	return saveIndex, isReadyGameState(game, gameMode, saveIndex, runtimeMarker), nil
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

	// Fallback: Context_Init has cleared the magic. Only accept a game once it
	// is in normal gameplay and points at one of the three real save slots.
	saveIndex, ready, err := d.detectReadyGame(GameOot)
	if err != nil {
		return 0, false, err
	}
	if ready {
		return saveIndex, true, nil
	}

	saveIndex, ready, err = d.detectReadyGame(GameMm)
	if err != nil {
		return 0, false, err
	}
	if ready {
		return saveIndex, true, nil
	}

	return 0, false, nil
}

// DetectActiveGame determines which game is currently running.
// It first checks that neither ComboContext still contains the init magic.
// If the magic is present at either address then Context_Init has not yet
// finished and the save data is unreliable — the same approach multi-client
// uses (checking NET_GLOBAL magic before any read).
func (d *Detector) DetectActiveGame() (ActiveGame, error) {
	// Guard: if the ComboCtx magic is still visible at either address the
	// game is mid-transition (Context_Init clears it when ready).
	if d.isComboCtxMagicPresent() {
		return GameNone, nil
	}

	_, ootReady, err := d.detectReadyGame(GameOot)
	if err != nil {
		return GameNone, fmt.Errorf("detect OoT readiness: %w", err)
	}

	_, mmReady, err := d.detectReadyGame(GameMm)
	if err != nil {
		return GameNone, fmt.Errorf("detect MM readiness: %w", err)
	}

	if ootReady {
		return GameOot, nil
	}
	if mmReady {
		return GameMm, nil
	}

	return GameNone, nil
}

// isComboCtxMagicPresent returns true when the ComboContext init magic is
// found at either the OoT or MM address.  This indicates Context_Init is
// still running and save data should not be read.
func (d *Detector) isComboCtxMagicPresent() bool {
	if data, err := d.mem.Read(AddrComboCtxOot, ComboCtxSize); err == nil {
		if string(data[CtxOffMagic:CtxOffMagic+8]) == "OoT+MM<3" {
			return true
		}
	}
	if data, err := d.mem.Read(AddrComboCtxMm, ComboCtxSize); err == nil {
		if string(data[CtxOffMagic:CtxOffMagic+8]) == "OoT+MM<3" {
			return true
		}
	}
	return false
}
