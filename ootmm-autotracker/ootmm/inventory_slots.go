package ootmm

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type inventorySlotFile struct {
	Oot []inventorySlotEntry `json:"oot"`
	Mm  []inventorySlotEntry `json:"mm"`
}

type inventorySlotEntry struct {
	Index  int    `json:"index"`
	Slot   string `json:"slot"`
	ItemID string `json:"itemId"`
}

var (
	ootInventorySlots []string
	mmInventorySlots  []string
	ootInventorySlotIndices map[string]int
	mmInventorySlotIndices  map[string]int
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