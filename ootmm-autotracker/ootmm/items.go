package ootmm

import "math/bits"

const (
	dungeonItemMapMask     = 0x04
	dungeonItemCompassMask = 0x02
	dungeonItemBossKeyMask = 0x01
)

// TrackedItem represents a single trackable item with its current quantity.
type TrackedItem struct {
	ID  string `json:"id"`
	Qty int    `json:"qty"`
}

// TrackedCheck represents a single location check.
type TrackedCheck struct {
	ID      string `json:"id"`
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
		if qty > 0 {
			items = append(items, TrackedItem{entry.ItemID, qty})
		}
	}

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
		if qty > 0 {
			items = append(items, TrackedItem{entry.ItemID, qty})
		}
	}

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
	items = appendCatalogItems(items, state)

	return items
}

// ExtractChecks extracts location checks from scene flags.
func ExtractChecks(state *GameState) []TrackedCheck {
	checks := make([]TrackedCheck, 0, 256)

	// OoT scene flag-based checks
	for sceneIdx := 0; sceneIdx < OotPermCount; sceneIdx++ {
		sf := &state.Oot.SceneFlags[sceneIdx]
		// Each set bit in the chests field = one opened chest
		for bit := 0; bit < 32; bit++ {
			if sf.Chests&(1<<uint(bit)) != 0 {
				id := ootSceneCheckID(sceneIdx, "chest", bit)
				checks = append(checks, TrackedCheck{id, true})
			}
		}
		for bit := 0; bit < 32; bit++ {
			if sf.Collectibles&(1<<uint(bit)) != 0 {
				id := ootSceneCheckID(sceneIdx, "collect", bit)
				checks = append(checks, TrackedCheck{id, true})
			}
		}
	}

	// MM scene flag-based checks (permanent only)
	for sceneIdx := 0; sceneIdx < MmPermCount; sceneIdx++ {
		sf := &state.Mm.SceneFlags[sceneIdx]
		for bit := 0; bit < 32; bit++ {
			if sf.Chests&(1<<uint(bit)) != 0 {
				id := mmSceneCheckID(sceneIdx, "chest", bit)
				checks = append(checks, TrackedCheck{id, true})
			}
		}
		for bit := 0; bit < 32; bit++ {
			if sf.Collectibles&(1<<uint(bit)) != 0 {
				id := mmSceneCheckID(sceneIdx, "collect", bit)
				checks = append(checks, TrackedCheck{id, true})
			}
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
	return "OOT_" + kind + "_" + itoa(scene) + "_" + itoa(bit)
}

func mmSceneCheckID(scene int, kind string, bit int) string {
	return "MM_" + kind + "_" + itoa(scene) + "_" + itoa(bit)
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
		"OOT_ICE_CAVERN", "OOT_GANONS_TOWER", "", "", "",
		"", "", "", "OOT_GERUDO_FORTRESS", "OOT_GERUDO_TRAINING",
		"OOT_GANONS_CASTLE",
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
	case 17:
		return "", "", ""
	case 18:
		return "", "OOT_COMPASS_GTG", "OOT_MAP_GTG"
	case 19:
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
	case 10:
		return "OOT_SMALL_KEY_GANON"
	case 17:
		return "OOT_SMALL_KEY_GF"
	case 18:
		return "OOT_SMALL_KEY_GTG"
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
		case "shared-bitmap-bit":
			qty = boolToInt(bitmapHasBit(state.Shared.Bitmap(entry.Source.Block), entry.Source.Bit))
		case "oot-extra-bit":
			if entry.Source.Record >= 0 && entry.Source.Record < len(state.Oot.ExtraRecords) {
				qty = boolToInt(state.Oot.ExtraRecords[entry.Source.Record]&(1<<uint(entry.Source.Bit)) != 0)
			}
		}
		items = append(items, TrackedItem{ID: entry.ItemID, Qty: qty})
	}
	return items
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
