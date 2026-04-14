package ootmm

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type locationFile struct {
	Scene   []sceneLocationEntry  `json:"scene"`
	Bitmap  []bitmapLocationEntry `json:"bitmap"`
	Symbols []symbolLocationEntry `json:"symbols"`
}

type sceneLocationEntry struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type bitmapLocationEntry struct {
	Block string `json:"block"`
	Bit   int    `json:"bit"`
	Name  string `json:"name"`
}

type symbolLocationEntry struct {
	Game   string `json:"game"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

var (
	checkNameTable         map[string]string
	npcCheckTables         map[string]map[int]string
	xflagCheckTables       map[string]map[int]string
	shopCheckTables        map[string]map[int]string
	scrubCheckTables       map[string]map[int]string
	silverRupeeCheckTables map[string]map[int]string
	npcSymbolTables        map[string]map[string]string
)

//go:embed locations.json
var embeddedLocations []byte

func init() {
	checkNameTable = map[string]string{}
	npcCheckTables = map[string]map[int]string{"OOT": {}, "MM": {}}
	xflagCheckTables = map[string]map[int]string{"OOT": {}, "MM": {}}
	shopCheckTables = map[string]map[int]string{"OOT": {}, "MM": {}}
	scrubCheckTables = map[string]map[int]string{"OOT": {}, "MM": {}}
	silverRupeeCheckTables = map[string]map[int]string{"OOT": {}, "MM": {}}
	npcSymbolTables = map[string]map[string]string{"OOT": {}, "MM": {}}

	var locations locationFile
	if err := json.Unmarshal(embeddedLocations, &locations); err != nil {
		panic(fmt.Sprintf("parse embedded location mapping: %v", err))
	}

	for _, entry := range locations.Scene {
		if entry.Key == "" || entry.Name == "" {
			panic("scene location entry is missing key or name")
		}
		if _, exists := checkNameTable[entry.Key]; exists {
			panic(fmt.Sprintf("duplicate scene location key %s", entry.Key))
		}
		checkNameTable[entry.Key] = entry.Name
	}

	for _, entry := range locations.Bitmap {
		if entry.Block == "" || entry.Name == "" {
			panic("bitmap location entry is missing block or name")
		}
		if entry.Bit < 0 {
			panic(fmt.Sprintf("bitmap location entry %s has invalid bit %d", entry.Block, entry.Bit))
		}
		table := tableForBitmapBlock(entry.Block)
		if _, exists := table[entry.Bit]; exists {
			panic(fmt.Sprintf("duplicate bitmap location for %s bit %d", entry.Block, entry.Bit))
		}
		table[entry.Bit] = entry.Name
	}

	for _, entry := range locations.Symbols {
		if entry.Game == "" || entry.Symbol == "" || entry.Name == "" {
			panic("symbol location entry is missing game, symbol, or name")
		}
		gameTable, ok := npcSymbolTables[entry.Game]
		if !ok {
			panic(fmt.Sprintf("unsupported symbol entry game %s", entry.Game))
		}
		if _, exists := gameTable[entry.Symbol]; exists {
			panic(fmt.Sprintf("duplicate symbol location for %s %s", entry.Game, entry.Symbol))
		}
		gameTable[entry.Symbol] = entry.Name
	}
}

func tableForBitmapBlock(block string) map[int]string {
	switch block {
	case "npcOot":
		return npcCheckTables["OOT"]
	case "npcMm":
		return npcCheckTables["MM"]
	case "xflagsOot":
		return xflagCheckTables["OOT"]
	case "xflagsMm":
		return xflagCheckTables["MM"]
	case "shopsOot":
		return shopCheckTables["OOT"]
	case "shopsMm":
		return shopCheckTables["MM"]
	case "scrubsOot":
		return scrubCheckTables["OOT"]
	case "srOot":
		return silverRupeeCheckTables["OOT"]
	default:
		panic(fmt.Sprintf("unsupported bitmap location block %s", block))
	}
}

func sceneCheckKey(game string, scene int, kind string, bit int) string {
	return game + "_" + kind + "_" + itoa(scene) + "_" + itoa(bit)
}

func lookupSceneCheckName(game string, scene int, kind string, bit int) (string, bool) {
	key := sceneCheckKey(game, scene, kind, bit)
	name, ok := checkNameTable[key]
	return name, ok
}

func sceneCheckName(game string, scene int, kind string, bit int) string {
	if name, ok := lookupSceneCheckName(game, scene, kind, bit); ok {
		return name
	}
	return sceneCheckKey(game, scene, kind, bit)
}

func npcCheckName(game string, id int) (string, bool) {
	if gameTable, ok := npcCheckTables[game]; ok {
		name, ok := gameTable[id]
		return name, ok
	}
	return "", false
}

func npcSymbolCheckName(game string, symbol string) (string, bool) {
	if gameTable, ok := npcSymbolTables[game]; ok {
		name, ok := gameTable[symbol]
		return name, ok
	}
	return "", false
}

func xflagCheckName(game string, bitPos int) (string, bool) {
	if gameTable, ok := xflagCheckTables[game]; ok {
		name, ok := gameTable[bitPos]
		return name, ok
	}
	return "", false
}

func shopCheckName(game string, bitPos int) (string, bool) {
	if gameTable, ok := shopCheckTables[game]; ok {
		name, ok := gameTable[bitPos]
		return name, ok
	}
	return "", false
}

func scrubCheckName(game string, bitPos int) (string, bool) {
	if gameTable, ok := scrubCheckTables[game]; ok {
		name, ok := gameTable[bitPos]
		return name, ok
	}
	return "", false
}

func silverRupeeCheckName(game string, bitPos int) (string, bool) {
	if gameTable, ok := silverRupeeCheckTables[game]; ok {
		name, ok := gameTable[bitPos]
		return name, ok
	}
	return "", false
}
