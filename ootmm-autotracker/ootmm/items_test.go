package ootmm

import "testing"

func TestExtractItemsNormalizesOotInventorySlots(t *testing.T) {
	state := &GameState{}
	state.Oot.Items[mustOotInventorySlotIndex("OOT_OCARINA")] = 0x08
	state.Oot.Items[mustOotInventorySlotIndex("OOT_HOOKSHOT")] = 0x0B
	state.Oot.Items[mustOotInventorySlotIndex("OOT_BOTTLE_1")] = 0x14
	state.Oot.Items[mustOotInventorySlotIndex("OOT_MAGIC_BEAN")] = 0x10
	state.Oot.Beans = 5

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_OCARINA"]; got != 2 {
		t.Fatalf("OOT_OCARINA = %d, want 2", got)
	}
	if got := items["OOT_HOOKSHOT"]; got != 2 {
		t.Fatalf("OOT_HOOKSHOT = %d, want 2", got)
	}
	if got := items["OOT_BOTTLE_1"]; got != 1 {
		t.Fatalf("OOT_BOTTLE_1 = %d, want 1", got)
	}
	if got := items["OOT_MAGIC_BEAN"]; got != 5 {
		t.Fatalf("OOT_MAGIC_BEAN = %d, want 5", got)
	}
}

func TestExtractItemsNormalizesMmInventorySlots(t *testing.T) {
	state := &GameState{}
	state.Mm.Items[mustMmInventorySlotIndex("MM_OCARINA")] = 0x00
	state.Mm.Items[mustMmInventorySlotIndex("MM_HOOKSHOT")] = 0x11
	state.Mm.Items[mustMmInventorySlotIndex("MM_GREAT_FAIRY_SWORD")] = 0xB5
	state.Mm.Items[mustMmInventorySlotIndex("MM_MASK_POSTMAN")] = 0x3E
	state.Mm.Items[mustMmInventorySlotIndex("MM_BOTTLE_1")] = 0x13

	items := itemQtyMap(ExtractItems(state))

	if got := items["MM_OCARINA"]; got != 2 {
		t.Fatalf("MM_OCARINA = %d, want 2", got)
	}
	if got := items["MM_HOOKSHOT"]; got != 1 {
		t.Fatalf("MM_HOOKSHOT = %d, want 1", got)
	}
	if got := items["MM_GREAT_FAIRY_SWORD"]; got != 2 {
		t.Fatalf("MM_GREAT_FAIRY_SWORD = %d, want 2", got)
	}
	if got := items["MM_MASK_POSTMAN"]; got != 1 {
		t.Fatalf("MM_MASK_POSTMAN = %d, want 1", got)
	}
	if got := items["MM_BOTTLE_1"]; got != 1 {
		t.Fatalf("MM_BOTTLE_1 = %d, want 1", got)
	}
}

func TestExtractItemsCountsOotEquipmentStages(t *testing.T) {
	state := &GameState{}
	state.Oot.Equipment = 0x1537

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_SWORD"]; got != 3 {
		t.Fatalf("OOT_SWORD = %d, want 3", got)
	}
	if got := items["OOT_SHIELD"]; got != 2 {
		t.Fatalf("OOT_SHIELD = %d, want 2", got)
	}
	if got := items["OOT_TUNIC"]; got != 3 {
		t.Fatalf("OOT_TUNIC = %d, want 3", got)
	}
	if got := items["OOT_BOOTS"]; got != 1 {
		t.Fatalf("OOT_BOOTS = %d, want 1", got)
	}
}

func TestExtractItemsReportsOotMasterSwordWithoutKokiri(t *testing.T) {
	state := &GameState{}
	state.Oot.Equipment = 0x0002

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_SWORD"]; got != 2 {
		t.Fatalf("OOT_SWORD = %d, want 2", got)
	}
}

func TestExtractItemsReportsOotBiggoronSword(t *testing.T) {
	state := &GameState{}
	state.Oot.Equipment = 0x0004
	state.Oot.IsBiggoronSword = true

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_SWORD"]; got != 4 {
		t.Fatalf("OOT_SWORD = %d, want 4", got)
	}
}

func TestExtractItemsTreatsBrokenKnifeAsStageThree(t *testing.T) {
	state := &GameState{}
	state.Oot.Equipment = 0x0008

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_SWORD"]; got != 3 {
		t.Fatalf("OOT_SWORD = %d, want 3", got)
	}
}

func TestParseOotSaveReadsBiggoronFlag(t *testing.T) {
	data := make([]byte, OotSaveSize)
	data[OotOffIsBiggoronSword] = 1

	var oot OotState
	if err := parseOotSave(&oot, data); err != nil {
		t.Fatalf("parseOotSave: %v", err)
	}
	if !oot.IsBiggoronSword {
		t.Fatal("expected Biggoron flag to be true")
	}
}

func TestExtractItemsReadsMmEquipmentLevels(t *testing.T) {
	state := &GameState{}
	state.Mm.Equipment = 0x0023

	items := itemQtyMap(ExtractItems(state))

	if got := items["MM_SWORD"]; got != 3 {
		t.Fatalf("MM_SWORD = %d, want 3", got)
	}
	if got := items["MM_SHIELD"]; got != 2 {
		t.Fatalf("MM_SHIELD = %d, want 2", got)
	}
}

func TestExtractItemsUsesOfficialOotBossKeyIDs(t *testing.T) {
	state := &GameState{}
	state.Oot.DungeonItems[4] = 0x01
	state.Oot.DungeonItems[6] = 0x01
	state.Oot.DungeonItems[7] = 0x01

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_BOSS_KEY_FIRE"]; got != 1 {
		t.Fatalf("OOT_BOSS_KEY_FIRE = %d, want 1", got)
	}
	if got := items["OOT_BOSS_KEY_SPIRIT"]; got != 1 {
		t.Fatalf("OOT_BOSS_KEY_SPIRIT = %d, want 1", got)
	}
	if got := items["OOT_BOSS_KEY_SHADOW"]; got != 1 {
		t.Fatalf("OOT_BOSS_KEY_SHADOW = %d, want 1", got)
	}
}

func TestExtractItemsUsesOfficialMmDungeonItemIDs(t *testing.T) {
	state := &GameState{}
	state.Mm.DungeonItems[0] = 0x07
	state.Mm.DungeonKeys[0] = 3
	state.Mm.StrayFairies[0] = 12
	state.Mm.TownStrayFairy = true

	items := itemQtyMap(ExtractItems(state))

	if got := items["MM_BOSS_KEY_WF"]; got != 1 {
		t.Fatalf("MM_BOSS_KEY_WF = %d, want 1", got)
	}
	if got := items["MM_COMPASS_WF"]; got != 1 {
		t.Fatalf("MM_COMPASS_WF = %d, want 1", got)
	}
	if got := items["MM_MAP_WF"]; got != 1 {
		t.Fatalf("MM_MAP_WF = %d, want 1", got)
	}
	if got := items["MM_SMALL_KEY_WF"]; got != 3 {
		t.Fatalf("MM_SMALL_KEY_WF = %d, want 3", got)
	}
	if got := items["MM_STRAY_FAIRY_WF"]; got != 12 {
		t.Fatalf("MM_STRAY_FAIRY_WF = %d, want 12", got)
	}
	if got := items["MM_STRAY_FAIRY_TOWN"]; got != 1 {
		t.Fatalf("MM_STRAY_FAIRY_TOWN = %d, want 1", got)
	}
}

func TestExtractItemsIncludesSoulBitmaps(t *testing.T) {
	state := &GameState{}
	ootSoul := mustCatalogItemSource("OOT_SOUL_ENEMY_STALFOS")
	mmSoul := mustCatalogItemSource("MM_SOUL_BOSS_GOHT")
	state.Shared.SetBit(ootSoul.Block, ootSoul.Bit)
	state.Shared.SetBit(mmSoul.Block, mmSoul.Bit)

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_SOUL_ENEMY_STALFOS"]; got != 1 {
		t.Fatalf("OOT_SOUL_ENEMY_STALFOS = %d, want 1", got)
	}
	if got := items["MM_SOUL_BOSS_GOHT"]; got != 1 {
		t.Fatalf("MM_SOUL_BOSS_GOHT = %d, want 1", got)
	}
	if got := items["MM_SOUL_MISC_GS"]; got != 0 {
		t.Fatalf("MM_SOUL_MISC_GS = %d, want 0", got)
	}
}

func TestExtractItemsReadsOotTradeFromExtraRecords(t *testing.T) {
	state := &GameState{}
	// OotExtraTrade: u16 child (upper 16) | u16 adult (lower 16).
	// Adult: bits 0 (Pocket Egg) + 2 (Cojiro) + 10 (Claim Check) = 0x0405
	// Child: bits 0 (Weird Egg) + 2 (Zelda's Letter) + 6 (Bunny Hood) = 0x0045
	state.Oot.ExtraRecords[ExtraIdxOotTrade] = 0x00450405

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_ADULT_TRADE"]; got != 0x0405 {
		t.Fatalf("OOT_ADULT_TRADE = %#x, want 0x0405", got)
	}
	if got := items["OOT_CHILD_TRADE"]; got != 0x0045 {
		t.Fatalf("OOT_CHILD_TRADE = %#x, want 0x0045", got)
	}
}

func TestExtractItemsReadsMmTradeFromExtraRecords(t *testing.T) {
	state := &GameState{}
	// MmExtraTrade (big-endian bitfields from MSB):
	//   trade1:6 | trade2:5 | trade3:5 | tradeObtained1:6 | tradeObtained2:5 | tradeObtained3:5
	// We only care about tradeObtained fields.
	// tradeObtained1 = 0x15 (bits 15-10), tradeObtained2 = 0x0B (bits 9-5), tradeObtained3 = 0x05 (bits 4-0)
	state.Oot.ExtraRecords[ExtraIdxMmTrade] = (0x15 << 10) | (0x0B << 5) | 0x05

	items := itemQtyMap(ExtractItems(state))

	if got := items["MM_TRADE_1"]; got != 0x15 {
		t.Fatalf("MM_TRADE_1 = %#x, want 0x15", got)
	}
	if got := items["MM_TRADE_2"]; got != 0x0B {
		t.Fatalf("MM_TRADE_2 = %#x, want 0x0B", got)
	}
	if got := items["MM_TRADE_3"]; got != 0x05 {
		t.Fatalf("MM_TRADE_3 = %#x, want 0x05", got)
	}
}

func TestExtractItemsTradeIgnoresInventorySlot(t *testing.T) {
	state := &GameState{}
	// Set inventory slot to a trade item but leave ExtraRecords empty.
	state.Oot.Items[mustOotInventorySlotIndex("OOT_ADULT_TRADE")] = 0x37 // Claim Check
	state.Mm.Items[mustMmInventorySlotIndex("MM_TRADE_1")] = 0x2C       // Ocean Title Deed

	items := itemQtyMap(ExtractItems(state))

	// Trade items should come from ExtraRecords (0), not inventory slot.
	if got := items["OOT_ADULT_TRADE"]; got != 0 {
		t.Fatalf("OOT_ADULT_TRADE = %d, want 0 (ExtraRecord empty)", got)
	}
	if got := items["MM_TRADE_1"]; got != 0 {
		t.Fatalf("MM_TRADE_1 = %d, want 0 (ExtraRecord empty)", got)
	}
}

func TestExtractItemsIncludesMmSpecialItems(t *testing.T) {
	state := &GameState{}
	for _, itemID := range []string{"MM_HAMMER", "MM_SPELL_FIRE", "MM_ROOM_KEY", "MM_WALLET5", "MM_STONE_OF_AGONY"} {
		source := mustCatalogItemSource(itemID)
		state.Oot.ExtraRecords[source.Record] |= 1 << uint(source.Bit)
	}

	items := itemQtyMap(ExtractItems(state))

	if got := items["MM_HAMMER"]; got != 1 {
		t.Fatalf("MM_HAMMER = %d, want 1", got)
	}
	if got := items["MM_SPELL_FIRE"]; got != 1 {
		t.Fatalf("MM_SPELL_FIRE = %d, want 1", got)
	}
	if got := items["MM_ROOM_KEY"]; got != 1 {
		t.Fatalf("MM_ROOM_KEY = %d, want 1", got)
	}
	if got := items["MM_WALLET5"]; got != 1 {
		t.Fatalf("MM_WALLET5 = %d, want 1", got)
	}
	if got := items["MM_STONE_OF_AGONY"]; got != 1 {
		t.Fatalf("MM_STONE_OF_AGONY = %d, want 1", got)
	}
	if got := items["MM_PENDANT_OF_MEMORIES"]; got != 0 {
		t.Fatalf("MM_PENDANT_OF_MEMORIES = %d, want 0", got)
	}
}

func itemQtyMap(items []TrackedItem) map[string]int {
	result := make(map[string]int, len(items))
	for _, item := range items {
		result[item.ID] = item.Qty
	}
	return result
}
