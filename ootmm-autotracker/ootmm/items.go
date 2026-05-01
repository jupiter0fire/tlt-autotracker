package ootmm

import "math/bits"

const (
	dungeonItemMapMask              = 0x04
	dungeonItemCompassMask          = 0x02
	dungeonItemBossKeyMask          = 0x01
	fishingPondLoachWeightMask      = 0x80
	fishingPondChildFishMinWeight   = 2
	fishingPondChildFishMaxWeight   = 14
	fishingPondAdultFishMinWeight   = 4
	fishingPondAdultFishMaxWeight   = 25
	fishingPondChildLoachMinWeight  = 14
	fishingPondChildLoachMaxWeight  = 19
	fishingPondAdultLoachMinWeight  = 29
	fishingPondAdultLoachMaxWeight  = 36
	sharedOcarinaButtonMaskDisabled = 0xffff
	sharedOcarinaButtonAMask        = 0x8000
	sharedOcarinaButtonCRightMask   = 0x0001
	sharedOcarinaButtonCLeftMask    = 0x0002
	sharedOcarinaButtonCUpMask      = 0x0004
	sharedOcarinaButtonCDownMask    = 0x0008
	// OotExtraFlags.greatFairies is stored MSB-first in raw bits 25..30.
	ootExtraFlagsGreatFairyMagicBit   = 25
	ootExtraFlagsGreatFairyMagic2Bit  = 26
	ootExtraFlagsGreatFairyDefenseBit = 27
	ootExtraFlagsGreatFairyWindBit    = 28
	ootExtraFlagsGreatFairyFireBit    = 29
	ootExtraFlagsGreatFairyLoveBit    = 30
	ootExtraFlagsChildWalletBit     = 17
	ootExtraFlagsBottomlessBit      = 7
	// MmExtraFlags.greatFairies is stored in raw extra-record bits 1..6.
	mmExtraFlagsGreatFairyTownBit     = 1
	mmExtraFlagsGreatFairyTownAltBit  = 2
	mmExtraFlagsGreatFairySwampBit    = 3
	mmExtraFlagsGreatFairyMountainBit = 4
	mmExtraFlagsGreatFairyOceanBit    = 5
	mmExtraFlagsGreatFairyValleyBit   = 6
	// OoT/MM extra-record bitfields are stored MSB-first on N64.
	mmExtraFlags2ChildWalletBit                = 31
	mmExtraFlags3BottomlessBit                 = 31
	mmExtraFlags2MaskKafei                     = 28
	mmExtraFlags2HoneyDarling                  = 27
	mmExtraFlags2RoomKey                       = 26
	mmExtraFlags2LetterKafei                   = 25
	mmExtraFlags2Pendant                       = 24
	mmExtraFlags2LetterMama                    = 23
	mmExtraFlags2Notebook                      = 22
	mmExtraFlags2MaskBlast                     = 21
	mmExtraFlags2DekuPlayground                = 20
	mmExtraFlags2MaskCouple                    = 19
	mmExtraFlags2MaskPostman                   = 17
	mmExtraFlags2MaskTroupeLeader              = 16
	mmExtraFlags2MaskFierceDeity               = 15
	mmExtraFlags2Ocarina                       = 14
	mmExtraFlags2SongOath                      = 13
	mmExtraFlags2MaskBremen                    = 10
	mmExtraFlags2MaskScents                    = 9
	mmExtraFlags2MaskKamaro                    = 8
	mmExtraFlags2MoonTear                      = 6
	mmExtraFlags2SongHealing                   = 5
	mmExtraFlags2TownStrayFairy                = 4
	mmExtraFlags3Lottery1                      = 29
	mmExtraFlags3Lottery2                      = 28
	mmExtraFlags3Lottery3                      = 27
	mmWeekEventTingleMapsByte                  = 0x118 >> 3
	mmWeekEventTingleMapClockTownMask          = 1 << (0x118 & 7)
	mmWeekEventTingleMapWoodfallMask           = 1 << (0x119 & 7)
	mmWeekEventTingleMapSnowheadMask           = 1 << (0x11a & 7)
	mmWeekEventTingleMapRanchMask              = 1 << (0x11b & 7)
	mmWeekEventTingleMapGreatBayMask           = 1 << (0x11c & 7)
	mmWeekEventTingleMapIkanaMask              = 1 << (0x11d & 7)
	mmWeekEventArcheryByte                     = 59
	mmWeekEventArcherySwampReward1Mask         = 0x10
	mmWeekEventArcheryTownReward1Mask          = 0x20
	mmWeekEventSwordsmanSchoolByte             = 63
	mmWeekEventSwordsmanSchoolMask             = 0x20
	mmOwlGreatBayBit                           = 0
	mmOwlZoraCapeBit                           = 1
	mmOwlSnowheadBit                           = 2
	mmOwlMountainVillageBit                    = 3
	mmOwlClockTownBit                          = 4
	mmOwlMilkRoadBit                           = 5
	mmOwlWoodfallBit                           = 6
	mmOwlSouthernSwampBit                      = 7
	mmOwlIkanaCanyonBit                        = 8
	mmOwlStoneTowerBit                         = 9
	mmOwlHiddenBit                             = 15
	ootEventSongSariaVanilla                   = 0x38
	ootEventSongSariaCustom                    = 0x58
	ootEventItemAnjuBottle                     = 0x0c
	ootEventItemTalonBottle                    = 0x02
	ootEventItemLostWoodsMemoryOrShootingChild = 0x0d
	ootEventItemLostWoodsMemory                = 0x17
	ootEventItemShootingGalleryAdult           = 0x0e
	ootEventItemGerudoArchery2                 = 0x0f
	ootEventItemLostWoodsTarget                = 0x1d
	ootEventItemGoronBracelet                  = 0x20
	ootEventItemPocketEgg                      = 0x2c
	ootEventItemMaskSellKeaton                 = 0x38
	ootEventItemMaskSellSkull                  = 0x39
	ootEventItemMaskSellSpooky                 = 0x3a
	ootEventItemMaskSellBunny                  = 0x3b
	ootEventMalonEgg                           = 0x12
	ootEventSongEpona                          = 0x62
	ootEventSongSunCustom                      = 0x5a
	ootEventFrogsGame                          = 0xd0
	ootEventFrogsZelda                         = 0xd1
	ootEventFrogsEpona                         = 0xd2
	ootEventFrogsSun                           = 0xd3
	ootEventFrogsSaria                         = 0xd4
	ootEventFrogsSongOfTime                    = 0xd5
	ootEventFrogsStorms                        = 0xd6
	ootEventSkulltulaHouse10                   = 0xda
	ootEventSkulltulaHouse20                   = 0xdb
	ootEventSkulltulaHouse30                   = 0xdc
	ootEventSkulltulaHouse40                   = 0xdd
	ootEventSkulltulaHouse50                   = 0xde
	ootEventMiscGerudoArchery1                 = 0x190
	ootEventMiscRichardHeartPiece              = 0x191
	ootEventMiscMedigoron                      = 0xb2
	ootChildTradeWeirdEggBit                   = 0
	ootChildTradeHatchMask                     = 0x1ffe
	ootChildTradeLetterMask                    = 0x1ffc
	ootAdultTradePocketEggBit                  = 0
	ootAdultTradePocketCuccoMask               = 0x07fe
	ootSceneShootingGallery                    = 0x42
	ootSceneShootingGallerySave                = 0x43
	ootEntranceChildArchery                    = 0x16d
	ootNpcLostWoodsMemoryBit                   = 12
	ootNpcShootingGalleryChildBit              = 31
	ootItemRutoLetter                          = 0x1b
	mmItemRutoLetter                           = 0xb6
)

type ootSymbolFlagCheck struct {
	symbol string
	flags  []int
}

type ootTradeSymbolCheck struct {
	symbol string
	mask   uint16
}

type ootQuestSymbolCheck struct {
	symbol string
	bit    int
}

var mmOwlItems = [...]struct {
	itemID string
	bit    uint
}{
	{"MM_OWL_GREAT_BAY", mmOwlGreatBayBit},
	{"MM_OWL_ZORA_CAPE", mmOwlZoraCapeBit},
	{"MM_OWL_SNOWHEAD", mmOwlSnowheadBit},
	{"MM_OWL_MOUNTAIN_VILLAGE", mmOwlMountainVillageBit},
	{"MM_OWL_CLOCK_TOWN", mmOwlClockTownBit},
	{"MM_OWL_MILK_ROAD", mmOwlMilkRoadBit},
	{"MM_OWL_WOODFALL", mmOwlWoodfallBit},
	{"MM_OWL_SOUTHERN_SWAMP", mmOwlSouthernSwampBit},
	{"MM_OWL_IKANA_CANYON", mmOwlIkanaCanyonBit},
	{"MM_OWL_STONE_TOWER", mmOwlStoneTowerBit},
	{"MM_OWL_HIDDEN", mmOwlHiddenBit},
}

var mmOwlCheckSymbols = [...]struct {
	symbol string
	bit    uint
}{
	{"OWL_GREAT_BAY", mmOwlGreatBayBit},
	{"OWL_ZORA_CAPE", mmOwlZoraCapeBit},
	{"OWL_SNOWHEAD", mmOwlSnowheadBit},
	{"OWL_MOUNTAIN_VILLAGE", mmOwlMountainVillageBit},
	{"OWL_CLOCK_TOWN", mmOwlClockTownBit},
	{"OWL_MILK_ROAD", mmOwlMilkRoadBit},
	{"OWL_WOODFALL", mmOwlWoodfallBit},
	{"OWL_SOUTHERN_SWAMP", mmOwlSouthernSwampBit},
	{"OWL_IKANA_CANYON", mmOwlIkanaCanyonBit},
	{"OWL_STONE_TOWER", mmOwlStoneTowerBit},
}

var mmExtraFlagsSymbolChecks = [...]struct {
	bit    int
	symbol string
}{
	{mmExtraFlagsGreatFairyTownBit, "GREAT_FAIRY_TOWN"},
	{mmExtraFlagsGreatFairyTownAltBit, "GREAT_FAIRY_TOWN_ALT"},
	{mmExtraFlagsGreatFairySwampBit, "GREAT_FAIRY_SWAMP"},
	{mmExtraFlagsGreatFairyMountainBit, "GREAT_FAIRY_MOUNTAIN"},
	{mmExtraFlagsGreatFairyOceanBit, "GREAT_FAIRY_OCEAN"},
	{mmExtraFlagsGreatFairyValleyBit, "GREAT_FAIRY_VALLEY"},
}

var mmExtraFlags2SymbolChecks = [...]struct {
	bit    int
	symbol string
}{
	{mmExtraFlags2MaskKafei, "MASK_KAFEI"},
	{mmExtraFlags2HoneyDarling, "HONEY_DARLING_1"},
	{mmExtraFlags2RoomKey, "ROOM_KEY"},
	{mmExtraFlags2LetterKafei, "LETTER_TO_KAFEI"},
	{mmExtraFlags2Pendant, "PENDANT_OF_MEMORIES"},
	{mmExtraFlags2LetterMama, "LETTER_TO_MAMA"},
	{mmExtraFlags2Notebook, "BOMBER_NOTEBOOK"},
	{mmExtraFlags2MaskBlast, "MASK_BLAST"},
	{mmExtraFlags2DekuPlayground, "DEKU_PLAYGROUND_1"},
	{mmExtraFlags2MaskCouple, "MASK_COUPLE"},
	{mmExtraFlags2MaskPostman, "MASK_POSTMAN"},
	{mmExtraFlags2MaskTroupeLeader, "MASK_TROUPE_LEADER"},
	{mmExtraFlags2MaskFierceDeity, "MASK_FIERCE_DEITY"},
	{mmExtraFlags2Ocarina, "SKULL_KID_OCARINA"},
	{mmExtraFlags2SongOath, "SONG_ORDER"},
	{mmExtraFlags2MaskBremen, "MASK_BREMEN"},
	{mmExtraFlags2MaskScents, "MASK_SCENTS"},
	{mmExtraFlags2MaskKamaro, "MASK_KAMARO"},
	{mmExtraFlags2MoonTear, "MOON_TEAR"},
	{mmExtraFlags2SongHealing, "SONG_HEALING"},
	{mmExtraFlags2TownStrayFairy, "STRAY_FAIRY_TOWN"},
}

var mmExtraFlags3SymbolChecks = [...]struct {
	bit    int
	symbol string
}{
	{mmExtraFlags3Lottery1, "LOTTERY_NIGHT_1"},
	{mmExtraFlags3Lottery2, "LOTTERY_NIGHT_2"},
	{mmExtraFlags3Lottery3, "LOTTERY_NIGHT_3"},
}

var mmWeekEventSymbolChecks = [...]struct {
	byteIndex int
	mask      uint8
	symbol    string
}{
	{mmWeekEventTingleMapsByte, mmWeekEventTingleMapClockTownMask, "TINGLE_MAP_CLOCK_TOWN"},
	{mmWeekEventTingleMapsByte, mmWeekEventTingleMapWoodfallMask, "TINGLE_MAP_WOODFALL"},
	{mmWeekEventTingleMapsByte, mmWeekEventTingleMapSnowheadMask, "TINGLE_MAP_SNOWHEAD"},
	{mmWeekEventTingleMapsByte, mmWeekEventTingleMapRanchMask, "TINGLE_MAP_ROMANI_RANCH"},
	{mmWeekEventTingleMapsByte, mmWeekEventTingleMapGreatBayMask, "TINGLE_MAP_GREAT_BAY"},
	{mmWeekEventTingleMapsByte, mmWeekEventTingleMapIkanaMask, "TINGLE_MAP_STONE_TOWER"},
	{mmWeekEventArcheryByte, mmWeekEventArcherySwampReward1Mask, "SHOOTING_GAME_SWAMP_1"},
	{mmWeekEventArcheryByte, mmWeekEventArcheryTownReward1Mask, "SHOOTING_GAME_TOWN_1"},
	{mmWeekEventSwordsmanSchoolByte, mmWeekEventSwordsmanSchoolMask, "SWORDSMAN_HEART_PIECE"},
}

var sharedOcarinaButtons = [...]struct {
	ootItemID string
	mmItemID  string
	mask      uint16
}{
	{"OOT_BUTTON_A", "MM_BUTTON_A", sharedOcarinaButtonAMask},
	{"OOT_BUTTON_C_RIGHT", "MM_BUTTON_C_RIGHT", sharedOcarinaButtonCRightMask},
	{"OOT_BUTTON_C_LEFT", "MM_BUTTON_C_LEFT", sharedOcarinaButtonCLeftMask},
	{"OOT_BUTTON_C_UP", "MM_BUTTON_C_UP", sharedOcarinaButtonCUpMask},
	{"OOT_BUTTON_C_DOWN", "MM_BUTTON_C_DOWN", sharedOcarinaButtonCDownMask},
}

var mmSkeletonKeyMaxKeys = [...]int{1, 3, 1, 4}

var ootFallbackMaxKeys = [...]int{0, 0, 0, 5, 7, 5, 5, 5, 3, 0, 0, 9, 4, 2, 0, 0, 0}

var ootFallbackSilverRupeeMaxCounts = [...]int{0, 5, 5, 5, 5, 5, 0, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}

var ootSilverRupeeItemIDs = [...][2]string{
	{"OOT_RUPEE_SILVER_DC", "OOT_RUPEE_SILVER_DC"},
	{"OOT_RUPEE_SILVER_BOTW", "OOT_RUPEE_SILVER_BOTW"},
	{"OOT_RUPEE_SILVER_SPIRIT_CHILD", "OOT_RUPEE_SILVER_SPIRIT_LOBBY"},
	{"OOT_RUPEE_SILVER_SPIRIT_SUN", "OOT_RUPEE_SILVER_SPIRIT_ADULT"},
	{"OOT_RUPEE_SILVER_SPIRIT_BOULDERS", "OOT_RUPEE_SILVER_SPIRIT_BOULDERS"},
	{"OOT_RUPEE_SILVER_SHADOW_SCYTHE", "OOT_RUPEE_SILVER_SHADOW_SCYTHE"},
	{"OOT_RUPEE_SILVER_SHADOW_BLADES", "OOT_RUPEE_SILVER_SHADOW_BLADES"},
	{"OOT_RUPEE_SILVER_SHADOW_PIT", "OOT_RUPEE_SILVER_SHADOW_PIT"},
	{"OOT_RUPEE_SILVER_SHADOW_SPIKES", "OOT_RUPEE_SILVER_SHADOW_SPIKES"},
	{"OOT_RUPEE_SILVER_IC_SCYTHE", "OOT_RUPEE_SILVER_IC_SCYTHE"},
	{"OOT_RUPEE_SILVER_IC_BLOCK", "OOT_RUPEE_SILVER_IC_BLOCK"},
	{"OOT_RUPEE_SILVER_GTG_SLOPES", "OOT_RUPEE_SILVER_GTG_SLOPES"},
	{"OOT_RUPEE_SILVER_GTG_LAVA", "OOT_RUPEE_SILVER_GTG_LAVA"},
	{"OOT_RUPEE_SILVER_GTG_WATER", "OOT_RUPEE_SILVER_GTG_WATER"},
	{"OOT_RUPEE_SILVER_GANON_SPIRIT", "OOT_RUPEE_SILVER_GANON_SHADOW"},
	{"OOT_RUPEE_SILVER_GANON_LIGHT", "OOT_RUPEE_SILVER_GANON_WATER"},
	{"OOT_RUPEE_SILVER_GANON_FIRE", "OOT_RUPEE_SILVER_GANON_FIRE"},
	{"OOT_RUPEE_SILVER_GANON_FOREST", "OOT_RUPEE_SILVER_GANON_FOREST"},
}

var ootEventSymbolChecks = [...]ootSymbolFlagCheck{
	{symbol: "MALON_EGG", flags: []int{ootEventMalonEgg}},
	{symbol: "MALON_SONG", flags: []int{ootEventSongEpona}},
	{symbol: "MASTER_SWORD", flags: []int{0x4f}},
	{symbol: "LIGHT_MEDALLION", flags: []int{0x45}},
	{symbol: "OCARINA_TIME_ITEM", flags: []int{0x43}},
	{symbol: "OCARINA_TIME_SONG", flags: []int{0xa9}},
	{symbol: "ROYAL_TOMB_SONG", flags: []int{ootEventSongSunCustom}},
	{symbol: "FROGS_GAME", flags: []int{ootEventFrogsGame}},
	{symbol: "FROGS_ZL", flags: []int{ootEventFrogsZelda}},
	{symbol: "FROGS_EPONA", flags: []int{ootEventFrogsEpona}},
	{symbol: "FROGS_SARIA", flags: []int{ootEventFrogsSaria}},
	{symbol: "FROGS_SUNS", flags: []int{ootEventFrogsSun}},
	{symbol: "FROGS_SOT", flags: []int{ootEventFrogsSongOfTime}},
	{symbol: "FROGS_STORMS", flags: []int{ootEventFrogsStorms}},
	{symbol: "SARIA_SONG", flags: []int{ootEventSongSariaVanilla, ootEventSongSariaCustom}},
	{symbol: "GS_10", flags: []int{ootEventSkulltulaHouse10}},
	{symbol: "GS_20", flags: []int{ootEventSkulltulaHouse20}},
	{symbol: "GS_30", flags: []int{ootEventSkulltulaHouse30}},
	{symbol: "GS_40", flags: []int{ootEventSkulltulaHouse40}},
	{symbol: "GS_50", flags: []int{ootEventSkulltulaHouse50}},
	{symbol: "SARIA_OCARINA", flags: []int{0xc1}},
	{symbol: "SHEIK_FOREST", flags: []int{0x50}},
	{symbol: "SHEIK_FIRE", flags: []int{0x51}},
	{symbol: "SHEIK_WATER", flags: []int{0x52}},
	{symbol: "SHEIK_SHADOW", flags: []int{0x54}},
	{symbol: "SHEIK_LIGHT", flags: []int{0x55}},
	{symbol: "SHEIK_SPIRIT", flags: []int{0xac}},
	{symbol: "SONG_STORMS", flags: []int{0x5b}},
	{symbol: "ZELDA_LETTER", flags: []int{0x40}},
	{symbol: "ZELDA_LIGHT_ARROW", flags: []int{0xc4}},
	{symbol: "ZELDA_SONG", flags: []int{0x59}},
}

var ootEventItemSymbolChecks = [...]ootSymbolFlagCheck{
	{symbol: "ANJU_BOTTLE", flags: []int{ootEventItemAnjuBottle}},
	{symbol: "TALON_BOTTLE", flags: []int{ootEventItemTalonBottle}},
	{symbol: "SHOOTING_GAME_ADULT", flags: []int{ootEventItemShootingGalleryAdult}},
	{symbol: "GERUDO_ARCHERY_2", flags: []int{ootEventItemGerudoArchery2}},
	{symbol: "LOST_WOODS_TARGET", flags: []int{ootEventItemLostWoodsTarget}},
	{symbol: "DARUNIA_BRACELET", flags: []int{ootEventItemGoronBracelet}},
	{symbol: "POCKET_EGG", flags: []int{ootEventItemPocketEgg}},
	{symbol: "MASK_SELL_KEATON", flags: []int{ootEventItemMaskSellKeaton}},
	{symbol: "MASK_SELL_SKULL", flags: []int{ootEventItemMaskSellSkull}},
	{symbol: "MASK_SELL_SPOOKY", flags: []int{ootEventItemMaskSellSpooky}},
	{symbol: "MASK_SELL_BUNNY", flags: []int{ootEventItemMaskSellBunny}},
}

var ootEventMiscSymbolChecks = [...]ootSymbolFlagCheck{
	{symbol: "GERUDO_ARCHERY_1", flags: []int{ootEventMiscGerudoArchery1}},
	{symbol: "DOG_LADY", flags: []int{ootEventMiscRichardHeartPiece}},
	{symbol: "MEDIGORON", flags: []int{ootEventMiscMedigoron}},
}

var ootTradeSymbolChecks = [...]ootTradeSymbolCheck{
	// OoTMM persists obtained adult-trade items in ExtraIdxOotTradeSave.
	// Pocket Cucco itself is unreliable as a standalone save bit in live files,
	// but any adult-trade progress beyond the initial Pocket Egg implies the egg
	// has already hatched.
	{symbol: "POCKET_EGG", mask: ootAdultTradePocketCuccoMask},
}

var ootChildTradeSymbolChecks = [...]ootTradeSymbolCheck{
	// Child trade progression is also persisted in ExtraIdxOotTradeSave. As
	// with Pocket Egg, later persistent progression implies the earlier checks
	// have already been completed even when the corresponding event bit is not
	// present in a live save snapshot.
	{symbol: "WEIRD_EGG", mask: ootChildTradeHatchMask},
	{symbol: "ZELDA_LETTER", mask: ootChildTradeLetterMask},
}

var ootQuestSymbolChecks = [...]ootQuestSymbolCheck{
	{symbol: "GERUDO_CARD", bit: QuestOotGerudoCard},
}

var ootExtraFlagsSymbolChecks = [...]struct {
	bit    int
	symbol string
}{
	{ootExtraFlagsGreatFairyMagicBit, "FAIRY_MAGIC_UPGRADE"},
	{ootExtraFlagsGreatFairyMagic2Bit, "FAIRY_MAGIC_UPGRADE2"},
	{ootExtraFlagsGreatFairyDefenseBit, "FAIRY_DEFENSE_UPGRADE"},
	{ootExtraFlagsGreatFairyWindBit, "FAIRY_SPELL_WIND"},
	{ootExtraFlagsGreatFairyFireBit, "FAIRY_SPELL_FIRE"},
	{ootExtraFlagsGreatFairyLoveBit, "FAIRY_SPELL_LOVE"},
}

// TrackedItem represents a single trackable item with its current quantity.
type TrackedItem struct {
	ID  string `json:"id"`
	Qty int    `json:"qty"`
}

// TrackedCheck represents a single location check.
type TrackedCheck struct {
	Key     string `json:"-"`
	Name    string `json:"name"`
	Checked bool   `json:"checked"`
}

func ootWalletLevel(oot *OotState) int {
	ootFlags := oot.ExtraRecords[ExtraIdxOotFlags]
	if ootFlags&(1<<ootExtraFlagsChildWalletBit) == 0 {
		return 0
	}
	if ootFlags&(1<<ootExtraFlagsBottomlessBit) != 0 {
		return 5
	}
	return GetUpgradeLevel(oot.Upgrades, 12, 2) + 1
}

func mmWalletLevel(state *GameState) int {
	if state.Oot.ExtraRecords[ExtraIdxMmFlags2]&(1<<mmExtraFlags2ChildWalletBit) == 0 {
		return 0
	}
	if state.Oot.ExtraRecords[ExtraIdxMmFlags3]&(1<<mmExtraFlags3BottomlessBit) != 0 {
		return 5
	}
	return GetUpgradeLevel(state.Mm.Upgrades, 12, 2) + 1
}

func mmExtraFlags2(state *GameState) uint32 {
	if state == nil {
		return 0
	}
	if state.Mm.ExtraFlags2 != 0 {
		return state.Mm.ExtraFlags2
	}
	return state.Oot.ExtraRecords[ExtraIdxMmFlags2]
}

func mmExtraFlags3(state *GameState) uint32 {
	if state == nil {
		return 0
	}
	return state.Oot.ExtraRecords[ExtraIdxMmFlags3]
}

func mmExtraFlags(state *GameState) uint32 {
	if state == nil {
		return 0
	}
	return state.Oot.ExtraRecords[ExtraIdxMmFlags]
}

func mmTownStrayFairyCollected(state *GameState) bool {
	return state.Mm.TownStrayFairy || mmExtraFlags2(state)&(1<<mmExtraFlags2TownStrayFairy) != 0
}

func hasMmWeekEventBit(state *GameState, byteIndex int, mask uint8) bool {
	if state == nil || byteIndex < 0 || byteIndex >= len(state.Mm.WeekEventReg) {
		return false
	}
	return state.Mm.WeekEventReg[byteIndex]&mask != 0
}

func magicUpgradeLevel(hasMagic bool, hasDoubleMagic bool) int {
	if !hasMagic {
		return 0
	}
	if hasDoubleMagic {
		return 2
	}
	return 1
}

// ExtractItems extracts all trackable items from the game state.
func ExtractItems(state *GameState) []TrackedItem {
	items := make([]TrackedItem, 0, 128)

	// === OOT ITEMS ===
	oot := &state.Oot

	// Medallions
	items = appendQuestBit(items, oot.QuestItems, QuestOotMedallionForest, "OOT_MEDALLION_FOREST")
	items = appendQuestBit(items, oot.QuestItems, QuestOotMedallionFire, "OOT_MEDALLION_FIRE")
	items = appendQuestBit(items, oot.QuestItems, QuestOotMedallionWater, "OOT_MEDALLION_WATER")
	items = appendQuestBit(items, oot.QuestItems, QuestOotMedallionSpirit, "OOT_MEDALLION_SPIRIT")
	items = appendQuestBit(items, oot.QuestItems, QuestOotMedallionShadow, "OOT_MEDALLION_SHADOW")
	items = appendQuestBit(items, oot.QuestItems, QuestOotMedallionLight, "OOT_MEDALLION_LIGHT")

	// Spiritual Stones
	items = appendQuestBit(items, oot.QuestItems, QuestOotStoneEmerald, "OOT_STONE_EMERALD")
	items = appendQuestBit(items, oot.QuestItems, QuestOotStoneRuby, "OOT_STONE_RUBY")
	items = appendQuestBit(items, oot.QuestItems, QuestOotStoneSapphire, "OOT_STONE_SAPPHIRE")

	// Songs
	items = appendQuestBit(items, oot.QuestItems, QuestOotSongMinuet, "OOT_SONG_MINUET")
	items = appendQuestBit(items, oot.QuestItems, QuestOotSongBolero, "OOT_SONG_BOLERO")
	items = appendQuestBit(items, oot.QuestItems, QuestOotSongSerenade, "OOT_SONG_SERENADE")
	items = appendQuestBit(items, oot.QuestItems, QuestOotSongRequiem, "OOT_SONG_REQUIEM")
	items = appendQuestBit(items, oot.QuestItems, QuestOotSongNocturne, "OOT_SONG_NOCTURNE")
	items = appendQuestBit(items, oot.QuestItems, QuestOotSongPrelude, "OOT_SONG_PRELUDE")
	items = appendQuestBit(items, oot.QuestItems, QuestOotSongLullaby, "OOT_SONG_LULLABY")
	items = appendQuestBit(items, oot.QuestItems, QuestOotSongEpona, "OOT_SONG_EPONA")
	items = appendQuestBit(items, oot.QuestItems, QuestOotSongSaria, "OOT_SONG_SARIA")
	items = appendQuestBit(items, oot.QuestItems, QuestOotSongSun, "OOT_SONG_SUN")
	items = appendQuestBit(items, oot.QuestItems, QuestOotSongTime, "OOT_SONG_TIME")
	items = appendQuestBit(items, oot.QuestItems, QuestOotSongStorms, "OOT_SONG_STORMS")

	// Other quest items
	items = appendQuestBit(items, oot.QuestItems, QuestOotAgony, "OOT_STONE_OF_AGONY")
	items = appendQuestBit(items, oot.QuestItems, QuestOotGerudoCard, "OOT_GERUDO_CARD")

	// Counts
	items = append(items, TrackedItem{"OOT_GOLD_TOKENS", int(oot.GoldTokens)})
	items = append(items, TrackedItem{"OOT_HEART_PIECES", int(oot.HeartPieces)})

	// OoT swords and shields are published as raw ownership bitmasks so the
	// tracker can distinguish Kokiri/Deku ownership from later upgrades.
	items = append(items, TrackedItem{"OOT_SWORD", int(oot.Equipment & 0xF)})
	items = append(items, TrackedItem{"OOT_SHIELD", int((oot.Equipment >> 4) & 0xF)})
	items = append(items, TrackedItem{"OOT_TUNIC", ootEquipmentLevel((oot.Equipment >> 8) & 0xF)})
	items = append(items, TrackedItem{"OOT_BOOTS", ootEquipmentLevel((oot.Equipment >> 12) & 0xF)})

	// Upgrades
	items = append(items, TrackedItem{"OOT_QUIVER", GetUpgradeLevel(oot.Upgrades, 0, 3)})
	items = append(items, TrackedItem{"OOT_BOMB_BAG", GetUpgradeLevel(oot.Upgrades, 3, 3)})
	items = append(items, TrackedItem{"OOT_STRENGTH", GetUpgradeLevel(oot.Upgrades, 6, 3)})
	items = append(items, TrackedItem{"OOT_SCALE", GetUpgradeLevel(oot.Upgrades, 9, 3)})
	items = append(items, TrackedItem{"OOT_MAGIC_UPGRADE", magicUpgradeLevel(oot.HasMagic, oot.HasDoubleMagic)})
	items = append(items, TrackedItem{"OOT_WALLET", ootWalletLevel(oot)})
	items = append(items, TrackedItem{"OOT_BULLET_BAG", GetUpgradeLevel(oot.Upgrades, 14, 3)})
	items = append(items, TrackedItem{"OOT_STICK_UPGRADE", GetUpgradeLevel(oot.Upgrades, 17, 3)})
	items = append(items, TrackedItem{"OOT_NUT_UPGRADE", GetUpgradeLevel(oot.Upgrades, 20, 3)})

	// Main inventory items (track which slots have items, 0xFF = empty)
	for i, itemID := range oot.Items {
		entry := ootInventorySlotEntry(i)
		if entry.ItemID == "" {
			continue
		}
		// Skip trade slots — tracked via ExtraRecord bitmasks below.
		if entry.ItemID == "OOT_ADULT_TRADE" || entry.ItemID == "OOT_CHILD_TRADE" {
			continue
		}
		qty := inventorySlotQty(entry, itemID, oot.Beans)
		if entry.ItemID == "OOT_BOMBCHUS" && qty == 0 && state.Shared.BombchuBagOot > 0 {
			qty = 1
		}
		if qty > 0 {
			items = append(items, TrackedItem{entry.ItemID, qty})
		}
	}
	items = append(items, TrackedItem{"OOT_BOTTLE_RUTO_LETTER", countOotBottleItem(oot.Items[:], ootItemRutoLetter)})

	// OoT trade items — bitmasks from ExtraRecords.
	// OotExtraTrade: u16 child (upper 16 bits) | u16 adult (lower 16 bits).
	ootTradeRec := oot.ExtraRecords[ExtraIdxOotTrade]
	items = append(items, TrackedItem{"OOT_CHILD_TRADE", int(ootTradeRec >> 16)})
	items = append(items, TrackedItem{"OOT_ADULT_TRADE", int(ootTradeRec & 0xFFFF)})

	// Dungeon items
	for i := 0; i < 20; i++ {
		di := oot.DungeonItems[i]
		bossKeyID, compassID, mapID := ootDungeonItemIDs(i)
		if bossKeyID == "" && compassID == "" && mapID == "" {
			continue
		}
		if bossKeyID != "" {
			items = append(items, TrackedItem{bossKeyID, boolToInt(di&dungeonItemBossKeyMask != 0)})
		}
		if compassID != "" {
			items = append(items, TrackedItem{compassID, boolToInt(di&dungeonItemCompassMask != 0)})
		}
		if mapID != "" {
			items = append(items, TrackedItem{mapID, boolToInt(di&dungeonItemMapMask != 0)})
		}
	}
	for i := 0; i < 19; i++ {
		keyID := ootDungeonSmallKeyID(i)
		if keyID == "" {
			continue
		}
		keys := oot.DungeonKeys[i]
		if keys < 0 {
			keys = 0
		}
		items = append(items, TrackedItem{keyID, int(keys)})
	}
	items = appendOotSilverRupeeItems(items, oot)

	// Extra records
	ootExtraFlags := oot.ExtraRecords[ExtraIdxOotFlags]
	items = append(items, TrackedItem{"OOT_GANON_BK", boolToInt(ootExtraFlags&1 != 0)})

	// === MM ITEMS ===
	mm := &state.Mm
	// Boss Remains
	items = appendQuestBit(items, mm.QuestItems, QuestMmRemainsOdolwa, "MM_REMAINS_ODOLWA")
	items = appendQuestBit(items, mm.QuestItems, QuestMmRemainsGoht, "MM_REMAINS_GOHT")
	items = appendQuestBit(items, mm.QuestItems, QuestMmRemainsGyorg, "MM_REMAINS_GYORG")
	items = appendQuestBit(items, mm.QuestItems, QuestMmRemainsTwinmold, "MM_REMAINS_TWINMOLD")

	// Songs
	items = appendQuestBit(items, mm.QuestItems, QuestMmSongAwakening, "MM_SONG_AWAKENING")
	items = appendQuestBit(items, mm.QuestItems, QuestMmSongGoron, "MM_SONG_GORON")
	items = appendQuestBit(items, mm.QuestItems, QuestMmSongZora, "MM_SONG_ZORA")
	items = appendQuestBit(items, mm.QuestItems, QuestMmSongEmptiness, "MM_SONG_EMPTINESS")
	items = appendQuestBit(items, mm.QuestItems, QuestMmSongOrder, "MM_SONG_ORDER")
	items = appendQuestBit(items, mm.QuestItems, QuestMmSongSaria, "MM_SONG_SARIA")
	items = appendQuestBit(items, mm.QuestItems, QuestMmSongTime, "MM_SONG_TIME")
	items = appendQuestBit(items, mm.QuestItems, QuestMmSongHealing, "MM_SONG_HEALING")
	items = appendQuestBit(items, mm.QuestItems, QuestMmSongEpona, "MM_SONG_EPONA")
	items = appendQuestBit(items, mm.QuestItems, QuestMmSongSoaring, "MM_SONG_SOARING")
	items = appendQuestBit(items, mm.QuestItems, QuestMmSongStorms, "MM_SONG_STORMS")
	items = appendQuestBit(items, mm.QuestItems, QuestMmSongSun, "MM_SONG_SUN")

	items = appendQuestBit(items, mm.QuestItems, QuestMmNotebook, "MM_NOTEBOOK")
	items = append(items, TrackedItem{"MM_HEART_PIECES", int(mm.HeartPieces)})
	items = append(items, TrackedItem{"MM_SWORD", int(mm.Equipment & 0xF)})
	items = append(items, TrackedItem{"MM_SHIELD", int((mm.Equipment >> 4) & 0xF)})

	// MM inventory items
	for i, itemID := range mm.Items {
		entry := mmInventorySlotEntry(i)
		if entry.ItemID == "" {
			continue
		}
		// Skip trade slots — tracked via ExtraRecord bitmasks below.
		if entry.ItemID == "MM_TRADE_1" || entry.ItemID == "MM_TRADE_2" || entry.ItemID == "MM_TRADE_3" {
			continue
		}
		qty := inventorySlotQty(entry, itemID, 0)
		if entry.ItemID == "MM_BOMBCHU" && qty == 0 && state.Shared.BombchuBagMm > 0 {
			qty = 1
		}
		if qty > 0 {
			items = append(items, TrackedItem{entry.ItemID, qty})
		}
	}
	items = append(items, TrackedItem{"MM_BOTTLE_RUTO_LETTER", countMmBottleItem(mm.Items[:], mmItemRutoLetter)})

	// MM trade items — bitmasks from ExtraRecords (stored in OoT save).
	// MmExtraTrade (big-endian): trade1:6|trade2:5|trade3:5|tradeObtained1:6|tradeObtained2:5|tradeObtained3:5
	// We use the "tradeObtained" fields which persist across MM cycles.
	mmTradeRec := oot.ExtraRecords[ExtraIdxMmTrade]
	items = append(items, TrackedItem{"MM_TRADE_1", int((mmTradeRec >> 10) & 0x3F)})
	items = append(items, TrackedItem{"MM_TRADE_2", int((mmTradeRec >> 5) & 0x1F)})
	items = append(items, TrackedItem{"MM_TRADE_3", int(mmTradeRec & 0x1F)})

	// MM Upgrades
	items = append(items, TrackedItem{"MM_QUIVER", GetUpgradeLevel(mm.Upgrades, 0, 3)})
	items = append(items, TrackedItem{"MM_BOMB_BAG", GetUpgradeLevel(mm.Upgrades, 3, 3)})
	items = append(items, TrackedItem{"MM_STRENGTH", GetUpgradeLevel(mm.Upgrades, 6, 3)})
	items = append(items, TrackedItem{"MM_SCALE", GetUpgradeLevel(mm.Upgrades, 9, 3)})
	items = append(items, TrackedItem{"MM_MAGIC_UPGRADE", magicUpgradeLevel(mm.HasMagic, mm.HasDoubleMagic)})
	items = append(items, TrackedItem{"MM_WALLET", mmWalletLevel(state)})

	// MM Dungeon items
	for i := 0; i < 10; i++ {
		di := mm.DungeonItems[i]
		bossKeyID, compassID, mapID := mmDungeonItemIDs(i)
		if bossKeyID == "" && compassID == "" && mapID == "" {
			continue
		}
		if bossKeyID != "" {
			items = append(items, TrackedItem{bossKeyID, boolToInt(di&dungeonItemBossKeyMask != 0)})
		}
		if compassID != "" {
			items = append(items, TrackedItem{compassID, boolToInt(di&dungeonItemCompassMask != 0)})
		}
		if mapID != "" {
			items = append(items, TrackedItem{mapID, boolToInt(di&dungeonItemMapMask != 0)})
		}
	}
	for i := 0; i < 9; i++ {
		keyID := mmDungeonSmallKeyID(i)
		if keyID == "" {
			continue
		}
		keys := mm.DungeonKeys[i]
		if keys < 0 {
			keys = 0
		}
		items = append(items, TrackedItem{keyID, int(keys)})
	}

	// MM Stray Fairies
	for i := 0; i < 10; i++ {
		fairyID := mmDungeonStrayFairyID(i)
		if fairyID == "" {
			continue
		}
		items = append(items, TrackedItem{fairyID, int(mm.StrayFairies[i])})
	}
	items = append(items, TrackedItem{"MM_STRAY_FAIRY_TOWN", boolToInt(mmTownStrayFairyCollected(state))})
	items = append(items, TrackedItem{"MM_GS_TOKEN_SWAMP", int(mm.SkullTokensSwamp)})
	items = append(items, TrackedItem{"MM_GS_TOKEN_OCEAN", int(mm.SkullTokensOcean)})
	for _, owl := range mmOwlItems {
		items = append(items, TrackedItem{owl.itemID, boolToInt(oot.ExtraRecords[ExtraIdxMmOwlFlags]&(1<<owl.bit) != 0)})
	}
	for _, button := range sharedOcarinaButtons {
		items = append(items, TrackedItem{button.ootItemID, boolToInt(sharedOcarinaButtonOwned(state.Shared.OcarinaButtonMaskOot, button.mask))})
		items = append(items, TrackedItem{button.mmItemID, boolToInt(sharedOcarinaButtonOwned(state.Shared.OcarinaButtonMaskMm, button.mask))})
	}
	items = appendOotFishingPondItems(items, &state.Shared)
	items = appendCatalogItems(items, state)

	return items
}

// ExtractChecks extracts location checks from scene flags.
func ExtractChecks(state *GameState) []TrackedCheck {
	checks := make([]TrackedCheck, 0, 256)
	seenNames := make(map[string]struct{}, 256)
	appendCheck := func(key, name string) {
		if name == "" {
			return
		}
		if _, seen := seenNames[name]; seen {
			return
		}
		seenNames[name] = struct{}{}
		checks = append(checks, TrackedCheck{Key: key, Name: name, Checked: true})
	}

	// OoT scene flag-based checks
	for sceneIdx := 0; sceneIdx < OotPermCount; sceneIdx++ {
		sf := &state.Oot.SceneFlags[sceneIdx]
		chests := sf.Chests
		collectibles := sf.Collectibles
		if state.Oot.HasLiveSceneFlags && sceneIdx == int(state.Oot.LiveSceneID) {
			chests |= state.Oot.LiveChestFlags
			collectibles |= state.Oot.LiveCollectFlags
			collectibles |= state.Oot.LiveTempCollectFlag
		}
		// Each set bit in the chests field = one opened chest
		for bit := 0; bit < 32; bit++ {
			if chests&(1<<uint(bit)) != 0 {
				if name, ok := ootSceneCheckNameForState(&state.Oot, sceneIdx, "chest", bit); ok {
					appendCheck(ootSceneCheckID(sceneIdx, "chest", bit), name)
				}
			}
		}
		for bit := 0; bit < 32; bit++ {
			if collectibles&(1<<uint(bit)) != 0 {
				if name, ok := ootSceneCheckNameForState(&state.Oot, sceneIdx, "collect", bit); ok {
					appendCheck(ootSceneCheckID(sceneIdx, "collect", bit), name)
				}
			}
		}
	}

	// MM scene flag-based checks: merge permanent + cycle flags.
	// In MM many collectibles only persist in cycle flags, so we OR
	// both to catch checks that haven't been flushed to permanent yet.
	for sceneIdx := 0; sceneIdx < MmPermCount; sceneIdx++ {
		sf := &state.Mm.SceneFlags[sceneIdx]
		chests := sf.Chests
		collectibles := sf.Collectibles
		if sceneIdx < len(state.Mm.CycleFlags) {
			chests |= state.Mm.CycleFlags[sceneIdx].Chests
			collectibles |= state.Mm.CycleFlags[sceneIdx].Collectibles
		}
		if state.Mm.HasLiveSceneFlags && sceneIdx == int(state.Mm.LiveSceneID) {
			chests |= state.Mm.LiveChestFlags
			collectibles |= state.Mm.LiveCollectFlags
		}
		for bit := 0; bit < 32; bit++ {
			if chests&(1<<uint(bit)) != 0 {
				if name, ok := lookupSceneCheckName("MM", sceneIdx, "chest", bit); ok {
					appendCheck(mmSceneCheckID(sceneIdx, "chest", bit), name)
				}
			}
		}
		for bit := 0; bit < 32; bit++ {
			if collectibles&(1<<uint(bit)) != 0 {
				if name, ok := lookupSceneCheckName("MM", sceneIdx, "collect", bit); ok {
					appendCheck(mmSceneCheckID(sceneIdx, "collect", bit), name)
				}
			}
		}
	}

	appendBitmapChecks := func(bitmap []uint8, game string, keyPrefix string, lookup func(string, int) (string, bool)) {
		for byteIndex, value := range bitmap {
			for bit := 0; bit < 8; bit++ {
				if value&(1<<uint(bit)) == 0 {
					continue
				}
				index := byteIndex*8 + bit
				if name, ok := lookup(game, index); ok {
					appendCheck(game+"_"+keyPrefix+"_"+itoa(index), name)
				}
			}
		}
	}

	appendBitmapChecks(state.Shared.Bitmap("npcOot"), "OOT", "npc", npcCheckName)
	appendBitmapChecks(state.Shared.Bitmap("npcMm"), "MM", "npc", npcCheckName)
	appendOotGsChecks(state.Oot.GsFlags[:], &state.Oot, appendCheck)
	appendOotXflagChecks(state.Shared.Bitmap("xflagsOot"), &state.Oot, appendCheck)
	appendBitmapChecks(state.Shared.Bitmap("xflagsMm"), "MM", "xflag", xflagCheckName)
	appendBitmapChecks(state.Shared.Bitmap("shopsOot"), "OOT", "shop", shopCheckName)
	appendBitmapChecks(state.Shared.Bitmap("shopsMm"), "MM", "shop", shopCheckName)
	appendBitmapChecks(state.Shared.Bitmap("scrubsOot"), "OOT", "scrub", scrubCheckName)
	appendBitmapChecks(state.Shared.Bitmap("srOot"), "OOT", "sr", silverRupeeCheckName)

	ootFlags := state.Oot.ExtraRecords[ExtraIdxOotFlags]
	for _, entry := range ootExtraFlagsSymbolChecks {
		if ootFlags&(1<<entry.bit) == 0 {
			continue
		}
		if name, ok := npcSymbolCheckName("OOT", entry.symbol); ok {
			appendCheck("OOT_extra_2_"+itoa(entry.bit), name)
		}
	}

	mmFlags := mmExtraFlags(state)
	for _, entry := range mmExtraFlagsSymbolChecks {
		if mmFlags&(1<<entry.bit) == 0 {
			continue
		}
		if name, ok := npcSymbolCheckName("MM", entry.symbol); ok {
			appendCheck("MM_extra_6_"+itoa(entry.bit), name)
		}
	}
	mmFlags2 := mmExtraFlags2(state)
	for _, entry := range mmExtraFlags2SymbolChecks {
		if mmFlags2&(1<<entry.bit) == 0 {
			continue
		}
		if name, ok := npcSymbolCheckName("MM", entry.symbol); ok {
			appendCheck("MM_extra_"+itoa(entry.bit), name)
		}
	}
	mmFlags3 := mmExtraFlags3(state)
	for _, entry := range mmExtraFlags3SymbolChecks {
		if mmFlags3&(1<<entry.bit) == 0 {
			continue
		}
		if name, ok := npcSymbolCheckName("MM", entry.symbol); ok {
			appendCheck("MM_extra_13_"+itoa(entry.bit), name)
		}
	}
	for _, entry := range mmWeekEventSymbolChecks {
		if !hasMmWeekEventBit(state, entry.byteIndex, entry.mask) {
			continue
		}
		if name, ok := npcSymbolCheckName("MM", entry.symbol); ok {
			appendCheck("MM_week_event_"+itoa(entry.byteIndex)+"_"+itoa(int(entry.mask)), name)
		}
	}
	for _, owl := range mmOwlCheckSymbols {
		if state.Mm.OwlActivationFlags&(1<<owl.bit) == 0 {
			continue
		}
		if name, ok := npcSymbolCheckName("MM", owl.symbol); ok {
			appendCheck("MM_owl_activation_"+owl.symbol, name)
		}
	}
	appendOotQuestSymbolChecks(state, appendCheck)
	appendOotChildTradeSymbolChecks(state, appendCheck)
	appendOotTradeSymbolChecks(state, appendCheck)
	appendOotEventSymbolChecks(state, appendCheck)
	appendOotEventItemSymbolChecks(state, appendCheck)
	appendOotAmbiguousEventItemChecks(state, appendCheck)
	appendOotEventMiscSymbolChecks(state, appendCheck)

	return checks
}

func appendOotQuestSymbolChecks(state *GameState, appendCheck func(string, string)) {
	for _, entry := range ootQuestSymbolChecks {
		if !HasQuestBit(state.Oot.QuestItems, entry.bit) {
			continue
		}
		if name, ok := npcSymbolCheckName("OOT", entry.symbol); ok {
			appendCheck("OOT_quest_"+entry.symbol, name)
		}
	}
}

func appendOotEventSymbolChecks(state *GameState, appendCheck func(string, string)) {
	appendOotSymbolChecksFromFlags(state, ootEventSymbolChecks[:], hasOotEventCheck, "OOT_event_", appendCheck)
}

func appendOotChildTradeSymbolChecks(state *GameState, appendCheck func(string, string)) {
	childTradeSave := uint16(state.Oot.ExtraRecords[ExtraIdxOotTradeSave] >> 16)
	for _, entry := range ootChildTradeSymbolChecks {
		if childTradeSave&entry.mask == 0 {
			continue
		}
		if name, ok := npcSymbolCheckName("OOT", entry.symbol); ok {
			appendCheck("OOT_child_trade_"+entry.symbol, name)
		}
	}
}

func appendOotTradeSymbolChecks(state *GameState, appendCheck func(string, string)) {
	adultTradeSave := uint16(state.Oot.ExtraRecords[ExtraIdxOotTradeSave] & 0xffff)
	for _, entry := range ootTradeSymbolChecks {
		if adultTradeSave&entry.mask == 0 {
			continue
		}
		if name, ok := npcSymbolCheckName("OOT", entry.symbol); ok {
			appendCheck("OOT_trade_"+entry.symbol, name)
		}
	}
}

func appendOotEventItemSymbolChecks(state *GameState, appendCheck func(string, string)) {
	appendOotSymbolChecksFromFlags(state, ootEventItemSymbolChecks[:], hasOotEventItemCheck, "OOT_event_item_", appendCheck)
}

func appendOotAmbiguousEventItemChecks(state *GameState, appendCheck func(string, string)) {
	npcOot := state.Shared.Bitmap("npcOot")
	if bitmapHasBit(npcOot, ootNpcLostWoodsMemoryBit) || hasOotEventItemCheck(state, ootEventItemLostWoodsMemory) || ootHasLostWoodsMemoryGameProgress(state) {
		if name, ok := npcSymbolCheckName("OOT", "LOST_WOODS_MEMORY"); ok {
			appendCheck("OOT_event_item_LOST_WOODS_MEMORY", name)
		}
	}

	if bitmapHasBit(npcOot, ootNpcShootingGalleryChildBit) || hasOotEventItemCheck(state, ootEventItemLostWoodsMemoryOrShootingChild) {
		if name, ok := npcSymbolCheckName("OOT", "SHOOTING_GAME_CHILD"); ok {
			appendCheck("OOT_event_item_SHOOTING_GAME_CHILD", name)
		}
	}
}

func ootHasLostWoodsMemoryGameProgress(state *GameState) bool {
	if state == nil {
		return false
	}
	return state.Oot.OcarinaGameRound > 0
}

func appendOotEventMiscSymbolChecks(state *GameState, appendCheck func(string, string)) {
	appendOotSymbolChecksFromFlags(state, ootEventMiscSymbolChecks[:], hasOotEventMiscCheck, "OOT_event_misc_", appendCheck)
}

func ootCurrentSceneID(state *GameState) uint16 {
	if state == nil {
		return 0xffff
	}
	if state.Oot.HasLiveSceneFlags && state.Oot.LiveSceneID < OotPermCount {
		return state.Oot.LiveSceneID
	}
	return state.Oot.SceneID
}

func appendOotSymbolChecksFromFlags(state *GameState, entries []ootSymbolFlagCheck, hasFlag func(*GameState, int) bool, keyPrefix string, appendCheck func(string, string)) {
	for _, entry := range entries {
		for _, flag := range entry.flags {
			if !hasFlag(state, flag) {
				continue
			}
			if name, ok := npcSymbolCheckName("OOT", entry.symbol); ok {
				appendCheck(keyPrefix+entry.symbol, name)
			}
			break
		}
	}
}

func hasOotEventCheck(state *GameState, flag int) bool {
	return hasOotEventBitmapFlag(state.Oot.EventsChk[:], flag)
}

func hasOotEventItemCheck(state *GameState, flag int) bool {
	return hasOotEventBitmapFlag(state.Oot.EventsItem[:], flag)
}

func hasOotEventMiscCheck(state *GameState, flag int) bool {
	return hasOotEventBitmapFlag(state.Oot.EventsMisc[:], flag)
}

func hasOotEventBitmapFlag(bitmap []uint16, flag int) bool {
	word := flag >> 4
	if word < 0 || word >= len(bitmap) {
		return false
	}
	return bitmap[word]&(1<<uint(flag&0xF)) != 0
}

func appendOotGsChecks(bitmap []uint32, oot *OotState, appendCheck func(string, string)) {
	for wordIndex, value := range bitmap {
		for bit := 0; bit < 32; bit++ {
			if value&(1<<uint(bit)) == 0 {
				continue
			}
			index := wordIndex*32 + bit
			if name, ok := gsCheckName("OOT", index); ok {
				appendCheck("OOT_gs_"+itoa(index), name)
				continue
			}
			if names, ok := ootConflictingGsCheckNames(oot, index); ok {
				for _, name := range names {
					appendCheck("OOT_gs_"+itoa(index), name)
				}
			}
		}
	}
}

func appendOotXflagChecks(bitmap []uint8, oot *OotState, appendCheck func(string, string)) {
	for byteIndex, value := range bitmap {
		for bit := 0; bit < 8; bit++ {
			if value&(1<<uint(bit)) == 0 {
				continue
			}
			index := byteIndex*8 + bit
			if name, ok := xflagCheckName("OOT", index); ok {
				appendCheck("OOT_xflag_"+itoa(index), name)
				continue
			}
			if names, ok := ootConflictingXflagCheckNames(oot, index); ok {
				for _, name := range names {
					appendCheck("OOT_xflag_"+itoa(index), name)
				}
			}
		}
	}
}

func appendQuestBit(items []TrackedItem, questItems uint32, bit int, id string) []TrackedItem {
	if HasQuestBit(questItems, bit) {
		return append(items, TrackedItem{id, 1})
	}
	return append(items, TrackedItem{id, 0})
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func ootSceneCheckID(scene int, kind string, bit int) string {
	return sceneCheckKey("OOT", scene, kind, bit)
}

func mmSceneCheckID(scene int, kind string, bit int) string {
	return sceneCheckKey("MM", scene, kind, bit)
}

func ootSwordLevel(oot *OotState) int {
	swords := oot.Equipment & 0xF
	if swords&(0x4|0x8) != 0 {
		if oot.IsBiggoronSword {
			return 4
		}
		return 3
	}
	return ootEquipmentLevel(swords)
}

func ootEquipmentLevel(mask uint16) int {
	return bits.Len16(mask)
}

func itoa(i int) string {
	// Simple int to string for small non-negative values
	if i < 0 {
		return "-" + itoa(-i)
	}
	if i < 10 {
		return string(rune('0' + i))
	}
	return itoa(i/10) + string(rune('0'+i%10))
}

// OoT dungeon names by index.
func ootDungeonName(idx int) string {
	names := [20]string{
		"OOT_DEKU_TREE", "OOT_DODONGOS_CAVERN", "OOT_JABU_JABU",
		"OOT_FOREST_TEMPLE", "OOT_FIRE_TEMPLE", "OOT_WATER_TEMPLE",
		"OOT_SPIRIT_TEMPLE", "OOT_SHADOW_TEMPLE", "OOT_BOTTOM_WELL",
		"OOT_ICE_CAVERN", "OOT_GANONS_TOWER", "OOT_GERUDO_TRAINING", "OOT_GERUDO_FORTRESS", "OOT_GANONS_CASTLE",
		"", "", "OOT_TREASURE_CHEST_GAME", "", "", "",
	}
	if idx >= 0 && idx < len(names) {
		return names[idx]
	}
	return ""
}

func ootDungeonItemIDs(idx int) (bossKeyID, compassID, mapID string) {
	switch idx {
	case 0:
		return "", "OOT_COMPASS_DT", "OOT_MAP_DT"
	case 1:
		return "", "OOT_COMPASS_DC", "OOT_MAP_DC"
	case 2:
		return "", "OOT_COMPASS_JJ", "OOT_MAP_JJ"
	case 3:
		return "OOT_BOSS_KEY_FOREST", "OOT_COMPASS_FOREST", "OOT_MAP_FOREST"
	case 4:
		return "OOT_BOSS_KEY_FIRE", "OOT_COMPASS_FIRE", "OOT_MAP_FIRE"
	case 5:
		return "OOT_BOSS_KEY_WATER", "OOT_COMPASS_WATER", "OOT_MAP_WATER"
	case 6:
		return "OOT_BOSS_KEY_SPIRIT", "OOT_COMPASS_SPIRIT", "OOT_MAP_SPIRIT"
	case 7:
		return "OOT_BOSS_KEY_SHADOW", "OOT_COMPASS_SHADOW", "OOT_MAP_SHADOW"
	case 8:
		return "", "OOT_COMPASS_BOTW", "OOT_MAP_BOTW"
	case 9:
		return "", "OOT_COMPASS_ICE", "OOT_MAP_ICE"
	case 10:
		return "OOT_BOSS_KEY_GANON", "", ""
	case 11:
		return "", "OOT_COMPASS_GTG", "OOT_MAP_GTG"
	case 13:
		return "", "OOT_COMPASS_GANON", "OOT_MAP_GANON"
	default:
		return "", "", ""
	}
}

func ootDungeonSmallKeyID(idx int) string {
	switch idx {
	case 3:
		return "OOT_SMALL_KEY_FOREST"
	case 4:
		return "OOT_SMALL_KEY_FIRE"
	case 5:
		return "OOT_SMALL_KEY_WATER"
	case 6:
		return "OOT_SMALL_KEY_SPIRIT"
	case 7:
		return "OOT_SMALL_KEY_SHADOW"
	case 8:
		return "OOT_SMALL_KEY_BOTW"
	case 11:
		return "OOT_SMALL_KEY_GTG"
	case 12:
		return "OOT_SMALL_KEY_GF"
	case 13:
		return "OOT_SMALL_KEY_GANON"
	case 16:
		return "OOT_SMALL_KEY_TCG"
	default:
		return ""
	}
}

// MM dungeon names.
func mmDungeonName(idx int) string {
	names := [10]string{
		"MM_WOODFALL_TEMPLE", "MM_SNOWHEAD_TEMPLE", "MM_GREAT_BAY_TEMPLE",
		"MM_STONE_TOWER_TEMPLE", "", "", "", "", "", "",
	}
	if idx >= 0 && idx < len(names) {
		return names[idx]
	}
	return ""
}

func mmDungeonItemIDs(idx int) (bossKeyID, compassID, mapID string) {
	switch idx {
	case 0:
		return "MM_BOSS_KEY_WF", "MM_COMPASS_WF", "MM_MAP_WF"
	case 1:
		return "MM_BOSS_KEY_SH", "MM_COMPASS_SH", "MM_MAP_SH"
	case 2:
		return "MM_BOSS_KEY_GB", "MM_COMPASS_GB", "MM_MAP_GB"
	case 3:
		return "MM_BOSS_KEY_ST", "MM_COMPASS_ST", "MM_MAP_ST"
	default:
		return "", "", ""
	}
}

func mmDungeonSmallKeyID(idx int) string {
	switch idx {
	case 0:
		return "MM_SMALL_KEY_WF"
	case 1:
		return "MM_SMALL_KEY_SH"
	case 2:
		return "MM_SMALL_KEY_GB"
	case 3:
		return "MM_SMALL_KEY_ST"
	default:
		return ""
	}
}

func mmDungeonStrayFairyID(idx int) string {
	switch idx {
	case 0:
		return "MM_STRAY_FAIRY_WF"
	case 1:
		return "MM_STRAY_FAIRY_SH"
	case 2:
		return "MM_STRAY_FAIRY_GB"
	case 3:
		return "MM_STRAY_FAIRY_ST"
	default:
		return ""
	}
}

const emptyInventoryItem = 0xFF

func appendCatalogItems(items []TrackedItem, state *GameState) []TrackedItem {
	for _, entry := range trackedCatalogItems {
		qty := 0
		switch entry.Source.Kind {
		case "oot-derived-key-ring":
			qty = boolToInt(hasOotKeyRing(&state.Oot, entry.Source.Record))
		case "oot-derived-skeleton-key":
			qty = boolToInt(hasOotSkeletonKey(&state.Oot))
		case "mm-derived-key-ring":
			qty = boolToInt(hasMmKeyRing(&state.Mm, entry.Source.Record))
		case "oot-derived-platinum-token":
			qty = boolToInt(hasOotPlatinumToken(&state.Oot))
		case "mm-derived-platinum-token":
			qty = boolToInt(hasMmPlatinumToken(&state.Mm))
		case "oot-derived-magical-rupee":
			qty = boolToInt(hasOotMagicalRupee(&state.Oot))
		case "shared-bitmap-bit":
			qty = boolToInt(bitmapHasBit(state.Shared.Bitmap(entry.Source.Block), entry.Source.Bit))
		case "oot-extra-bit":
			if entry.Source.Record >= 0 && entry.Source.Record < len(state.Oot.ExtraRecords) {
				qty = boolToInt(state.Oot.ExtraRecords[entry.Source.Record]&(1<<uint(entry.Source.Bit)) != 0)
			}
		case "mm-week-event-bit":
			if entry.Source.Byte >= 0 && entry.Source.Byte < len(state.Mm.WeekEventReg) {
				qty = boolToInt(state.Mm.WeekEventReg[entry.Source.Byte]&(1<<uint(entry.Source.Bit)) != 0)
			}
		case "mm-derived-skeleton-key":
			qty = boolToInt(hasMmSkeletonKey(&state.Mm))
		case "mm-derived-transcendent-fairy":
			qty = boolToInt(hasMmTranscendentFairy(state))
		}
		items = append(items, TrackedItem{ID: entry.ItemID, Qty: qty})
	}
	return items
}

func appendOotFishingPondItems(items []TrackedItem, shared *SharedCustomState) []TrackedItem {
	items = appendFishingPondWeightItems(items, shared.CaughtChildFishWeights[:], false)
	items = appendFishingPondWeightItems(items, shared.CaughtAdultFishWeights[:], true)
	return items
}

func appendFishingPondWeightItems(items []TrackedItem, caughtWeights []uint8, adult bool) []TrackedItem {
	if len(caughtWeights) == 0 {
		return items
	}

	count := int(caughtWeights[0])
	if count > len(caughtWeights)-1 {
		count = len(caughtWeights) - 1
	}
	if count <= 0 {
		return items
	}

	counts := make(map[uint8]int, count)
	for index := 1; index <= count; index++ {
		itemID := ootFishingPondItemID(adult, caughtWeights[index])
		if itemID == "" {
			continue
		}
		counts[caughtWeights[index]]++
	}

	if adult {
		for weight := uint8(fishingPondAdultFishMinWeight); weight <= fishingPondAdultFishMaxWeight; weight++ {
			if qty := counts[weight]; qty > 0 {
				items = append(items, TrackedItem{ID: ootFishingPondItemID(true, weight), Qty: qty})
			}
		}
		for weight := uint8(fishingPondAdultLoachMinWeight); weight <= fishingPondAdultLoachMaxWeight; weight++ {
			rawWeight := weight | fishingPondLoachWeightMask
			if qty := counts[rawWeight]; qty > 0 {
				items = append(items, TrackedItem{ID: ootFishingPondItemID(true, rawWeight), Qty: qty})
			}
		}
		return items
	}

	for weight := uint8(fishingPondChildFishMinWeight); weight <= fishingPondChildFishMaxWeight; weight++ {
		if qty := counts[weight]; qty > 0 {
			items = append(items, TrackedItem{ID: ootFishingPondItemID(false, weight), Qty: qty})
		}
	}
	for weight := uint8(fishingPondChildLoachMinWeight); weight <= fishingPondChildLoachMaxWeight; weight++ {
		rawWeight := weight | fishingPondLoachWeightMask
		if qty := counts[rawWeight]; qty > 0 {
			items = append(items, TrackedItem{ID: ootFishingPondItemID(false, rawWeight), Qty: qty})
		}
	}

	return items
}

func ootFishingPondItemID(adult bool, rawWeight uint8) string {
	weight := rawWeight &^ fishingPondLoachWeightMask
	if rawWeight&fishingPondLoachWeightMask != 0 {
		if adult {
			if weight < fishingPondAdultLoachMinWeight || weight > fishingPondAdultLoachMaxWeight {
				return ""
			}
			return "OOT_FISHING_POND_ADULT_LOACH_" + itoa(int(weight)) + "LBS"
		}
		if weight < fishingPondChildLoachMinWeight || weight > fishingPondChildLoachMaxWeight {
			return ""
		}
		return "OOT_FISHING_POND_CHILD_LOACH_" + itoa(int(weight)) + "LBS"
	}

	if adult {
		if weight < fishingPondAdultFishMinWeight || weight > fishingPondAdultFishMaxWeight {
			return ""
		}
		return "OOT_FISHING_POND_ADULT_FISH_" + itoa(int(weight)) + "LBS"
	}
	if weight < fishingPondChildFishMinWeight || weight > fishingPondChildFishMaxWeight {
		return ""
	}
	return "OOT_FISHING_POND_CHILD_FISH_" + itoa(int(weight)) + "LBS"
}

func hasMmSkeletonKey(mm *MmState) bool {
	for index, want := range mmSkeletonKeyMaxKeys {
		if dungeonMaxKeys(mm.DungeonItems[index]) < want {
			return false
		}
	}
	return true
}

func hasOotKeyRing(oot *OotState, dungeonIndex int) bool {
	if dungeonIndex < 0 || dungeonIndex >= OotRuntimeSceneCount {
		return false
	}
	want := ootMaxKeyLimit(oot, dungeonIndex)
	if want == 0 {
		return false
	}
	return dungeonMaxKeys(oot.DungeonItems[dungeonIndex]) >= want
}

func hasOotSkeletonKey(oot *OotState) bool {
	totalWant := 0
	for sceneID := 0; sceneID < OotRuntimeSceneCount; sceneID++ {
		want := ootMaxKeyLimit(oot, sceneID)
		if want == 0 {
			continue
		}
		totalWant += want
		if dungeonMaxKeys(oot.DungeonItems[sceneID]) < want {
			return false
		}
	}
	return totalWant > 0
}

func hasMmKeyRing(mm *MmState, dungeonIndex int) bool {
	if dungeonIndex < 0 || dungeonIndex >= len(mmSkeletonKeyMaxKeys) {
		return false
	}
	return dungeonMaxKeys(mm.DungeonItems[dungeonIndex]) >= mmSkeletonKeyMaxKeys[dungeonIndex]
}

func hasOotPlatinumToken(oot *OotState) bool {
	return oot.GoldTokens >= 100
}

func hasMmPlatinumToken(mm *MmState) bool {
	return mm.SkullTokensSwamp >= 30 && mm.SkullTokensOcean >= 30
}

func appendOotSilverRupeeItems(items []TrackedItem, oot *OotState) []TrackedItem {
	for silverRupeeID := 0; silverRupeeID < OotSilverRupeeSetCount; silverRupeeID++ {
		qty := ootSilverRupeeCount(oot, silverRupeeID)
		if qty <= 0 {
			continue
		}
		itemID := ootSilverRupeeItemID(oot, silverRupeeID)
		if itemID == "" {
			continue
		}
		items = append(items, TrackedItem{itemID, qty})
	}
	return items
}

func ootSilverRupeeItemID(oot *OotState, silverRupeeID int) string {
	if silverRupeeID < 0 || silverRupeeID >= len(ootSilverRupeeItemIDs) {
		return ""
	}
	variant := 0
	switch silverRupeeID {
	case 2, 3:
		if ootIsMqSpirit(oot) {
			variant = 1
		}
	case 14, 15:
		if ootIsMqGanonCastle(oot) {
			variant = 1
		}
	}
	return ootSilverRupeeItemIDs[silverRupeeID][variant]
}

func ootMqDungeonState(oot *OotState, dungeonID int) (bool, bool) {
	if oot == nil {
		return false, false
	}
	if oot.HasRuntimeMqBits {
		return oot.RuntimeMqBits&(1<<uint(dungeonID)) != 0, true
	}

	switch dungeonID {
	case OotMqDodongosCavern:
		return ootRuntimeSilverLimitMqState(oot, 0, 5, 0)
	case OotMqTempleForest:
		return ootRuntimeMaxKeyMqState(oot, OotSceneTempleForest, 6, 5)
	case OotMqTempleFire:
		return ootRuntimeMaxKeyMqState(oot, OotSceneTempleFire, 5, 7, 8)
	case OotMqTempleWater:
		return ootRuntimeMaxKeyMqState(oot, OotSceneTempleWater, 2, 5)
	case OotMqTempleSpirit:
		if mq, ok := ootRuntimeSilverLimitMqState(oot, 4, 0, 5); ok {
			return mq, true
		}
		return ootRuntimeMaxKeyMqState(oot, OotSceneTempleSpirit, 7, 5)
	case OotMqTempleShadow:
		return ootRuntimeMaxKeyMqState(oot, OotSceneTempleShadow, 6, 5)
	case OotMqBottomOfTheWell:
		return ootRuntimeMaxKeyMqState(oot, OotSceneBottomOfTheWell, 2, 3)
	case OotMqIceCavern:
		return ootRuntimeSilverLimitMqState(oot, 9, 0, 5)
	case OotMqGerudoTrainingGrounds:
		if mq, ok := ootRuntimeMaxKeyMqState(oot, OotSceneGerudoTrainingGround, 3, 9); ok {
			return mq, true
		}
		return ootRuntimeSilverLimitMqState(oot, 12, 6, 5)
	case OotMqGanonCastle:
		if mq, ok := ootRuntimeSilverLimitMqState(oot, 17, 0, 5); ok {
			return mq, true
		}
		return ootRuntimeMaxKeyMqState(oot, OotSceneInsideGanonCastle, 3, 2)
	default:
		return false, false
	}
}

func ootRuntimeMaxKeyMqState(oot *OotState, sceneID int, mqValue int, vanillaValues ...int) (bool, bool) {
	if oot == nil || !oot.HasRuntimeMaxKeys || sceneID < 0 || sceneID >= len(oot.RuntimeMaxKeys) {
		return false, false
	}
	value := int(oot.RuntimeMaxKeys[sceneID])
	if value == mqValue {
		return true, true
	}
	for _, vanillaValue := range vanillaValues {
		if value == vanillaValue {
			return false, true
		}
	}
	return false, false
}

func ootRuntimeSilverLimitMqState(oot *OotState, silverRupeeID int, mqValue int, vanillaValues ...int) (bool, bool) {
	if oot == nil || !oot.HasRuntimeSilverRupeeCounts || silverRupeeID < 0 || silverRupeeID >= len(oot.RuntimeSilverRupeeCounts) {
		return false, false
	}
	value := int(oot.RuntimeSilverRupeeCounts[silverRupeeID])
	if value == mqValue {
		return true, true
	}
	for _, vanillaValue := range vanillaValues {
		if value == vanillaValue {
			return false, true
		}
	}
	return false, false
}

func ootIsMqSpirit(oot *OotState) bool {
	mq, _ := ootMqDungeonState(oot, OotMqTempleSpirit)
	return mq
}

func ootIsMqGanonCastle(oot *OotState) bool {
	mq, _ := ootMqDungeonState(oot, OotMqGanonCastle)
	return mq
}

func hasOotMagicalRupee(oot *OotState) bool {
	for silverRupeeID := 0; silverRupeeID < OotSilverRupeeSetCount; silverRupeeID++ {
		want := ootSilverRupeeLimit(oot, silverRupeeID)
		if want == 0 {
			continue
		}
		if ootSilverRupeeCount(oot, silverRupeeID) < want {
			return false
		}
	}
	return true
}

func ootSilverRupeeCount(oot *OotState, silverRupeeID int) int {
	if silverRupeeID < 0 {
		return 0
	}
	recordIndex := ExtraIdxOotSilver1 + silverRupeeID/4
	if recordIndex < 0 || recordIndex >= len(oot.ExtraRecords) {
		return 0
	}
	shift := uint((silverRupeeID % 4) * 8)
	return int((oot.ExtraRecords[recordIndex] >> shift) & 0xff)
}

func hasMmTranscendentFairy(state *GameState) bool {
	if !mmTownStrayFairyCollected(state) {
		return false
	}
	for index := 0; index < len(mmSkeletonKeyMaxKeys); index++ {
		if state.Mm.StrayFairies[index] < 15 {
			return false
		}
	}
	return true
}

func sharedOcarinaButtonOwned(buttonMask uint16, mask uint16) bool {
	if buttonMask == sharedOcarinaButtonMaskDisabled {
		return false
	}
	return buttonMask&mask != 0
}

func ootMaxKeyLimit(oot *OotState, sceneID int) int {
	if sceneID < 0 || sceneID >= OotRuntimeSceneCount {
		return 0
	}
	if oot.HasRuntimeMaxKeys {
		return int(oot.RuntimeMaxKeys[sceneID])
	}
	return ootFallbackMaxKeys[sceneID]
}

func ootSilverRupeeLimit(oot *OotState, silverRupeeID int) int {
	if silverRupeeID < 0 || silverRupeeID >= OotSilverRupeeSetCount {
		return 0
	}
	if oot.HasRuntimeSilverRupeeCounts {
		return int(oot.RuntimeSilverRupeeCounts[silverRupeeID])
	}
	return ootFallbackSilverRupeeMaxCounts[silverRupeeID]
}

func dungeonMaxKeys(dungeonItem uint8) int {
	return int(dungeonItem >> 3)
}

func bitmapHasBit(bitmap []uint8, bit int) bool {
	byteIndex := bit / 8
	if byteIndex < 0 || byteIndex >= len(bitmap) {
		return false
	}
	return bitmap[byteIndex]&(1<<uint(bit%8)) != 0
}

func inventorySlotQty(entry inventorySlotEntry, itemID, beans uint8) int {
	if itemID == emptyInventoryItem {
		return 0
	}

	if entry.Quantity.UseBeansCount {
		if beans > 0 {
			return int(beans)
		}
		return 1
	}

	if len(entry.Quantity.Stages) > 0 {
		if entry.Quantity.MaxWithBottle && isOotBottleItem(itemID) {
			return len(entry.Quantity.Stages)
		}
		return stageQty(itemID, entry.Quantity.Stages)
	}

	return 1
}

func stageQty(itemID uint8, stages []uint8) int {
	for index, stageID := range stages {
		if stageID == itemID {
			return index + 1
		}
	}
	return 1
}

func isOotBottleItem(itemID uint8) bool {
	switch {
	case itemID >= 0x14 && itemID <= 0x20:
		return true
	case itemID == 0x82:
		return true
	case itemID >= 0x9E && itemID <= 0xA5:
		return true
	default:
		return false
	}
}

func countOotBottleItem(items []uint8, target uint8) int {
	return countBottleItem(items, target, ootInventorySlotEntry, "OOT_BOTTLE_")
}

func countMmBottleItem(items []uint8, target uint8) int {
	return countBottleItem(items, target, mmInventorySlotEntry, "MM_BOTTLE_")
}

func countBottleItem(items []uint8, target uint8, slotEntry func(int) inventorySlotEntry, bottlePrefix string) int {
	count := 0
	for index, itemID := range items {
		if itemID != target {
			continue
		}
		entry := slotEntry(index)
		if len(entry.ItemID) >= len(bottlePrefix) && entry.ItemID[:len(bottlePrefix)] == bottlePrefix {
			count++
		}
	}
	return count
}
