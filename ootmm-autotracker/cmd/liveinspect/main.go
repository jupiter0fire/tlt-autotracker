package main

import (
	"fmt"
	"os"

	"github.com/ootmm-autotracker/n64"
	"github.com/ootmm-autotracker/ootmm"
	"github.com/ootmm-autotracker/retroarch"
)

const inspectBit = 170

func main() {
	client := retroarch.NewClient(retroarch.DefaultHost, retroarch.DefaultPort)
	if err := client.Connect(); err != nil {
		fatalf("connect RetroArch: %v", err)
	}
	defer client.Close()

	mem := n64.NewMemory(client)
	if err := mem.Probe(); err != nil {
		fatalf("probe memory: %v", err)
	}

	reader := ootmm.NewReader(mem)
	var (
		state *ootmm.GameState
		err   error
	)
	for attempt := 0; attempt < 8; attempt++ {
		state, err = reader.ReadState()
		if err != nil {
			fatalf("read state: %v", err)
		}
		if state != nil && state.Valid && state.ActiveGame != ootmm.GameNone {
			break
		}
	}

	if state == nil {
		fatalf("reader returned nil state")
	}

	fmt.Printf("state valid=%v active=%s save=%d ootScene=%#x liveXflag170=%v\n",
		state.Valid,
		state.ActiveGame,
		state.SaveIndex,
		state.Oot.SceneID,
		isBitSet(state.Shared.Bitmap("xflagsOot"), inspectBit),
	)
	fmt.Printf("live scene=%#x chest=%#08x collect=%#08x tempCollect=%#08x hasLive=%v\n",
		state.Oot.LiveSceneID,
		state.Oot.LiveChestFlags,
		state.Oot.LiveCollectFlags,
		state.Oot.LiveTempCollectFlag,
		state.Oot.HasLiveSceneFlags,
	)
	target := "Dodongo Cavern Heart Miniboss Lava"
	hasTarget := false
	for _, check := range ootmm.ExtractChecks(state) {
		if check.Name == target {
			hasTarget = true
			break
		}
	}
	fmt.Printf("target=%q present=%v\n", target, hasTarget)
}

func isBitSet(bitmap []byte, bit int) bool {
	byteIndex := bit / 8
	if byteIndex < 0 || byteIndex >= len(bitmap) {
		return false
	}
	return bitmap[byteIndex]&(1<<uint(bit%8)) != 0
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}