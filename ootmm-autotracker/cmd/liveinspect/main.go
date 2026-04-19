package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"ootmm-autotracker/n64"
	"ootmm-autotracker/ootmm"
	"ootmm-autotracker/retroarch"
)

func main() {
	const (
		targetChestKey  = "MM_chest_45_1"
		targetXflagKey  = "MM_xflag_2036"
		targetXflagByte = 2036 / 8
		targetXflagMask = 1 << (2036 % 8)
	)

	client := retroarch.NewClient(retroarch.DefaultHost, retroarch.DefaultPort)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "connect RetroArch: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	mem := n64.NewMemory(client)
	if err := mem.Probe(); err != nil {
		fmt.Fprintf(os.Stderr, "probe memory: %v\n", err)
		os.Exit(1)
	}

	reader := ootmm.NewReader(mem)

	// Poll rapidly to catch transitions.
	// Preserve the last valid MM snapshot across GameNone frames so we can
	// distinguish real removals from transition gaps.
	prev := make(map[string]struct{})
	for i := 0; i < 600; i++ { // ~60 seconds at 100ms
		state, err := reader.ReadState()
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %d: %v\n", i, err)
			continue
		}

		current := make(map[string]struct{})
		isValidMM := state.ActiveGame == ootmm.GameMm
		if state.ActiveGame != ootmm.GameNone {
			for _, check := range ootmm.ExtractChecks(state) {
				name := strings.ToLower(check.Name)
				if strings.Contains(name, "termina field") {
					current[check.Key] = struct{}{}
				}
			}
		}

		xflagSet := false
		xflagsMm := state.Shared.Bitmap("xflagsMm")
		if targetXflagByte < len(xflagsMm) {
			xflagSet = xflagsMm[targetXflagByte]&targetXflagMask != 0
		}

		dumpState := func(prefix string) {
			fmt.Printf(
				"[%3d] %s game=%s mmScene=%d hasLive=%v liveChest=%#08x cycleChest45=%#08x permChest45=%#08x xflag2036=%v checks(chest=%v wonder=%v totalTermina=%d)\n",
				i,
				prefix,
				state.ActiveGame,
				state.Mm.LiveSceneID,
				state.Mm.HasLiveSceneFlags,
				state.Mm.LiveChestFlags,
				state.Mm.CycleFlags[45].Chests,
				state.Mm.SceneFlags[45].Chests,
				xflagSet,
				contains(current, targetChestKey),
				contains(current, targetXflagKey),
				len(current),
			)
		}

		if !isValidMM {
			if state.ActiveGame == ootmm.GameNone && len(prev) > 0 {
				dumpState("game=None")
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		prevChest := contains(prev, targetChestKey)
		prevWonder := contains(prev, targetXflagKey)
		currentChest := contains(current, targetChestKey)
		currentWonder := contains(current, targetXflagKey)

		if prevChest != currentChest {
			if currentChest {
				dumpState("+ADD " + targetChestKey)
			} else {
				dumpState("-DEL " + targetChestKey)
			}
		}
		if prevWonder != currentWonder {
			if currentWonder {
				dumpState("+ADD " + targetXflagKey)
			} else {
				dumpState("-DEL " + targetXflagKey)
			}
		}
		if len(prev) != len(current) && prevChest == currentChest && prevWonder == currentWonder {
			dumpState("Termina count changed")
		}
		if state.ActiveGame == ootmm.GameNone && len(prev) > 0 {
			dumpState("game=None")
		}

		prev = current
		// sleep ~100ms
		time.Sleep(100 * time.Millisecond)
	}
}

func contains(values map[string]struct{}, key string) bool {
	_, ok := values[key]
	return ok
}
