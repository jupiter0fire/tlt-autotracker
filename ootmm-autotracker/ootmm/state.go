package ootmm

// GameState holds the complete tracked state from both games.
type GameState struct {
	ActiveGame ActiveGame
	SaveIndex  uint32
	Valid      bool

	Oot    OotState
	Mm     MmState
	Shared SharedCustomState
}

// SharedCustomState holds the subset of OoTMM's shared custom save that is
// relevant for item tracking across both games.
type SharedCustomState struct {
	Bitmaps                map[string][]uint8
	OcarinaButtonMaskOot   uint16
	OcarinaButtonMaskMm    uint16
	BombchuBagOot          uint8
	BombchuBagMm           uint8
	SongNotes              [sharedSongNoteCount]uint8
	CaughtChildFishWeights [20]uint8
	CaughtAdultFishWeights [20]uint8
}

func (s SharedCustomState) Clone() SharedCustomState {
	if len(s.Bitmaps) == 0 {
		return SharedCustomState{
			OcarinaButtonMaskOot:   s.OcarinaButtonMaskOot,
			OcarinaButtonMaskMm:    s.OcarinaButtonMaskMm,
			BombchuBagOot:          s.BombchuBagOot,
			BombchuBagMm:           s.BombchuBagMm,
			SongNotes:              s.SongNotes,
			CaughtChildFishWeights: s.CaughtChildFishWeights,
			CaughtAdultFishWeights: s.CaughtAdultFishWeights,
		}
	}

	clone := SharedCustomState{
		Bitmaps:                make(map[string][]uint8, len(s.Bitmaps)),
		OcarinaButtonMaskOot:   s.OcarinaButtonMaskOot,
		OcarinaButtonMaskMm:    s.OcarinaButtonMaskMm,
		BombchuBagOot:          s.BombchuBagOot,
		BombchuBagMm:           s.BombchuBagMm,
		SongNotes:              s.SongNotes,
		CaughtChildFishWeights: s.CaughtChildFishWeights,
		CaughtAdultFishWeights: s.CaughtAdultFishWeights,
	}
	for name, bitmap := range s.Bitmaps {
		clone.Bitmaps[name] = append([]uint8(nil), bitmap...)
	}
	return clone
}

func (s SharedCustomState) Bitmap(name string) []uint8 {
	if s.Bitmaps == nil {
		return nil
	}
	return s.Bitmaps[name]
}

func (s *SharedCustomState) SetBitmap(name string, bitmap []uint8) {
	if s.Bitmaps == nil {
		s.Bitmaps = make(map[string][]uint8)
	}
	s.Bitmaps[name] = append([]uint8(nil), bitmap...)
}

func (s *SharedCustomState) SetBit(name string, bit int) {
	if bit < 0 {
		return
	}
	if s.Bitmaps == nil {
		s.Bitmaps = make(map[string][]uint8)
	}
	byteIndex := bit / 8
	bitmap := s.Bitmaps[name]
	if len(bitmap) <= byteIndex {
		grown := make([]uint8, byteIndex+1)
		copy(grown, bitmap)
		bitmap = grown
	}
	bitmap[byteIndex] |= 1 << uint(bit%8)
	s.Bitmaps[name] = bitmap
}

// OotState holds all tracked OoT data.
type OotState struct {
	Entrance         uint32
	SceneID          uint16
	LiveSceneID      uint16
	Age              uint32 // 0=adult, 1=child
	GameMode         uint32
	OcarinaGameRound uint8
	HasMagic         bool
	HasDoubleMagic   bool
	IsBiggoronSword  bool

	LiveChestFlags      uint32
	LiveCollectFlags    uint32
	LiveTempCollectFlag uint32
	HasLiveSceneFlags   bool

	// Inventory
	Items [24]uint8
	Ammo  [15]uint8
	Beans uint8

	// Equipment bitfield: boots:4, tunics:4, shields:4, swords:4
	Equipment uint16
	// Upgrades bitfield
	Upgrades uint32
	// Quest items bitfield (medallions, songs, stones, etc.)
	QuestItems uint32
	// Heart pieces (top 4 bits of questItems)
	HeartPieces uint8

	// Per-dungeon packed byte: boss key/compass/map use low bits, max keys use upper bits.
	DungeonItems [20]uint8
	DungeonKeys  [19]int8

	GoldTokens                  uint16
	GsFlags                     [6]uint32
	RuntimeMqBits               uint32
	RuntimeMaxKeys              [OotRuntimeSceneCount]uint8
	RuntimeSilverRupeeCounts    [OotSilverRupeeSetCount]uint8
	HasRuntimeMqBits            bool
	HasRuntimeMaxKeys           bool
	HasRuntimeSilverRupeeCounts bool

	// Scene flags: chests/switches/collectibles per scene
	SceneFlags [124]SceneFlags

	// Extra records
	ExtraRecords [20]uint32

	// Event flags
	EventsChk  [14]uint16
	EventsItem [4]uint16
	EventsMisc [30]uint16
}

// MmState holds all tracked MM data.
type MmState struct {
	PlayerForm     uint8
	Day            uint32
	Time           uint16
	GameMode       uint32
	HasMagic       bool
	HasDoubleMagic bool

	LiveSceneID       uint16
	LiveChestFlags    uint32
	LiveCollectFlags  uint32
	HasLiveSceneFlags bool

	// Inventory
	Items [48]uint8
	Ammo  [24]int8

	// Equipment bitfield: boots:4, tunic:4, shield:4, sword:4.
	// In practice MM tracking currently uses sword and shield levels.
	Equipment uint16

	// Upgrades bitfield
	Upgrades uint32
	// Quest items (remains, songs, notebook, heart pieces)
	QuestItems  uint32
	HeartPieces uint8
	// Vanilla MM owl statue activation bits. In OoTMM these still represent the
	// persistent activated/check state even when owl statue shuffle is enabled.
	OwlActivationFlags uint16

	// Per-dungeon packed byte: boss key/compass/map use low bits, max keys use upper bits.
	DungeonItems     [10]uint8
	DungeonKeys      [9]int8
	StrayFairies     [10]int8
	WeekEventReg     [100]uint8
	TownStrayFairy   bool
	ExtraFlags2      uint32
	SkullTokensSwamp uint16
	SkullTokensOcean uint16

	SceneFlags [120]SceneFlags

	// Cycle flags (reset each 3-day cycle)
	CycleFlags [120]CycleSceneFlags
}

// SceneFlags represents the permanent flags for a single scene.
type SceneFlags struct {
	Chests        uint32
	Switches      uint32
	RoomClear     uint32
	Collectibles  uint32
	Unused        uint32
	VisitedRooms  uint32
	VisitedFloors uint32
}

// CycleSceneFlags represents MM's per-cycle scene flags.
type CycleSceneFlags struct {
	Chests       uint32
	Switch0      uint32
	Switch1      uint32
	ClearedRoom  uint32
	Collectibles uint32
}

// HasQuestBit checks if a specific quest bit is set.
func HasQuestBit(questItems uint32, bit int) bool {
	return questItems&(1<<uint(bit)) != 0
}

// GetUpgradeLevel extracts a multi-bit upgrade field.
func GetUpgradeLevel(upgrades uint32, shift, bits int) int {
	mask := uint32((1 << bits) - 1)
	return int((upgrades >> uint(shift)) & mask)
}
