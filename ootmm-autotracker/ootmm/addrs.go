package ootmm

// N64 RDRAM addresses for OoTMM structures.
// All addresses are virtual (0x80-prefixed); the n64.Memory layer translates them.
// Source: OoTMM/packages/generator/include/combo/defs.h

const (
	// ComboContext — shared between OoT and MM, written to a known fixed address.
	AddrComboCtxOot uint32 = 0x80006584
	AddrComboCtxMm  uint32 = 0x80098280
	ComboCtxSize    int    = 0x20 // 32 bytes

	// OoT PlayState / GamePlay context. OoTMM currently runs on the NTSC 1.0
	// gameplay state layout, but we probe a small set of known retail/debug
	// addresses before caching the live PlayState base.
	AddrOotPlayStateNtsc10 uint32 = 0x801C84A0
	AddrOotPlayStateNtsc11 uint32 = 0x801C8660
	AddrOotPlayStateNtsc12 uint32 = 0x801C8D60
	AddrOotPlayStatePal10  uint32 = 0x801C64E0
	AddrOotPlayStatePal11  uint32 = 0x801C6520
	AddrOotPlayStateDebug  uint32 = 0x80212020

	// MM PlayState heap address (deterministic SystemArena allocation).
	AddrMmPlayState1 uint32 = 0x803E6B20

	// OoT gSaveContext
	AddrOotSaveCtx uint32 = 0x8011A5D0
	OotSaveCtxSize int    = 0x1450 // 5200 bytes
	OotSaveSize    int    = 0x1354 // OotSave within context

	// MM gSaveContext
	AddrMmSaveCtx uint32 = 0x801EF670
	MmSaveCtxSize int    = 0x48D0 // 18640 bytes
	MmSaveSize    int    = 0x3CA0 // MmSave within context

	// Payload region where gSharedCustomSave lives (OoT side)
	AddrOotPayload uint32 = 0x80400000
	OotPayloadSize int    = 0x80000 // 512KB

	// Payload region in MM where the foreign OoT save is kept.
	AddrMmPayload uint32 = 0x80730000
	MmPayloadSize int    = 0x50000 // 320KB

	// SharedCustomSave is 0x870 bytes in the current OoTMM build.
	// It sits immediately before the inactive game's foreign save inside
	// the payload, matching the layout produced by common/save.c.
	SharedCustomSaveSize uint32 = 0x870

	// Runtime OoT config objects read from the live payload.
	OotRuntimeSceneCount      = 17
	OotSilverRupeeSetCount    = 18
	OotSilverRupeeDataSize    = OotSilverRupeeSetCount * 4
	OotMaxKeysBlockSize       = OotRuntimeSceneCount + 4
	OotComboConfigSize        = 0x2dc
	ootMaxKeysFromSilverDelta = 0x13068

	OotComboConfigMqOffset                    = 0x09c
	OotComboConfigTriforcePiecesOffset        = 0x276
	OotComboConfigTriforceGoalOffset          = 0x278
	OotComboConfigSpecialOffset               = 0x12c
	OotComboConfigSpecialCount                = 5
	OotComboConfigSpecialSize                 = 8
	OotComboConfigPricesOffset                = 0x15c
	OotComboConfigPriceCount                  = 141
	OotComboConfigPriceMax                    = 4995
	OotComboConfigStaticHintsOffset           = 0x2a4
	OotComboConfigStaticHintCount             = 20
	OotComboConfigBossOffset                  = 0x2ba
	OotComboConfigBossCount                   = 12
	OotComboConfigStrayFairyRewardCountOffset = 0x2c6
	OotComboConfigBombchuBehaviorOotOffset    = 0x2c7
	OotComboConfigBombchuBehaviorMmOffset     = 0x2c8
	OotComboConfigSongEventsOffset            = 0x2c9
	OotComboConfigSongEventCount              = 18
)

// OoT scene IDs used by OoTMM's runtime key and silver-rupee metadata.
const (
	OotSceneDekuTree = iota
	OotSceneDodongosCavern
	OotSceneInsideJabuJabu
	OotSceneTempleForest
	OotSceneTempleFire
	OotSceneTempleWater
	OotSceneTempleSpirit
	OotSceneTempleShadow
	OotSceneBottomOfTheWell
	OotSceneIceCavern
	OotSceneGanonTower
	OotSceneGerudoTrainingGround
	OotSceneThievesHideout
	OotSceneInsideGanonCastle
	OotSceneUnused14
	OotSceneUnused15
	OotSceneTreasureShop
)

// OoT MQ dungeon IDs from OoTMM's combo/dungeon.h.
const (
	OotMqDekuTree = iota
	OotMqDodongosCavern
	OotMqJabuJabu
	OotMqTempleForest
	OotMqTempleFire
	OotMqTempleWater
	OotMqTempleSpirit
	OotMqTempleShadow
	OotMqBottomOfTheWell
	OotMqIceCavern
	OotMqGerudoTrainingGrounds
	OotMqGanonCastle
	OotMqDungeonCount
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
	OotOffEntrance        = 0x00 // u32
	OotOffAge             = 0x04 // u32 (0=adult, 1=child)
	OotOffSceneID         = 0x66 // u16 — current scene
	OotOffMagicAcquired   = 0x3A // u8 OotSave.info.playerData.isMagicAcquired
	OotOffDoubleMagic     = 0x3C // u8 OotSave.info.playerData.isDoubleMagicAcquired
	OotOffIsBiggoronSword = 0x3E // u8 in OotSaveInfo
	OotOffEquips          = 0x68 // OotItemEquips (0x0A bytes)

	// OotInventory starts at OotSave + 0x74 (after equips + padding)
	// But ASSERT_OFFSET says equipment is at 0x9C
	OotOffInvItems     = 0x74   // u8[0x18] = 24 item slots
	OotOffInvAmmo      = 0x8C   // u8[0x0F] = 15 ammo types
	OotOffInvBeans     = 0x9B   // u8
	OotOffEquipment    = 0x9C   // u16 bitfield {boots:4, tunics:4, shields:4, swords:4}
	OotOffUpgrades     = 0xA0   // u32 bitfield
	OotOffQuestItems   = 0xA4   // u32 bitfield (medallions, songs, stones)
	OotOffDungeonItems = 0xA8   // OotDungeonItems[0x14]
	OotOffDungeonKeys  = 0xBC   // s8[0x13]
	OotOffGoldTokens   = 0xD0   // u16
	OotOffGsFlags      = 0xE9C  // u8[24] (s32[6])
	OotOffChecksum     = 0x1352 // u16 additive checksum over OotSave with checksum zeroed

	// Scene flags
	OotOffPerm       = 0xD4 // OotPermanentSceneFlags[124], each 0x1C bytes
	OotPermEntrySize = 0x1C
	OotPermCount     = 124

	// Event flags
	OotOffEventsChk  = 0xED4 // u16[14]
	OotOffEventsItem = 0xEF0 // u16[4]
	OotOffEventsMisc = 0xEF8 // u16[30]

	// Extra records stored in perm[N].raw + 0x10
	OotPermExtraOff = 0x10

	// OotSaveContext fields beyond OotSave
	OotCtxOffFileNum  = 0x1354 // u32 active save slot
	OotCtxOffGameMode = 0x135C // s32 (0 = GAMEMODE_NORMAL)

	// Live OoT PlayState fields.
	OotPlayOffSceneID        = 0x00A4
	OotPlayOffActorTotal     = 0x1C2C
	OotPlayOffChestFlags     = 0x1D38
	OotPlayOffCollectFlags   = 0x1D44
	OotPlayOffTempCollect    = 0x1D48
	OotPlayOffCurrentRoom    = 0x11CBC
	OotPlayOffGameplayFrames = 0x11DE4
	OotPlayOffLinkAgeOnLoad  = 0x11DE8
)

// MmSave offsets
// Source: OoTMM/packages/generator/include/combo/mm/save.h (ASSERT_OFFSET verified)
const (
	MmOffEntrance   = 0x00 // s32
	MmOffMask       = 0x04 // u8 equippedMask
	MmOffPlayerForm = 0x20 // u8 (0=FD, 1=Goron, 2=Zora, 3=Deku, 4=Human)
	MmOffDay        = 0x18 // u32
	MmOffTime       = 0x0C // u16

	// MmSaveInfo starts at 0x24
	// MmSavePlayerData at 0x24 (size 0x28)
	MmOffMagicAcquired      = 0x40 // u8 MmSavePlayerData.isMagicAcquired
	MmOffDoubleMagic        = 0x41 // u8 MmSavePlayerData.isDoubleMagicAcquired
	MmOffOwlActivationFlags = 0x46 // u16 MmSavePlayerData.owlActivationFlags
	// MmItemEquips starts at 0x4C and is followed by 2 bytes of padding so
	// the 4-byte aligned MmInventory begins at 0x70.
	MmOffEquipment    = 0x6C   // u16 bitfield {boots:4, tunic:4, shield:4, sword:4}
	MmOffInvItems     = 0x70   // u8[48]
	MmOffInvAmmo      = 0xA0   // s8[24]
	MmOffInvUpgrades  = 0xB8   // u32 bitfield
	MmOffInvQuest     = 0xBC   // u32 MmQuestItems
	MmOffDungeonItems = 0xC0   // MmDungeonItems[10]
	MmOffDungeonKeys  = 0xCA   // s8[9]
	MmOffStrayFairies = 0xD4   // s8[10]
	MmOffWeekEventReg = 0xEF8  // u8[100]
	MmOffChecksum     = 0x100A // u16 additive checksum over MmSave with checksum zeroed

	// Permanent scene flags (within MmSaveInfo, offset from MmSave start)
	// MmSaveInfo starts at 0x24, then playerData(0x28) + itemEquips(0x24 with padding)
	// + inventory(0x88 with padding)
	MmOffPermScenes = 0xF8
	MmPermEntrySize = 0x1C
	MmPermCount     = 120

	// Skull counts in MmSave
	MmOffSkullSwamp = 0xEC0
	MmOffSkullOcean = 0xEC2

	// MmSaveContext fields beyond MmSave
	MmCtxOffFileNum    = 0x3CA0 // u32 active save slot
	MmCtxOffGameMode   = 0x3CA8 // s32
	MmCtxOffCycleFlags = 0x3F68 // MmCycleSceneFlags[120], each 0x14 bytes

	// Live MM PlayState fields.
	// MM PlayState is heap-allocated (SystemArena), so we probe a small
	// set of known heap addresses (deterministic for a given OoTMM build).
	// Source: zeldaret/mm ActorContext (actorCtx at +0x1CA0)
	//   ActorContextSceneFlags at actorCtx+0x1B8:
	//     switches[4] +0x00, chest +0x10, clearedRoom +0x14,
	//     clearedRoomTemp +0x18, collectible[4] +0x1C
	MmPlayOffSceneID        = 0x00A4
	MmPlayOffActorTotal     = 0x1CAE // actorCtx + 0x0E (totalLoadedActors)
	MmPlayOffChestFlags     = 0x1E68 // actorCtx + 0x1C8 (sceneFlags.chest)
	MmPlayOffCollectFlags   = 0x1E74 // actorCtx + 0x1D4 (sceneFlags.collectible[0])
	MmPlayOffCurrentRoom    = 0x186E0
	MmPlayOffGameplayFrames = 0x18840
)

// Foreign save copies inside the payload areas use the same flash layout as
// Save_ReadForeign in OoTMM common/save.c and must be indexed by fileNum.
const (
	ForeignOotSaveBaseOff uint32 = 0x20
	ForeignOotSaveStride  uint32 = 0x1450
	ForeignMmSaveBaseOff  uint32 = 0x8000
	ForeignMmSaveStride   uint32 = 0x8000
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
	ExtraIdxOotTrade     = 0
	ExtraIdxOotItems     = 1
	ExtraIdxOotFlags     = 2
	ExtraIdxMmBoss       = 3
	ExtraIdxMmItems      = 4
	ExtraIdxMmTrade      = 5
	ExtraIdxMmFlags      = 6
	ExtraIdxMmFlags2     = 7
	ExtraIdxMiscFlags    = 8
	ExtraIdxCowFlags     = 9
	ExtraIdxOotTradeSave = 10
	ExtraIdxMmOwlFlags   = 11
	ExtraIdxLedgerBase   = 12
	ExtraIdxMmFlags3     = 13
	ExtraIdxOotSilver1   = 14
	ExtraIdxOotSilver2   = 15
	ExtraIdxOotSilver3   = 16
	ExtraIdxOotSilver4   = 17
	ExtraIdxOotSilver5   = 18
	ExtraIdxTriforce     = 19
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
	QuestMmRemainsOdolwa   = 0
	QuestMmRemainsGoht     = 1
	QuestMmRemainsGyorg    = 2
	QuestMmRemainsTwinmold = 3
	QuestMmSongAwakening   = 6
	QuestMmSongGoron       = 7
	QuestMmSongZora        = 8
	QuestMmSongEmptiness   = 9
	QuestMmSongOrder       = 10
	QuestMmSongSaria       = 11
	QuestMmSongTime        = 12
	QuestMmSongHealing     = 13
	QuestMmSongEpona       = 14
	QuestMmSongSoaring     = 15
	QuestMmSongStorms      = 16
	QuestMmSongSun         = 17
	QuestMmNotebook        = 18
)

const (
	mmWeekEventTownStrayFairyByte = 8
	mmWeekEventTownStrayFairyMask = 0x80
)
