package ootmm

// High-level fixed region sizes for the raw OoTMM protocol.
// Detailed field offsets live next to the narrow internal code that still uses them.
const (
	ComboCtxSize int = 0x20

	OotSaveCtxSize int = 0x1450
	OotSaveSize    int = 0x1354

	MmSaveCtxSize int = 0x48D0
	MmSaveSize    int = 0x3CA0

	OotPayloadSize int = 0x80000
	MmPayloadSize  int = 0x50000

	SharedCustomSaveSize uint32 = 0x870
	OotComboConfigSize   int    = 0x2DC
)
