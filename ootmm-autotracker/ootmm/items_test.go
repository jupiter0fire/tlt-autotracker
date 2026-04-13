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

func TestExtractItemsPublishesOotSwordAndShieldBitmasks(t *testing.T) {
	state := &GameState{}
	state.Oot.Equipment = 0x1537

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_SWORD"]; got != 7 {
		t.Fatalf("OOT_SWORD = %d, want 7", got)
	}
	if got := items["OOT_SHIELD"]; got != 3 {
		t.Fatalf("OOT_SHIELD = %d, want 3", got)
	}
	if got := items["OOT_TUNIC"]; got != 3 {
		t.Fatalf("OOT_TUNIC = %d, want 3", got)
	}
	if got := items["OOT_BOOTS"]; got != 1 {
		t.Fatalf("OOT_BOOTS = %d, want 1", got)
	}
}

func TestExtractItemsReportsOotSwordBitmaskWithoutKokiri(t *testing.T) {
	state := &GameState{}
	state.Oot.Equipment = 0x0002

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_SWORD"]; got != 2 {
		t.Fatalf("OOT_SWORD = %d, want 2", got)
	}
}

func TestExtractItemsReportsOotSwordBitmaskWithBiggoron(t *testing.T) {
	state := &GameState{}
	state.Oot.Equipment = 0x0006
	state.Oot.IsBiggoronSword = true

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_SWORD"]; got != 6 {
		t.Fatalf("OOT_SWORD = %d, want 6", got)
	}
}

func TestExtractItemsReportsBrokenKnifeBitmask(t *testing.T) {
	state := &GameState{}
	state.Oot.Equipment = 0x0008

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_SWORD"]; got != 8 {
		t.Fatalf("OOT_SWORD = %d, want 8", got)
	}
}

func TestExtractItemsPublishesOotKokiriSwordAndDekuShieldBits(t *testing.T) {
	state := &GameState{}
	state.Oot.Equipment = 0x0031

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_SWORD"]; got != 1 {
		t.Fatalf("OOT_SWORD = %d, want 1", got)
	}
	if got := items["OOT_SHIELD"]; got != 3 {
		t.Fatalf("OOT_SHIELD = %d, want 3", got)
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

func TestExtractItemsIncludesSpecificRequestedSouls(t *testing.T) {
	state := &GameState{}
	for _, itemID := range []string{"OOT_SOUL_ENEMY_ANUBIS", "OOT_SOUL_NPC_ANJU"} {
		source := mustCatalogItemSource(itemID)
		state.Shared.SetBit(source.Block, source.Bit)
	}

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_SOUL_ENEMY_ANUBIS"]; got != 1 {
		t.Fatalf("OOT_SOUL_ENEMY_ANUBIS = %d, want 1", got)
	}
	if got := items["OOT_SOUL_NPC_ANJU"]; got != 1 {
		t.Fatalf("OOT_SOUL_NPC_ANJU = %d, want 1", got)
	}
}

func TestExtractItemsKeepsMmComponentsWhenSpecialItemsPresent(t *testing.T) {
	state := &GameState{}
	state.Mm.DungeonItems[0] = 0x08
	state.Mm.DungeonItems[1] = 0x18
	state.Mm.DungeonItems[2] = 0x08
	state.Mm.DungeonItems[3] = 0x20
	state.Mm.DungeonKeys[0] = 1
	state.Mm.DungeonKeys[1] = 3
	state.Mm.DungeonKeys[2] = 1
	state.Mm.DungeonKeys[3] = 4
	state.Mm.StrayFairies[0] = 15
	state.Mm.StrayFairies[1] = 15
	state.Mm.StrayFairies[2] = 15
	state.Mm.StrayFairies[3] = 15
	state.Mm.TownStrayFairy = true

	items := itemQtyMap(ExtractItems(state))

	if _, ok := items["MM_SKELETON_KEY"]; ok {
		t.Fatal("MM_SKELETON_KEY should not be exported")
	}
	if _, ok := items["MM_TRANSCENDENT_FAIRY"]; ok {
		t.Fatal("MM_TRANSCENDENT_FAIRY should not be exported")
	}
	if got := items["MM_SMALL_KEY_WF"]; got != 1 {
		t.Fatalf("MM_SMALL_KEY_WF = %d, want 1", got)
	}
	if got := items["MM_SMALL_KEY_SH"]; got != 3 {
		t.Fatalf("MM_SMALL_KEY_SH = %d, want 3", got)
	}
	if got := items["MM_SMALL_KEY_GB"]; got != 1 {
		t.Fatalf("MM_SMALL_KEY_GB = %d, want 1", got)
	}
	if got := items["MM_SMALL_KEY_ST"]; got != 4 {
		t.Fatalf("MM_SMALL_KEY_ST = %d, want 4", got)
	}
	if got := items["MM_STRAY_FAIRY_WF"]; got != 15 {
		t.Fatalf("MM_STRAY_FAIRY_WF = %d, want 15", got)
	}
	if got := items["MM_STRAY_FAIRY_SH"]; got != 15 {
		t.Fatalf("MM_STRAY_FAIRY_SH = %d, want 15", got)
	}
	if got := items["MM_STRAY_FAIRY_GB"]; got != 15 {
		t.Fatalf("MM_STRAY_FAIRY_GB = %d, want 15", got)
	}
	if got := items["MM_STRAY_FAIRY_ST"]; got != 15 {
		t.Fatalf("MM_STRAY_FAIRY_ST = %d, want 15", got)
	}
	if got := items["MM_STRAY_FAIRY_TOWN"]; got != 1 {
		t.Fatalf("MM_STRAY_FAIRY_TOWN = %d, want 1", got)
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
	state.Mm.Items[mustMmInventorySlotIndex("MM_TRADE_1")] = 0x2C        // Ocean Title Deed

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

func TestExtractItemsIgnoresActiveMmTradeBitsForSpecialItems(t *testing.T) {
	state := &GameState{}
	// The upper 16 bits of MmExtraTrade store active slot contents, not obtained items.
	state.Oot.ExtraRecords[ExtraIdxMmTrade] = 1 << 27

	items := itemQtyMap(ExtractItems(state))

	if got := items["MM_SPELL_LOVE"]; got != 0 {
		t.Fatalf("MM_SPELL_LOVE = %d, want 0 when only active MM trade bits are set", got)
	}
	if got := items["MM_BOOTS_IRON"]; got != 0 {
		t.Fatalf("MM_BOOTS_IRON = %d, want 0 when only active MM trade bits are set", got)
	}
	if got := items["MM_TRADE_1"]; got != 0 {
		t.Fatalf("MM_TRADE_1 = %#x, want 0 when only active MM trade bits are set", got)
	}
}

func TestExtractItemsDerivesRequestedMmSpecialItems(t *testing.T) {
	state := &GameState{}
	state.Mm.TownStrayFairy = true
	for index := 0; index < len(mmSkeletonKeyMaxKeys); index++ {
		state.Mm.StrayFairies[index] = 15
		state.Mm.DungeonItems[index] = uint8(mmSkeletonKeyMaxKeys[index] << 3)
	}

	items := itemQtyMap(ExtractItems(state))

	if _, ok := items["MM_SKELETON_KEY"]; ok {
		t.Fatal("MM_SKELETON_KEY should not be exported")
	}
	if _, ok := items["MM_TRANSCENDENT_FAIRY"]; ok {
		t.Fatal("MM_TRANSCENDENT_FAIRY should not be exported")
	}

	state.Mm.DungeonItems[0] = 0
	state.Mm.StrayFairies[0] = 0
	items = itemQtyMap(ExtractItems(state))

	if _, ok := items["MM_SKELETON_KEY"]; ok {
		t.Fatal("MM_SKELETON_KEY should stay omitted after clearing dungeon state")
	}
	if _, ok := items["MM_TRANSCENDENT_FAIRY"]; ok {
		t.Fatal("MM_TRANSCENDENT_FAIRY should stay omitted after clearing fairy state")
	}
}

func TestExtractItemsDerivesOotCombinedItems(t *testing.T) {
	state := &GameState{}
	state.Oot.GoldTokens = 100
	state.Oot.RuntimeMaxKeys = [OotRuntimeSceneCount]uint8{0, 0, 0, 5, 7, 5, 5, 5, 3, 0, 0, 9, 4, 2, 0, 0, 6}
	state.Oot.HasRuntimeMaxKeys = true
	state.Oot.DungeonItems[3] = uint8(state.Oot.RuntimeMaxKeys[3] << 3)
	state.Oot.DungeonItems[4] = uint8(state.Oot.RuntimeMaxKeys[4] << 3)
	state.Oot.DungeonItems[5] = uint8(state.Oot.RuntimeMaxKeys[5] << 3)
	state.Oot.DungeonItems[6] = uint8(state.Oot.RuntimeMaxKeys[6] << 3)
	state.Oot.DungeonItems[7] = uint8(state.Oot.RuntimeMaxKeys[7] << 3)
	state.Oot.DungeonItems[8] = uint8(state.Oot.RuntimeMaxKeys[8] << 3)
	state.Oot.DungeonItems[11] = uint8(state.Oot.RuntimeMaxKeys[11] << 3)
	state.Oot.DungeonItems[12] = uint8(state.Oot.RuntimeMaxKeys[12] << 3)
	state.Oot.DungeonItems[13] = uint8(state.Oot.RuntimeMaxKeys[13] << 3)
	state.Oot.DungeonItems[16] = uint8(state.Oot.RuntimeMaxKeys[16] << 3)
	state.Oot.RuntimeSilverRupeeCounts = [OotSilverRupeeSetCount]uint8{5, 5, 5, 5, 5, 5, 0, 5, 5, 5, 5, 5, 6, 3, 5, 5, 5, 0}
	state.Oot.HasRuntimeSilverRupeeCounts = true
	for silverRupeeID := 0; silverRupeeID < OotSilverRupeeSetCount; silverRupeeID++ {
		want := int(state.Oot.RuntimeSilverRupeeCounts[silverRupeeID])
		if want == 0 {
			continue
		}
		setOotSilverRupeeCount(state, silverRupeeID, want)
	}

	items := itemQtyMap(ExtractItems(state))

	for _, itemID := range []string{"OOT_KEY_RING_FOREST", "OOT_KEY_RING_FIRE", "OOT_KEY_RING_WATER", "OOT_KEY_RING_SPIRIT", "OOT_KEY_RING_SHADOW", "OOT_KEY_RING_BOTW", "OOT_KEY_RING_GTG", "OOT_KEY_RING_GF", "OOT_KEY_RING_GANON", "OOT_KEY_RING_TCG", "OOT_SKELETON_KEY", "OOT_PLATINUM_TOKEN", "OOT_RUPEE_MAGICAL"} {
		if _, ok := items[itemID]; ok {
			t.Fatalf("%s should not be exported", itemID)
		}
	}

	state.Oot.GoldTokens = 99
	setOotSilverRupeeCount(state, 1, int(state.Oot.RuntimeSilverRupeeCounts[1])-1)
	state.Oot.DungeonItems[3] = 0
	items = itemQtyMap(ExtractItems(state))

	for _, itemID := range []string{"OOT_KEY_RING_FOREST", "OOT_KEY_RING_FIRE", "OOT_KEY_RING_WATER", "OOT_KEY_RING_SPIRIT", "OOT_KEY_RING_SHADOW", "OOT_KEY_RING_BOTW", "OOT_KEY_RING_GTG", "OOT_KEY_RING_GF", "OOT_KEY_RING_GANON", "OOT_KEY_RING_TCG", "OOT_SKELETON_KEY", "OOT_PLATINUM_TOKEN", "OOT_RUPEE_MAGICAL"} {
		if _, ok := items[itemID]; ok {
			t.Fatalf("%s should stay omitted after lowering source state", itemID)
		}
	}
}

func TestExtractItemsDerivesMmCombinedItems(t *testing.T) {
	state := &GameState{}
	for index := 0; index < len(mmSkeletonKeyMaxKeys); index++ {
		state.Mm.DungeonItems[index] = uint8(mmSkeletonKeyMaxKeys[index] << 3)
	}
	state.Mm.SkullTokensSwamp = 30
	state.Mm.SkullTokensOcean = 30

	items := itemQtyMap(ExtractItems(state))

	for _, itemID := range []string{"MM_KEY_RING_WF", "MM_KEY_RING_SH", "MM_KEY_RING_GB", "MM_KEY_RING_ST", "MM_PLATINUM_TOKEN"} {
		if _, ok := items[itemID]; ok {
			t.Fatalf("%s should not be exported", itemID)
		}
	}

	state.Mm.SkullTokensOcean = 29
	state.Mm.DungeonItems[0] = 0
	items = itemQtyMap(ExtractItems(state))

	for _, itemID := range []string{"MM_KEY_RING_WF", "MM_KEY_RING_SH", "MM_KEY_RING_GB", "MM_KEY_RING_ST", "MM_PLATINUM_TOKEN"} {
		if _, ok := items[itemID]; ok {
			t.Fatalf("%s should stay omitted after lowering source state", itemID)
		}
	}
}

func TestExtractItemsTreatsSharedBombchuBagAsOwnedBombchu(t *testing.T) {
	state := &GameState{}
	for i := range state.Oot.Items {
		state.Oot.Items[i] = emptyInventoryItem
	}
	for i := range state.Mm.Items {
		state.Mm.Items[i] = emptyInventoryItem
	}
	state.Shared.BombchuBagOot = 2
	state.Shared.BombchuBagMm = 1

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_BOMBCHUS"]; got != 1 {
		t.Fatalf("OOT_BOMBCHUS = %d, want 1 from shared bombchu bag", got)
	}
	if got := items["MM_BOMBCHU"]; got != 1 {
		t.Fatalf("MM_BOMBCHU = %d, want 1 from shared bombchu bag", got)
	}
}

func setOotSilverRupeeCount(state *GameState, silverRupeeID, count int) {
	recordIndex := ExtraIdxOotSilver1 + silverRupeeID/4
	shift := uint((silverRupeeID % 4) * 8)
	mask := uint32(0xff) << shift
	state.Oot.ExtraRecords[recordIndex] = (state.Oot.ExtraRecords[recordIndex] &^ mask) | (uint32(count) << shift)
}

func itemQtyMap(items []TrackedItem) map[string]int {
	result := make(map[string]int, len(items))
	for _, item := range items {
		result[item.ID] = item.Qty
	}
	return result
}

func TestExtractChecksUsesResolvedNamesWhenAvailable(t *testing.T) {
	original := checkNameTable
	checkNameTable = map[string]string{
		"OOT_chest_40_0": "Mido's House Top Left",
		"OOT_chest_40_1": "Mido's House Top Right",
		"OOT_chest_40_2": "Mido's House Bottom Left",
		"OOT_chest_40_3": "Mido's House Bottom Right",
	}
	t.Cleanup(func() {
		checkNameTable = original
	})

	state := &GameState{}
	state.Oot.SceneFlags[0x28].Chests = 0x0F

	checks := checkNameSet(ExtractChecks(state))
	for _, name := range []string{
		"Mido's House Top Left",
		"Mido's House Top Right",
		"Mido's House Bottom Left",
		"Mido's House Bottom Right",
	} {
		if _, ok := checks[name]; !ok {
			t.Fatalf("missing named check %q", name)
		}
	}
}

func TestExtractChecksFallsBackToStableKey(t *testing.T) {
	original := checkNameTable
	checkNameTable = map[string]string{}
	t.Cleanup(func() {
		checkNameTable = original
	})

	state := &GameState{}
	state.Oot.SceneFlags[1].Chests = 1 << 7

	checks := ExtractChecks(state)
	if len(checks) != 0 {
		t.Fatalf("len(checks) = %d, want 0 for unmapped scene flags", len(checks))
	}
}

func TestExtractChecksIncludesNpcBitmapChecks(t *testing.T) {
	originalNpc := npcCheckTables
	npcCheckTables = map[string]map[int]string{
		"OOT": {},
		"MM":  {0x4a: "Clock Town Blast Mask"},
	}
	t.Cleanup(func() {
		npcCheckTables = originalNpc
	})

	state := &GameState{}
	state.Shared.SetBit("npcMm", 0x4a)

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Clock Town Blast Mask"]; !ok {
		t.Fatal("missing MM npc check from shared bitmap")
	}
}

func TestExtractChecksIncludesXflagBitmapChecks(t *testing.T) {
	originalXflags := xflagCheckTables
	xflagCheckTables = map[string]map[int]string{
		"OOT": {17: "Kokiri Forest Grass Adult Near Crawl 1"},
		"MM":  {},
	}
	t.Cleanup(func() {
		xflagCheckTables = originalXflags
	})

	state := &GameState{}
	state.Shared.SetBit("xflagsOot", 17)

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Kokiri Forest Grass Adult Near Crawl 1"]; !ok {
		t.Fatal("missing OoT xflag check from shared bitmap")
	}
}

func TestExtractChecksIncludesMmExtraFlagChecks(t *testing.T) {
	originalSymbols := npcSymbolTables
	npcSymbolTables = map[string]map[string]string{
		"OOT": {},
		"MM": {
			"MASK_BLAST":      "Clock Town Blast Mask",
			"BOMBER_NOTEBOOK": "Clock Town Bomber Notebook",
		},
	}
	t.Cleanup(func() {
		npcSymbolTables = originalSymbols
	})

	state := &GameState{}
	state.Oot.ExtraRecords[ExtraIdxMmFlags2] = (1 << mmExtraFlags2Notebook) | (1 << mmExtraFlags2MaskBlast)

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Clock Town Blast Mask"]; !ok {
		t.Fatal("missing MM extra-flag check for Blast Mask")
	}
	if _, ok := checks["Clock Town Bomber Notebook"]; !ok {
		t.Fatal("missing MM extra-flag check for Bomber Notebook")
	}
}

func checkNameSet(checks []TrackedCheck) map[string]struct{} {
	result := make(map[string]struct{}, len(checks))
	for _, check := range checks {
		result[check.Name] = struct{}{}
	}
	return result
}
