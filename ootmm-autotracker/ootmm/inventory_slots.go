package ootmm

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type inventorySlotFile struct {
	Oot     []inventorySlotEntry `json:"oot"`
	Mm      []inventorySlotEntry `json:"mm"`
	Catalog itemCatalog          `json:"catalog"`
}

type inventorySlotEntry struct {
	Index  int    `json:"index"`
	Slot   string `json:"slot"`
	ItemID string `json:"itemId"`
}

type itemCatalog struct {
	Souls   soulCatalog    `json:"souls"`
	Special specialCatalog `json:"special"`
}

type soulCatalog struct {
	Oot soulGroups `json:"oot"`
	Mm  soulGroups `json:"mm"`
}

type soulGroups struct {
	Enemy  []string `json:"enemy"`
	Boss   []string `json:"boss"`
	Npc    []string `json:"npc"`
	Animal []string `json:"animal"`
	Misc   []string `json:"misc"`
}

type specialCatalog struct {
	MmItems  []string `json:"mmItems"`
	MmTrade1 []string `json:"mmTrade1"`
	MmTrade2 []string `json:"mmTrade2"`
	MmTrade3 []string `json:"mmTrade3"`
	MmFlags3 []string `json:"mmFlags3"`
}

var (
	ootInventorySlots       []string
	mmInventorySlots        []string
	ootInventorySlotIndices map[string]int
	mmInventorySlotIndices  map[string]int
	ootSoulEnemyIDs         []string
	ootSoulBossIDs          []string
	ootSoulNpcIDs           []string
	ootSoulAnimalIDs        []string
	ootSoulMiscIDs          []string
	mmSoulEnemyIDs          []string
	mmSoulBossIDs           []string
	mmSoulNpcIDs            []string
	mmSoulAnimalIDs         []string
	mmSoulMiscIDs           []string
	mmItemIDs               []string
	mmTrade1ItemIDs         []string
	mmTrade2ItemIDs         []string
	mmTrade3ItemIDs         []string
	mmFlags3ItemIDs         []string
)

//go:embed inventory_slots.json
var embeddedInventorySlots []byte

func init() {
	var slotFile inventorySlotFile
	if err := json.Unmarshal(embeddedInventorySlots, &slotFile); err != nil {
		panic(fmt.Sprintf("parse embedded inventory slot mapping: %v", err))
	}

	ootInventorySlots, ootInventorySlotIndices = buildInventorySlotTable("OOT", slotFile.Oot)
	mmInventorySlots, mmInventorySlotIndices = buildInventorySlotTable("MM", slotFile.Mm)
	ootSoulEnemyIDs = slotFile.Catalog.Souls.Oot.Enemy
	ootSoulBossIDs = slotFile.Catalog.Souls.Oot.Boss
	ootSoulNpcIDs = slotFile.Catalog.Souls.Oot.Npc
	ootSoulAnimalIDs = slotFile.Catalog.Souls.Oot.Animal
	ootSoulMiscIDs = slotFile.Catalog.Souls.Oot.Misc
	mmSoulEnemyIDs = slotFile.Catalog.Souls.Mm.Enemy
	mmSoulBossIDs = slotFile.Catalog.Souls.Mm.Boss
	mmSoulNpcIDs = slotFile.Catalog.Souls.Mm.Npc
	mmSoulAnimalIDs = slotFile.Catalog.Souls.Mm.Animal
	mmSoulMiscIDs = slotFile.Catalog.Souls.Mm.Misc
	mmItemIDs = slotFile.Catalog.Special.MmItems
	mmTrade1ItemIDs = slotFile.Catalog.Special.MmTrade1
	mmTrade2ItemIDs = slotFile.Catalog.Special.MmTrade2
	mmTrade3ItemIDs = slotFile.Catalog.Special.MmTrade3
	mmFlags3ItemIDs = slotFile.Catalog.Special.MmFlags3
}

func buildInventorySlotTable(game string, entries []inventorySlotEntry) ([]string, map[string]int) {
	maxIndex := -1
	for _, entry := range entries {
		if entry.Index > maxIndex {
			maxIndex = entry.Index
		}
	}
	if maxIndex < 0 {
		return nil, nil
	}

	table := make([]string, maxIndex+1)
	indices := make(map[string]int, len(entries))
	for _, entry := range entries {
		if entry.Index < 0 || entry.Index >= len(table) {
			panic(fmt.Sprintf("invalid %s slot index %d for %s", game, entry.Index, entry.Slot))
		}
		if table[entry.Index] != "" {
			panic(fmt.Sprintf("duplicate %s slot index %d", game, entry.Index))
		}
		if entry.ItemID == "" {
			panic(fmt.Sprintf("empty tracker ID for %s", entry.Slot))
		}
		if _, exists := indices[entry.ItemID]; exists {
			panic(fmt.Sprintf("duplicate %s tracker ID %s", game, entry.ItemID))
		}
		table[entry.Index] = entry.ItemID
		indices[entry.ItemID] = entry.Index
	}

	for index, trackerID := range table {
		if trackerID == "" {
			panic(fmt.Sprintf("missing %s tracker ID for slot %d", game, index))
		}
	}

	return table, indices
}

func mustOotInventorySlotIndex(itemID string) int {
	return mustInventorySlotIndex("OOT", ootInventorySlotIndices, itemID)
}

func mustMmInventorySlotIndex(itemID string) int {
	return mustInventorySlotIndex("MM", mmInventorySlotIndices, itemID)
}

func mustInventorySlotIndex(game string, indices map[string]int, itemID string) int {
	index, ok := indices[itemID]
	if !ok {
		panic(fmt.Sprintf("missing %s inventory slot for %s", game, itemID))
	}
	return index
}
