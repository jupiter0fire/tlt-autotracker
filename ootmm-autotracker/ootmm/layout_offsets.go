package ootmm

// Internal offsets that still matter for raw-frame stability and readiness checks.
// These stay private so addrs.go remains limited to high-level region sizes.
const (
	comboCtxMagicOffset     = 0x00
	comboCtxMagicSize       = 8
	comboCtxValidOffset     = 0x08
	comboCtxSaveIndexOffset = 0x0C

	ootSaveCtxFileNumOffset  = 0x1354
	ootSaveCtxGameModeOffset = 0x135C

	mmSaveCtxFileNumOffset    = 0x3CA0
	mmSaveCtxGameModeOffset   = 0x3CA8
	mmSaveCtxCycleFlagsOffset = 0x3F68

	mmCycleSceneFlagCount  = 120
	mmCycleSceneFlagStride = 0x14

	gameModeNormal      = 0
	gameModeTitleScreen = 1
	gameModeFileSelect  = 2
)