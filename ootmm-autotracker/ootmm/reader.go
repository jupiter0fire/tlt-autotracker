package ootmm

import (
	"encoding/binary"
	"fmt"

	"github.com/ootmm-autotracker/n64"
)

const maxForeignMmChecksumDelta = 0x400

// Reader reads OoTMM game state from N64 RDRAM.
type Reader struct {
	mem      *n64.Memory
	detector *Detector

	foreignOotSaveAddr uint32
	foreignMmSaveAddr  uint32
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

	// Event flags — these are within OotSaveInfo which starts at OotSave+0x20
	// ASSERT_OFFSET(OotSave, info.eventsMisc, 0xef8) — this is from OotSave start
	eventsChkOff := 0xED8 // eventsChk offset from OotSave start
	for i := 0; i < 14; i++ {
		oot.EventsChk[i] = binary.BigEndian.Uint16(data[eventsChkOff+i*2:])
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
	if err := r.readSharedStateNearForeignSave(game, shared); err == nil {
		return nil
	}

	if err := r.readSharedStateFromPayload(AddrOotPayload, OotPayloadSize, saveIndex, shared); err == nil {
		return nil
	}

	if err := r.readSharedStateFromPayload(AddrMmPayload, MmPayloadSize, saveIndex, shared); err == nil {
		return nil
	}

	if r.hasLastKnownShared {
		*shared = r.lastKnownShared.Clone()
	}

	return nil
}

func (r *Reader) readSharedStateNearForeignSave(game ActiveGame, shared *SharedCustomState) error {
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
		return fmt.Errorf("no active game")
	}

	if foreignSaveAddr == 0 {
		return fmt.Errorf("foreign save address unavailable")
	}
	if foreignSaveAddr < payloadBase+SharedCustomSaveSize {
		return fmt.Errorf("foreign save %#x is before shared custom save window", foreignSaveAddr)
	}

	addr := foreignSaveAddr - SharedCustomSaveSize
	end := uint64(addr-payloadBase) + uint64(sharedStorage.TrackedSize)
	if end > uint64(payloadSize) {
		return fmt.Errorf("shared custom save near %#x exceeds payload bounds", foreignSaveAddr)
	}

	data, err := r.mem.Read(addr, sharedStorage.TrackedSize)
	if err != nil {
		return fmt.Errorf("read shared custom save near %#x: %w", foreignSaveAddr, err)
	}

	parsed := SharedCustomState{}
	for _, bitmap := range sharedStorage.Bitmaps {
		end := bitmap.Offset + bitmap.Size
		if bitmap.Offset < 0 || end > len(data) {
			return fmt.Errorf("shared bitmap %s out of bounds near foreign save", bitmap.Name)
		}
		parsed.SetBitmap(bitmap.Name, data[bitmap.Offset:end])
	}

	if !isPlausibleSharedState(parsed) {
		return fmt.Errorf("shared custom save near %#x failed plausibility checks", foreignSaveAddr)
	}

	*shared = parsed
	return nil
}

func (r *Reader) readSharedStateFromPayload(payloadBase uint32, payloadSize int, saveIndex uint32, shared *SharedCustomState) error {
	addr, err := sharedSaveAddr(payloadBase, payloadSize, saveIndex)
	if err != nil {
		return err
	}

	data, err := r.mem.Read(addr, sharedStorage.TrackedSize)
	if err != nil {
		return fmt.Errorf("read shared custom save from %#x: %w", payloadBase, err)
	}

	parsed := SharedCustomState{}
	for _, bitmap := range sharedStorage.Bitmaps {
		end := bitmap.Offset + bitmap.Size
		if bitmap.Offset < 0 || end > len(data) {
			return fmt.Errorf("shared bitmap %s out of bounds in payload copy", bitmap.Name)
		}
		parsed.SetBitmap(bitmap.Name, data[bitmap.Offset:end])
	}

	if !isPlausibleSharedState(parsed) {
		return fmt.Errorf("shared custom save at %#x failed plausibility checks", addr)
	}

	*shared = parsed
	return nil
}

func (r *Reader) rememberOotState(oot OotState) {
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
	return foreignSaveAddr(payloadBase, payloadSize, sharedStorage.BaseOffset, sharedStorage.Stride, sharedStorage.TrackedSize, saveIndex)
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
