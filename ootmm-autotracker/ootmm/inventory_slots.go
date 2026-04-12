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
	Index    int              `json:"index"`
	Slot     string           `json:"slot"`
	ItemID   string           `json:"itemId"`
	Quantity slotQuantityRule `json:"quantity,omitempty"`
}

type slotQuantityRule struct {
	Stages        []uint8 `json:"stages,omitempty"`
	MaxWithBottle bool    `json:"maxWithBottle,omitempty"`
	UseBeansCount bool    `json:"useBeansCount,omitempty"`
}

type itemCatalog struct {
	Shared sharedStorageLayout `json:"shared"`
	Items  []catalogItemEntry  `json:"items"`
}

type sharedStorageLayout struct {
	BaseOffset  uint32             `json:"baseOffset"`
	Stride      uint32             `json:"stride"`
	TrackedSize int                `json:"trackedSize"`
	Bitmaps     []sharedBitmapInfo `json:"bitmaps"`
}

type sharedBitmapInfo struct {
	Name   string `json:"name"`
	Offset int    `json:"offset"`
	Size   int    `json:"size"`
}

type catalogItemEntry struct {
	ItemID string            `json:"itemId"`
	Source catalogItemSource `json:"source"`
}

type catalogItemSource struct {
	Kind   string `json:"kind"`
	Block  string `json:"block,omitempty"`
	Record int    `json:"record,omitempty"`
	Bit    int    `json:"bit,omitempty"`
}

var (
	ootInventoryEntries     []inventorySlotEntry
	mmInventoryEntries      []inventorySlotEntry
	ootInventorySlotIndices map[string]int
	mmInventorySlotIndices  map[string]int
	sharedStorage           sharedStorageLayout
	sharedBitmaps           map[string]sharedBitmapInfo
	sharedBitmapUsedBits    map[string]int
	trackedCatalogItems     []catalogItemEntry
	catalogItemSources      map[string]catalogItemSource
)

//go:embed inventory_slots.json
var embeddedInventorySlots []byte

func init() {
	var slotFile inventorySlotFile
	if err := json.Unmarshal(embeddedInventorySlots, &slotFile); err != nil {
		panic(fmt.Sprintf("parse embedded inventory slot mapping: %v", err))
	}

	ootInventoryEntries, ootInventorySlotIndices = buildInventorySlotTable("OOT", slotFile.Oot)
	mmInventoryEntries, mmInventorySlotIndices = buildInventorySlotTable("MM", slotFile.Mm)
	sharedStorage = slotFile.Catalog.Shared
	sharedBitmaps = buildSharedBitmapTable(slotFile.Catalog.Shared)
	trackedCatalogItems, catalogItemSources, sharedBitmapUsedBits = buildCatalogTables(slotFile.Catalog.Items, sharedBitmaps)
}

func buildInventorySlotTable(game string, entries []inventorySlotEntry) ([]inventorySlotEntry, map[string]int) {
	maxIndex := -1
	for _, entry := range entries {
		if entry.Index > maxIndex {
			maxIndex = entry.Index
		}
	}
	if maxIndex < 0 {
		return nil, nil
	}

	table := make([]inventorySlotEntry, maxIndex+1)
	indices := make(map[string]int, len(entries))
	for _, entry := range entries {
		if entry.Index < 0 || entry.Index >= len(table) {
			panic(fmt.Sprintf("invalid %s slot index %d for %s", game, entry.Index, entry.Slot))
		}
		if table[entry.Index].ItemID != "" {
			panic(fmt.Sprintf("duplicate %s slot index %d", game, entry.Index))
		}
		if entry.ItemID == "" {
			panic(fmt.Sprintf("empty tracker ID for %s", entry.Slot))
		}
		if _, exists := indices[entry.ItemID]; exists {
			panic(fmt.Sprintf("duplicate %s tracker ID %s", game, entry.ItemID))
		}
		table[entry.Index] = entry
		indices[entry.ItemID] = entry.Index
	}

	for index, entry := range table {
		if entry.ItemID == "" {
			panic(fmt.Sprintf("missing %s tracker ID for slot %d", game, index))
		}
	}

	return table, indices
}

func buildSharedBitmapTable(layout sharedStorageLayout) map[string]sharedBitmapInfo {
	bitmaps := make(map[string]sharedBitmapInfo, len(layout.Bitmaps))
	for _, bitmap := range layout.Bitmaps {
		if bitmap.Name == "" {
			panic("shared bitmap is missing a name")
		}
		if bitmap.Size <= 0 {
			panic(fmt.Sprintf("shared bitmap %s has invalid size %d", bitmap.Name, bitmap.Size))
		}
		if bitmap.Offset < 0 || bitmap.Offset+bitmap.Size > layout.TrackedSize {
			panic(fmt.Sprintf("shared bitmap %s is out of bounds", bitmap.Name))
		}
		if _, exists := bitmaps[bitmap.Name]; exists {
			panic(fmt.Sprintf("duplicate shared bitmap %s", bitmap.Name))
		}
		bitmaps[bitmap.Name] = bitmap
	}
	return bitmaps
}

func buildCatalogTables(items []catalogItemEntry, bitmaps map[string]sharedBitmapInfo) ([]catalogItemEntry, map[string]catalogItemSource, map[string]int) {
	tracked := make([]catalogItemEntry, 0, len(items))
	sources := make(map[string]catalogItemSource, len(items))
	usedBits := make(map[string]int)
	for _, item := range items {
		if item.ItemID == "" {
			panic("catalog item is missing itemId")
		}
		if _, exists := sources[item.ItemID]; exists {
			panic(fmt.Sprintf("duplicate catalog item %s", item.ItemID))
		}
		source := item.Source
		switch source.Kind {
		case "shared-bitmap-bit":
			bitmap, ok := bitmaps[source.Block]
			if !ok {
				panic(fmt.Sprintf("catalog item %s references unknown shared bitmap %s", item.ItemID, source.Block))
			}
			if source.Bit < 0 || source.Bit >= bitmap.Size*8 {
				panic(fmt.Sprintf("catalog item %s references out-of-range bit %d for %s", item.ItemID, source.Bit, source.Block))
			}
			if source.Bit+1 > usedBits[source.Block] {
				usedBits[source.Block] = source.Bit + 1
			}
		case "oot-extra-bit":
			if source.Record < 0 {
				panic(fmt.Sprintf("catalog item %s has invalid extra record %d", item.ItemID, source.Record))
			}
			if source.Bit < 0 || source.Bit >= 32 {
				panic(fmt.Sprintf("catalog item %s has invalid extra record bit %d", item.ItemID, source.Bit))
			}
		case "oot-derived-key-ring", "mm-derived-key-ring":
			if source.Record < 0 {
				panic(fmt.Sprintf("catalog item %s has invalid dungeon record %d", item.ItemID, source.Record))
			}
		case "oot-derived-skeleton-key", "oot-derived-platinum-token", "mm-derived-platinum-token", "oot-derived-magical-rupee", "mm-derived-skeleton-key", "mm-derived-transcendent-fairy":
		default:
			panic(fmt.Sprintf("catalog item %s has unsupported source kind %s", item.ItemID, source.Kind))
		}
		sources[item.ItemID] = source
		if shouldTrackCatalogItem(source) {
			tracked = append(tracked, item)
		}
	}
	return tracked, sources, usedBits
}

func shouldTrackCatalogItem(source catalogItemSource) bool {
	switch source.Kind {
	case "oot-derived-key-ring", "mm-derived-key-ring", "oot-derived-skeleton-key", "oot-derived-platinum-token", "mm-derived-platinum-token", "oot-derived-magical-rupee", "mm-derived-skeleton-key", "mm-derived-transcendent-fairy":
		return false
	default:
		return true
	}
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

func ootInventorySlotEntry(slot int) inventorySlotEntry {
	if slot >= 0 && slot < len(ootInventoryEntries) {
		return ootInventoryEntries[slot]
	}
	return inventorySlotEntry{}
}

func mmInventorySlotEntry(slot int) inventorySlotEntry {
	if slot >= 0 && slot < len(mmInventoryEntries) {
		return mmInventoryEntries[slot]
	}
	return inventorySlotEntry{}
}

func ootInventorySlotName(slot int) string {
	return ootInventorySlotEntry(slot).ItemID
}

func mmInventorySlotName(slot int) string {
	return mmInventorySlotEntry(slot).ItemID
}

func mustCatalogItemSource(itemID string) catalogItemSource {
	source, ok := catalogItemSources[itemID]
	if !ok {
		panic(fmt.Sprintf("missing catalog source for %s", itemID))
	}
	return source
}
