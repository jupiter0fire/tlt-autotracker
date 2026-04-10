package ootmm

import "testing"

func TestExtractItemsNormalizesOotInventorySlots(t *testing.T) {
	state := &GameState{}
	state.Oot.Items[mustOotInventorySlotIndex("OOT_OCARINA")] = 0x08
	state.Oot.Items[mustOotInventorySlotIndex("OOT_HOOKSHOT")] = 0x0B
	state.Oot.Items[mustOotInventorySlotIndex("OOT_BOTTLE_1")] = 0x14
	state.Oot.Items[mustOotInventorySlotIndex("OOT_ADULT_TRADE")] = 0x37
	state.Oot.Items[mustOotInventorySlotIndex("OOT_MAGIC_BEANS")] = 0x10
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
	if got := items["OOT_ADULT_TRADE"]; got != 11 {
		t.Fatalf("OOT_ADULT_TRADE = %d, want 11", got)
	}
	if got := items["OOT_MAGIC_BEANS"]; got != 5 {
		t.Fatalf("OOT_MAGIC_BEANS = %d, want 5", got)
	}
}

func TestExtractItemsNormalizesMmInventorySlots(t *testing.T) {
	state := &GameState{}
	state.Mm.Items[mustMmInventorySlotIndex("MM_OCARINA")] = 0x00
	state.Mm.Items[mustMmInventorySlotIndex("MM_TRADE_1")] = 0x2C
	state.Mm.Items[mustMmInventorySlotIndex("MM_TRADE_2")] = 0xB3
	state.Mm.Items[mustMmInventorySlotIndex("MM_HOOKSHOT")] = 0x11
	state.Mm.Items[mustMmInventorySlotIndex("MM_GREAT_FAIRY_SWORD")] = 0xB5
	state.Mm.Items[mustMmInventorySlotIndex("MM_TRADE_3")] = 0x30
	state.Mm.Items[mustMmInventorySlotIndex("MM_MASK_POSTMAN")] = 0x3E
	state.Mm.Items[mustMmInventorySlotIndex("MM_BOTTLE_1")] = 0x13

	items := itemQtyMap(ExtractItems(state))

	if got := items["MM_OCARINA"]; got != 2 {
		t.Fatalf("MM_OCARINA = %d, want 2", got)
	}
	if got := items["MM_TRADE_1"]; got != 6 {
		t.Fatalf("MM_TRADE_1 = %d, want 6", got)
	}
	if got := items["MM_TRADE_2"]; got != 3 {
		t.Fatalf("MM_TRADE_2 = %d, want 3", got)
	}
	if got := items["MM_HOOKSHOT"]; got != 1 {
		t.Fatalf("MM_HOOKSHOT = %d, want 1", got)
	}
	if got := items["MM_GREAT_FAIRY_SWORD"]; got != 2 {
		t.Fatalf("MM_GREAT_FAIRY_SWORD = %d, want 2", got)
	}
	if got := items["MM_TRADE_3"]; got != 5 {
		t.Fatalf("MM_TRADE_3 = %d, want 5", got)
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
	if got := items["OOT_TUNIC"]; got != 2 {
		t.Fatalf("OOT_TUNIC = %d, want 2", got)
	}
	if got := items["OOT_BOOTS"]; got != 1 {
		t.Fatalf("OOT_BOOTS = %d, want 1", got)
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

func TestExtractItemsIncludesSoulBitmaps(t *testing.T) {
	state := &GameState{}
	state.Shared.SoulsEnemyOot[0] = 1 << 0
	state.Shared.SoulsBossMm[0] = 1 << 1

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

func TestExtractItemsIncludesMmSpecialItems(t *testing.T) {
	state := &GameState{}
	state.Oot.ExtraRecords[ExtraIdxMmItems] = 1 << mmExtraHammerShift
	state.Oot.ExtraRecords[ExtraIdxMmTrade] = (1 << mmExtraTradeObtained1Shift) | (1 << (mmExtraTradeObtained2Shift + 3))
	state.Oot.ExtraRecords[ExtraIdxMmFlags3] = (1 << mmExtraFlags3Bottomless) | (1 << mmExtraFlags3StoneAgony)

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
