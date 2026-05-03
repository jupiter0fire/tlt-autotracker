package ootmm

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type locationFile struct {
	Scene           []sceneLocationEntry  `json:"scene"`
	SceneConflicts  []sceneConflictEntry  `json:"scene_conflicts"`
	Bitmap          []bitmapLocationEntry `json:"bitmap"`
	BitmapConflicts []bitmapConflictEntry `json:"bitmap_conflicts"`
	Symbols         []symbolLocationEntry `json:"symbols"`
}

type sceneLocationEntry struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type sceneConflictEntry struct {
	Key       string `json:"key"`
	DungeonMq int    `json:"dungeonMq"`
	Vanilla   string `json:"vanilla"`
	Mq        string `json:"mq"`
}

type bitmapLocationEntry struct {
	Block string `json:"block"`
	Bit   int    `json:"bit"`
	Name  string `json:"name"`
}

type bitmapConflictEntry struct {
	Block     string   `json:"block"`
	Bit       int      `json:"bit"`
	DungeonMq int      `json:"dungeonMq"`
	Vanilla   []string `json:"vanilla"`
	Mq        []string `json:"mq"`
}

type symbolLocationEntry struct {
	Game   string `json:"game"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

var (
	checkNameTable         map[string]string
	ootSceneConflictTable  map[string]sceneConflictEntry
	npcCheckTables         map[string]map[int]string
	gsCheckTables          map[string]map[int]string
	xflagCheckTables       map[string]map[int]string
	ootBitmapConflictTable map[string]map[int]bitmapConflictEntry
	shopCheckTables        map[string]map[int]string
	scrubCheckTables       map[string]map[int]string
	silverRupeeCheckTables map[string]map[int]string
	npcSymbolTables        map[string]map[string]string
	sceneCheckFallbacks    map[string]string
)

//go:embed locations.json
var embeddedLocations []byte

func init() {
	checkNameTable = map[string]string{}
	ootSceneConflictTable = map[string]sceneConflictEntry{}
	npcCheckTables = map[string]map[int]string{"OOT": {}, "MM": {}}
	gsCheckTables = map[string]map[int]string{"OOT": {}}
	xflagCheckTables = map[string]map[int]string{"OOT": {}, "MM": {}}
	ootBitmapConflictTable = map[string]map[int]bitmapConflictEntry{
		"gsOot":     {},
		"xflagsOot": {},
	}
	shopCheckTables = map[string]map[int]string{"OOT": {}, "MM": {}}
	scrubCheckTables = map[string]map[int]string{"OOT": {}, "MM": {}}
	silverRupeeCheckTables = map[string]map[int]string{"OOT": {}, "MM": {}}
	npcSymbolTables = map[string]map[string]string{"OOT": {}, "MM": {}}
	sceneCheckFallbacks = map[string]string{
		sceneCheckKey("OOT", 1, "collect", 24): "Dodongo Cavern Heart Miniboss Lava",
	}

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

	for _, entry := range locations.SceneConflicts {
		if entry.Key == "" {
			panic("scene conflict entry is missing key")
		}
		if entry.DungeonMq < 0 || entry.DungeonMq >= OotMqDungeonCount {
			panic(fmt.Sprintf("scene conflict entry %s has invalid dungeon MQ id %d", entry.Key, entry.DungeonMq))
		}
		if entry.Vanilla == "" || entry.Mq == "" {
			panic(fmt.Sprintf("scene conflict entry %s is missing variant names", entry.Key))
		}
		if _, exists := ootSceneConflictTable[entry.Key]; exists {
			panic(fmt.Sprintf("duplicate scene conflict entry for %s", entry.Key))
		}
		ootSceneConflictTable[entry.Key] = entry
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

	for _, entry := range locations.BitmapConflicts {
		conflicts, ok := ootBitmapConflictTable[entry.Block]
		if !ok {
			panic(fmt.Sprintf("unsupported bitmap conflict block %s", entry.Block))
		}
		if entry.Bit < 0 {
			panic(fmt.Sprintf("bitmap conflict entry %s has invalid bit %d", entry.Block, entry.Bit))
		}
		if entry.DungeonMq < 0 || entry.DungeonMq >= OotMqDungeonCount {
			panic(fmt.Sprintf("bitmap conflict entry %s bit %d has invalid dungeon MQ id %d", entry.Block, entry.Bit, entry.DungeonMq))
		}
		if len(entry.Vanilla) == 0 || len(entry.Mq) == 0 {
			panic(fmt.Sprintf("bitmap conflict entry %s bit %d is missing variant names", entry.Block, entry.Bit))
		}
		if _, exists := conflicts[entry.Bit]; exists {
			panic(fmt.Sprintf("duplicate bitmap conflict entry for %s bit %d", entry.Block, entry.Bit))
		}
		conflicts[entry.Bit] = entry
	}

	for _, entry := range locations.Symbols {
		if entry.Game == "" || entry.Symbol == "" || entry.Name == "" {
			panic("symbol location entry is missing game, symbol, or name")
		}
		if entry.Game == "MM" {
			// MM symbol check names are now loaded from special_locations_mm.json
			continue
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
	case "gsOot":
		return gsCheckTables["OOT"]
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
	if !ok {
		name, ok = sceneCheckFallbacks[key]
	}
	return name, ok
}

func sceneCheckName(game string, scene int, kind string, bit int) string {
	if name, ok := lookupSceneCheckName(game, scene, kind, bit); ok {
		return name
	}
	return sceneCheckKey(game, scene, kind, bit)
}

func ootSceneCheckNameForState(oot *OotState, scene int, kind string, bit int) (string, bool) {
	key := sceneCheckKey("OOT", scene, kind, bit)
	if name, ok := checkNameTable[key]; ok {
		return name, true
	}
	if name, ok := ootConflictingSceneCheckName(oot, key); ok {
		return name, true
	}
	name, ok := sceneCheckFallbacks[key]
	return name, ok
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

func gsCheckName(game string, bitPos int) (string, bool) {
	if gameTable, ok := gsCheckTables[game]; ok {
		name, ok := gameTable[bitPos]
		return name, ok
	}
	return "", false
}

func ootConflictingBitmapCheckNames(oot *OotState, block string, bitPos int) ([]string, bool) {
	entries, ok := ootBitmapConflictTable[block]
	if !ok {
		return nil, false
	}
	entry, ok := entries[bitPos]
	if !ok {
		return nil, false
	}
	mq, known := ootMqDungeonState(oot, entry.DungeonMq)
	if !known {
		return nil, false
	}
	if mq {
		return entry.Mq, true
	}
	return entry.Vanilla, true
}

func ootConflictingXflagCheckNames(oot *OotState, bitPos int) ([]string, bool) {
	return ootConflictingBitmapCheckNames(oot, "xflagsOot", bitPos)
}

func ootConflictingGsCheckNames(oot *OotState, bitPos int) ([]string, bool) {
	return ootConflictingBitmapCheckNames(oot, "gsOot", bitPos)
}

func ootConflictingSceneCheckName(oot *OotState, key string) (string, bool) {
	entry, ok := ootSceneConflictTable[key]
	if !ok {
		return "", false
	}
	mq, known := ootMqDungeonState(oot, entry.DungeonMq)
	if !known {
		return "", false
	}
	if mq {
		return entry.Mq, true
	}
	return entry.Vanilla, true
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
