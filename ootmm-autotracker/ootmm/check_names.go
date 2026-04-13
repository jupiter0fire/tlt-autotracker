package ootmm

import (
	"encoding/binary"
	"encoding/csv"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/ootmm-autotracker/n64"
)

const (
	ootSceneTableSize = 101
	mmSceneTableSize  = 113

	xflagTableOotScenesAddr = 0x80948d0
	xflagTableOotSetupsAddr = 0x80949a0
	xflagTableOotRoomsAddr  = 0x8094ac0
	xflagTableMmScenesAddr  = 0x8097ba0
	xflagTableMmSetupsAddr  = 0x8097c90
	xflagTableMmRoomsAddr   = 0x8097d80
)

type xflagDefinition struct {
	SceneID int
	SetupID int
	RoomID  int
	SliceID int
	ActorID int
	Name    string
}

var (
	checkNameTable    map[string]string
	npcCheckTables    map[string]map[int]string
	npcSymbolTables   map[string]map[string]string
	xflagDefinitions  map[string][]xflagDefinition
	xflagCheckTables  = map[string]map[int]string{"OOT": {}, "MM": {}}
	checkTablesLoaded bool
	xflagTablesOnce   sync.Once
)

func init() {
	loadCheckTables()
}

func loadCheckTables() {
	for _, root := range candidateOoTMMDataRoots() {
		scenes, npcs, defs, err := loadCheckTablesFromRoot(root)
		if err == nil && (len(scenes) > 0 || len(npcs["OOT"]) > 0 || len(npcs["MM"]) > 0) {
			checkNameTable = scenes
			npcCheckTables = npcs
			npcSymbolTables = buildNpcSymbolTables(root)
			xflagDefinitions = defs
			checkTablesLoaded = true
			return
		}
	}

	checkNameTable = map[string]string{}
	npcCheckTables = map[string]map[int]string{"OOT": {}, "MM": {}}
	npcSymbolTables = map[string]map[string]string{"OOT": {}, "MM": {}}
	xflagDefinitions = map[string][]xflagDefinition{"OOT": nil, "MM": nil}
}

func buildNpcSymbolTables(root string) map[string]map[string]string {
	result := map[string]map[string]string{"OOT": {}, "MM": {}}
	for game, path := range map[string]string{
		"OOT": filepath.Join(root, "pool", "pool_oot.csv"),
		"MM":  filepath.Join(root, "pool", "pool_mm.csv"),
	} {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		reader := csv.NewReader(file)
		reader.TrimLeadingSpace = true
		reader.FieldsPerRecord = -1
		rows, err := reader.ReadAll()
		file.Close()
		if err != nil {
			continue
		}
		for _, row := range rows[1:] {
			if len(row) < 5 || strings.TrimSpace(row[1]) != "npc" {
				continue
			}
			location := strings.TrimSpace(row[0])
			symbol := strings.TrimSpace(row[4])
			if location == "" || symbol == "" {
				continue
			}
			if _, exists := result[game][symbol]; !exists {
				result[game][symbol] = location
			}
		}
	}
	return result
}

func candidateOoTMMDataRoots() []string {
	candidates := []string{}
	seen := map[string]struct{}{}
	add := func(path string) {
		if path == "" {
			return
		}
		clean := filepath.Clean(path)
		if _, ok := seen[clean]; ok {
			return
		}
		seen[clean] = struct{}{}
		candidates = append(candidates, clean)
	}

	if cwd, err := os.Getwd(); err == nil {
		add(filepath.Join(cwd, "..", "OoTMM", "packages", "data", "src"))
		add(filepath.Join(cwd, "OoTMM", "packages", "data", "src"))
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		add(filepath.Join(exeDir, "..", "OoTMM", "packages", "data", "src"))
		add(filepath.Join(exeDir, "OoTMM", "packages", "data", "src"))
	}

	if _, file, _, ok := runtime.Caller(0); ok {
		sourceDir := filepath.Dir(file)
		add(filepath.Join(sourceDir, "..", "..", "OoTMM", "packages", "data", "src"))
	}

	return candidates
}

func loadCheckTablesFromRoot(root string) (map[string]string, map[string]map[int]string, map[string][]xflagDefinition, error) {
	scenes, err := loadSceneIDs(filepath.Join(root, "defs", "scenes.yml"))
	if err != nil {
		return nil, nil, nil, err
	}

	npcs, err := loadSymbolIDs(filepath.Join(root, "defs", "npc.yml"))
	if err != nil {
		return nil, nil, nil, err
	}

	table := map[string]string{}
	npcTable := map[string]map[int]string{"OOT": {}, "MM": {}}
	xflags := map[string][]xflagDefinition{"OOT": nil, "MM": nil}
	if err := addCheckNamesFromCSV(table, npcTable, xflags, filepath.Join(root, "pool", "pool_oot.csv"), "OOT", scenes, npcs); err != nil {
		return nil, nil, nil, err
	}
	if err := addCheckNamesFromCSV(table, npcTable, xflags, filepath.Join(root, "pool", "pool_mm.csv"), "MM", scenes, npcs); err != nil {
		return nil, nil, nil, err
	}
	return table, npcTable, xflags, nil
}

func loadSceneIDs(path string) (map[string]int, error) {
	return loadSymbolIDs(path)
}

func loadSymbolIDs(path string) (map[string]int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	values := map[string]int{}
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		name, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		name = strings.TrimSpace(name)
		if !strings.HasPrefix(name, "OOT_") && !strings.HasPrefix(name, "MM_") {
			continue
		}

		sceneID, err := strconv.ParseInt(strings.TrimSpace(value), 0, 64)
		if err != nil {
			continue
		}
		values[name] = int(sceneID)
	}

	return values, nil
}

func addCheckNamesFromCSV(table map[string]string, npcTable map[string]map[int]string, xflags map[string][]xflagDefinition, path, game string, scenes map[string]int, npcs map[string]int) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	rows, err := reader.ReadAll()
	if err != nil {
		return err
	}

	byKey := map[string]string{}
	conflicts := map[string]struct{}{}
	npcByID := map[int]string{}
	npcConflicts := map[int]struct{}{}

	for _, row := range rows[1:] {
		if len(row) < 5 {
			continue
		}

		location := strings.TrimSpace(row[0])
		checkType := strings.TrimSpace(row[1])
		sceneName := strings.TrimSpace(row[3])
		value := strings.TrimSpace(row[4])
		if location == "" || sceneName == "" || value == "" {
			continue
		}

		switch checkType {
		case "chest", "collectible":
			sceneID, ok := scenes[game+"_"+sceneName]
			if !ok {
				continue
			}
			flagValue, err := strconv.ParseInt(value, 0, 64)
			if err != nil || flagValue < 0 || flagValue >= 32 {
				continue
			}
			kind := "collect"
			if checkType == "chest" {
				kind = "chest"
			}
			key := sceneCheckKey(game, sceneID, kind, int(flagValue))
			if existing, ok := byKey[key]; ok && existing != location {
				conflicts[key] = struct{}{}
				continue
			}
			byKey[key] = location
		case "npc":
			npcID, ok := npcs[game+"_"+value]
			if !ok {
				continue
			}
			if existing, ok := npcByID[npcID]; ok && existing != location {
				npcConflicts[npcID] = struct{}{}
				continue
			}
			npcByID[npcID] = location
		default:
			if !isXflagCheckType(checkType) {
				continue
			}
			sceneID, ok := scenes[game+"_"+sceneName]
			if !ok {
				continue
			}
			rawID, err := strconv.ParseInt(value, 0, 64)
			if err != nil || rawID < 0 {
				continue
			}
			xflags[game] = append(xflags[game], xflagDefinition{
				SceneID: sceneID,
				SetupID: int((rawID >> 14) & 0x3),
				RoomID:  int((rawID >> 8) & 0x3f),
				SliceID: int(rawID >> 16),
				ActorID: int(rawID & 0xff),
				Name:    location,
			})
		}
	}

	for key, name := range byKey {
		if _, conflicted := conflicts[key]; conflicted {
			continue
		}
		table[key] = name
	}
	for id, name := range npcByID {
		if _, conflicted := npcConflicts[id]; conflicted {
			continue
		}
		npcTable[game][id] = name
	}

	return nil
}

func isXflagCheckType(checkType string) bool {
	switch checkType {
	case "pot", "crate", "barrel", "grass", "tree", "bush", "rock", "soil", "fairy", "snowball", "hive", "rupee", "heart", "fairy_spot", "wonder", "butterfly", "redboulder", "icicle", "redice":
		return true
	default:
		return false
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

func ensureRuntimeCheckTables(mem *n64.Memory) {
	if mem == nil || !checkTablesLoaded {
		return
	}

	xflagTablesOnce.Do(func() {
		xflagCheckTables["OOT"] = buildRuntimeXflagCheckTable(mem, xflagDefinitions["OOT"], ootSceneTableSize, XflagsCountOot*8, xflagTableOotScenesAddr, xflagTableOotSetupsAddr, xflagTableOotRoomsAddr)
		xflagCheckTables["MM"] = buildRuntimeXflagCheckTable(mem, xflagDefinitions["MM"], mmSceneTableSize, XflagsCountMm*8, xflagTableMmScenesAddr, xflagTableMmSetupsAddr, xflagTableMmRoomsAddr)
	})
}

func buildRuntimeXflagCheckTable(mem *n64.Memory, defs []xflagDefinition, sceneCount int, bitLimit int, scenesAddr uint32, setupsAddr uint32, roomsAddr uint32) map[int]string {
	if len(defs) == 0 {
		return map[int]string{}
	}

	scenesTable, ok := readUint16Table(mem, scenesAddr, sceneCount)
	if !ok {
		return map[int]string{}
	}

	maxSetupIndex := -1
	for _, def := range defs {
		if def.SceneID < 0 || def.SceneID >= len(scenesTable) {
			continue
		}
		setupIndex := int(scenesTable[def.SceneID]) + def.SetupID
		if setupIndex > maxSetupIndex {
			maxSetupIndex = setupIndex
		}
	}
	if maxSetupIndex < 0 {
		return map[int]string{}
	}

	setupsTable, ok := readUint16Table(mem, setupsAddr, maxSetupIndex+1)
	if !ok {
		return map[int]string{}
	}

	maxRoomIndex := -1
	for _, def := range defs {
		if def.SceneID < 0 || def.SceneID >= len(scenesTable) {
			continue
		}
		setupIndex := int(scenesTable[def.SceneID]) + def.SetupID
		if setupIndex < 0 || setupIndex >= len(setupsTable) {
			continue
		}
		roomIndex := int(setupsTable[setupIndex]) + def.RoomID*12 + def.SliceID
		if roomIndex > maxRoomIndex {
			maxRoomIndex = roomIndex
		}
	}
	if maxRoomIndex < 0 {
		return map[int]string{}
	}

	roomsTable, ok := readInt16Table(mem, roomsAddr, maxRoomIndex+1)
	if !ok {
		return map[int]string{}
	}

	result := map[int]string{}
	conflicts := map[int]struct{}{}
	for _, def := range defs {
		if def.SceneID < 0 || def.SceneID >= len(scenesTable) {
			continue
		}
		setupIndex := int(scenesTable[def.SceneID]) + def.SetupID
		if setupIndex < 0 || setupIndex >= len(setupsTable) {
			continue
		}
		roomIndex := int(setupsTable[setupIndex]) + def.RoomID*12 + def.SliceID
		if roomIndex < 0 || roomIndex >= len(roomsTable) {
			continue
		}
		bitPos := int(roomsTable[roomIndex]) + def.ActorID
		if bitPos < 0 || bitPos >= bitLimit {
			continue
		}
		if existing, ok := result[bitPos]; ok && existing != def.Name {
			conflicts[bitPos] = struct{}{}
			continue
		}
		result[bitPos] = def.Name
	}
	for bitPos := range conflicts {
		delete(result, bitPos)
	}
	return result
}

func readUint16Table(mem *n64.Memory, addr uint32, count int) ([]uint16, bool) {
	data, err := mem.Read(addr, count*2)
	if err != nil || len(data) != count*2 {
		return nil, false
	}
	values := make([]uint16, count)
	for i := 0; i < count; i++ {
		values[i] = binary.BigEndian.Uint16(data[i*2:])
	}
	return values, true
}

func readInt16Table(mem *n64.Memory, addr uint32, count int) ([]int16, bool) {
	data, err := mem.Read(addr, count*2)
	if err != nil || len(data) != count*2 {
		return nil, false
	}
	values := make([]int16, count)
	for i := 0; i < count; i++ {
		values[i] = int16(binary.BigEndian.Uint16(data[i*2:]))
	}
	return values, true
}
