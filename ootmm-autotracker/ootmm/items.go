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
	sharedOcarinaButtonAMask        = 0x0010
	sharedOcarinaButtonCRightMask   = 0x0001
	sharedOcarinaButtonCLeftMask    = 0x0002
	sharedOcarinaButtonCUpMask      = 0x0004
	sharedOcarinaButtonCDownMask    = 0x0008
	mmExtraFlags2Notebook           = 9
	mmExtraFlags2MaskBlast          = 10
	mmOwlGreatBayBit                = 0
	mmOwlZoraCapeBit                = 1
	mmOwlSnowheadBit                = 2
	mmOwlMountainVillageBit         = 3
	mmOwlClockTownBit               = 4
	mmOwlMilkRoadBit                = 5
	mmOwlWoodfallBit                = 6
	mmOwlSouthernSwampBit           = 7
	mmOwlIkanaCanyonBit             = 8
	mmOwlStoneTowerBit              = 9
	mmOwlHiddenBit                  = 15
	ootEventSongSariaVanilla        = 0x38
	ootEventSongSariaCustom         = 0x58
	ootEventSongSunCustom           = 0x5a
	ootItemRutoLetter               = 0x1b
	mmItemRutoLetter                = 0xb6
)

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
	items = append(items, TrackedItem{"OOT_WALLET", GetUpgradeLevel(oot.Upgrades, 12, 2)})
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
	items = append(items, TrackedItem{"MM_WALLET", GetUpgradeLevel(mm.Upgrades, 12, 2)})

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
	items = append(items, TrackedItem{"MM_STRAY_FAIRY_TOWN", boolToInt(mm.TownStrayFairy)})
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
				if name, ok := lookupSceneCheckName("OOT", sceneIdx, "chest", bit); ok {
					appendCheck(ootSceneCheckID(sceneIdx, "chest", bit), name)
				}
			}
		}
		for bit := 0; bit < 32; bit++ {
			if collectibles&(1<<uint(bit)) != 0 {
				if name, ok := lookupSceneCheckName("OOT", sceneIdx, "collect", bit); ok {
					appendCheck(ootSceneCheckID(sceneIdx, "collect", bit), name)
				}
			}
		}
	}

	// MM scene flag-based checks (permanent only)
	for sceneIdx := 0; sceneIdx < MmPermCount; sceneIdx++ {
		sf := &state.Mm.SceneFlags[sceneIdx]
		for bit := 0; bit < 32; bit++ {
			if sf.Chests&(1<<uint(bit)) != 0 {
				if name, ok := lookupSceneCheckName("MM", sceneIdx, "chest", bit); ok {
					appendCheck(mmSceneCheckID(sceneIdx, "chest", bit), name)
				}
			}
		}
		for bit := 0; bit < 32; bit++ {
			if sf.Collectibles&(1<<uint(bit)) != 0 {
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
	appendBitmapChecks(state.Shared.Bitmap("xflagsOot"), "OOT", "xflag", xflagCheckName)
	appendBitmapChecks(state.Shared.Bitmap("xflagsMm"), "MM", "xflag", xflagCheckName)
	appendBitmapChecks(state.Shared.Bitmap("shopsOot"), "OOT", "shop", shopCheckName)
	appendBitmapChecks(state.Shared.Bitmap("shopsMm"), "MM", "shop", shopCheckName)
	appendBitmapChecks(state.Shared.Bitmap("scrubsOot"), "OOT", "scrub", scrubCheckName)
	appendBitmapChecks(state.Shared.Bitmap("srOot"), "OOT", "sr", silverRupeeCheckName)

	mmFlags2 := state.Oot.ExtraRecords[ExtraIdxMmFlags2]
	if mmFlags2&(1<<mmExtraFlags2Notebook) != 0 {
		if name, ok := npcSymbolCheckName("MM", "BOMBER_NOTEBOOK"); ok {
			appendCheck("MM_extra_"+itoa(mmExtraFlags2Notebook), name)
		}
	}
	if mmFlags2&(1<<mmExtraFlags2MaskBlast) != 0 {
		if name, ok := npcSymbolCheckName("MM", "MASK_BLAST"); ok {
			appendCheck("MM_extra_"+itoa(mmExtraFlags2MaskBlast), name)
		}
	}

	hasOotEventCheck := func(flag int) bool {
		word := flag >> 4
		if word < 0 || word >= len(state.Oot.EventsChk) {
			return false
		}
		return state.Oot.EventsChk[word]&(1<<uint(flag&0xF)) != 0
	}

	if hasOotEventCheck(ootEventSongSariaVanilla) || hasOotEventCheck(ootEventSongSariaCustom) {
		if name, ok := npcSymbolCheckName("OOT", "SARIA_SONG"); ok {
			appendCheck("OOT_event_song_saria", name)
		}
	}
	if hasOotEventCheck(ootEventSongSunCustom) {
		if name, ok := npcSymbolCheckName("OOT", "ROYAL_TOMB_SONG"); ok {
			appendCheck("OOT_event_song_sun", name)
		}
	}

	return checks
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
		case "mm-derived-skeleton-key":
			qty = boolToInt(hasMmSkeletonKey(&state.Mm))
		case "mm-derived-transcendent-fairy":
			qty = boolToInt(hasMmTranscendentFairy(&state.Mm))
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

func ootIsMqSpirit(oot *OotState) bool {
	return ootSilverRupeeLimit(oot, 2) > 0 && ootSilverRupeeLimit(oot, 3) > 0 && ootSilverRupeeLimit(oot, 4) == 0
}

func ootIsMqGanonCastle(oot *OotState) bool {
	return ootSilverRupeeLimit(oot, 14) > 0 && ootSilverRupeeLimit(oot, 15) > 0 && ootSilverRupeeLimit(oot, 16) > 0 && ootSilverRupeeLimit(oot, 17) == 0
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

func hasMmTranscendentFairy(mm *MmState) bool {
	if !mm.TownStrayFairy {
		return false
	}
	for index := 0; index < len(mmSkeletonKeyMaxKeys); index++ {
		if mm.StrayFairies[index] < 15 {
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
