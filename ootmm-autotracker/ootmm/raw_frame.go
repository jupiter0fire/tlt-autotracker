package ootmm

import "fmt"

const (
	ootRawPlayStateCoreSize = OotPlayOffTempCollect + 4 - OotPlayOffSceneID
	ootRawPlayStateTailSize = OotPlayOffLinkAgeOnLoad + 1 - OotPlayOffCurrentRoom
	mmRawPlayStateCoreSize  = MmPlayOffCollectFlags + 4 - MmPlayOffSceneID
	mmRawPlayStateTailSize  = MmPlayOffGameplayFrames + 4 - MmPlayOffCurrentRoom
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
	ootSaveCtxUsedSize      = OotCtxOffGameMode + 4
	mmSaveCtxUsedSize       = MmCtxOffCycleFlags + MmPermCount*0x14
)

// RawChunk is a single opaque memory range exported for the TypeScript-side
// parser. JSON marshaling encodes Data as base64 automatically.
type RawChunk struct {
	Name    string `json:"name"`
	Address uint32 `json:"address"`
	Length  int    `json:"length"`
	Data    []byte `json:"data"`
}

// RawFrame is a complete raw snapshot for one stable OoTMM poll.
type RawFrame struct {
	Valid      bool
	ActiveGame ActiveGame
	SaveIndex  uint32
	Chunks     []RawChunk
}

type rawChunkSpec struct {
	name    string
	address uint32
	length  int
}

// ReadRawFrame captures a stable raw snapshot without running item/check
// extraction. It mirrors ReadState's game/save stability checks so raw clients
// do not observe mixed-transition frames.
func (r *Reader) ReadRawFrame() (*RawFrame, error) {
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

	chunks, err := r.readRawChunks(game)
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
// established a stable game/save pair through ReadState.
func (r *Reader) ReadRawFrameForStableState(game ActiveGame, saveIndex uint32) (*RawFrame, error) {
	chunks, err := r.readRawChunks(game)
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

func (r *Reader) readRawChunks(game ActiveGame) ([]RawChunk, error) {
	specs := append(rawChunkSpecs(game), r.selectedPlayStateChunkSpecs(game)...)
	chunks := make([]RawChunk, 0, len(specs))
	for _, spec := range specs {
		data, err := r.mem.Read(spec.address, spec.length)
		if err != nil {
			return nil, fmt.Errorf("read raw chunk %s at %#x: %w", spec.name, spec.address, err)
		}
		chunks = append(chunks, RawChunk{
			Name:    spec.name,
			Address: spec.address,
			Length:  spec.length,
			Data:    data,
		})
	}
	return chunks, nil
}

func (r *Reader) selectedPlayStateChunkSpecs(game ActiveGame) []rawChunkSpec {
	switch game {
	case GameOot:
		if _, ok := r.readOotPlayStateSampleCached(); !ok || r.ootPlayStateAddr == 0 {
			return nil
		}
		return []rawChunkSpec{
			{
				name:    ootPlayStateCoreChunk,
				address: r.ootPlayStateAddr + uint32(OotPlayOffSceneID),
				length:  ootRawPlayStateCoreSize,
			},
			{
				name:    ootPlayStateTailChunk,
				address: r.ootPlayStateAddr + uint32(OotPlayOffCurrentRoom),
				length:  ootRawPlayStateTailSize,
			},
		}
	case GameMm:
		if _, ok := r.readMmPlayStateSampleCached(); !ok || r.mmPlayStateAddr == 0 {
			return nil
		}
		return []rawChunkSpec{
			{
				name:    mmPlayStateCoreChunk,
				address: r.mmPlayStateAddr + uint32(MmPlayOffSceneID),
				length:  mmRawPlayStateCoreSize,
			},
			{
				name:    mmPlayStateTailChunk,
				address: r.mmPlayStateAddr + uint32(MmPlayOffCurrentRoom),
				length:  mmRawPlayStateTailSize,
			},
		}
	default:
		return nil
	}
}

func rawChunkSpecs(game ActiveGame) []rawChunkSpec {
	switch game {
	case GameOot:
		return []rawChunkSpec{
			{name: ootSaveCtxChunk, address: AddrOotSaveCtx, length: maxInt(OotSaveSize, ootSaveCtxUsedSize)},
			{name: ootForeignMmSaveChunk, address: AddrOotForeignMmSaveLive, length: MmSaveSize},
			{name: ootSharedSaveChunk, address: AddrOotSharedCustomSaveLive, length: sharedStateReadSize()},
			{name: ootRuntimeComboChunk, address: AddrOotRuntimeOotComboConfigLive, length: OotComboConfigSize},
			{name: ootRuntimeSilverChunk, address: AddrOotRuntimeSilverRupeeDataLive, length: OotSilverRupeeDataSize},
			{name: ootRuntimeMaxKeysChunk, address: AddrOotRuntimeMaxKeysLive, length: OotMaxKeysBlockSize},
		}
	case GameMm:
		return []rawChunkSpec{
			{name: mmSaveCtxChunk, address: AddrMmSaveCtx, length: maxInt(MmSaveSize, mmSaveCtxUsedSize)},
			{name: mmForeignOotSaveChunk, address: AddrMmForeignOotSaveLive, length: OotSaveSize},
			{name: mmSharedSaveChunk, address: AddrMmSharedCustomSaveLive, length: sharedStateReadSize()},
			{name: mmRuntimeComboChunk, address: AddrMmRuntimeOotComboConfigLive, length: OotComboConfigSize},
		}
	default:
		return nil
	}
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
