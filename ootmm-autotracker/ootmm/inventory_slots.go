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
)

//go:embed inventory_slots.json
var embeddedInventorySlots []byte

func init() {
	var slotFile inventorySlotFile
	if err := json.Unmarshal(embeddedInventorySlots, &slotFile); err != nil {
		panic(fmt.Sprintf("parse embedded inventory slot mapping: %v", err))
	}

	ootInventorySlots = buildInventorySlotTable("OOT", slotFile.Oot)
	mmInventorySlots = buildInventorySlotTable("MM", slotFile.Mm)
}

func buildInventorySlotTable(game string, entries []inventorySlotEntry) []string {
	maxIndex := -1
	for _, entry := range entries {
		if entry.Index > maxIndex {
			maxIndex = entry.Index
		}
	}
	if maxIndex < 0 {
		return nil
	}

	table := make([]string, maxIndex+1)
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
		table[entry.Index] = entry.ItemID
	}

	for index, trackerID := range table {
		if trackerID == "" {
			panic(fmt.Sprintf("missing %s tracker ID for slot %d", game, index))
		}
	}

	return table
}