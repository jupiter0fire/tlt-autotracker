package ootmm

import "testing"

func TestIsPlayableSaveStateRequiresNormalGameplayAndRealSaveSlot(t *testing.T) {
	tests := []struct {
		name      string
		game      ActiveGame
		gameMode  uint32
		saveIndex uint32
		runtime   uint32
		want      bool
	}{
		{name: "oot slot 0", game: GameOot, gameMode: gameModeNormal, saveIndex: 0, runtime: 1, want: true},
		{name: "oot slot 2", game: GameOot, gameMode: gameModeNormal, saveIndex: 2, runtime: 1, want: true},
		{name: "mm menu runtime not initialized", game: GameMm, gameMode: gameModeNormal, saveIndex: 0, runtime: 0, want: false},
		{name: "mm loaded save", game: GameMm, gameMode: gameModeNormal, saveIndex: 0, runtime: 0x801c0000, want: true},
		{name: "file select slot", game: GameMm, gameMode: gameModeFileSelect, saveIndex: 1, runtime: 0x801c0000, want: false},
		{name: "title screen", game: GameMm, gameMode: gameModeTitleScreen, saveIndex: 0, runtime: 0x801c0000, want: false},
		{name: "copy button index", game: GameMm, gameMode: gameModeNormal, saveIndex: 3, runtime: 0x801c0000, want: false},
		{name: "options button index", game: GameMm, gameMode: gameModeNormal, saveIndex: 4, runtime: 0x801c0000, want: false},
		{name: "empty sentinel", game: GameMm, gameMode: gameModeNormal, saveIndex: 0xff, runtime: 0x801c0000, want: false},
	}

	for _, test := range tests {
		if got := isReadyGameState(test.game, test.gameMode, test.saveIndex, test.runtime); got != test.want {
			t.Fatalf("%s: got %v want %v", test.name, got, test.want)
		}
	}
}