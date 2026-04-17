package main

import (
"fmt"
"os"
"strings"

"ootmm-autotracker/n64"
"ootmm-autotracker/ootmm"
"ootmm-autotracker/retroarch"
)

func main() {
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
state, err := reader.ReadState()
if err != nil {
fmt.Fprintf(os.Stderr, "read state: %v\n", err)
os.Exit(1)
}

if state == nil {
fmt.Fprintf(os.Stderr, "reader returned nil state\n")
os.Exit(1)
}

fmt.Printf("active game: %s\n", state.ActiveGame)
fmt.Printf("current OoT scene: %#x\n", state.Oot.SceneID)
fmt.Printf("live chest flags: %#08x\n", state.Oot.LiveChestFlags)
if len(state.Oot.SceneFlags) > 1 {
fmt.Printf("stable scene 1 chest flags: %#08x\n", state.Oot.SceneFlags[1].Chests)
}

foundCompass := false
for _, check := range ootmm.ExtractChecks(state) {
if strings.Contains(strings.ToLower(check.Name), "dodongo") && strings.Contains(strings.ToLower(check.Name), "compass") {
fmt.Printf("ExtractChecks includes: %s\n", check.Name)
foundCompass = true
}
}
if !foundCompass {
fmt.Println("ExtractChecks does NOT include any Dodongo Cavern Compass chest name.")
}
}
