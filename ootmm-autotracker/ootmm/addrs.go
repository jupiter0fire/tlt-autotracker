package ootmm

// N64 RDRAM addresses for OoTMM structures.
// All addresses are virtual (0x80-prefixed); the n64.Memory layer translates them.
// Source: OoTMM/packages/generator/include/combo/defs.h

const (
	// ComboContext — shared between OoT and MM, written to a known fixed address.
	AddrComboCtxOot uint32 = 0x80006584
	AddrComboCtxMm  uint32 = 0x80098280
	ComboCtxSize    int    = 0x20 // 32 bytes

	// OoT gSaveContext
	AddrOotSaveCtx     uint32 = 0x8011A5D0
	OotSaveCtxSize     int    = 0x1450 // 5200 bytes
	OotSaveSize        int    = 0x1354 // OotSave within context

	// MM gSaveContext
	AddrMmSaveCtx      uint32 = 0x801EF670
	MmSaveCtxSize      int    = 0x48D0 // 18640 bytes
	MmSaveSize         int    = 0x3CA0 // MmSave within context

	// Payload region where gSharedCustomSave lives (OoT side)
	AddrOotPayload     uint32 = 0x80400000
	OotPayloadSize     int    = 0x80000 // 512KB

	// Payload region in MM where the foreign OoT save is kept.
	AddrMmPayload      uint32 = 0x80730000
	MmPayloadSize      int    = 0x50000 // 320KB
)

// ComboContext offsets (32 bytes, PACKED ALIGNED(4))
// Source: OoTMM/packages/generator/include/combo/context.h
const (
	CtxOffMagic     = 0x00 // char[8] "OoT+MM<3"
	CtxOffValid     = 0x08 // u32
	CtxOffSaveIndex = 0x0C // u32
	CtxOffEntrance  = 0x10 // u32
)

// OotSave offsets
// Source: OoTMM/packages/generator/include/combo/oot/save.h (ASSERT_OFFSET verified)
const (
	// Within OotSave
	OotOffEntrance     = 0x00  // u32
	OotOffAge          = 0x04  // u32 (0=adult, 1=child)
	OotOffSceneID      = 0x66  // u16 — current scene
	OotOffEquips       = 0x68  // OotItemEquips (0x0A bytes)

	// OotInventory starts at OotSave + 0x74 (after equips + padding)
	// But ASSERT_OFFSET says equipment is at 0x9C
	OotOffInvItems     = 0x74  // u8[0x18] = 24 item slots
	OotOffInvAmmo      = 0x8C  // u8[0x0F] = 15 ammo types
	OotOffInvBeans     = 0x9B  // u8
	OotOffEquipment    = 0x9C  // u16 bitfield {boots:4, tunics:4, shields:4, swords:4}
	OotOffUpgrades     = 0xA0  // u32 bitfield
	OotOffQuestItems   = 0xA4  // u32 bitfield (medallions, songs, stones)
	OotOffDungeonItems = 0xA8  // OotDungeonItems[0x14]
	OotOffDungeonKeys  = 0xBC  // s8[0x13]
	OotOffGoldTokens   = 0xD0  // u16

	// Scene flags
	OotOffPerm         = 0xD4  // OotPermanentSceneFlags[124], each 0x1C bytes
	OotPermEntrySize   = 0x1C
	OotPermCount       = 124

	// Event flags
	OotOffEventsChk    = 0xED8 // u16[14] — from info start (OotSave + 0x20)
	OotOffEventsMisc   = 0xEF8 // u16[30]

	// Extra records stored in perm[N].raw + 0x10
	OotPermExtraOff    = 0x10

	// OotSaveContext fields beyond OotSave
	OotCtxOffGameMode  = 0x135C // s32 (0 = GAMEMODE_NORMAL)
)

// MmSave offsets
// Source: OoTMM/packages/generator/include/combo/mm/save.h (ASSERT_OFFSET verified)
const (
	MmOffEntrance    = 0x00  // s32
	MmOffMask        = 0x04  // u8 equippedMask
	MmOffPlayerForm  = 0x20  // u8 (0=FD, 1=Goron, 2=Zora, 3=Deku, 4=Human)
	MmOffDay         = 0x18  // u32
	MmOffTime        = 0x0C  // u16

	// MmSaveInfo starts at 0x24
	// MmSavePlayerData at 0x24 (size 0x28)
	// MmItemEquips starts at 0x4C and is followed by 2 bytes of padding so
	// the 4-byte aligned MmInventory begins at 0x70.
	MmOffInvItems      = 0x70  // u8[48]
	MmOffInvAmmo       = 0xA0  // s8[24]
	MmOffInvUpgrades   = 0xB8  // u32 bitfield
	MmOffInvQuest      = 0xBC  // u32 MmQuestItems
	MmOffDungeonItems  = 0xC0  // MmDungeonItems[10]
	MmOffDungeonKeys   = 0xCA  // s8[9]
	MmOffStrayFairies  = 0xD4  // s8[10]

	// Permanent scene flags (within MmSaveInfo, offset from MmSave start)
	// MmSaveInfo starts at 0x24, then playerData(0x28) + itemEquips(0x24 with padding)
	// + inventory(0x88 with padding)
	MmOffPermScenes    = 0xF8
	MmPermEntrySize    = 0x1C
	MmPermCount        = 120

	// Skull counts in MmSaveInfo
	MmOffSkullSwamp  = 0x24 + 0xEA2
	MmOffSkullOcean  = 0x24 + 0xEA4

	// MmSaveContext fields beyond MmSave
	MmCtxOffGameMode   = 0x3CA8 // s32
	MmCtxOffCycleFlags = 0x3F68 // MmCycleSceneFlags[120], each 0x14 bytes
)

// XFlags counts
// Source: OoTMM/packages/generator/include/combo/xflags_data.h
const (
	XflagsCountOot = 0x2E8 // 744 bytes
	XflagsCountMm  = 0x34A // 842 bytes
)

// Game mode constants
const (
	GameModeNormal      = 0
	GameModeTitleScreen = 1
	GameModeFileSelect  = 2
)

// Extra record indices (stored in OotSave.info.perm[N].raw + 0x10)
const (
	ExtraIdxOotTrade      = 0
	ExtraIdxOotItems      = 1
	ExtraIdxOotFlags      = 2
	ExtraIdxMmBoss        = 3
	ExtraIdxMmItems       = 4
	ExtraIdxMmTrade       = 5
	ExtraIdxMmFlags       = 6
	ExtraIdxMmFlags2      = 7
	ExtraIdxMiscFlags     = 8
	ExtraIdxCowFlags      = 9
	ExtraIdxOotTradeSave  = 10
	ExtraIdxMmOwlFlags    = 11
	ExtraIdxLedgerBase    = 12
	ExtraIdxMmFlags3      = 13
	ExtraIdxOotSilver1    = 14
	ExtraIdxOotSilver2    = 15
	ExtraIdxOotSilver3    = 16
	ExtraIdxOotSilver4    = 17
	ExtraIdxOotSilver5    = 18
	ExtraIdxTriforce      = 19
)

// OoT Quest item bit positions
const (
	QuestOotMedallionForest = 0
	QuestOotMedallionFire   = 1
	QuestOotMedallionWater  = 2
	QuestOotMedallionSpirit = 3
	QuestOotMedallionShadow = 4
	QuestOotMedallionLight  = 5
	QuestOotSongMinuet      = 6
	QuestOotSongBolero      = 7
	QuestOotSongSerenade    = 8
	QuestOotSongRequiem     = 9
	QuestOotSongNocturne    = 10
	QuestOotSongPrelude     = 11
	QuestOotSongLullaby     = 12
	QuestOotSongEpona       = 13
	QuestOotSongSaria       = 14
	QuestOotSongSun         = 15
	QuestOotSongTime        = 16
	QuestOotSongStorms      = 17
	QuestOotStoneEmerald    = 18
	QuestOotStoneRuby       = 19
	QuestOotStoneSapphire   = 20
	QuestOotAgony           = 21
	QuestOotGerudoCard      = 22
	QuestOotGoldToken       = 23
)

// MM Quest item bit positions
const (
	QuestMmRemainsOdolwa    = 0
	QuestMmRemainsGoht      = 1
	QuestMmRemainsGyorg     = 2
	QuestMmRemainsTwinmold  = 3
	QuestMmSongAwakening    = 6
	QuestMmSongGoron        = 7
	QuestMmSongZora         = 8
	QuestMmSongEmptiness    = 9
	QuestMmSongOrder        = 10
	QuestMmSongSaria        = 11
	QuestMmSongTime         = 12
	QuestMmSongHealing      = 13
	QuestMmSongEpona        = 14
	QuestMmSongSoaring      = 15
	QuestMmSongStorms       = 16
	QuestMmSongSun          = 17
	QuestMmNotebook         = 18
)
