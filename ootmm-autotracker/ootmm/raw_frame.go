package ootmm

import "fmt"

const (
	ootRawPlayStateCoreSize = 0x1CA8
	ootRawPlayStateTailSize = 0x12D
	mmRawPlayStateCoreSize  = 0x1DD4
	mmRawPlayStateTailSize  = 0x164
	ootSaveCtxChunk         = "oot_save_ctx"
	mmSaveCtxChunk          = "mm_save_ctx"
	ootForeignMmSaveChunk   = "oot_foreign_mm_save"
	mmForeignOotSaveChunk   = "mm_foreign_oot_save"
	ootSharedSaveChunk      = "oot_shared_custom_save"
	mmSharedSaveChunk       = "mm_shared_custom_save"
	ootRuntimeComboChunk    = "oot_runtime_combo_config"
	mmRuntimeComboChunk     = "mm_runtime_combo_config"
	ootRuntimeSilverChunk   = "oot_runtime_silver_rupee_data"
	ootRuntimeMaxKeysChunk  = "oot_runtime_max_keys"
	ootPlayStateCoreChunk   = "oot_playstate_core"
	ootPlayStateTailChunk   = "oot_playstate_tail"
	mmPlayStateCoreChunk    = "mm_playstate_core"
	mmPlayStateTailChunk    = "mm_playstate_tail"
	ootSaveCtxUsedSize      = ootSaveCtxGameModeOffset + 4
	mmSaveCtxUsedSize       = mmSaveCtxCycleFlagsOffset + mmCycleSceneFlagCount*mmCycleSceneFlagStride
)

// RawChunk is a single opaque memory range exported for the TypeScript-side
// parser. JSON marshaling encodes Data as base64 automatically.
type RawChunk struct {
	Name    string `json:"name"`
	Address uint32 `json:"address"`
	Length  int    `json:"length"`
	Data    []byte `json:"data"`
}

type RawChunkSpec struct {
	Name    string `json:"name"`
	Address uint32 `json:"address"`
	Length  int    `json:"length"`
}

type RawChunkSpecSelection struct {
	Oot []RawChunkSpec
	Mm  []RawChunkSpec
}

// RawFrame is a complete raw snapshot for one stable OoTMM poll.
type RawFrame struct {
	Valid      bool
	ActiveGame ActiveGame
	SaveIndex  uint32
	Chunks     []RawChunk
}

// ReadRawFrame captures a stable raw snapshot without running item/check
// extraction. It uses the same two-frame game/save stability guard as the
// tracker runtime so raw clients do not observe mixed-transition frames.
func (r *Reader) ReadRawFrame() (*RawFrame, error) {
	return r.ReadRawFrameWithSelection(defaultRawChunkSelection())
}

func (r *Reader) ReadRawFrameWithSelection(selection RawChunkSpecSelection) (*RawFrame, error) {
	frame := &RawFrame{}

	saveIndex, valid, err := r.detector.DetectOoTMM()
	if err != nil {
		return nil, fmt.Errorf("detect OoTMM: %w", err)
	}
	frame.Valid = valid
	frame.SaveIndex = saveIndex

	if !valid {
		r.resetPendingState()
		return frame, nil
	}

	game, err := r.detector.DetectActiveGame()
	if err != nil {
		return nil, fmt.Errorf("detect game: %w", err)
	}
	if game == GameNone {
		r.resetPendingState()
		return frame, nil
	}

	activeSaveIndex, err := r.readActiveSaveIndex(game)
	if err != nil {
		return nil, fmt.Errorf("read active save index: %w", err)
	}
	frame.SaveIndex = activeSaveIndex

	chunks, err := r.readRawChunks(selection.specsForGame(game))
	if err != nil {
		return nil, err
	}

	gameAfter, err := r.detector.DetectActiveGame()
	if err != nil {
		return nil, fmt.Errorf("re-detect game: %w", err)
	}
	if gameAfter != game {
		r.resetPendingState()
		return frame, nil
	}

	saveIndexAfter, err := r.readActiveSaveIndex(gameAfter)
	if err != nil {
		return nil, fmt.Errorf("re-read active save index: %w", err)
	}
	if saveIndexAfter != activeSaveIndex {
		r.resetPendingState()
		return frame, nil
	}

	if !r.acceptStableState(game, activeSaveIndex) {
		return frame, nil
	}

	frame.ActiveGame = game
	frame.Chunks = chunks
	return frame, nil
}

// ReadRawFrameForStableState captures a raw snapshot after a caller already
// established a stable game/save pair through its own polling logic.
func (r *Reader) ReadRawFrameForStableState(game ActiveGame, saveIndex uint32, specs []RawChunkSpec) (*RawFrame, error) {
	chunks, err := r.readRawChunks(specs)
	if err != nil {
		return nil, err
	}

	return &RawFrame{
		Valid:      true,
		ActiveGame: game,
		SaveIndex:  saveIndex,
		Chunks:     chunks,
	}, nil
}

func defaultRawChunkSelection() RawChunkSpecSelection {
	return RawChunkSpecSelection{
		Oot: []RawChunkSpec{{
			Name:    ootSaveCtxChunk,
			Address: AddrOotSaveCtx,
			Length:  maxInt(OotSaveSize, ootSaveCtxUsedSize),
		}},
		Mm: []RawChunkSpec{{
			Name:    mmSaveCtxChunk,
			Address: AddrMmSaveCtx,
			Length:  maxInt(MmSaveSize, mmSaveCtxUsedSize),
		}},
	}
}

func (selection RawChunkSpecSelection) specsForGame(game ActiveGame) []RawChunkSpec {
	switch game {
	case GameOot:
		return selection.Oot
	case GameMm:
		return selection.Mm
	default:
		return nil
	}
}

func (r *Reader) readRawChunks(specs []RawChunkSpec) ([]RawChunk, error) {
	chunks := make([]RawChunk, 0, len(specs))
	for _, spec := range specs {
		if spec.Length <= 0 || spec.Name == "" {
			continue
		}
		data, err := r.mem.Read(spec.Address, spec.Length)
		if err != nil {
			return nil, fmt.Errorf("read raw chunk %s at %#x: %w", spec.Name, spec.Address, err)
		}
		chunks = append(chunks, RawChunk{
			Name:    spec.Name,
			Address: spec.Address,
			Length:  spec.Length,
			Data:    data,
		})
	}
	return chunks, nil
}

func maxInt(values ...int) int {
	max := 0
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}
