package ootmm

import "math/bits"

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

	// Equipment levels (from bitfield: boots:4, tunics:4, shields:4, swords:4)
	items = append(items, TrackedItem{"OOT_SWORD", bits.OnesCount16(oot.Equipment & 0xF)})
	items = append(items, TrackedItem{"OOT_SHIELD", bits.OnesCount16((oot.Equipment >> 4) & 0xF)})
	items = append(items, TrackedItem{"OOT_TUNIC", bits.OnesCount16((oot.Equipment >> 8) & 0xF)})
	items = append(items, TrackedItem{"OOT_BOOTS", bits.OnesCount16((oot.Equipment >> 12) & 0xF)})

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
		name := ootInventorySlotName(i)
		if name == "" {
			continue
		}
		qty := ootInventorySlotQty(i, itemID, oot.Beans)
		if qty > 0 {
			items = append(items, TrackedItem{name, qty})
		}
	}

	// Dungeon items
	for i := 0; i < 20; i++ {
		di := oot.DungeonItems[i]
		prefix := ootDungeonName(i)
		if prefix == "" {
			continue
		}
		items = append(items, TrackedItem{prefix + "_BOSS_KEY", boolToInt(di&0x80 != 0)})
		items = append(items, TrackedItem{prefix + "_COMPASS", boolToInt(di&0x40 != 0)})
		items = append(items, TrackedItem{prefix + "_MAP", boolToInt(di&0x20 != 0)})
	}
	for i := 0; i < 19; i++ {
		name := ootDungeonName(i)
		if name == "" {
			continue
		}
		keys := oot.DungeonKeys[i]
		if keys < 0 {
			keys = 0
		}
		items = append(items, TrackedItem{name + "_KEYS", int(keys)})
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

	// MM inventory items
	for i, itemID := range mm.Items {
		name := mmInventorySlotName(i)
		if name == "" {
			continue
		}
		qty := mmInventorySlotQty(i, itemID)
		if qty > 0 {
			items = append(items, TrackedItem{name, qty})
		}
	}

	// MM Upgrades
	items = append(items, TrackedItem{"MM_QUIVER", GetUpgradeLevel(mm.Upgrades, 0, 3)})
	items = append(items, TrackedItem{"MM_BOMB_BAG", GetUpgradeLevel(mm.Upgrades, 3, 3)})
	items = append(items, TrackedItem{"MM_STRENGTH", GetUpgradeLevel(mm.Upgrades, 6, 3)})
	items = append(items, TrackedItem{"MM_SCALE", GetUpgradeLevel(mm.Upgrades, 9, 3)})
	items = append(items, TrackedItem{"MM_WALLET", GetUpgradeLevel(mm.Upgrades, 12, 2)})

	// MM Dungeon items
	for i := 0; i < 10; i++ {
		di := mm.DungeonItems[i]
		prefix := mmDungeonName(i)
		if prefix == "" {
			continue
		}
		items = append(items, TrackedItem{prefix + "_BOSS_KEY", boolToInt(di&0x80 != 0)})
		items = append(items, TrackedItem{prefix + "_COMPASS", boolToInt(di&0x40 != 0)})
		items = append(items, TrackedItem{prefix + "_MAP", boolToInt(di&0x20 != 0)})
	}
	for i := 0; i < 9; i++ {
		name := mmDungeonName(i)
		if name == "" {
			continue
		}
		keys := mm.DungeonKeys[i]
		if keys < 0 {
			keys = 0
		}
		items = append(items, TrackedItem{name + "_KEYS", int(keys)})
	}

	// MM Stray Fairies
	for i := 0; i < 10; i++ {
		name := mmDungeonName(i)
		if name == "" {
			continue
		}
		items = append(items, TrackedItem{name + "_STRAY_FAIRIES", int(mm.StrayFairies[i])})
	}

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

// OoT inventory slot names. Returns "" for unused/unknown slots.
func ootInventorySlotName(slot int) string {
	if slot >= 0 && slot < len(ootInventorySlots) {
		return ootInventorySlots[slot]
	}
	return ""
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

// MM inventory slot names.
func mmInventorySlotName(slot int) string {
	if slot >= 0 && slot < len(mmInventorySlots) {
		return mmInventorySlots[slot]
	}
	return ""
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

const emptyInventoryItem = 0xFF

const (
	ootSlotOcarina   = 7
	ootSlotHookshot  = 9
	ootSlotBeans     = 14
	ootSlotTradeAdult = 22
	ootSlotTradeChild = 23

	mmSlotOcarina   = 0
	mmSlotTrade1    = 5
	mmSlotTrade2    = 11
	mmSlotHookshot  = 15
	mmSlotGFS       = 16
	mmSlotTrade3    = 17
)

var (
	ootOcarinaStages = []uint8{0x07, 0x08}
	ootHookshotStages = []uint8{0x0A, 0x0B}
	ootAdultTradeStages = []uint8{0x2D, 0x2E, 0x2F, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x14}
	ootChildTradeStages = []uint8{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x9C, 0x9D, 0x14}

	mmOcarinaStages = []uint8{0x05, 0x00}
	mmTrade1Stages  = []uint8{0xB0, 0x28, 0x29, 0x2A, 0x2B, 0x2C}
	mmTrade2Stages  = []uint8{0xAE, 0xB1, 0xB3, 0x2D, 0x2E}
	mmHookshotStages = []uint8{0x11, 0x0F}
	mmGFSStages      = []uint8{0x10, 0xB5}
	mmTrade3Stages   = []uint8{0xAF, 0xB2, 0xB4, 0x2F, 0x30}
)

func ootInventorySlotQty(slot int, itemID, beans uint8) int {
	if itemID == emptyInventoryItem {
		return 0
	}

	switch slot {
	case ootSlotOcarina:
		return stageQty(itemID, ootOcarinaStages)
	case ootSlotHookshot:
		return stageQty(itemID, ootHookshotStages)
	case ootSlotBeans:
		if beans > 0 {
			return int(beans)
		}
		return 1
	case ootSlotTradeAdult:
		if isOotBottleItem(itemID) {
			return len(ootAdultTradeStages)
		}
		return stageQty(itemID, ootAdultTradeStages)
	case ootSlotTradeChild:
		if isOotBottleItem(itemID) {
			return len(ootChildTradeStages)
		}
		return stageQty(itemID, ootChildTradeStages)
	default:
		return 1
	}
}

func mmInventorySlotQty(slot int, itemID uint8) int {
	if itemID == emptyInventoryItem {
		return 0
	}

	switch slot {
	case mmSlotOcarina:
		return stageQty(itemID, mmOcarinaStages)
	case mmSlotTrade1:
		return stageQty(itemID, mmTrade1Stages)
	case mmSlotTrade2:
		return stageQty(itemID, mmTrade2Stages)
	case mmSlotHookshot:
		return stageQty(itemID, mmHookshotStages)
	case mmSlotGFS:
		return stageQty(itemID, mmGFSStages)
	case mmSlotTrade3:
		return stageQty(itemID, mmTrade3Stages)
	default:
		return 1
	}
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
