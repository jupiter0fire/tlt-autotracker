package ootmm

import "fmt"

const (
	ootRawPlayStateCoreSize = OotPlayOffTempCollect + 4 - OotPlayOffSceneID
	ootRawPlayStateTailSize = OotPlayOffLinkAgeOnLoad + 1 - OotPlayOffCurrentRoom
	mmRawPlayStateCoreSize  = MmPlayOffCollectFlags + 4 - MmPlayOffSceneID
	mmRawPlayStateTailSize  = MmPlayOffGameplayFrames + 4 - MmPlayOffCurrentRoom
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

	chunks, err := r.readRawChunks()
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
	chunks, err := r.readRawChunks()
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

func (r *Reader) readRawChunks() ([]RawChunk, error) {
	specs := rawChunkSpecs()
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

func rawChunkSpecs() []rawChunkSpec {
	specs := []rawChunkSpec{
		{name: "combo_ctx_oot", address: AddrComboCtxOot, length: ComboCtxSize},
		{name: "combo_ctx_mm", address: AddrComboCtxMm, length: ComboCtxSize},
		{name: "oot_save_ctx", address: AddrOotSaveCtx, length: OotSaveCtxSize},
		{name: "mm_save_ctx", address: AddrMmSaveCtx, length: MmSaveCtxSize},
		{name: "oot_payload", address: AddrOotPayload, length: OotPayloadSize},
		{name: "mm_payload", address: AddrMmPayload, length: MmPayloadSize},
	}

	for index, addr := range ootPlayStateCandidateAddrs {
		specs = append(specs,
			rawChunkSpec{
				name:    fmt.Sprintf("oot_playstate_candidate_%d_core", index),
				address: addr + uint32(OotPlayOffSceneID),
				length:  ootRawPlayStateCoreSize,
			},
			rawChunkSpec{
				name:    fmt.Sprintf("oot_playstate_candidate_%d_tail", index),
				address: addr + uint32(OotPlayOffCurrentRoom),
				length:  ootRawPlayStateTailSize,
			},
		)
	}

	for index, addr := range mmPlayStateCandidateAddrs {
		specs = append(specs,
			rawChunkSpec{
				name:    fmt.Sprintf("mm_playstate_candidate_%d_core", index),
				address: addr + uint32(MmPlayOffSceneID),
				length:  mmRawPlayStateCoreSize,
			},
			rawChunkSpec{
				name:    fmt.Sprintf("mm_playstate_candidate_%d_tail", index),
				address: addr + uint32(MmPlayOffCurrentRoom),
				length:  mmRawPlayStateTailSize,
			},
		)
	}

	return specs
}
