package ootmm

import (
	"fmt"

	"ootmm-autotracker/n64"
)

// Reader reads stable raw OoTMM slices from N64 RDRAM.
type Reader struct {
	mem      *n64.Memory
	detector *Detector

	stableGame       ActiveGame
	stableSaveIndex  uint32
	pendingGame      ActiveGame
	pendingSaveIndex uint32
	hasPendingState  bool
}

func NewReader(mem *n64.Memory) *Reader {
	return &Reader{
		mem:      mem,
		detector: NewDetector(mem),
	}
}

func (r *Reader) readActiveSaveIndex(game ActiveGame) (uint32, error) {
	switch game {
	case GameOot:
		return r.mem.ReadU32BE(AddrOotSaveCtx + ootSaveCtxFileNumOffset)
	case GameMm:
		return r.mem.ReadU32BE(AddrMmSaveCtx + mmSaveCtxFileNumOffset)
	default:
		return 0, fmt.Errorf("no active game")
	}
}

func (r *Reader) resetPendingState() {
	r.hasPendingState = false
}

func (r *Reader) acceptStableState(game ActiveGame, saveIndex uint32) bool {
	if game == r.stableGame && saveIndex == r.stableSaveIndex {
		r.resetPendingState()
		return true
	}

	if r.hasPendingState && game == r.pendingGame && saveIndex == r.pendingSaveIndex {
		r.stableGame = game
		r.stableSaveIndex = saveIndex
		r.resetPendingState()
		return true
	}

	r.pendingGame = game
	r.pendingSaveIndex = saveIndex
	r.hasPendingState = true
	return false
}
