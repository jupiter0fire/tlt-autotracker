package ootmm

import (
	"encoding/binary"
	"fmt"
	"math/bits"

	"ootmm-autotracker/n64"
)

const (
	maxForeignMmChecksumDelta         = 0x400
	sharedOcarinaButtonMaskOotOffset  = 0x7c8
	sharedOcarinaButtonMaskMmOffset   = 0x7ca
	sharedCaughtChildFishWeightOffset = 2037
	sharedCaughtAdultFishWeightOffset = 2057
	sharedCaughtFishWeightCount       = 20

	// SharedCustomSave stores bombchu bag progression in the flag byte after
	// the souls, fish, and respawn blocks.
	sharedBombchuBagFlagsOffset = 2114
	sharedBombchuBagOotShift    = 4
	sharedBombchuBagMmShift     = 6
	sharedBombchuBagMask        = 0x3
)

var sharedCheckBitmapNames = [...]string{
	"xflagsOot",
	"npcOot",
	"shopsOot",
	"scrubsOot",
	"srOot",
	"xflagsMm",
	"npcMm",
	"shopsMm",
}

type sharedStateCandidate struct {
	source string
	state  SharedCustomState
}

// Reader reads OoTMM game state from N64 RDRAM.
type Reader struct {
	mem      *n64.Memory
	detector *Detector

	foreignOotSaveAddr uint32
	foreignMmSaveAddr  uint32
	ootPlayStateAddr   uint32
	comboConfigOotAddr uint32
	comboConfigMmAddr  uint32
	ootMaxKeysAddr     uint32
	ootSilverDataAddr  uint32
	lastKnownOot       OotState
	lastKnownMm        MmState
	lastKnownShared    SharedCustomState
	hasLastKnownOot    bool
	hasLastKnownMm     bool
	hasLastKnownShared bool
	stableGame         ActiveGame
	stableSaveIndex    uint32
	pendingGame        ActiveGame
	pendingSaveIndex   uint32
	hasPendingState    bool
}

func NewReader(mem *n64.Memory) *Reader {
	return &Reader{
		mem:      mem,
		detector: NewDetector(mem),
	}
}

// ReadState performs a full read of the current game state.
// It guards against race conditions during game switches by verifying the
// active game has not changed after reading is complete. When a change is
// detected mid-read the data is discarded (ActiveGame set to GameNone).
func (r *Reader) ReadState() (*GameState, error) {
	state := &GameState{}

	// Detect OoTMM
	saveIndex, valid, err := r.detector.DetectOoTMM()
	if err != nil {
		return nil, fmt.Errorf("detect OoTMM: %w", err)
	}
	state.SaveIndex = saveIndex
	state.Valid = valid

	if !valid {
		r.resetPendingState()
		return state, nil
	}

	// Step 1: Detect active game — if no game is active, skip this frame
	game, err := r.detector.DetectActiveGame()
	if err != nil {
		return nil, fmt.Errorf("detect game: %w", err)
	}
	if game == GameNone {
		r.resetPendingState()
		return state, nil
	}
	state.ActiveGame = game

	activeSaveIndex, err := r.readActiveSaveIndex(game)
	if err != nil {
		return nil, fmt.Errorf("read active save index: %w", err)
	}
	state.SaveIndex = activeSaveIndex

	// Step 2: Read state for the detected game
	switch game {
	case GameOot:
		if err := r.readOotSave(&state.Oot); err != nil {
			return nil, fmt.Errorf("read OoT save: %w", err)
		}
		r.readOotLiveState(&state.Oot)
		if err := r.readForeignMmState(&state.Mm); err != nil {
			return nil, fmt.Errorf("read foreign MM save: %w", err)
		}
	case GameMm:
		if err := r.readMmSave(&state.Mm); err != nil {
			return nil, fmt.Errorf("read MM save: %w", err)
		}
		if err := r.readForeignOotState(&state.Oot); err != nil {
			return nil, fmt.Errorf("read foreign OoT save: %w", err)
		}
	}
	if err := r.readSharedState(game, activeSaveIndex, &state.Shared); err != nil {
		return nil, fmt.Errorf("read shared custom save: %w", err)
	}
	r.readOotRuntimeConfig(game, &state.Oot)

	// Step 3: Re-verify the active game hasn't changed during the read
	gameAfter, err := r.detector.DetectActiveGame()
	if err != nil {
		return nil, fmt.Errorf("re-detect game: %w", err)
	}
	if gameAfter != game {
		// Game changed mid-read — the data is unreliable, discard it
		r.resetPendingState()
		state.ActiveGame = GameNone
		return state, nil
	}

	saveIndexAfter, err := r.readActiveSaveIndex(gameAfter)
	if err != nil {
		return nil, fmt.Errorf("re-read active save index: %w", err)
	}
	if saveIndexAfter != activeSaveIndex {
		r.resetPendingState()
		state.ActiveGame = GameNone
		return state, nil
	}

	if !r.acceptStableState(game, activeSaveIndex) {
		state.ActiveGame = GameNone
		return state, nil
	}

	r.rememberOotState(state.Oot)
	r.rememberMmState(state.Mm)
	r.rememberSharedState(state.Shared)

	return state, nil
}

func (r *Reader) readOotSave(oot *OotState) error {
	// Read the entire OotSaveContext in one large read
	data, err := r.mem.Read(AddrOotSaveCtx, OotSaveCtxSize)
	if err != nil {
		return fmt.Errorf("read OoT save context: %w", err)
	}
	return parseOotSave(oot, data)
}

func parseOotSave(oot *OotState, data []byte) error {
	if len(data) < OotSaveSize {
		return fmt.Errorf("OoT save too small: got %#x bytes", len(data))
	}

	// Parse OotSave (starts at offset 0 within OotSaveContext)
	oot.Age = binary.BigEndian.Uint32(data[OotOffAge:])
	oot.SceneID = binary.BigEndian.Uint16(data[OotOffSceneID:])
	oot.IsBiggoronSword = data[OotOffIsBiggoronSword] != 0

	// Inventory
	copy(oot.Items[:], data[OotOffInvItems:OotOffInvItems+24])
	copy(oot.Ammo[:], data[OotOffInvAmmo:OotOffInvAmmo+15])
	oot.Beans = data[OotOffInvBeans]

	// Equipment and upgrades
	oot.Equipment = binary.BigEndian.Uint16(data[OotOffEquipment:])
	oot.Upgrades = binary.BigEndian.Uint32(data[OotOffUpgrades:])
	oot.QuestItems = binary.BigEndian.Uint32(data[OotOffQuestItems:])
	oot.HeartPieces = uint8((oot.QuestItems >> 28) & 0xF)

	// Dungeon items
	copy(oot.DungeonItems[:], data[OotOffDungeonItems:OotOffDungeonItems+20])
	for i := 0; i < 19; i++ {
		oot.DungeonKeys[i] = int8(data[OotOffDungeonKeys+i])
	}
	oot.GoldTokens = binary.BigEndian.Uint16(data[OotOffGoldTokens:])
	for i := range oot.GsFlags {
		oot.GsFlags[i] = binary.BigEndian.Uint32(data[OotOffGsFlags+i*4:])
	}

	// Scene flags
	for i := 0; i < OotPermCount; i++ {
		off := OotOffPerm + i*OotPermEntrySize
		oot.SceneFlags[i] = SceneFlags{
			Chests:        binary.BigEndian.Uint32(data[off:]),
			Switches:      binary.BigEndian.Uint32(data[off+4:]),
			RoomClear:     binary.BigEndian.Uint32(data[off+8:]),
			Collectibles:  binary.BigEndian.Uint32(data[off+12:]),
			Unused:        binary.BigEndian.Uint32(data[off+16:]),
			VisitedRooms:  binary.BigEndian.Uint32(data[off+20:]),
			VisitedFloors: binary.BigEndian.Uint32(data[off+24:]),
		}
	}

	// Extra records stored in perm[N].raw + 0x10
	for i := 0; i < 20; i++ {
		off := OotOffPerm + i*OotPermEntrySize + OotPermExtraOff
		oot.ExtraRecords[i] = binary.BigEndian.Uint32(data[off:])
	}

	// Event flags
	for i := 0; i < 14; i++ {
		oot.EventsChk[i] = binary.BigEndian.Uint16(data[OotOffEventsChk+i*2:])
	}
	eventsMiscOff := OotOffEventsMisc
	for i := 0; i < 30; i++ {
		oot.EventsMisc[i] = binary.BigEndian.Uint16(data[eventsMiscOff+i*2:])
	}

	if len(data) >= OotCtxOffGameMode+4 {
		oot.GameMode = binary.BigEndian.Uint32(data[OotCtxOffGameMode:])
	} else {
		oot.GameMode = 0
	}

	return nil
}

func (r *Reader) readMmSave(mm *MmState) error {
	// Read the entire MmSaveContext in one large read
	data, err := r.mem.Read(AddrMmSaveCtx, MmSaveCtxSize)
	if err != nil {
		return fmt.Errorf("read MM save context: %w", err)
	}
	return parseMmSave(mm, data)
}

func parseMmSave(mm *MmState, data []byte) error {
	if len(data) < MmSaveSize {
		return fmt.Errorf("MM save too small: got %#x bytes", len(data))
	}

	// Basic state
	mm.PlayerForm = data[MmOffPlayerForm]
	mm.Day = binary.BigEndian.Uint32(data[MmOffDay:])
	mm.Time = binary.BigEndian.Uint16(data[MmOffTime:])

	// Inventory
	copy(mm.Items[:], data[MmOffInvItems:MmOffInvItems+48])
	for i := 0; i < 24; i++ {
		mm.Ammo[i] = int8(data[MmOffInvAmmo+i])
	}

	mm.Equipment = binary.BigEndian.Uint16(data[MmOffEquipment:])
	mm.Upgrades = binary.BigEndian.Uint32(data[MmOffInvUpgrades:])
	mm.QuestItems = binary.BigEndian.Uint32(data[MmOffInvQuest:])
	mm.HeartPieces = uint8((mm.QuestItems >> 28) & 0xF)
	mm.SkullTokensSwamp = binary.BigEndian.Uint16(data[MmOffSkullSwamp:])
	mm.SkullTokensOcean = binary.BigEndian.Uint16(data[MmOffSkullOcean:])

	// Dungeon items
	copy(mm.DungeonItems[:], data[MmOffDungeonItems:MmOffDungeonItems+10])
	for i := 0; i < 9; i++ {
		mm.DungeonKeys[i] = int8(data[MmOffDungeonKeys+i])
	}
	for i := 0; i < 10; i++ {
		mm.StrayFairies[i] = int8(data[MmOffStrayFairies+i])
	}
	copy(mm.WeekEventReg[:], data[MmOffWeekEventReg:MmOffWeekEventReg+len(mm.WeekEventReg)])
	if MmOffWeekEventReg+mmWeekEventTownStrayFairyByte < len(data) {
		mm.TownStrayFairy = data[MmOffWeekEventReg+mmWeekEventTownStrayFairyByte]&mmWeekEventTownStrayFairyMask != 0
	}

	// Permanent scene flags
	permOff := MmOffPermScenes
	for i := 0; i < MmPermCount; i++ {
		off := permOff + i*MmPermEntrySize
		if off+MmPermEntrySize > len(data) {
			break
		}
		mm.SceneFlags[i] = SceneFlags{
			Chests:        binary.BigEndian.Uint32(data[off:]),
			Switches:      binary.BigEndian.Uint32(data[off+4:]),
			RoomClear:     binary.BigEndian.Uint32(data[off+8:]),
			Collectibles:  binary.BigEndian.Uint32(data[off+12:]),
			Unused:        binary.BigEndian.Uint32(data[off+16:]),
			VisitedRooms:  binary.BigEndian.Uint32(data[off+20:]),
			VisitedFloors: binary.BigEndian.Uint32(data[off+24:]),
		}
	}

	// Cycle scene flags (at MmSaveContext offset, not MmSave)
	cycleOff := MmCtxOffCycleFlags
	for i := 0; i < MmPermCount; i++ {
		off := cycleOff + i*0x14
		if off+0x14 > len(data) {
			break
		}
		mm.CycleFlags[i] = CycleSceneFlags{
			Chests:       binary.BigEndian.Uint32(data[off:]),
			Switch0:      binary.BigEndian.Uint32(data[off+4:]),
			Switch1:      binary.BigEndian.Uint32(data[off+8:]),
			ClearedRoom:  binary.BigEndian.Uint32(data[off+12:]),
			Collectibles: binary.BigEndian.Uint32(data[off+16:]),
		}
	}

	if len(data) >= MmCtxOffGameMode+4 {
		mm.GameMode = binary.BigEndian.Uint32(data[MmCtxOffGameMode:])
	} else {
		mm.GameMode = 0
	}

	return nil
}

func (r *Reader) readOotRuntimeConfig(activeGame ActiveGame, oot *OotState) {
	if oot == nil {
		return
	}

	r.readOotComboConfig(activeGame, oot)
	if activeGame != GameOot {
		return
	}

	var payload []byte
	if r.ootSilverDataAddr == 0 || r.ootMaxKeysAddr == 0 {
		data, err := r.mem.Read(AddrOotPayload, OotPayloadSize)
		if err == nil {
			payload = data
		}
	}

	if r.ootSilverDataAddr == 0 && len(payload) >= OotSilverRupeeDataSize {
		if off, ok := locateSilverRupeeData(payload); ok {
			r.ootSilverDataAddr = AddrOotPayload + uint32(off)
		}
	}
	if r.ootSilverDataAddr != 0 {
		if data, err := r.mem.Read(r.ootSilverDataAddr, OotSilverRupeeDataSize); err == nil && validateSilverRupeeData(data) {
			for index := 0; index < OotSilverRupeeSetCount; index++ {
				oot.RuntimeSilverRupeeCounts[index] = data[index*4+3]
			}
			oot.HasRuntimeSilverRupeeCounts = true
		}
	}

	if r.ootMaxKeysAddr == 0 && r.ootSilverDataAddr != 0 {
		candidate := r.ootSilverDataAddr + ootMaxKeysFromSilverDelta
		if data, err := r.mem.Read(candidate, OotMaxKeysBlockSize); err == nil && validateOotMaxKeyBlock(data) {
			r.ootMaxKeysAddr = candidate
		}
	}
	if r.ootMaxKeysAddr == 0 && len(payload) >= OotMaxKeysBlockSize {
		if off, ok := locateOotMaxKeys(payload); ok {
			r.ootMaxKeysAddr = AddrOotPayload + uint32(off)
		}
	}
	if r.ootMaxKeysAddr != 0 {
		if data, err := r.mem.Read(r.ootMaxKeysAddr, OotMaxKeysBlockSize); err == nil && validateOotMaxKeyBlock(data) {
			copy(oot.RuntimeMaxKeys[:], data[:OotRuntimeSceneCount])
			oot.HasRuntimeMaxKeys = true
		}
	}
}

func (r *Reader) readOotComboConfig(activeGame ActiveGame, oot *OotState) {
	if oot == nil {
		return
	}

	payloadAddr := uint32(0)
	payloadSize := 0
	cachedAddr := (*uint32)(nil)
	switch activeGame {
	case GameOot:
		payloadAddr = AddrOotPayload
		payloadSize = OotPayloadSize
		cachedAddr = &r.comboConfigOotAddr
	case GameMm:
		payloadAddr = AddrMmPayload
		payloadSize = MmPayloadSize
		cachedAddr = &r.comboConfigMmAddr
	default:
		return
	}

	if *cachedAddr != 0 {
		if data, err := r.mem.Read(*cachedAddr, OotComboConfigSize); err == nil && validateOotComboConfig(data) {
			oot.RuntimeMqBits = binary.BigEndian.Uint32(data[OotComboConfigMqOffset:])
			oot.HasRuntimeMqBits = true
			return
		}
		*cachedAddr = 0
	}

	payload, err := r.mem.Read(payloadAddr, payloadSize)
	if err != nil || len(payload) < OotComboConfigSize {
		return
	}

	off, ok := locateOotComboConfig(payload)
	if !ok {
		return
	}
	*cachedAddr = payloadAddr + uint32(off)
	oot.RuntimeMqBits = binary.BigEndian.Uint32(payload[off+OotComboConfigMqOffset:])
	oot.HasRuntimeMqBits = true
}

type ootPlayStateSample struct {
	sceneID        uint16
	actorTotal     uint8
	currentRoom    uint8
	linkAgeOnLoad  uint8
	gameplayFrames uint32
	chestFlags     uint32
	collectFlags   uint32
	tempCollect    uint32
}

var ootPlayStateCandidateAddrs = [...]uint32{
	AddrOotPlayStateNtsc10,
	AddrOotPlayStateNtsc11,
	AddrOotPlayStateNtsc12,
	AddrOotPlayStatePal10,
	AddrOotPlayStatePal11,
	AddrOotPlayStateDebug,
}

func (r *Reader) readOotLiveState(oot *OotState) {
	if oot == nil {
		return
	}

	sample, ok := r.readOotPlayStateSampleCached()
	if !ok {
		return
	}
	if sample.sceneID >= OotPermCount {
		return
	}

	oot.LiveSceneID = sample.sceneID
	oot.LiveChestFlags = sample.chestFlags
	oot.LiveCollectFlags = sample.collectFlags
	oot.LiveTempCollectFlag = sample.tempCollect
	oot.HasLiveSceneFlags = true
}

func (r *Reader) readOotPlayStateSampleCached() (ootPlayStateSample, bool) {
	if r.ootPlayStateAddr != 0 {
		sample, err := r.readOotPlayStateSample(r.ootPlayStateAddr)
		if err == nil && isPlausibleOotPlayStateSample(sample) {
			return sample, true
		}
		r.ootPlayStateAddr = 0
	}

	for _, addr := range ootPlayStateCandidateAddrs {
		sample, err := r.readOotPlayStateSample(addr)
		if err != nil || !isPlausibleOotPlayStateSample(sample) {
			continue
		}
		r.ootPlayStateAddr = addr
		return sample, true
	}

	return ootPlayStateSample{}, false
}

func (r *Reader) readOotPlayStateSample(addr uint32) (ootPlayStateSample, error) {
	sceneID, err := r.mem.ReadU16BE(addr + uint32(OotPlayOffSceneID))
	if err != nil {
		return ootPlayStateSample{}, err
	}
	actorTotal, err := r.mem.ReadU8(addr + uint32(OotPlayOffActorTotal))
	if err != nil {
		return ootPlayStateSample{}, err
	}
	currentRoom, err := r.mem.ReadU8(addr + uint32(OotPlayOffCurrentRoom))
	if err != nil {
		return ootPlayStateSample{}, err
	}
	linkAgeOnLoad, err := r.mem.ReadU8(addr + uint32(OotPlayOffLinkAgeOnLoad))
	if err != nil {
		return ootPlayStateSample{}, err
	}
	gameplayFrames, err := r.mem.ReadU32BE(addr + uint32(OotPlayOffGameplayFrames))
	if err != nil {
		return ootPlayStateSample{}, err
	}
	chestFlags, err := r.mem.ReadU32BE(addr + uint32(OotPlayOffChestFlags))
	if err != nil {
		return ootPlayStateSample{}, err
	}
	collectFlags, err := r.mem.ReadU32BE(addr + uint32(OotPlayOffCollectFlags))
	if err != nil {
		return ootPlayStateSample{}, err
	}
	tempCollect, err := r.mem.ReadU32BE(addr + uint32(OotPlayOffTempCollect))
	if err != nil {
		return ootPlayStateSample{}, err
	}

	return ootPlayStateSample{
		sceneID:        sceneID,
		actorTotal:     actorTotal,
		currentRoom:    currentRoom,
		linkAgeOnLoad:  linkAgeOnLoad,
		gameplayFrames: gameplayFrames,
		chestFlags:     chestFlags,
		collectFlags:   collectFlags,
		tempCollect:    tempCollect,
	}, nil
}

func isPlausibleOotPlayStateSample(sample ootPlayStateSample) bool {
	if sample.sceneID >= OotPermCount {
		return false
	}
	if sample.actorTotal == 0 || sample.actorTotal > 200 {
		return false
	}
	if sample.currentRoom >= 0x40 {
		return false
	}
	if sample.linkAgeOnLoad > 1 {
		return false
	}
	return sample.gameplayFrames > 0
}

type silverRupeeFlagCount struct {
	flag  byte
	count byte
}

var ootSilverRupeeAllowed = [...][2]silverRupeeFlagCount{
	{{0x00, 0x00}, {0x25, 0x05}},
	{{0x1f, 0x05}, {0x00, 0x00}},
	{{0x05, 0x05}, {0x37, 0x05}},
	{{0x0a, 0x05}, {0x00, 0x00}},
	{{0x02, 0x05}, {0x00, 0x00}},
	{{0x01, 0x05}, {0x01, 0x05}},
	{{0x00, 0x00}, {0x03, 0x0a}},
	{{0x09, 0x05}, {0x11, 0x05}},
	{{0x08, 0x05}, {0x08, 0x0a}},
	{{0x08, 0x05}, {0x00, 0x00}},
	{{0x09, 0x05}, {0x00, 0x00}},
	{{0x1c, 0x05}, {0x1c, 0x05}},
	{{0x0c, 0x05}, {0x0c, 0x06}},
	{{0x1b, 0x05}, {0x1b, 0x03}},
	{{0x0b, 0x05}, {0x0b, 0x05}},
	{{0x12, 0x05}, {0x02, 0x05}},
	{{0x09, 0x05}, {0x01, 0x05}},
	{{0x0e, 0x05}, {0x00, 0x00}},
}

func locateSilverRupeeData(payload []byte) (int, bool) {
	for off := 0; off+OotSilverRupeeDataSize <= len(payload); off++ {
		if validateSilverRupeeData(payload[off : off+OotSilverRupeeDataSize]) {
			return off, true
		}
	}
	return 0, false
}

func validateSilverRupeeData(data []byte) bool {
	if len(data) < OotSilverRupeeDataSize {
		return false
	}
	for index := 0; index < OotSilverRupeeSetCount; index++ {
		flag := data[index*4+2]
		count := data[index*4+3]
		allowed := ootSilverRupeeAllowed[index]
		if (flag != allowed[0].flag || count != allowed[0].count) && (flag != allowed[1].flag || count != allowed[1].count) {
			return false
		}
	}
	return sameSilverScene(data, 2, 3) && sameSilverScene(data, 3, 4) &&
		sameSilverScene(data, 5, 6) && sameSilverScene(data, 6, 7) && sameSilverScene(data, 7, 8) &&
		sameSilverScene(data, 9, 10) &&
		sameSilverScene(data, 11, 12) && sameSilverScene(data, 12, 13) &&
		sameSilverScene(data, 14, 15) && sameSilverScene(data, 15, 16) && sameSilverScene(data, 16, 17)
}

func sameSilverScene(data []byte, left, right int) bool {
	leftOff := left * 4
	rightOff := right * 4
	return data[leftOff] == data[rightOff] && data[leftOff+1] == data[rightOff+1]
}

func locateOotMaxKeys(payload []byte) (int, bool) {
	bestOff := -1
	bestScore := -1
	for off := 0; off+OotMaxKeysBlockSize <= len(payload); off++ {
		block := payload[off : off+OotMaxKeysBlockSize]
		if !validateOotMaxKeyBlock(block) {
			continue
		}
		score := maxKeyBlockScore(payload, off)
		if score <= 0 {
			continue
		}
		if score > bestScore {
			bestScore = score
			bestOff = off
		}
	}
	return bestOff, bestOff >= 0
}

func maxKeyBlockScore(payload []byte, off int) int {
	end := off + OotMaxKeysBlockSize
	score := 0
	for _, value := range payload[off:end] {
		score += int(value)
	}
	for index := off - 8; index < off; index++ {
		if index >= 0 && payload[index] != 0 {
			score++
		}
	}
	for index := end; index < end+12 && index < len(payload); index++ {
		if payload[index] != 0 {
			score++
		}
	}
	return score
}

func validateOotMaxKeyBlock(data []byte) bool {
	if len(data) < OotMaxKeysBlockSize {
		return false
	}
	for sceneID := 0; sceneID < OotRuntimeSceneCount; sceneID++ {
		if !ootMaxKeyValueAllowed(sceneID, data[sceneID]) {
			return false
		}
	}
	return mmMaxKeyValueAllowed(0, data[17]) &&
		mmMaxKeyValueAllowed(1, data[18]) &&
		mmMaxKeyValueAllowed(2, data[19]) &&
		mmMaxKeyValueAllowed(3, data[20])
}

func locateOotComboConfig(payload []byte) (int, bool) {
	bestOff := -1
	bestScore := -1
	for off := 0; off+OotComboConfigSize <= len(payload); off += 4 {
		block := payload[off : off+OotComboConfigSize]
		if !validateOotComboConfig(block) {
			continue
		}
		score := ootComboConfigScore(block)
		if score > bestScore {
			bestScore = score
			bestOff = off
		}
	}
	return bestOff, bestOff >= 0
}

func validateOotComboConfig(data []byte) bool {
	if len(data) < OotComboConfigSize {
		return false
	}
	if data[0] == 0 || data[1] != 0 || data[2] != 0 || data[3] != 0 {
		return false
	}

	mqBits := binary.BigEndian.Uint32(data[OotComboConfigMqOffset:])
	if mqBits&^uint32((1<<OotMqDungeonCount)-1) != 0 {
		return false
	}

	triforcePieces := binary.BigEndian.Uint16(data[OotComboConfigTriforcePiecesOffset:])
	triforceGoal := binary.BigEndian.Uint16(data[OotComboConfigTriforceGoalOffset:])
	if triforcePieces != 0 && triforceGoal > triforcePieces {
		return false
	}

	for index := 0; index < OotComboConfigSpecialCount; index++ {
		off := OotComboConfigSpecialOffset + index*OotComboConfigSpecialSize
		flags := binary.BigEndian.Uint32(data[off:])
		count := binary.BigEndian.Uint16(data[off+4:])
		zero := binary.BigEndian.Uint16(data[off+6:])
		if zero != 0 || flags>>19 != 0 || count > 0x400 {
			return false
		}
	}

	for index := 0; index < OotComboConfigPriceCount; index++ {
		off := OotComboConfigPricesOffset + index*2
		price := binary.BigEndian.Uint16(data[off:])
		if price > OotComboConfigPriceMax || price%5 != 0 {
			return false
		}
	}

	for index := 0; index < OotComboConfigStaticHintCount; index++ {
		value := int8(data[OotComboConfigStaticHintsOffset+index])
		if value < -1 || value > 3 {
			return false
		}
	}

	var seenBosses [OotComboConfigBossCount]bool
	for index := 0; index < OotComboConfigBossCount; index++ {
		bossID := int(data[OotComboConfigBossOffset+index])
		if bossID < 0 || bossID >= OotComboConfigBossCount || seenBosses[bossID] {
			return false
		}
		seenBosses[bossID] = true
	}

	if data[OotComboConfigStrayFairyRewardCountOffset] > 15 {
		return false
	}
	if data[OotComboConfigBombchuBehaviorOotOffset] > 3 || data[OotComboConfigBombchuBehaviorMmOffset] > 3 {
		return false
	}
	for index := 0; index < OotComboConfigSongEventCount; index++ {
		if data[OotComboConfigSongEventsOffset+index] > 5 {
			return false
		}
	}

	return true
}

func ootComboConfigScore(data []byte) int {
	score := int(data[0])
	for off := 4; off < OotComboConfigPricesOffset; off += 4 {
		if binary.BigEndian.Uint32(data[off:]) != 0 {
			score++
		}
	}
	for index := 0; index < OotComboConfigPriceCount; index++ {
		if binary.BigEndian.Uint16(data[OotComboConfigPricesOffset+index*2:]) != 0 {
			score++
		}
	}
	return score
}

func ootMaxKeyValueAllowed(sceneID int, value byte) bool {
	switch sceneID {
	case OotSceneDekuTree, OotSceneDodongosCavern, OotSceneInsideJabuJabu, OotSceneIceCavern, OotSceneGanonTower, OotSceneUnused14, OotSceneUnused15:
		return value == 0
	case OotSceneTempleForest:
		return value == 0 || value == 5 || value == 6
	case OotSceneTempleFire:
		return value == 0 || value == 5 || value == 7 || value == 8
	case OotSceneTempleWater:
		return value == 0 || value == 2 || value == 5
	case OotSceneTempleSpirit:
		return value == 0 || value == 5 || value == 7
	case OotSceneTempleShadow:
		return value == 0 || value == 5 || value == 6
	case OotSceneBottomOfTheWell:
		return value == 0 || value == 2 || value == 3
	case OotSceneGerudoTrainingGround:
		return value == 0 || value == 3 || value == 9
	case OotSceneThievesHideout:
		return value == 0 || value == 1 || value == 4
	case OotSceneInsideGanonCastle:
		return value == 0 || value == 2 || value == 3
	case OotSceneTreasureShop:
		return value == 0 || value == 6
	default:
		return false
	}
}

func mmMaxKeyValueAllowed(dungeonIndex int, value byte) bool {
	switch dungeonIndex {
	case 0, 2:
		return value == 0 || value == 1
	case 1:
		return value == 0 || value == 3
	case 3:
		return value == 0 || value == 4
	default:
		return false
	}
}

func (r *Reader) readActiveSaveIndex(game ActiveGame) (uint32, error) {
	switch game {
	case GameOot:
		return r.mem.ReadU32BE(AddrOotSaveCtx + uint32(OotCtxOffFileNum))
	case GameMm:
		return r.mem.ReadU32BE(AddrMmSaveCtx + uint32(MmCtxOffFileNum))
	default:
		return 0, fmt.Errorf("no active game")
	}
}

func (r *Reader) readForeignOotSave(oot *OotState) error {
	if r.foreignOotSaveAddr != 0 {
		if err := r.readOotSaveAt(r.foreignOotSaveAddr, oot); err == nil {
			return nil
		}
		r.foreignOotSaveAddr = 0
	}

	addr, err := r.findForeignOotSaveAddr()
	if err != nil {
		return err
	}

	return r.readOotSaveAt(addr, oot)
}

func (r *Reader) readOotSaveAt(addr uint32, oot *OotState) error {
	data, err := r.mem.Read(addr, OotSaveSize)
	if err != nil {
		return fmt.Errorf("read foreign OoT save: %w", err)
	}
	if err := validateOotSave(data); err != nil {
		return fmt.Errorf("validate foreign OoT save: %w", err)
	}

	return parseOotSave(oot, data)
}

func (r *Reader) readForeignMmSave(mm *MmState) error {
	if r.foreignMmSaveAddr != 0 {
		if err := r.readMmSaveAt(r.foreignMmSaveAddr, mm); err == nil {
			return nil
		}
		r.foreignMmSaveAddr = 0
	}

	addr, err := r.findForeignMmSaveAddr()
	if err != nil {
		return err
	}

	return r.readMmSaveAt(addr, mm)
}

func (r *Reader) readMmSaveAt(addr uint32, mm *MmState) error {
	data, err := r.mem.Read(addr, MmSaveSize)
	if err != nil {
		return fmt.Errorf("read foreign MM save: %w", err)
	}
	if err := validateForeignMmSave(data); err != nil {
		return fmt.Errorf("validate foreign MM save: %w", err)
	}

	return parseMmSave(mm, data)
}

func (r *Reader) readForeignOotState(oot *OotState) error {
	if err := r.readForeignOotSave(oot); err == nil {
		return nil
	}

	if r.hasLastKnownOot {
		*oot = r.lastKnownOot
		return nil
	}

	resetEmptyOotState(oot)

	return nil
}

func (r *Reader) readForeignMmState(mm *MmState) error {
	if err := r.readForeignMmSave(mm); err == nil {
		return nil
	}

	if r.hasLastKnownMm {
		*mm = r.lastKnownMm
		return nil
	}

	resetEmptyMmState(mm)

	return nil
}

func (r *Reader) readSharedState(game ActiveGame, saveIndex uint32, shared *SharedCustomState) error {
	candidates := make([]sharedStateCandidate, 0, 3)
	var nearForeignChecks *SharedCustomState
	if liveChecks, err := r.readSharedCheckStateNearForeign(game); err == nil {
		liveChecksCopy := liveChecks.Clone()
		nearForeignChecks = &liveChecksCopy
	}
	if candidate, err := r.readSharedStateNearForeignSaveCandidate(game); err == nil {
		candidates = append(candidates, candidate)
	}
	if candidate, err := r.readSharedStateFromPayloadCandidate(AddrOotPayload, OotPayloadSize, saveIndex); err == nil {
		candidates = append(candidates, candidate)
	}
	if candidate, err := r.readSharedStateFromPayloadCandidate(AddrMmPayload, MmPayloadSize, saveIndex); err == nil {
		candidates = append(candidates, candidate)
	}

	if candidate, ok := r.selectSharedStateCandidate(candidates); ok {
		*shared = candidate.state
		overlaySharedCheckBitmaps(shared, nearForeignChecks)
		return nil
	}

	if r.hasLastKnownShared {
		*shared = r.lastKnownShared.Clone()
	}
	overlaySharedCheckBitmaps(shared, nearForeignChecks)

	return nil
}

func overlaySharedCheckBitmaps(dst *SharedCustomState, src *SharedCustomState) {
	if dst == nil || src == nil {
		return
	}

	for _, name := range sharedCheckBitmapNames {
		srcBitmap := src.Bitmap(name)
		if len(srcBitmap) == 0 {
			continue
		}

		dstBitmap := append([]uint8(nil), dst.Bitmap(name)...)
		if len(dstBitmap) < len(srcBitmap) {
			grown := make([]uint8, len(srcBitmap))
			copy(grown, dstBitmap)
			dstBitmap = grown
		}
		for index, value := range srcBitmap {
			dstBitmap[index] |= value
		}
		dst.SetBitmap(name, dstBitmap)
	}
}

func (r *Reader) readSharedStateNearForeignRaw(game ActiveGame) ([]byte, uint32, error) {
	var foreignSaveAddr uint32
	var payloadBase uint32
	var payloadSize int

	switch game {
	case GameOot:
		foreignSaveAddr = r.foreignMmSaveAddr
		payloadBase = AddrOotPayload
		payloadSize = OotPayloadSize
	case GameMm:
		foreignSaveAddr = r.foreignOotSaveAddr
		payloadBase = AddrMmPayload
		payloadSize = MmPayloadSize
	default:
		return nil, 0, fmt.Errorf("no active game")
	}

	if foreignSaveAddr == 0 {
		return nil, 0, fmt.Errorf("foreign save address unavailable")
	}
	if foreignSaveAddr < payloadBase+SharedCustomSaveSize {
		return nil, 0, fmt.Errorf("foreign save %#x is before shared custom save window", foreignSaveAddr)
	}

	addr := foreignSaveAddr - SharedCustomSaveSize
	readSize := sharedStateReadSize()
	end := uint64(addr-payloadBase) + uint64(readSize)
	if end > uint64(payloadSize) {
		return nil, 0, fmt.Errorf("shared custom save near %#x exceeds payload bounds", foreignSaveAddr)
	}

	data, err := r.mem.Read(addr, readSize)
	if err != nil {
		return nil, 0, fmt.Errorf("read shared custom save near %#x: %w", foreignSaveAddr, err)
	}

	return data, foreignSaveAddr, nil
}

func (r *Reader) readSharedStateNearForeignSave(game ActiveGame, shared *SharedCustomState) error {
	data, foreignSaveAddr, err := r.readSharedStateNearForeignRaw(game)
	if err != nil {
		return err
	}

	parsed, err := parseSharedState(data)
	if err != nil {
		return fmt.Errorf("parse shared custom save near %#x: %w", foreignSaveAddr, err)
	}

	*shared = parsed
	return nil
}

func (r *Reader) readSharedStateNearForeignSaveCandidate(game ActiveGame) (sharedStateCandidate, error) {
	var shared SharedCustomState
	if err := r.readSharedStateNearForeignSave(game, &shared); err != nil {
		return sharedStateCandidate{}, err
	}
	return sharedStateCandidate{source: "near-foreign", state: shared}, nil
}

func (r *Reader) readSharedCheckStateNearForeign(game ActiveGame) (SharedCustomState, error) {
	data, foreignSaveAddr, err := r.readSharedStateNearForeignRaw(game)
	if err != nil {
		return SharedCustomState{}, err
	}
	shared, err := parseSharedCheckState(data)
	if err != nil {
		return SharedCustomState{}, fmt.Errorf("parse shared custom save near %#x for checks: %w", foreignSaveAddr, err)
	}
	return shared, nil
}

func (r *Reader) readSharedStateFromPayload(payloadBase uint32, payloadSize int, saveIndex uint32, shared *SharedCustomState) error {
	addr, err := sharedSaveAddr(payloadBase, payloadSize, saveIndex)
	if err != nil {
		return err
	}

	data, err := r.mem.Read(addr, sharedStateReadSize())
	if err != nil {
		return fmt.Errorf("read shared custom save from %#x: %w", payloadBase, err)
	}

	parsed, err := parseSharedState(data)
	if err != nil {
		return fmt.Errorf("parse shared custom save at %#x: %w", addr, err)
	}

	*shared = parsed
	return nil
}

func (r *Reader) readSharedStateFromPayloadCandidate(payloadBase uint32, payloadSize int, saveIndex uint32) (sharedStateCandidate, error) {
	var shared SharedCustomState
	if err := r.readSharedStateFromPayload(payloadBase, payloadSize, saveIndex, &shared); err != nil {
		return sharedStateCandidate{}, err
	}
	return sharedStateCandidate{source: fmt.Sprintf("payload-%08x", payloadBase), state: shared}, nil
}

func (r *Reader) selectSharedStateCandidate(candidates []sharedStateCandidate) (sharedStateCandidate, bool) {
	if len(candidates) == 0 {
		return sharedStateCandidate{}, false
	}

	best := candidates[0]
	bestScore := r.scoreSharedStateCandidate(best)
	for _, candidate := range candidates[1:] {
		score := r.scoreSharedStateCandidate(candidate)
		if score > bestScore || (score == bestScore && sharedCandidateSourceRank(candidate.source) < sharedCandidateSourceRank(best.source)) {
			best = candidate
			bestScore = score
		}
	}

	return best, true
}

func (r *Reader) scoreSharedStateCandidate(candidate sharedStateCandidate) int {
	score := sharedCheckBitmapScore(candidate.state)
	if r.hasLastKnownShared {
		score += sharedStateContinuityScore(candidate.state, r.lastKnownShared)
	}
	return score + sharedCandidateSourceBonus(candidate.source)
}

func sharedCheckBitmapScore(shared SharedCustomState) int {
	score := 0
	for _, name := range sharedCheckBitmapNames {
		for _, value := range shared.Bitmap(name) {
			score += bits.OnesCount8(value)
		}
	}
	return score * 8
}

func sharedStateContinuityScore(current SharedCustomState, last SharedCustomState) int {
	score := 0
	for _, name := range sharedCheckBitmapNames {
		currentBitmap := current.Bitmap(name)
		lastBitmap := last.Bitmap(name)
		limit := len(currentBitmap)
		if len(lastBitmap) > limit {
			limit = len(lastBitmap)
		}
		for i := 0; i < limit; i++ {
			var currentByte uint8
			if i < len(currentBitmap) {
				currentByte = currentBitmap[i]
			}
			var lastByte uint8
			if i < len(lastBitmap) {
				lastByte = lastBitmap[i]
			}

			overlap := currentByte & lastByte
			newBits := currentByte &^ lastByte
			missingBits := lastByte &^ currentByte

			score += bits.OnesCount8(overlap) * 32
			score += bits.OnesCount8(newBits) * 4
			score -= bits.OnesCount8(missingBits) * 256
		}
	}
	return score
}

func sharedCandidateSourceBonus(source string) int {
	switch source {
	case "payload-80400000", "payload-80730000":
		return 2
	case "near-foreign":
		return 0
	default:
		return 1
	}
}

func sharedCandidateSourceRank(source string) int {
	switch source {
	case "payload-80400000":
		return 0
	case "payload-80730000":
		return 1
	case "near-foreign":
		return 2
	default:
		return 3
	}
}

func sharedStateReadSize() int {
	readSize := int(SharedCustomSaveSize)
	if readSize < sharedStorage.TrackedSize {
		readSize = sharedStorage.TrackedSize
	}
	if readSize < sharedBombchuBagFlagsOffset+1 {
		readSize = sharedBombchuBagFlagsOffset + 1
	}
	return readSize
}

func parseSharedState(data []byte) (SharedCustomState, error) {
	parsed, err := parseSharedStateUnchecked(data)
	if err != nil {
		return SharedCustomState{}, err
	}
	if !isPlausibleSharedState(parsed) {
		return SharedCustomState{}, fmt.Errorf("shared custom save failed plausibility checks")
	}
	return parsed, nil
}

func parseSharedStateUnchecked(data []byte) (SharedCustomState, error) {
	parsed := SharedCustomState{}
	for _, bitmap := range sharedStorage.Bitmaps {
		end := bitmap.Offset + bitmap.Size
		if bitmap.Offset < 0 || end > len(data) {
			return SharedCustomState{}, fmt.Errorf("shared bitmap %s out of bounds", bitmap.Name)
		}
		parsed.SetBitmap(bitmap.Name, data[bitmap.Offset:end])
	}
	if len(data) >= sharedOcarinaButtonMaskMmOffset+2 {
		parsed.OcarinaButtonMaskOot = binary.BigEndian.Uint16(data[sharedOcarinaButtonMaskOotOffset:])
		parsed.OcarinaButtonMaskMm = binary.BigEndian.Uint16(data[sharedOcarinaButtonMaskMmOffset:])
	}
	if len(data) >= sharedCaughtChildFishWeightOffset+sharedCaughtFishWeightCount {
		copy(parsed.CaughtChildFishWeights[:], data[sharedCaughtChildFishWeightOffset:sharedCaughtChildFishWeightOffset+sharedCaughtFishWeightCount])
	}
	if len(data) >= sharedCaughtAdultFishWeightOffset+sharedCaughtFishWeightCount {
		copy(parsed.CaughtAdultFishWeights[:], data[sharedCaughtAdultFishWeightOffset:sharedCaughtAdultFishWeightOffset+sharedCaughtFishWeightCount])
	}
	if len(data) > sharedBombchuBagFlagsOffset {
		flags := data[sharedBombchuBagFlagsOffset]
		parsed.BombchuBagOot = uint8((flags >> sharedBombchuBagOotShift) & sharedBombchuBagMask)
		parsed.BombchuBagMm = uint8((flags >> sharedBombchuBagMmShift) & sharedBombchuBagMask)
	}
	return parsed, nil
}

func parseSharedCheckState(data []byte) (SharedCustomState, error) {
	parsed, err := parseSharedStateUnchecked(data)
	if err != nil {
		return SharedCustomState{}, err
	}

	filtered := SharedCustomState{}
	for _, name := range sharedCheckBitmapNames {
		bitmap := parsed.Bitmap(name)
		if len(bitmap) == 0 {
			continue
		}
		filtered.SetBitmap(name, bitmap)
	}
	return filtered, nil
}

func (r *Reader) rememberOotState(oot OotState) {
	oot.LiveSceneID = 0
	oot.LiveChestFlags = 0
	oot.LiveCollectFlags = 0
	oot.LiveTempCollectFlag = 0
	oot.HasLiveSceneFlags = false
	r.lastKnownOot = oot
	r.hasLastKnownOot = true
}

func (r *Reader) rememberMmState(mm MmState) {
	r.lastKnownMm = mm
	r.hasLastKnownMm = true
}

func (r *Reader) rememberSharedState(shared SharedCustomState) {
	r.lastKnownShared = shared.Clone()
	r.hasLastKnownShared = true
}

func (r *Reader) resetPendingState() {
	r.hasPendingState = false
}

func (r *Reader) acceptStableState(game ActiveGame, saveIndex uint32) bool {
	if game == r.stableGame && saveIndex == r.stableSaveIndex {
		r.resetPendingState()
		return true
	}

	if r.hasPendingState && game == r.pendingGame && saveIndex == r.pendingSaveIndex {
		r.stableGame = game
		r.stableSaveIndex = saveIndex
		r.resetPendingState()
		return true
	}

	r.pendingGame = game
	r.pendingSaveIndex = saveIndex
	r.hasPendingState = true
	return false
}

func foreignOotSaveAddr(saveIndex uint32) (uint32, error) {
	return foreignSaveAddr(AddrMmPayload, MmPayloadSize, ForeignOotSaveBaseOff, ForeignOotSaveStride, OotSaveSize, saveIndex)
}

func (r *Reader) findForeignOotSaveAddr() (uint32, error) {
	if r.foreignOotSaveAddr != 0 {
		return r.foreignOotSaveAddr, nil
	}

	payload, err := r.mem.Read(AddrMmPayload, MmPayloadSize)
	if err != nil {
		return 0, fmt.Errorf("read MM payload: %w", err)
	}

	addr, ok := locateForeignOotSave(payload, AddrMmPayload)
	if !ok {
		return 0, fmt.Errorf("foreign OoT save not found in MM payload")
	}

	r.foreignOotSaveAddr = addr
	return addr, nil
}

func (r *Reader) findForeignMmSaveAddr() (uint32, error) {
	if r.foreignMmSaveAddr != 0 {
		return r.foreignMmSaveAddr, nil
	}

	payload, err := r.mem.Read(AddrOotPayload, OotPayloadSize)
	if err != nil {
		return 0, fmt.Errorf("read OoT payload: %w", err)
	}

	addr, ok := locateForeignMmSave(payload, AddrOotPayload)
	if !ok {
		return 0, fmt.Errorf("foreign MM save not found in OoT payload")
	}

	r.foreignMmSaveAddr = addr
	return addr, nil
}

func foreignMmSaveAddr(saveIndex uint32) (uint32, error) {
	return foreignSaveAddr(AddrOotPayload, OotPayloadSize, ForeignMmSaveBaseOff, ForeignMmSaveStride, MmSaveSize, saveIndex)
}

func sharedSaveAddr(payloadBase uint32, payloadSize int, saveIndex uint32) (uint32, error) {
	return foreignSaveAddr(payloadBase, payloadSize, sharedStorage.BaseOffset, sharedStorage.Stride, sharedStateReadSize(), saveIndex)
}

func foreignSaveAddr(payloadBase uint32, payloadSize int, baseOffset, stride uint32, saveSize int, saveIndex uint32) (uint32, error) {
	offset := uint64(baseOffset) + uint64(stride)*uint64(saveIndex)
	end := offset + uint64(saveSize)
	if end > uint64(payloadSize) {
		return 0, fmt.Errorf("foreign save index %d out of range for payload %#x", saveIndex, payloadBase)
	}
	return payloadBase + uint32(offset), nil
}

func locateForeignOotSave(payload []byte, payloadBase uint32) (uint32, bool) {
	for offset := 0; offset+OotSaveSize <= len(payload); offset += 16 {
		candidate := payload[offset : offset+OotSaveSize]
		expected := binary.BigEndian.Uint16(candidate[OotOffChecksum:])
		if expected == 0 {
			continue
		}
		if err := validateOotSave(candidate); err != nil {
			continue
		}
		if !isPlausibleOotSave(candidate) {
			continue
		}
		return payloadBase + uint32(offset), true
	}

	return 0, false
}

func isPlausibleOotSave(data []byte) bool {
	age := binary.BigEndian.Uint32(data[OotOffAge:])
	if age > 1 {
		return false
	}

	emptySlots := 0
	for _, itemID := range data[OotOffInvItems : OotOffInvItems+24] {
		if itemID == emptyInventoryItem {
			emptySlots++
		}
	}

	return emptySlots >= 8
}

func locateForeignMmSave(payload []byte, payloadBase uint32) (uint32, bool) {
	for offset := 0; offset+MmSaveSize <= len(payload); offset += 16 {
		candidate := payload[offset : offset+MmSaveSize]
		if err := validateMmSave(candidate); err != nil {
			continue
		}
		if !isPlausibleMmSave(candidate) {
			continue
		}
		return payloadBase + uint32(offset), true
	}

	bestOffset := -1
	bestDelta := maxForeignMmChecksumDelta + 1
	for offset := 0; offset+MmSaveSize <= len(payload); offset += 16 {
		candidate := payload[offset : offset+MmSaveSize]
		delta, ok := mmChecksumDelta(candidate)
		if !ok || delta > maxForeignMmChecksumDelta {
			continue
		}
		if !isPlausibleMmSave(candidate) {
			continue
		}
		if delta < bestDelta {
			bestDelta = delta
			bestOffset = offset
		}
	}
	if bestOffset >= 0 {
		return payloadBase + uint32(bestOffset), true
	}

	return 0, false
}

func isPlausibleMmSave(data []byte) bool {
	playerForm := data[MmOffPlayerForm]
	if playerForm > 4 {
		return false
	}

	day := binary.BigEndian.Uint32(data[MmOffDay:])
	if day > 4 {
		return false
	}

	emptySlots := 0
	for _, itemID := range data[MmOffInvItems : MmOffInvItems+48] {
		if itemID == emptyInventoryItem {
			emptySlots++
		}
	}
	if emptySlots < 16 {
		return false
	}
	for i := 0; i < 9; i++ {
		keys := int8(data[MmOffDungeonKeys+i])
		if keys < -1 || keys > 9 {
			return false
		}
	}
	for i := 0; i < 10; i++ {
		fairies := int8(data[MmOffStrayFairies+i])
		if fairies < 0 || fairies > 15 {
			return false
		}
	}

	return true
}

func isPlausibleSharedState(shared SharedCustomState) bool {
	for name, bitmapInfo := range sharedBitmaps {
		if !sharedBitmapHasNoUnusedBits(shared.Bitmap(name), sharedBitmapUsedBits[name], bitmapInfo.Size) {
			return false
		}
	}
	return true
}

func mmChecksumDelta(data []byte) (int, bool) {
	if len(data) < MmSaveSize {
		return 0, false
	}

	expected := binary.BigEndian.Uint16(data[MmOffChecksum:])
	if expected == 0 {
		return 0, false
	}

	checksum := mmChecksum(data)
	delta := int(checksum) - int(expected)
	if delta < 0 {
		delta = -delta
	}
	if delta > 0x8000 {
		delta = 0x10000 - delta
	}
	return delta, true
}

func mmChecksum(data []byte) uint16 {
	checksum := uint16(0)
	for i := 0; i < MmSaveSize; i++ {
		if i == MmOffChecksum || i == MmOffChecksum+1 {
			continue
		}
		checksum += uint16(data[i])
	}
	return checksum
}

func resetEmptyOotState(oot *OotState) {
	*oot = OotState{}
	for i := range oot.Items {
		oot.Items[i] = emptyInventoryItem
	}
	for i := range oot.DungeonKeys {
		oot.DungeonKeys[i] = -1
	}
}

func resetEmptyMmState(mm *MmState) {
	*mm = MmState{}
	for i := range mm.Items {
		mm.Items[i] = emptyInventoryItem
	}
	for i := range mm.DungeonKeys {
		mm.DungeonKeys[i] = -1
	}
}

func sharedBitmapHasNoUnusedBits(bitmap []uint8, usedBits, expectedSize int) bool {
	if len(bitmap) != expectedSize {
		return false
	}
	if usedBits < 0 || usedBits > len(bitmap)*8 {
		return false
	}

	fullBytes := usedBits / 8
	remainingBits := usedBits % 8
	if remainingBits != 0 && fullBytes < len(bitmap) {
		mask := uint8(0xFF << uint(remainingBits))
		if bitmap[fullBytes]&mask != 0 {
			return false
		}
		fullBytes++
	}

	for _, value := range bitmap[fullBytes:] {
		if value != 0 {
			return false
		}
	}

	return true
}

func validateOotSave(data []byte) error {
	if len(data) < OotSaveSize {
		return fmt.Errorf("OoT save too small: got %#x bytes", len(data))
	}

	expected := binary.BigEndian.Uint16(data[OotOffChecksum:])
	checksum := uint16(0)
	for i := 0; i < OotSaveSize; i += 2 {
		if i == OotOffChecksum {
			continue
		}
		checksum += binary.BigEndian.Uint16(data[i:])
	}
	if checksum != expected {
		return fmt.Errorf("invalid OoT checksum: got %04x want %04x", expected, checksum)
	}

	return nil
}

func validateMmSave(data []byte) error {
	if len(data) < MmSaveSize {
		return fmt.Errorf("MM save too small: got %#x bytes", len(data))
	}

	expected := binary.BigEndian.Uint16(data[MmOffChecksum:])
	checksum := mmChecksum(data)
	if checksum != expected {
		return fmt.Errorf("invalid MM checksum: got %04x want %04x", expected, checksum)
	}

	return nil
}

func validateForeignMmSave(data []byte) error {
	if err := validateMmSave(data); err == nil {
		return nil
	}

	delta, ok := mmChecksumDelta(data)
	if !ok {
		return fmt.Errorf("invalid MM checksum")
	}
	if delta > maxForeignMmChecksumDelta {
		return fmt.Errorf("invalid MM checksum delta %d", delta)
	}
	if !isPlausibleMmSave(data) {
		return fmt.Errorf("MM save failed plausibility checks")
	}

	return nil
}
