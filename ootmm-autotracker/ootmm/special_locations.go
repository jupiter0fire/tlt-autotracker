package ootmm

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/bits"
	"strconv"
	"strings"
)

const mmExtraBossLegacyDungeonIndexBase = 8

type specialLocationEntry struct {
	Symbol    string                       `json:"symbol"`
	Name      string                       `json:"name"`
	Sources   []specialLocationSourceEntry `json:"sources"`
	Bits      []int                        `json:"bits"`
	ByteIndex int                          `json:"byteIndex"`
	Mask      uint8                        `json:"mask"`
}

type specialLocationSourceEntry struct {
	Group string `json:"group"`
	Field string `json:"field"`
	Mask  string `json:"mask"`
}

func getSourceInfo(group string) (mmSymbolCheckSource, string) {
	switch group {
	case "gMmExtraFlags":
		return mmSymbolCheckSourceExtraFlags, "MM_extra_6_"
	case "gMmExtraFlags2":
		return mmSymbolCheckSourceExtraFlags2, "MM_extra_"
	case "gMmExtraFlags3":
		return mmSymbolCheckSourceExtraFlags3, "MM_extra_13_"
	case "gMmExtraBoss":
		return mmSymbolCheckSourceExtraBoss, "MM_boss_remains_dungeon_"
	case "weekEventReg":
		return mmSymbolCheckSourceWeekEvent, "MM_week_event_"
	case "gMmOwlFlags":
		return mmSymbolCheckSourceOwlActivation, "MM_owl_activation_"
	}
	return 0, ""
}

//go:embed special_locations_mm.json
var embeddedSpecialLocations []byte

//go:embed special_locations_oot.json
var embeddedOotSpecialLocations []byte

type ootSpecialLocationEntry struct {
	Symbol  string                       `json:"symbol"`
	Name    string                       `json:"name"`
	Sources []ootSpecialLocationSourceEntry `json:"sources"`
	Note    string                       `json:"note"`
}

type ootSpecialLocationSourceEntry struct {
	Group string `json:"group"`
	Field string `json:"field"`
	Bit   int    `json:"bit"`
	Flag  int    `json:"flag"`
	Mask  string `json:"mask"`
}

func getOotSourceInfo(group string, field string) (ootSymbolCheckSource, string) {
	switch group {
	case "gOotExtraFlags":
		return ootSymbolCheckSourceExtraFlags, "OOT_extra_2_"
	case "inventoryQuest":
		return ootSymbolCheckSourceQuest, "OOT_quest_"
	case "gOotTradeSave":
		if strings.Contains(field, "child") {
			return ootSymbolCheckSourceChildTrade, "OOT_child_trade_"
		}
		return ootSymbolCheckSourceTrade, "OOT_trade_"
	case "eventsChk":
		return ootSymbolCheckSourceEvent, "OOT_event_"
	case "eventsItem":
		return ootSymbolCheckSourceEventItem, "OOT_event_item_"
	case "eventsMisc":
		return ootSymbolCheckSourceEventMisc, "OOT_event_misc_"
	}
	return 0, ""
}

// loadMmSymbolChecks parses special_locations_mm.json and builds the mmSymbolChecks slice.
func loadMmSymbolChecks() []mmSymbolCheck {
	var entries []specialLocationEntry
	if err := json.Unmarshal(embeddedSpecialLocations, &entries); err != nil {
		panic(fmt.Sprintf("parse embedded special locations: %v", err))
	}

	var result []mmSymbolCheck
	for _, entry := range entries {
		if entry.Name == "" {
			continue
		}

		bareSymbol := strings.TrimPrefix(entry.Symbol, "MM_")

		for _, src := range entry.Sources {
			source, keyPrefix := getSourceInfo(src.Group)
			if keyPrefix == "" {
				continue
			}

			check := mmSymbolCheck{
				source:    source,
				symbol:    bareSymbol,
				name:      entry.Name,
				keyPrefix: keyPrefix,
				bit:       0,
				byteIndex: 0,
				mask:      0,
			}

			switch source {
			case mmSymbolCheckSourceExtraBoss:
				if keyBit, sourceMask, ok := parseExtraBossSource(src); ok {
					check.bit = keyBit
					check.mask = sourceMask
				} else {
					continue
				}
			case mmSymbolCheckSourceWeekEvent:
				if byteIndex, mask, ok := parseWeekEventSource(entry, src); ok {
					check.byteIndex = byteIndex
					check.mask = mask
				} else {
					continue
				}
			default:
				if len(entry.Bits) == 0 {
					continue
				}
				check.bit = entry.Bits[0]
			}

			result = append(result, check)
		}
	}
	return result
}

func parseExtraBossSource(src specialLocationSourceEntry) (int, uint8, bool) {
	if src.Mask == "" {
		return 0, 0, false
	}
	value, err := strconv.ParseUint(src.Mask, 0, 8)
	if err != nil {
		return 0, 0, false
	}
	mask := uint8(value)
	if mask == 0 || mask&(mask-1) != 0 {
		return 0, 0, false
	}
	return bits.TrailingZeros8(mask) + mmExtraBossLegacyDungeonIndexBase, mask, true
}

func parseWeekEventSource(entry specialLocationEntry, src specialLocationSourceEntry) (int, uint8, bool) {
	byteIndex := entry.ByteIndex
	mask := entry.Mask
	hasByteIndex := entry.Mask != 0

	if src.Field != "" {
		if start := strings.LastIndex(src.Field, "["); start >= 0 {
			if end := strings.LastIndex(src.Field, "]"); end > start {
				if value, err := strconv.Atoi(src.Field[start+1 : end]); err == nil {
					byteIndex = value
					hasByteIndex = true
				}
			}
		}
	}
	if src.Mask != "" {
		if value, err := strconv.ParseUint(src.Mask, 0, 8); err == nil {
			mask = uint8(value)
		}
	}

	if !hasByteIndex || byteIndex < 0 || mask == 0 {
		return 0, 0, false
	}
	return byteIndex, mask, true
}

// loadOotSymbolChecks parses special_locations_oot.json and builds the ootSymbolChecks slice.
func loadOotSymbolChecks() []ootSymbolCheck {
	var entries []ootSpecialLocationEntry
	if err := json.Unmarshal(embeddedOotSpecialLocations, &entries); err != nil {
		panic(fmt.Sprintf("parse embedded OoT special locations: %v", err))
	}

	var result []ootSymbolCheck
	for _, entry := range entries {
		if entry.Symbol == "" {
			continue
		}

		for _, src := range entry.Sources {
			source, keyPrefix := getOotSourceInfo(src.Group, src.Field)
			if keyPrefix == "" {
				continue
			}

			check := ootSymbolCheck{
				source:    source,
				symbol:    entry.Symbol,
				keyPrefix: keyPrefix,
				bit:       0,
				flags:     nil,
				mask:      0,
			}

			switch source {
			case ootSymbolCheckSourceExtraFlags:
				check.bit = src.Bit
			case ootSymbolCheckSourceQuest:
				check.bit = src.Bit
			case ootSymbolCheckSourceChildTrade, ootSymbolCheckSourceTrade:
				if src.Mask != "" {
					if value, err := strconv.ParseUint(src.Mask, 0, 16); err == nil {
						check.mask = uint16(value)
					}
				}
			case ootSymbolCheckSourceEvent, ootSymbolCheckSourceEventItem, ootSymbolCheckSourceEventMisc:
				check.flags = []int{src.Flag}
			}

			result = append(result, check)
		}
	}
	return result
}
