package ootmm

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/ootmm-autotracker/n64"
)

// Reader reads OoTMM game state from N64 RDRAM.
type Reader struct {
	mem      *n64.Memory
	detector *Detector

	foreignOotSaveAddr uint32
	foreignMmSaveAddr  uint32
	lastKnownOot       OotState
	lastKnownMm        MmState
	hasLastKnownOot    bool
	hasLastKnownMm     bool
}

func NewReader(mem *n64.Memory) *Reader {
	return &Reader{
		mem:      mem,
		detector: NewDetector(mem),
	}
}

// ReadState performs a full read of the current game state.
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
		return state, nil
	}

	// Detect active game
	game, err := r.detector.DetectActiveGame()
	if err != nil {
		return nil, fmt.Errorf("detect game: %w", err)
	}
	state.ActiveGame = game

	// Read OoT save (always in memory, even when MM is active)
	switch state.ActiveGame {
	case GameOot:
		if err := r.readOotSave(&state.Oot); err != nil {
			return nil, fmt.Errorf("read OoT save: %w", err)
		}
		r.rememberOotState(state.Oot)
		if err := r.readForeignMmState(&state.Mm); err != nil {
			return nil, fmt.Errorf("read foreign MM save: %w", err)
		}
	case GameMm:
		if err := r.readMmSave(&state.Mm); err != nil {
			return nil, fmt.Errorf("read MM save: %w", err)
		}
		r.rememberMmState(state.Mm)
		if err := r.readForeignOotState(&state.Oot); err != nil {
			return nil, fmt.Errorf("read foreign OoT save: %w", err)
		}
	default:
		if err := r.readOotSave(&state.Oot); err != nil {
			return nil, fmt.Errorf("read OoT save: %w", err)
		}
		r.rememberOotState(state.Oot)
		if err := r.readMmSave(&state.Mm); err != nil {
			return nil, fmt.Errorf("read MM save: %w", err)
		}
		r.rememberMmState(state.Mm)
	}

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

	// Dungeon items
	copy(mm.DungeonItems[:], data[MmOffDungeonItems:MmOffDungeonItems+10])
	for i := 0; i < 9; i++ {
		mm.DungeonKeys[i] = int8(data[MmOffDungeonKeys+i])
	}
	for i := 0; i < 10; i++ {
		mm.StrayFairies[i] = int8(data[MmOffStrayFairies+i])
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

func (r *Reader) readForeignOotSave(oot *OotState) error {
	addr, err := r.findForeignOotSaveAddr()
	if err != nil {
		return err
	}

	data, err := r.mem.Read(addr, OotSaveSize)
	if err != nil {
		return fmt.Errorf("read foreign OoT save: %w", err)
	}

	return parseOotSave(oot, data)
}

func (r *Reader) readForeignMmSave(mm *MmState) error {
	addr, err := r.findForeignMmSaveAddr()
	if err != nil {
		return err
	}

	data, err := r.mem.Read(addr, MmSaveSize)
	if err != nil {
		return fmt.Errorf("read foreign MM save: %w", err)
	}

	return parseMmSave(mm, data)
}

func (r *Reader) readForeignOotState(oot *OotState) error {
	if err := r.readForeignOotSave(oot); err == nil {
		r.rememberOotState(*oot)
		return nil
	}

	if r.hasLastKnownOot {
		*oot = r.lastKnownOot
		return nil
	}

	return nil
}

func (r *Reader) readForeignMmState(mm *MmState) error {
	if err := r.readForeignMmSave(mm); err == nil {
		r.rememberMmState(*mm)
		return nil
	}

	if r.hasLastKnownMm {
		*mm = r.lastKnownMm
		return nil
	}

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

func (r *Reader) findForeignOotSaveAddr() (uint32, error) {
	if r.foreignOotSaveAddr != 0 {
		return r.foreignOotSaveAddr, nil
	}

	payload, err := r.mem.Read(AddrMmPayload, MmPayloadSize)
	if err != nil {
		return 0, fmt.Errorf("read MM payload: %w", err)
	}

	addr, ok := locateForeignSave(payload, AddrMmPayload, OotSaveSize, 0x20)
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

	addr, ok := locateForeignSave(payload, AddrOotPayload, MmSaveSize, 0x24)
	if !ok {
		return 0, fmt.Errorf("foreign MM save not found in OoT payload")
	}

	r.foreignMmSaveAddr = addr
	return addr, nil
}

func locateForeignSave(payload []byte, payloadBase uint32, saveSize, nameOffset int) (uint32, bool) {
	if len(payload) < saveSize || nameOffset < 0 || nameOffset+5 > saveSize {
		return 0, false
	}

	for offset := 0; offset+saveSize <= len(payload); offset += 16 {
		if !bytes.HasPrefix(payload[offset+nameOffset:offset+nameOffset+5], []byte("ZELDA")) {
			continue
		}
		return payloadBase + uint32(offset), true
	}

	return 0, false
}
