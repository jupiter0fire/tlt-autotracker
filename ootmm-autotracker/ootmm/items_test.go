package ootmm

import (
	"encoding/binary"
	"testing"
)

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

func TestExtractItemsReconstructsWalletLevelsFromFlags(t *testing.T) {
	state := &GameState{}
	state.Oot.Upgrades = 2 << 12
	state.Oot.ExtraRecords[ExtraIdxOotFlags] = 1 << ootExtraFlagsChildWalletBit
	state.Mm.Upgrades = 1 << 12
	state.Oot.ExtraRecords[ExtraIdxMmFlags2] = 1 << mmExtraFlags2ChildWalletBit

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_WALLET"]; got != 3 {
		t.Fatalf("OOT_WALLET = %d, want 3", got)
	}
	if got := items["MM_WALLET"]; got != 2 {
		t.Fatalf("MM_WALLET = %d, want 2", got)
	}

	state.Oot.ExtraRecords[ExtraIdxOotFlags] |= 1 << ootExtraFlagsBottomlessBit
	state.Oot.ExtraRecords[ExtraIdxMmFlags3] = 1 << mmExtraFlags3BottomlessBit
	items = itemQtyMap(ExtractItems(state))

	if got := items["OOT_WALLET"]; got != 5 {
		t.Fatalf("OOT_WALLET = %d, want 5", got)
	}
	if got := items["MM_WALLET"]; got != 5 {
		t.Fatalf("MM_WALLET = %d, want 5", got)
	}
}

func TestExtractItemsReconstructsWalletLevelsFromLiveRawFlags(t *testing.T) {
	state := &GameState{}
	state.Oot.Upgrades = 0x00162040
	state.Oot.ExtraRecords[ExtraIdxOotFlags] = 0x00020020
	state.Mm.Upgrades = 1 << 12
	state.Oot.ExtraRecords[ExtraIdxMmFlags2] = 0x80000000

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_WALLET"]; got != 3 {
		t.Fatalf("OOT_WALLET = %d, want 3", got)
	}
	if got := items["MM_WALLET"]; got != 2 {
		t.Fatalf("MM_WALLET = %d, want 2", got)
	}
}

func TestParseOotSaveReadsMagicFlags(t *testing.T) {
	data := make([]byte, OotSaveSize)
	data[OotOffMagicAcquired] = 1
	data[OotOffDoubleMagic] = 1

	var oot OotState
	if err := parseOotSave(&oot, data); err != nil {
		t.Fatalf("parseOotSave: %v", err)
	}
	if !oot.HasMagic {
		t.Fatal("expected OoT magic-acquired flag to be true")
	}
	if !oot.HasDoubleMagic {
		t.Fatal("expected OoT double-magic flag to be true")
	}
}

func TestParseOotSaveReadsOcarinaGameRound(t *testing.T) {
	data := make([]byte, OotSaveSize)
	data[OotOffOcarinaGameRound] = 1

	var oot OotState
	if err := parseOotSave(&oot, data); err != nil {
		t.Fatalf("parseOotSave: %v", err)
	}
	if oot.OcarinaGameRound != 1 {
		t.Fatalf("OcarinaGameRound = %d, want 1", oot.OcarinaGameRound)
	}
}

func TestExtractItemsReportsMagicUpgradeProgression(t *testing.T) {
	state := &GameState{}

	items := itemQtyMap(ExtractItems(state))
	if got := items["OOT_MAGIC_UPGRADE"]; got != 0 {
		t.Fatalf("OOT_MAGIC_UPGRADE = %d, want 0", got)
	}
	if got := items["MM_MAGIC_UPGRADE"]; got != 0 {
		t.Fatalf("MM_MAGIC_UPGRADE = %d, want 0", got)
	}

	state.Oot.HasMagic = true
	state.Mm.HasMagic = true
	items = itemQtyMap(ExtractItems(state))
	if got := items["OOT_MAGIC_UPGRADE"]; got != 1 {
		t.Fatalf("OOT_MAGIC_UPGRADE = %d, want 1", got)
	}
	if got := items["MM_MAGIC_UPGRADE"]; got != 1 {
		t.Fatalf("MM_MAGIC_UPGRADE = %d, want 1", got)
	}

	state.Oot.HasDoubleMagic = true
	state.Mm.HasDoubleMagic = true
	items = itemQtyMap(ExtractItems(state))
	if got := items["OOT_MAGIC_UPGRADE"]; got != 2 {
		t.Fatalf("OOT_MAGIC_UPGRADE = %d, want 2", got)
	}
	if got := items["MM_MAGIC_UPGRADE"]; got != 2 {
		t.Fatalf("MM_MAGIC_UPGRADE = %d, want 2", got)
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

func TestExtractItemsIncludesSpecialCatalogItems(t *testing.T) {
	state := &GameState{}
	for _, itemID := range []string{"MM_HAMMER", "MM_SPELL_FIRE", "MM_ROOM_KEY", "OOT_WALLET5", "MM_WALLET5", "MM_STONE_OF_AGONY"} {
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
	if got := items["OOT_WALLET5"]; got != 1 {
		t.Fatalf("OOT_WALLET5 = %d, want 1", got)
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

func TestExtractItemsIncludesSpinUpgradeSpecialItems(t *testing.T) {
	state := &GameState{}
	ootSpin := mustCatalogItemSource("OOT_SPIN_UPGRADE")
	mmSpin := mustCatalogItemSource("MM_SPIN_UPGRADE")
	state.Oot.ExtraRecords[ootSpin.Record] |= 1 << uint(ootSpin.Bit)
	state.Mm.WeekEventReg[mmSpin.Byte] |= 1 << uint(mmSpin.Bit)

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_SPIN_UPGRADE"]; got != 1 {
		t.Fatalf("OOT_SPIN_UPGRADE = %d, want 1", got)
	}
	if got := items["MM_SPIN_UPGRADE"]; got != 1 {
		t.Fatalf("MM_SPIN_UPGRADE = %d, want 1", got)
	}
	if got := items["MM_STONE_OF_AGONY"]; got != 0 {
		t.Fatalf("MM_STONE_OF_AGONY = %d, want 0", got)
	}
	if got := items["OOT_GANON_BK"]; got != 0 {
		t.Fatalf("OOT_GANON_BK = %d, want 0", got)
	}
	if got := items["MM_STRAY_FAIRY_TOWN"]; got != 0 {
		t.Fatalf("MM_STRAY_FAIRY_TOWN = %d, want 0", got)
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

func TestExtractItemsIncludesRequestedMissingItems(t *testing.T) {
	state := &GameState{}
	state.Mm.SkullTokensSwamp = 17
	state.Mm.SkullTokensOcean = 11
	state.Oot.HasMagic = true
	state.Mm.HasMagic = true
	state.Oot.ExtraRecords[ExtraIdxMmOwlFlags] = allMmOwlFlags()
	state.Shared.OcarinaButtonMaskOot = allSharedOcarinaButtonMasks()
	state.Shared.OcarinaButtonMaskMm = allSharedOcarinaButtonMasks()

	items := itemQtyMap(ExtractItems(state))

	if got := items["MM_GS_TOKEN_SWAMP"]; got != 17 {
		t.Fatalf("MM_GS_TOKEN_SWAMP = %d, want 17", got)
	}
	if got := items["MM_GS_TOKEN_OCEAN"]; got != 11 {
		t.Fatalf("MM_GS_TOKEN_OCEAN = %d, want 11", got)
	}
	if got := items["OOT_MAGIC_UPGRADE"]; got != 1 {
		t.Fatalf("OOT_MAGIC_UPGRADE = %d, want 1", got)
	}
	if got := items["MM_MAGIC_UPGRADE"]; got != 1 {
		t.Fatalf("MM_MAGIC_UPGRADE = %d, want 1", got)
	}
	for _, owl := range mmOwlItems {
		if got := items[owl.itemID]; got != 1 {
			t.Fatalf("%s = %d, want 1", owl.itemID, got)
		}
	}
	for _, button := range sharedOcarinaButtons {
		if got := items[button.ootItemID]; got != 1 {
			t.Fatalf("%s = %d, want 1", button.ootItemID, got)
		}
		if got := items[button.mmItemID]; got != 1 {
			t.Fatalf("%s = %d, want 1", button.mmItemID, got)
		}
	}
}

func TestExtractItemsIgnoresDisabledSharedOcarinaButtons(t *testing.T) {
	state := &GameState{}
	state.Shared.OcarinaButtonMaskOot = sharedOcarinaButtonMaskDisabled
	state.Shared.OcarinaButtonMaskMm = sharedOcarinaButtonMaskDisabled

	items := itemQtyMap(ExtractItems(state))

	for _, button := range sharedOcarinaButtons {
		if got := items[button.ootItemID]; got != 0 {
			t.Fatalf("%s = %d, want 0 when buttons are disabled by config", button.ootItemID, got)
		}
		if got := items[button.mmItemID]; got != 0 {
			t.Fatalf("%s = %d, want 0 when buttons are disabled by config", button.mmItemID, got)
		}
	}
}

func TestExtractItemsIncludesSharedOcarinaAButtonFromLiveMask(t *testing.T) {
	state := &GameState{}
	state.Shared.OcarinaButtonMaskOot = 0x800d
	state.Shared.OcarinaButtonMaskMm = 0x800d

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_BUTTON_A"]; got != 1 {
		t.Fatalf("OOT_BUTTON_A = %d, want 1 for live mask 0x800d", got)
	}
	if got := items["MM_BUTTON_A"]; got != 1 {
		t.Fatalf("MM_BUTTON_A = %d, want 1 for live mask 0x800d", got)
	}
	if got := items["OOT_BUTTON_C_LEFT"]; got != 0 {
		t.Fatalf("OOT_BUTTON_C_LEFT = %d, want 0 for live mask 0x800d", got)
	}
	if got := items["MM_BUTTON_C_LEFT"]; got != 0 {
		t.Fatalf("MM_BUTTON_C_LEFT = %d, want 0 for live mask 0x800d", got)
	}
}

func TestExtractItemsIncludesOotFishingPondItems(t *testing.T) {
	state := &GameState{}
	state.Shared.CaughtChildFishWeights[0] = 3
	state.Shared.CaughtChildFishWeights[1] = 2
	state.Shared.CaughtChildFishWeights[2] = fishingPondLoachWeightMask | 19
	state.Shared.CaughtChildFishWeights[3] = fishingPondLoachWeightMask | 19
	state.Shared.CaughtAdultFishWeights[0] = 3
	state.Shared.CaughtAdultFishWeights[1] = 25
	state.Shared.CaughtAdultFishWeights[2] = fishingPondLoachWeightMask | 36
	state.Shared.CaughtAdultFishWeights[3] = 1

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_FISHING_POND_CHILD_FISH_2LBS"]; got != 1 {
		t.Fatalf("OOT_FISHING_POND_CHILD_FISH_2LBS = %d, want 1", got)
	}
	if got := items["OOT_FISHING_POND_CHILD_LOACH_19LBS"]; got != 2 {
		t.Fatalf("OOT_FISHING_POND_CHILD_LOACH_19LBS = %d, want 2", got)
	}
	if got := items["OOT_FISHING_POND_ADULT_FISH_25LBS"]; got != 1 {
		t.Fatalf("OOT_FISHING_POND_ADULT_FISH_25LBS = %d, want 1", got)
	}
	if got := items["OOT_FISHING_POND_ADULT_LOACH_36LBS"]; got != 1 {
		t.Fatalf("OOT_FISHING_POND_ADULT_LOACH_36LBS = %d, want 1", got)
	}
	if _, ok := items["OOT_FISHING_POND_ADULT_FISH_1LBS"]; ok {
		t.Fatal("OOT_FISHING_POND_ADULT_FISH_1LBS should not be exported")
	}
}

func TestExtractItemsOnlyChangesRequestedMissingItems(t *testing.T) {
	base := &GameState{}
	base.Oot.QuestItems = 1 << QuestOotMedallionForest
	base.Mm.TownStrayFairy = true
	base.Mm.StrayFairies[0] = 7
	base.Oot.Equipment = 0x0031
	base.Shared.BombchuBagOot = 1
	base.Shared.OcarinaButtonMaskOot = sharedOcarinaButtonMaskDisabled
	base.Shared.OcarinaButtonMaskMm = sharedOcarinaButtonMaskDisabled

	updated := *base
	updated.Mm.SkullTokensSwamp = 9
	updated.Mm.SkullTokensOcean = 13
	updated.Oot.ExtraRecords[ExtraIdxMmOwlFlags] = allMmOwlFlags()
	updated.Shared.OcarinaButtonMaskMm = allSharedOcarinaButtonMasks()

	before := itemQtyMap(ExtractItems(base))
	after := itemQtyMap(ExtractItems(&updated))

	changed := make(map[string][2]int)
	for itemID, beforeQty := range before {
		afterQty := after[itemID]
		if beforeQty != afterQty {
			changed[itemID] = [2]int{beforeQty, afterQty}
		}
		delete(after, itemID)
	}
	for itemID, afterQty := range after {
		if afterQty != 0 {
			changed[itemID] = [2]int{0, afterQty}
		}
	}

	if len(changed) != 18 {
		t.Fatalf("changed item count = %d, want 18: %#v", len(changed), changed)
	}
	if diff := changed["MM_GS_TOKEN_SWAMP"]; diff != [2]int{0, 9} {
		t.Fatalf("MM_GS_TOKEN_SWAMP diff = %#v, want [0 9]", diff)
	}
	if diff := changed["MM_GS_TOKEN_OCEAN"]; diff != [2]int{0, 13} {
		t.Fatalf("MM_GS_TOKEN_OCEAN diff = %#v, want [0 13]", diff)
	}
	for _, owl := range mmOwlItems {
		if diff := changed[owl.itemID]; diff != [2]int{0, 1} {
			t.Fatalf("%s diff = %#v, want [0 1]", owl.itemID, diff)
		}
	}
	for _, button := range sharedOcarinaButtons {
		if _, ok := changed[button.ootItemID]; ok {
			t.Fatalf("%s should not change: %#v", button.ootItemID, changed[button.ootItemID])
		}
		if diff := changed[button.mmItemID]; diff != [2]int{0, 1} {
			t.Fatalf("%s diff = %#v, want [0 1]", button.mmItemID, diff)
		}
	}
}

func allMmOwlFlags() uint32 {
	var flags uint32
	for _, owl := range mmOwlItems {
		flags |= 1 << owl.bit
	}
	return flags
}

func allSharedOcarinaButtonMasks() uint16 {
	var mask uint16
	for _, button := range sharedOcarinaButtons {
		mask |= button.mask
	}
	return mask
}

func TestExtractItemsIncludesOotSilverRupeeAnywhereItem(t *testing.T) {
	state := &GameState{}
	state.Oot.Equipment = 0x0030
	setOotSilverRupeeCount(state, 13, 1)

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_RUPEE_SILVER_GTG_WATER"]; got != 1 {
		t.Fatalf("OOT_RUPEE_SILVER_GTG_WATER = %d, want 1", got)
	}
	if got := items["OOT_SHIELD"]; got != 3 {
		t.Fatalf("OOT_SHIELD = %d, want 3", got)
	}
	if _, ok := items["OOT_POUCH_SILVER_GTG_WATER"]; ok {
		t.Fatal("OOT_POUCH_SILVER_GTG_WATER should not be exported for silver-rupee counts")
	}
}

func TestExtractItemsResolvesMqSilverRupeeIDsFromRuntimeCounts(t *testing.T) {
	state := &GameState{}
	state.Oot.RuntimeSilverRupeeCounts = [OotSilverRupeeSetCount]uint8{0, 5, 5, 5, 0, 5, 10, 5, 10, 0, 0, 5, 6, 3, 5, 5, 5, 0}
	state.Oot.HasRuntimeSilverRupeeCounts = true
	setOotSilverRupeeCount(state, 2, 2)
	setOotSilverRupeeCount(state, 3, 1)
	setOotSilverRupeeCount(state, 14, 1)
	setOotSilverRupeeCount(state, 15, 3)

	items := itemQtyMap(ExtractItems(state))

	if got := items["OOT_RUPEE_SILVER_SPIRIT_LOBBY"]; got != 2 {
		t.Fatalf("OOT_RUPEE_SILVER_SPIRIT_LOBBY = %d, want 2", got)
	}
	if got := items["OOT_RUPEE_SILVER_SPIRIT_ADULT"]; got != 1 {
		t.Fatalf("OOT_RUPEE_SILVER_SPIRIT_ADULT = %d, want 1", got)
	}
	if got := items["OOT_RUPEE_SILVER_GANON_SHADOW"]; got != 1 {
		t.Fatalf("OOT_RUPEE_SILVER_GANON_SHADOW = %d, want 1", got)
	}
	if got := items["OOT_RUPEE_SILVER_GANON_WATER"]; got != 3 {
		t.Fatalf("OOT_RUPEE_SILVER_GANON_WATER = %d, want 3", got)
	}
	if _, ok := items["OOT_RUPEE_SILVER_GANON_SPIRIT"]; ok {
		t.Fatal("OOT_RUPEE_SILVER_GANON_SPIRIT should not be exported for MQ Ganon Castle")
	}
	if _, ok := items["OOT_RUPEE_SILVER_GANON_LIGHT"]; ok {
		t.Fatal("OOT_RUPEE_SILVER_GANON_LIGHT should not be exported for MQ Ganon Castle")
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

func TestExtractChecksIncludesLiveOotChestFlags(t *testing.T) {
	original := checkNameTable
	checkNameTable = map[string]string{
		"OOT_chest_40_2": "Mido's House Bottom Left",
	}
	t.Cleanup(func() {
		checkNameTable = original
	})

	state := &GameState{}
	state.Oot.LiveSceneID = 0x28
	state.Oot.LiveChestFlags = 1 << 2
	state.Oot.HasLiveSceneFlags = true

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Mido's House Bottom Left"]; !ok {
		t.Fatal("missing live OoT chest check")
	}
}

func TestExtractChecksIncludesLiveOotCollectibleFlags(t *testing.T) {
	original := checkNameTable
	checkNameTable = map[string]string{
		"OOT_collect_55_1": "Kokiri Forest GS House of Twins",
	}
	t.Cleanup(func() {
		checkNameTable = original
	})

	state := &GameState{}
	state.Oot.LiveSceneID = 0x37
	state.Oot.LiveTempCollectFlag = 1 << 1
	state.Oot.HasLiveSceneFlags = true

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Kokiri Forest GS House of Twins"]; !ok {
		t.Fatal("missing live OoT collectible check")
	}
}

func TestExtractChecksIncludesDodongoHeartMinibossLavaCollectibleFallback(t *testing.T) {
	state := &GameState{}
	state.Oot.LiveSceneID = 1
	state.Oot.LiveCollectFlags = 1 << 24
	state.Oot.HasLiveSceneFlags = true

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Dodongo Cavern Heart Miniboss Lava"]; !ok {
		t.Fatal("missing Dodongo Cavern Heart Miniboss Lava live collectible fallback")
	}

	state = &GameState{}
	state.Oot.SceneFlags[1].Collectibles = 1 << 24

	checks = checkNameSet(ExtractChecks(state))
	if _, ok := checks["Dodongo Cavern Heart Miniboss Lava"]; !ok {
		t.Fatal("missing Dodongo Cavern Heart Miniboss Lava stable collectible fallback")
	}
}

func TestExtractChecksResolvesDodongoCompassChestSceneConflict(t *testing.T) {
	state := &GameState{}
	state.Oot.HasRuntimeMqBits = true
	state.Oot.SceneFlags[1].Chests = 1 << 5

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Dodongo Cavern Compass Chest"]; !ok {
		t.Fatal("missing Dodongo Cavern Compass Chest scene conflict resolution")
	}

	state.Oot.RuntimeMqBits = 1 << OotMqDodongosCavern
	checks = checkNameSet(ExtractChecks(state))
	if _, ok := checks["MQ Dodongo Cavern Compass Chest"]; !ok {
		t.Fatal("missing MQ Dodongo Cavern Compass Chest scene conflict resolution")
	}
	if _, ok := checks["Dodongo Cavern Compass Chest"]; ok {
		t.Fatal("vanilla Dodongo Cavern Compass Chest should not be exported for MQ")
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

func TestExtractChecksResolvesOotXflagConflictsFromRuntimeMqBits(t *testing.T) {
	originalXflags := xflagCheckTables
	originalConflicts := ootBitmapConflictTable["xflagsOot"]
	xflagCheckTables = map[string]map[int]string{
		"OOT": {176: "Dodongo Cavern Pot Miniboss 4"},
		"MM":  {},
	}
	ootBitmapConflictTable["xflagsOot"] = map[int]bitmapConflictEntry{
		173: {Block: "xflagsOot", Bit: 173, DungeonMq: OotMqDodongosCavern, Vanilla: []string{"Dodongo Cavern Pot Miniboss 1"}, Mq: []string{"MQ Dodongo Cavern Pot Miniboss 2"}},
		174: {Block: "xflagsOot", Bit: 174, DungeonMq: OotMqDodongosCavern, Vanilla: []string{"Dodongo Cavern Pot Miniboss 2"}, Mq: []string{"MQ Dodongo Cavern Pot Miniboss 3"}},
		175: {Block: "xflagsOot", Bit: 175, DungeonMq: OotMqDodongosCavern, Vanilla: []string{"Dodongo Cavern Pot Miniboss 3"}, Mq: []string{"MQ Dodongo Cavern Pot Miniboss 4"}},
	}
	defer func() {
		xflagCheckTables = originalXflags
		ootBitmapConflictTable["xflagsOot"] = originalConflicts
	}()

	state := &GameState{}
	state.Oot.HasRuntimeMqBits = true
	for _, bit := range []int{173, 174, 175, 176} {
		state.Shared.SetBit("xflagsOot", bit)
	}

	checks := checkNameSet(ExtractChecks(state))
	for _, name := range []string{
		"Dodongo Cavern Pot Miniboss 1",
		"Dodongo Cavern Pot Miniboss 2",
		"Dodongo Cavern Pot Miniboss 3",
		"Dodongo Cavern Pot Miniboss 4",
	} {
		if _, ok := checks[name]; !ok {
			t.Fatalf("missing Dodongo Cavern miniboss pot check %q", name)
		}
	}

	state.Oot.RuntimeMqBits = 1 << OotMqDodongosCavern
	checks = checkNameSet(ExtractChecks(state))
	if _, ok := checks["MQ Dodongo Cavern Pot Miniboss 2"]; !ok {
		t.Fatal("missing MQ Dodongo xflag conflict resolution")
	}
}

func TestExtractChecksResolvesDodongoGsConflictsFromRuntimeMqBits(t *testing.T) {
	originalGs := gsCheckTables
	originalConflicts := ootBitmapConflictTable["gsOot"]
	gsCheckTables = map[string]map[int]string{
		"OOT": {9: "Dodongo Cavern GS Scarecrow"},
	}
	ootBitmapConflictTable["gsOot"] = map[int]bitmapConflictEntry{
		8:  {Block: "gsOot", Bit: 8, DungeonMq: OotMqDodongosCavern, Vanilla: []string{"Dodongo Cavern GS Stairs Vines"}, Mq: []string{"MQ Dodongo Cavern GS Near Boss"}},
		10: {Block: "gsOot", Bit: 10, DungeonMq: OotMqDodongosCavern, Vanilla: []string{"Dodongo Cavern GS Stairs Top"}, Mq: []string{"MQ Dodongo Cavern GS Upper Lizalfos"}},
		11: {Block: "gsOot", Bit: 11, DungeonMq: OotMqDodongosCavern, Vanilla: []string{"Dodongo Cavern GS Near Boss"}, Mq: []string{"MQ Dodongo Cavern GS Poe Room Side"}},
	}
	defer func() {
		gsCheckTables = originalGs
		ootBitmapConflictTable["gsOot"] = originalConflicts
	}()

	state := &GameState{}
	state.Oot.HasRuntimeMqBits = true
	for _, bit := range []int{8, 9, 10, 11} {
		setWordBitmapBit(state.Oot.GsFlags[:], bit)
	}

	checks := checkNameSet(ExtractChecks(state))
	for _, name := range []string{
		"Dodongo Cavern GS Stairs Vines",
		"Dodongo Cavern GS Stairs Top",
		"Dodongo Cavern GS Near Boss",
		"Dodongo Cavern GS Scarecrow",
	} {
		if _, ok := checks[name]; !ok {
			t.Fatalf("missing vanilla Dodongo GS check %q", name)
		}
	}

	state.Oot.RuntimeMqBits = 1 << OotMqDodongosCavern
	checks = checkNameSet(ExtractChecks(state))
	for _, name := range []string{
		"MQ Dodongo Cavern GS Near Boss",
		"MQ Dodongo Cavern GS Upper Lizalfos",
		"MQ Dodongo Cavern GS Poe Room Side",
	} {
		if _, ok := checks[name]; !ok {
			t.Fatalf("missing MQ Dodongo GS check %q", name)
		}
	}
	if _, ok := checks["Dodongo Cavern GS Stairs Vines"]; ok {
		t.Fatal("vanilla Dodongo Cavern GS Stairs Vines should not be exported for MQ")
	}
}

func TestDodongoGsConflictsGenerated(t *testing.T) {
	assertConflictContains := func(vanilla string, mq string) {
		t.Helper()
		for _, entry := range ootBitmapConflictTable["gsOot"] {
			if containsString(entry.Vanilla, vanilla) && containsString(entry.Mq, mq) {
				return
			}
		}
		t.Fatalf("missing generated Dodongo GS conflict %q <-> %q", vanilla, mq)
	}

	assertConflictContains("Dodongo Cavern GS Stairs Vines", "MQ Dodongo Cavern GS Near Boss")
	assertConflictContains("Dodongo Cavern GS Stairs Top", "MQ Dodongo Cavern GS Upper Lizalfos")
	assertConflictContains("Dodongo Cavern GS Near Boss", "MQ Dodongo Cavern GS Time Blocks")
}

func TestExtractChecksIncludesDodongoGsChecksFromGeneratedConflicts(t *testing.T) {
	state := &GameState{}
	state.Oot.HasRuntimeMqBits = true
	for _, bit := range []int{8, 9, 10, 11, 12} {
		setWordBitmapBit(state.Oot.GsFlags[:], bit)
	}

	checks := checkNameSet(ExtractChecks(state))
	for _, name := range []string{
		"Dodongo Cavern GS Stairs Vines",
		"Dodongo Cavern GS Scarecrow",
		"Dodongo Cavern GS Stairs Top",
		"Dodongo Cavern GS Near Boss",
		"Dodongo Cavern GS Side Room",
	} {
		if _, ok := checks[name]; !ok {
			t.Fatalf("missing vanilla Dodongo GS check %q from generated conflicts", name)
		}
	}

	state.Oot.RuntimeMqBits = 1 << OotMqDodongosCavern
	checks = checkNameSet(ExtractChecks(state))
	for _, name := range []string{
		"MQ Dodongo Cavern GS Near Boss",
		"MQ Dodongo Cavern GS Poe Room Side",
		"MQ Dodongo Cavern GS Upper Lizalfos",
		"MQ Dodongo Cavern GS Time Blocks",
		"MQ Dodongo Cavern GS Larve Room",
	} {
		if _, ok := checks[name]; !ok {
			t.Fatalf("missing MQ Dodongo GS check %q from generated conflicts", name)
		}
	}
}

func TestParseOotSaveReadsGsFlags(t *testing.T) {
	data := make([]byte, OotSaveSize)
	data[OotOffInvItems] = 0x00
	data[OotOffInvItems+15] = 0x0F
	binary.BigEndian.PutUint32(data[OotOffGsFlags+4:], 0x12345678)
	binary.BigEndian.PutUint16(data[OotOffGoldTokens:], 42)

	var oot OotState
	if err := parseOotSave(&oot, data); err != nil {
		t.Fatalf("parseOotSave: %v", err)
	}
	if oot.GoldTokens != 42 {
		t.Fatalf("unexpected OoT gold token count: %d", oot.GoldTokens)
	}
	if oot.GsFlags[1] != 0x12345678 {
		t.Fatalf("unexpected OoT gs flags word: %#x", oot.GsFlags[1])
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func setWordBitmapBit(bitmap []uint32, bit int) {
	bitmap[bit/32] |= 1 << uint(bit%32)
}

func TestExtractChecksIncludesShopBitmapChecks(t *testing.T) {
	originalShops := shopCheckTables
	shopCheckTables = map[string]map[int]string{
		"OOT": {3: "Kokiri Shop Item 4"},
		"MM":  {4: "Curiosity Shop All-Night Mask"},
	}
	t.Cleanup(func() {
		shopCheckTables = originalShops
	})

	state := &GameState{}
	state.Shared.SetBit("shopsOot", 3)
	state.Shared.SetBit("shopsMm", 4)

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Kokiri Shop Item 4"]; !ok {
		t.Fatal("missing OoT shop check from shared bitmap")
	}
	if _, ok := checks["Curiosity Shop All-Night Mask"]; !ok {
		t.Fatal("missing MM shop check from shared bitmap")
	}
}

func TestExtractChecksIncludesScrubAndSilverRupeeBitmapChecks(t *testing.T) {
	originalScrubs := scrubCheckTables
	originalSilver := silverRupeeCheckTables
	scrubCheckTables = map[string]map[int]string{
		"OOT": {0x1F: "Dodongo Cavern Lobby Scrub"},
		"MM":  {},
	}
	silverRupeeCheckTables = map[string]map[int]string{
		"OOT": {0x37: "Ice Cavern SR Scythe Left"},
		"MM":  {},
	}
	t.Cleanup(func() {
		scrubCheckTables = originalScrubs
		silverRupeeCheckTables = originalSilver
	})

	state := &GameState{}
	state.Shared.SetBit("scrubsOot", 0x1F)
	state.Shared.SetBit("srOot", 0x37)

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Dodongo Cavern Lobby Scrub"]; !ok {
		t.Fatal("missing OoT scrub check from shared bitmap")
	}
	if _, ok := checks["Ice Cavern SR Scythe Left"]; !ok {
		t.Fatal("missing OoT silver rupee check from shared bitmap")
	}
}

func TestExtractChecksIncludesMmExtraFlagChecks(t *testing.T) {
	originalSymbols := npcSymbolTables
	npcSymbolTables = map[string]map[string]string{
		"OOT": {},
		"MM": {
			"HONEY_DARLING_1":   "Honey & Darling Reward Any Day",
			"MASK_BLAST":        "Clock Town Blast Mask",
			"BOMBER_NOTEBOOK":   "Clock Town Bomber Notebook",
			"DEKU_PLAYGROUND_1": "Deku Playground Reward Any Day",
			"SONG_HEALING":      "Initial Song of Healing",
			"STRAY_FAIRY_TOWN":  "Clock Town Stray Fairy",
		},
	}
	t.Cleanup(func() {
		npcSymbolTables = originalSymbols
	})

	state := &GameState{}
	state.Oot.ExtraRecords[ExtraIdxMmFlags2] = (1 << mmExtraFlags2HoneyDarling) | (1 << mmExtraFlags2Notebook) | (1 << mmExtraFlags2MaskBlast) | (1 << mmExtraFlags2DekuPlayground) | (1 << mmExtraFlags2SongHealing) | (1 << mmExtraFlags2TownStrayFairy)

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Honey & Darling Reward Any Day"]; !ok {
		t.Fatal("missing MM extra-flag check for Honey & Darling Reward Any Day")
	}
	if _, ok := checks["Clock Town Blast Mask"]; !ok {
		t.Fatal("missing MM extra-flag check for Blast Mask")
	}
	if _, ok := checks["Clock Town Bomber Notebook"]; !ok {
		t.Fatal("missing MM extra-flag check for Bomber Notebook")
	}
	if _, ok := checks["Deku Playground Reward Any Day"]; !ok {
		t.Fatal("missing MM extra-flag check for Deku Playground Reward Any Day")
	}
	if _, ok := checks["Initial Song of Healing"]; !ok {
		t.Fatal("missing MM extra-flag check for Initial Song of Healing")
	}
	if _, ok := checks["Clock Town Stray Fairy"]; !ok {
		t.Fatal("missing MM extra-flag check for Clock Town Stray Fairy")
	}
}

func TestExtractItemsIncludesMmTownStrayFairyFromExtraFlags2(t *testing.T) {
	state := &GameState{}
	state.Mm.ExtraFlags2 = 1 << mmExtraFlags2TownStrayFairy

	items := itemQtyMap(ExtractItems(state))
	if got := items["MM_STRAY_FAIRY_TOWN"]; got != 1 {
		t.Fatalf("MM_STRAY_FAIRY_TOWN = %d, want 1 from MM extra flags", got)
	}
}

func TestExtractChecksIncludesOotSongEventFallbacks(t *testing.T) {
	originalSymbols := npcSymbolTables
	npcSymbolTables = map[string]map[string]string{
		"OOT": {
			"SARIA_SONG":      "Saria's Song",
			"ROYAL_TOMB_SONG": "Graveyard Royal Tomb Song",
		},
		"MM": {},
	}
	t.Cleanup(func() {
		npcSymbolTables = originalSymbols
	})

	state := &GameState{}
	state.Oot.EventsChk[ootEventSongSariaVanilla>>4] = 1 << (ootEventSongSariaVanilla & 0xF)
	state.Oot.EventsChk[ootEventSongSunCustom>>4] = 1 << (ootEventSongSunCustom & 0xF)

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Saria's Song"]; !ok {
		t.Fatal("missing OoT Saria song check from event fallback")
	}
	if _, ok := checks["Graveyard Royal Tomb Song"]; !ok {
		t.Fatal("missing OoT Royal Tomb Song check from event fallback")
	}
}

func TestExtractChecksIncludesOotMalonEventFallbacks(t *testing.T) {
	originalSymbols := npcSymbolTables
	npcSymbolTables = map[string]map[string]string{
		"OOT": {
			"MALON_EGG":  "Malon Egg",
			"MALON_SONG": "Lon Lon Ranch Malon Song",
		},
		"MM": {},
	}
	t.Cleanup(func() {
		npcSymbolTables = originalSymbols
	})

	state := &GameState{}
	state.Oot.EventsChk[ootEventMalonEgg>>4] = 1 << (ootEventMalonEgg & 0xF)
	state.Oot.EventsChk[ootEventSongEpona>>4] = 1 << (ootEventSongEpona & 0xF)

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Malon Egg"]; !ok {
		t.Fatal("missing OoT Malon Egg check from event fallback")
	}
	if _, ok := checks["Lon Lon Ranch Malon Song"]; !ok {
		t.Fatal("missing OoT Malon Song check from event fallback")
	}
}

func TestExtractChecksIncludesOotEventItemSymbolFallbacks(t *testing.T) {
	originalSymbols := npcSymbolTables
	npcSymbolTables = map[string]map[string]string{
		"OOT": {
			"DARUNIA_BRACELET":    "Darunia",
			"GERUDO_ARCHERY_2":    "Gerudo Fortress Archery Reward 2",
			"LOST_WOODS_MEMORY":   "Lost Woods Memory Game",
			"LOST_WOODS_TARGET":   "Lost Woods Target",
			"POCKET_EGG":          "Hatch Pocket Cucco",
			"TALON_BOTTLE":        "Lon Lon Ranch Talon Bottle",
			"SHOOTING_GAME_ADULT": "Shooting Gallery Adult",
			"SHOOTING_GAME_CHILD": "Shooting Gallery Child",
			"MASK_SELL_BUNNY":     "Hyrule Field Sell Bunny Mask",
			"MASK_SELL_KEATON":    "Kakariko Sell Keaton Mask",
			"MASK_SELL_SKULL":     "Lost Woods Sell Skull Mask",
			"MASK_SELL_SPOOKY":    "Graveyard Sell Spooky Mask",
		},
		"MM": {},
	}
	t.Cleanup(func() {
		npcSymbolTables = originalSymbols
	})

	state := &GameState{}
	state.Oot.Entrance = ootEntranceChildArchery
	state.Oot.SceneID = ootSceneShootingGallerySave
	state.Oot.EventsItem[ootEventItemLostWoodsMemoryOrShootingChild>>4] =
		(1 << (ootEventItemTalonBottle & 0xF)) |
			(1 << (ootEventItemLostWoodsMemoryOrShootingChild & 0xF)) |
			(1 << (ootEventItemShootingGalleryAdult & 0xF)) |
			(1 << (ootEventItemGerudoArchery2 & 0xF))
	state.Oot.EventsItem[ootEventItemLostWoodsTarget>>4] = 1 << (ootEventItemLostWoodsTarget & 0xF)
	state.Oot.EventsItem[ootEventItemGoronBracelet>>4] = 1 << (ootEventItemGoronBracelet & 0xF)
	state.Oot.EventsItem[ootEventItemPocketEgg>>4] |= 1 << (ootEventItemPocketEgg & 0xF)
	state.Oot.EventsItem[ootEventItemMaskSellKeaton>>4] =
		(1 << (ootEventItemMaskSellKeaton & 0xF)) |
			(1 << (ootEventItemMaskSellSkull & 0xF)) |
			(1 << (ootEventItemMaskSellSpooky & 0xF)) |
			(1 << (ootEventItemMaskSellBunny & 0xF))

	checks := checkNameSet(ExtractChecks(state))
	for _, name := range []string{
		"Darunia",
		"Gerudo Fortress Archery Reward 2",
		"Hatch Pocket Cucco",
		"Lon Lon Ranch Talon Bottle",
		"Lost Woods Target",
		"Shooting Gallery Adult",
		"Shooting Gallery Child",
		"Hyrule Field Sell Bunny Mask",
		"Kakariko Sell Keaton Mask",
		"Lost Woods Sell Skull Mask",
		"Graveyard Sell Spooky Mask",
	} {
		if _, ok := checks[name]; !ok {
			t.Fatalf("missing OoT event-item symbol check: %s", name)
		}
	}
}

func TestExtractChecksIncludesOotMemoryGameFromRoundProgress(t *testing.T) {
	originalSymbols := npcSymbolTables
	npcSymbolTables = map[string]map[string]string{
		"OOT": {
			"LOST_WOODS_MEMORY":   "Lost Woods Memory Game",
			"SHOOTING_GAME_CHILD": "Shooting Gallery Child",
		},
		"MM": {},
	}
	t.Cleanup(func() {
		npcSymbolTables = originalSymbols
	})

	state := &GameState{}
	state.Oot.OcarinaGameRound = 1

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Lost Woods Memory Game"]; !ok {
		t.Fatal("missing Lost Woods Memory Game check from OoT memory-game progress fallback")
	}
	if _, ok := checks["Shooting Gallery Child"]; ok {
		t.Fatal("unexpected Shooting Gallery Child check from OoT memory-game progress fallback")
	}
}

func TestExtractChecksIncludesOotMemoryAndChildShootingFromSeparateSignals(t *testing.T) {
	originalSymbols := npcSymbolTables
	npcSymbolTables = map[string]map[string]string{
		"OOT": {
			"LOST_WOODS_MEMORY":   "Lost Woods Memory Game",
			"SHOOTING_GAME_CHILD": "Shooting Gallery Child",
		},
		"MM": {},
	}
	t.Cleanup(func() {
		npcSymbolTables = originalSymbols
	})

	state := &GameState{}
	state.Oot.OcarinaGameRound = 1
	state.Oot.EventsItem[ootEventItemLostWoodsMemoryOrShootingChild>>4] = 1 << (ootEventItemLostWoodsMemoryOrShootingChild & 0xF)

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Lost Woods Memory Game"]; !ok {
		t.Fatal("missing Lost Woods Memory Game check from separate OoT fallback signals")
	}
	if _, ok := checks["Shooting Gallery Child"]; !ok {
		t.Fatal("missing Shooting Gallery Child check from separate OoT fallback signals")
	}
}

func TestExtractChecksIncludesOotPocketEggTradeFallback(t *testing.T) {
	originalSymbols := npcSymbolTables
	npcSymbolTables = map[string]map[string]string{
		"OOT": {
			"POCKET_EGG": "Hatch Pocket Cucco",
		},
		"MM": {},
	}
	t.Cleanup(func() {
		npcSymbolTables = originalSymbols
	})

	state := &GameState{}
	state.Oot.ExtraRecords[ExtraIdxOotTradeSave] = 1 << ootAdultTradePocketEggBit
	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Hatch Pocket Cucco"]; ok {
		t.Fatal("unexpected Pocket Cucco check from Pocket Egg alone")
	}

	state.Oot.ExtraRecords[ExtraIdxOotTradeSave] |= 1 << (ootAdultTradePocketEggBit + 3)
	checks = checkNameSet(ExtractChecks(state))
	if _, ok := checks["Hatch Pocket Cucco"]; !ok {
		t.Fatal("missing Pocket Cucco check from persistent adult trade progression")
	}
}

func TestExtractChecksIncludesOotChildTradeFallbacks(t *testing.T) {
	originalSymbols := npcSymbolTables
	npcSymbolTables = map[string]map[string]string{
		"OOT": {
			"WEIRD_EGG":    "Hatch Chicken",
			"ZELDA_LETTER": "Zelda's Letter",
		},
		"MM": {},
	}
	t.Cleanup(func() {
		npcSymbolTables = originalSymbols
	})

	state := &GameState{}
	state.Oot.ExtraRecords[ExtraIdxOotTradeSave] = uint32(1 << (16 + ootChildTradeWeirdEggBit))
	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Hatch Chicken"]; ok {
		t.Fatal("unexpected Hatch Chicken check from Weird Egg alone")
	}
	if _, ok := checks["Zelda's Letter"]; ok {
		t.Fatal("unexpected Zelda's Letter check from Weird Egg alone")
	}

	state.Oot.ExtraRecords[ExtraIdxOotTradeSave] |= uint32(1 << (16 + 1))
	checks = checkNameSet(ExtractChecks(state))
	if _, ok := checks["Hatch Chicken"]; !ok {
		t.Fatal("missing Hatch Chicken check from persistent child trade progression")
	}
	if _, ok := checks["Zelda's Letter"]; ok {
		t.Fatal("unexpected Zelda's Letter check before Zelda's Letter progression")
	}

	state.Oot.ExtraRecords[ExtraIdxOotTradeSave] |= uint32(1 << (16 + 2))
	checks = checkNameSet(ExtractChecks(state))
	if _, ok := checks["Zelda's Letter"]; !ok {
		t.Fatal("missing Zelda's Letter check from persistent child trade progression")
	}
}

func TestExtractChecksIncludesOotFrogSongEventFallbacks(t *testing.T) {
	originalSymbols := npcSymbolTables
	npcSymbolTables = map[string]map[string]string{
		"OOT": {
			"FROGS_GAME":   "Zora River Frogs Game",
			"FROGS_ZL":     "Zora River Frogs Zeldas Lullaby",
			"FROGS_EPONA":  "Zora River Frogs Eponas Song",
			"FROGS_SARIA":  "Zora River Frogs Sarias Song",
			"FROGS_SUNS":   "Zora River Frogs Suns Song",
			"FROGS_SOT":    "Zora River Frogs Song of Time",
			"FROGS_STORMS": "Zora River Frogs Storms",
		},
		"MM": {},
	}
	t.Cleanup(func() {
		npcSymbolTables = originalSymbols
	})

	state := &GameState{}
	state.Oot.EventsChk[ootEventFrogsGame>>4] =
		(1 << (ootEventFrogsGame & 0xF)) |
			(1 << (ootEventFrogsZelda & 0xF)) |
			(1 << (ootEventFrogsEpona & 0xF)) |
			(1 << (ootEventFrogsSun & 0xF)) |
			(1 << (ootEventFrogsSaria & 0xF)) |
			(1 << (ootEventFrogsSongOfTime & 0xF)) |
			(1 << (ootEventFrogsStorms & 0xF))

	checks := checkNameSet(ExtractChecks(state))
	for _, name := range []string{
		"Zora River Frogs Game",
		"Zora River Frogs Zeldas Lullaby",
		"Zora River Frogs Eponas Song",
		"Zora River Frogs Sarias Song",
		"Zora River Frogs Suns Song",
		"Zora River Frogs Song of Time",
		"Zora River Frogs Storms",
	} {
		if _, ok := checks[name]; !ok {
			t.Fatalf("missing OoT frog song check from event fallback: %s", name)
		}
	}
}

func TestExtractChecksIncludesTempleOfTimeEventFallbacks(t *testing.T) {
	originalSymbols := npcSymbolTables
	npcSymbolTables = map[string]map[string]string{
		"OOT": {
			"LIGHT_MEDALLION": "Temple of Time Medallion",
			"MASTER_SWORD":    "Temple of Time Master Sword",
		},
		"MM": {},
	}
	t.Cleanup(func() {
		npcSymbolTables = originalSymbols
	})

	state := &GameState{}
	state.Oot.EventsChk[0x45>>4] = (1 << (0x45 & 0xF)) | (1 << (0x4f & 0xF))

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Temple of Time Medallion"]; !ok {
		t.Fatal("missing Temple of Time medallion check from event fallback")
	}
	if _, ok := checks["Temple of Time Master Sword"]; !ok {
		t.Fatal("missing Temple of Time master sword check from event fallback")
	}
}

func TestExtractChecksIncludesSkulltulaHouseEventFallbacks(t *testing.T) {
	originalSymbols := npcSymbolTables
	npcSymbolTables = map[string]map[string]string{
		"OOT": {
			"GS_10": "Skulltula House 10 Tokens",
			"GS_20": "Skulltula House 20 Tokens",
			"GS_30": "Skulltula House 30 Tokens",
			"GS_40": "Skulltula House 40 Tokens",
			"GS_50": "Skulltula House 50 Tokens",
		},
		"MM": {},
	}
	t.Cleanup(func() {
		npcSymbolTables = originalSymbols
	})

	state := &GameState{}
	state.Oot.EventsChk[ootEventSkulltulaHouse10>>4] =
		(1 << (ootEventSkulltulaHouse10 & 0xF)) |
			(1 << (ootEventSkulltulaHouse20 & 0xF)) |
			(1 << (ootEventSkulltulaHouse30 & 0xF)) |
			(1 << (ootEventSkulltulaHouse40 & 0xF)) |
			(1 << (ootEventSkulltulaHouse50 & 0xF))

	checks := checkNameSet(ExtractChecks(state))
	for _, name := range []string{
		"Skulltula House 10 Tokens",
		"Skulltula House 20 Tokens",
		"Skulltula House 30 Tokens",
		"Skulltula House 40 Tokens",
		"Skulltula House 50 Tokens",
	} {
		if _, ok := checks[name]; !ok {
			t.Fatalf("missing OoT Skulltula House check from event fallback: %s", name)
		}
	}
}

func TestExtractChecksIncludesOtherOotEventSymbolFallbacks(t *testing.T) {
	originalSymbols := npcSymbolTables
	npcSymbolTables = map[string]map[string]string{
		"OOT": {
			"ZELDA_LIGHT_ARROW": "Temple of Time Light Arrows",
		},
		"MM": {},
	}
	t.Cleanup(func() {
		npcSymbolTables = originalSymbols
	})

	state := &GameState{}
	state.Oot.EventsChk[0xc4>>4] = 1 << (0xc4 & 0xF)

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Temple of Time Light Arrows"]; !ok {
		t.Fatal("missing OoT event-symbol check for Light Arrows")
	}
}

func TestExtractChecksIncludesOotEventMiscSymbolFallbacks(t *testing.T) {
	originalSymbols := npcSymbolTables
	npcSymbolTables = map[string]map[string]string{
		"OOT": {
			"GERUDO_ARCHERY_1": "Gerudo Fortress Archery Reward 1",
			"DOG_LADY":         "Market Dog Lady HP",
			"MEDIGORON":        "Goron City Medigoron Giant Knife",
		},
		"MM": {},
	}
	t.Cleanup(func() {
		npcSymbolTables = originalSymbols
	})

	state := &GameState{}
	state.Oot.EventsMisc[ootEventMiscGerudoArchery1>>4] |= 1 << (ootEventMiscGerudoArchery1 & 0xF)
	state.Oot.EventsMisc[ootEventMiscRichardHeartPiece>>4] |= 1 << (ootEventMiscRichardHeartPiece & 0xF)
	state.Oot.EventsMisc[ootEventMiscMedigoron>>4] = 1 << (ootEventMiscMedigoron & 0xF)

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Gerudo Fortress Archery Reward 1"]; !ok {
		t.Fatal("missing OoT event-misc symbol check for Gerudo Archery Reward 1")
	}
	if _, ok := checks["Market Dog Lady HP"]; !ok {
		t.Fatal("missing OoT event-misc symbol check for Market Dog Lady HP")
	}
	if _, ok := checks["Goron City Medigoron Giant Knife"]; !ok {
		t.Fatal("missing OoT event-misc symbol check for Medigoron")
	}
}

func checkNameSet(checks []TrackedCheck) map[string]struct{} {
	result := make(map[string]struct{}, len(checks))
	for _, check := range checks {
		result[check.Name] = struct{}{}
	}
	return result
}

func TestExtractChecksIncludesMmCycleFlagCollectible(t *testing.T) {
	state := &GameState{}
	// Scene 111 (0x6F), collectible bit 10 = Clock Town Platform HP.
	// In MM many collectibles only live in cycle flags.
	state.Mm.CycleFlags[111].Collectibles = 1 << 10

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Clock Town Platform HP"]; !ok {
		t.Fatal("missing Clock Town Platform HP from MM cycle flags")
	}
}

func TestExtractChecksIncludesMmCycleFlagChest(t *testing.T) {
	state := &GameState{}
	// Cycle flag chests should also be detected.
	state.Mm.CycleFlags[108].Chests = 1 << 10

	checks := ExtractChecks(state)
	found := false
	for _, c := range checks {
		if c.Key == "MM_chest_108_10" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("missing MM_chest_108_10 from MM cycle flags")
	}
}

func TestExtractChecksDoesNotDuplicateMmPermAndCycleFlags(t *testing.T) {
	state := &GameState{}
	// Set same bit in both perm and cycle flags.
	state.Mm.SceneFlags[111].Collectibles = 1 << 10
	state.Mm.CycleFlags[111].Collectibles = 1 << 10

	checks := ExtractChecks(state)
	count := 0
	for _, c := range checks {
		if c.Name == "Clock Town Platform HP" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("Clock Town Platform HP appeared %d times, want 1", count)
	}
}

func TestExtractChecksIncludesLiveMmChestFlags(t *testing.T) {
	state := &GameState{}
	// Scene 108 (0x6C), chest bit 10 — live PlayState chest flag.
	state.Mm.LiveSceneID = 108
	state.Mm.LiveChestFlags = 1 << 10
	state.Mm.HasLiveSceneFlags = true

	checks := ExtractChecks(state)
	found := false
	for _, c := range checks {
		if c.Key == "MM_chest_108_10" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("missing MM_chest_108_10 from MM live chest flags")
	}
}

func TestExtractChecksIncludesLiveMmCollectibleFlags(t *testing.T) {
	state := &GameState{}
	// Scene 111 (0x6F), collectible bit 10 — live PlayState collectible flag.
	state.Mm.LiveSceneID = 111
	state.Mm.LiveCollectFlags = 1 << 10
	state.Mm.HasLiveSceneFlags = true

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Clock Town Platform HP"]; !ok {
		t.Fatal("missing Clock Town Platform HP from MM live collectible flags")
	}
}

func TestExtractChecksIncludesMmOwlActivationFlags(t *testing.T) {
	state := &GameState{}
	state.Mm.OwlActivationFlags = 1 << mmOwlClockTownBit

	checks := checkNameSet(ExtractChecks(state))
	if _, ok := checks["Clock Town Owl Statue"]; !ok {
		t.Fatal("missing Clock Town Owl Statue from MM owl activation flags")
	}
}

func TestExtractChecksDoesNotDuplicateMmLiveAndCycleFlags(t *testing.T) {
	state := &GameState{}
	// Same bit in cycle flags AND live flags — should appear only once.
	state.Mm.CycleFlags[108].Chests = 1 << 10
	state.Mm.LiveSceneID = 108
	state.Mm.LiveChestFlags = 1 << 10
	state.Mm.HasLiveSceneFlags = true

	checks := ExtractChecks(state)
	count := 0
	for _, c := range checks {
		if c.Key == "MM_chest_108_10" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("MM_chest_108_10 appeared %d times, want 1", count)
	}
}
