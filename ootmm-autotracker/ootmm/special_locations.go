package ootmm

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

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
	case "weekEventReg":
		return mmSymbolCheckSourceWeekEvent, "MM_week_event_"
	case "gMmOwlFlags":
		return mmSymbolCheckSourceOwlActivation, "MM_owl_activation_"
	}
	return 0, ""
}

//go:embed special_locations.json
var embeddedSpecialLocations []byte

// loadMmSymbolChecks parses special_locations.json and builds the mmSymbolChecks slice.
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
