package ootmm

import (
	"encoding/binary"
	"fmt"

	"github.com/ootmm-autotracker/n64"
)

// Reader reads OoTMM game state from N64 RDRAM.
type Reader struct {
	mem      *n64.Memory
	detector *Detector
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
	if err := r.readOotSave(&state.Oot); err != nil {
		return nil, fmt.Errorf("read OoT save: %w", err)
	}

	// Read MM save
	if err := r.readMmSave(&state.Mm); err != nil {
		return nil, fmt.Errorf("read MM save: %w", err)
	}

	return state, nil
}

func (r *Reader) readOotSave(oot *OotState) error {
	// Read the entire OotSaveContext in one large read
	data, err := r.mem.Read(AddrOotSaveCtx, OotSaveCtxSize)
	if err != nil {
		return fmt.Errorf("read OoT save context: %w", err)
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

	oot.GameMode = binary.BigEndian.Uint32(data[OotCtxOffGameMode:])

	return nil
}

func (r *Reader) readMmSave(mm *MmState) error {
	// Read the entire MmSaveContext in one large read
	data, err := r.mem.Read(AddrMmSaveCtx, MmSaveCtxSize)
	if err != nil {
		return fmt.Errorf("read MM save context: %w", err)
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

	mm.GameMode = binary.BigEndian.Uint32(data[MmCtxOffGameMode:])

	return nil
}
