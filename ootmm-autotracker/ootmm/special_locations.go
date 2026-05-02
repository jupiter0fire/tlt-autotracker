package ootmm

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

type specialLocationEntry struct {
	Symbol  string                      `json:"symbol"`
	Name    string                      `json:"name"`
	Sources []specialLocationSourceEntry `json:"sources"`
}

type specialLocationSourceEntry struct {
	Group string `json:"group"`
	Field string `json:"field"`
	Mask  string `json:"mask"`
}

//go:embed special_locations.json
var embeddedSpecialLocations []byte

// mmSymbolBitMap maps (group, symbol) -> bit position for MM checks.
// This preserves the hardcoded bit positions and pairs them with JSON-sourced names.
var mmSymbolBitMap = map[string]map[string]mmSymbolCheckInfo{
	"gMmExtraFlags": {
		"KOUME_PICTOGRAPH_BOX": {source: mmSymbolCheckSourceExtraFlags, bit: 31, keyPrefix: "MM_extra_6_"},
		"SCRUB_LAND":           {source: mmSymbolCheckSourceExtraFlags, bit: 20, keyPrefix: "MM_extra_6_"},
		"SCRUB_SWAMP":          {source: mmSymbolCheckSourceExtraFlags, bit: 19, keyPrefix: "MM_extra_6_"},
		"SCRUB_MOUNTAIN":       {source: mmSymbolCheckSourceExtraFlags, bit: 18, keyPrefix: "MM_extra_6_"},
		"SCRUB_OCEAN":          {source: mmSymbolCheckSourceExtraFlags, bit: 17, keyPrefix: "MM_extra_6_"},
		"SCRUB_VALLEY":         {source: mmSymbolCheckSourceExtraFlags, bit: 16, keyPrefix: "MM_extra_6_"},
		"SCRUB_BOMB_BAG":       {source: mmSymbolCheckSourceExtraFlags, bit: 15, keyPrefix: "MM_extra_6_"},
		"GREAT_FAIRY_TOWN":     {source: mmSymbolCheckSourceExtraFlags, bit: 1, keyPrefix: "MM_extra_6_"},
		"GREAT_FAIRY_TOWN_ALT": {source: mmSymbolCheckSourceExtraFlags, bit: 2, keyPrefix: "MM_extra_6_"},
		"GREAT_FAIRY_SWAMP":    {source: mmSymbolCheckSourceExtraFlags, bit: 3, keyPrefix: "MM_extra_6_"},
		"GREAT_FAIRY_MOUNTAIN": {source: mmSymbolCheckSourceExtraFlags, bit: 4, keyPrefix: "MM_extra_6_"},
		"GREAT_FAIRY_OCEAN":    {source: mmSymbolCheckSourceExtraFlags, bit: 5, keyPrefix: "MM_extra_6_"},
		"GREAT_FAIRY_VALLEY":   {source: mmSymbolCheckSourceExtraFlags, bit: 6, keyPrefix: "MM_extra_6_"},
	},
	"gMmExtraFlags2": {
		"MASK_KAFEI":         {source: mmSymbolCheckSourceExtraFlags2, bit: 28, keyPrefix: "MM_extra_"},
		"HONEY_DARLING_1":    {source: mmSymbolCheckSourceExtraFlags2, bit: 27, keyPrefix: "MM_extra_"},
		"ROOM_KEY":           {source: mmSymbolCheckSourceExtraFlags2, bit: 26, keyPrefix: "MM_extra_"},
		"LETTER_TO_KAFEI":    {source: mmSymbolCheckSourceExtraFlags2, bit: 25, keyPrefix: "MM_extra_"},
		"PENDANT_OF_MEMORIES": {source: mmSymbolCheckSourceExtraFlags2, bit: 24, keyPrefix: "MM_extra_"},
		"LETTER_TO_MAMA":     {source: mmSymbolCheckSourceExtraFlags2, bit: 23, keyPrefix: "MM_extra_"},
		"MASK_BLAST":         {source: mmSymbolCheckSourceExtraFlags2, bit: 21, keyPrefix: "MM_extra_"},
		"BOMBER_NOTEBOOK":    {source: mmSymbolCheckSourceExtraFlags2, bit: 22, keyPrefix: "MM_extra_"},
		"DEKU_PLAYGROUND_1":  {source: mmSymbolCheckSourceExtraFlags2, bit: 20, keyPrefix: "MM_extra_"},
		"MASK_COUPLE":        {source: mmSymbolCheckSourceExtraFlags2, bit: 19, keyPrefix: "MM_extra_"},
		"MASK_POSTMAN":       {source: mmSymbolCheckSourceExtraFlags2, bit: 17, keyPrefix: "MM_extra_"},
		"MASK_TROUPE_LEADER": {source: mmSymbolCheckSourceExtraFlags2, bit: 16, keyPrefix: "MM_extra_"},
		"MASK_FIERCE_DEITY":  {source: mmSymbolCheckSourceExtraFlags2, bit: 15, keyPrefix: "MM_extra_"},
		"SKULL_KID_OCARINA":  {source: mmSymbolCheckSourceExtraFlags2, bit: 14, keyPrefix: "MM_extra_"},
		"SONG_ORDER":         {source: mmSymbolCheckSourceExtraFlags2, bit: 13, keyPrefix: "MM_extra_"},
		"MASK_BREMEN":        {source: mmSymbolCheckSourceExtraFlags2, bit: 10, keyPrefix: "MM_extra_"},
		"MASK_SCENTS":        {source: mmSymbolCheckSourceExtraFlags2, bit: 9, keyPrefix: "MM_extra_"},
		"MASK_KAMARO":        {source: mmSymbolCheckSourceExtraFlags2, bit: 8, keyPrefix: "MM_extra_"},
		"MOON_TEAR":          {source: mmSymbolCheckSourceExtraFlags2, bit: 6, keyPrefix: "MM_extra_"},
		"SONG_HEALING":       {source: mmSymbolCheckSourceExtraFlags2, bit: 5, keyPrefix: "MM_extra_"},
		"STRAY_FAIRY_TOWN":   {source: mmSymbolCheckSourceExtraFlags2, bit: 4, keyPrefix: "MM_extra_"},
	},
	"gMmExtraFlags3": {
		"LOTTERY_NIGHT_1": {source: mmSymbolCheckSourceExtraFlags3, bit: 29, keyPrefix: "MM_extra_13_"},
		"LOTTERY_NIGHT_2": {source: mmSymbolCheckSourceExtraFlags3, bit: 28, keyPrefix: "MM_extra_13_"},
		"LOTTERY_NIGHT_3": {source: mmSymbolCheckSourceExtraFlags3, bit: 27, keyPrefix: "MM_extra_13_"},
	},
	"weekEventReg": {
		"TINGLE_MAP_CLOCK_TOWN":   {source: mmSymbolCheckSourceWeekEvent, byteIndex: 0x118 >> 3, mask: 1 << (0x118 & 7), keyPrefix: "MM_week_event_"},
		"TINGLE_MAP_WOODFALL":     {source: mmSymbolCheckSourceWeekEvent, byteIndex: 0x119 >> 3, mask: 1 << (0x119 & 7), keyPrefix: "MM_week_event_"},
		"TINGLE_MAP_SNOWHEAD":     {source: mmSymbolCheckSourceWeekEvent, byteIndex: 0x11a >> 3, mask: 1 << (0x11a & 7), keyPrefix: "MM_week_event_"},
		"TINGLE_MAP_ROMANI_RANCH": {source: mmSymbolCheckSourceWeekEvent, byteIndex: 0x11b >> 3, mask: 1 << (0x11b & 7), keyPrefix: "MM_week_event_"},
		"TINGLE_MAP_GREAT_BAY":    {source: mmSymbolCheckSourceWeekEvent, byteIndex: 0x11c >> 3, mask: 1 << (0x11c & 7), keyPrefix: "MM_week_event_"},
		"TINGLE_MAP_STONE_TOWER":  {source: mmSymbolCheckSourceWeekEvent, byteIndex: 0x11d >> 3, mask: 1 << (0x11d & 7), keyPrefix: "MM_week_event_"},
		"SHOOTING_GAME_SWAMP_1":   {source: mmSymbolCheckSourceWeekEvent, byteIndex: 59, mask: 0x10, keyPrefix: "MM_week_event_"},
		"SHOOTING_GAME_TOWN_1":    {source: mmSymbolCheckSourceWeekEvent, byteIndex: 59, mask: 0x20, keyPrefix: "MM_week_event_"},
		"SWORDSMAN_HEART_PIECE":   {source: mmSymbolCheckSourceWeekEvent, byteIndex: 63, mask: 0x20, keyPrefix: "MM_week_event_"},
	},
	"gMmOwlFlags": {
		"OWL_GREAT_BAY":        {source: mmSymbolCheckSourceOwlActivation, bit: 0, keyPrefix: "MM_owl_activation_"},
		"OWL_ZORA_CAPE":        {source: mmSymbolCheckSourceOwlActivation, bit: 1, keyPrefix: "MM_owl_activation_"},
		"OWL_SNOWHEAD":         {source: mmSymbolCheckSourceOwlActivation, bit: 2, keyPrefix: "MM_owl_activation_"},
		"OWL_MOUNTAIN_VILLAGE": {source: mmSymbolCheckSourceOwlActivation, bit: 3, keyPrefix: "MM_owl_activation_"},
		"OWL_CLOCK_TOWN":       {source: mmSymbolCheckSourceOwlActivation, bit: 4, keyPrefix: "MM_owl_activation_"},
		"OWL_MILK_ROAD":        {source: mmSymbolCheckSourceOwlActivation, bit: 5, keyPrefix: "MM_owl_activation_"},
		"OWL_WOODFALL":         {source: mmSymbolCheckSourceOwlActivation, bit: 6, keyPrefix: "MM_owl_activation_"},
		"OWL_SOUTHERN_SWAMP":   {source: mmSymbolCheckSourceOwlActivation, bit: 7, keyPrefix: "MM_owl_activation_"},
		"OWL_IKANA_CANYON":     {source: mmSymbolCheckSourceOwlActivation, bit: 8, keyPrefix: "MM_owl_activation_"},
		"OWL_STONE_TOWER":      {source: mmSymbolCheckSourceOwlActivation, bit: 9, keyPrefix: "MM_owl_activation_"},
	},
}

type mmSymbolCheckInfo struct {
	source    mmSymbolCheckSource
	bit       int
	byteIndex int
	mask      uint8
	keyPrefix string
}

// loadMmSymbolChecks parses special_locations.json and builds the mmSymbolChecks slice.
// Names come from JSON; bit positions come from the hardcoded map.
func loadMmSymbolChecks() []mmSymbolCheck {
	var entries []specialLocationEntry
	if err := json.Unmarshal(embeddedSpecialLocations, &entries); err != nil {
		panic(fmt.Sprintf("parse embedded special locations: %v", err))
	}

	var result []mmSymbolCheck
	for _, entry := range entries {
		if entry.Name == "" {
			// No trackable check name (e.g. MM_OWL_HIDDEN, MM_MAJORA)
			continue
		}

		bareSymbol := strings.TrimPrefix(entry.Symbol, "MM_")

		// Find the first supported source group to get the check type
		for _, src := range entry.Sources {
			groupMap, ok := mmSymbolBitMap[src.Group]
			if !ok {
				continue // Skip unsupported groups
			}

			info, ok := groupMap[bareSymbol]
			if !ok {
				continue // Symbol not in hardcoded map
			}

			check := mmSymbolCheck{
				source:    info.source,
				symbol:    bareSymbol,
				name:      entry.Name,
				keyPrefix: info.keyPrefix,
				bit:       info.bit,
				byteIndex: info.byteIndex,
				mask:      info.mask,
			}
			result = append(result, check)
			break // Use only the first supported source
		}
	}
	return result
}

