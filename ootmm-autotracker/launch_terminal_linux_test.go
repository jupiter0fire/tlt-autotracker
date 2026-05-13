//go:build linux

package main

import (
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestShouldRelaunchInTerminal(t *testing.T) {
	tests := []struct {
		name  string
		state terminalRelaunchState
		want  bool
	}{
		{
			name: "desktop launch without tty",
			state: terminalRelaunchState{
				display: ":1",
			},
			want: true,
		},
		{
			name: "existing terminal session",
			state: terminalRelaunchState{
				stdoutTTY: true,
				display:   ":1",
			},
			want: false,
		},
		{
			name: "redirected shell keeps current session",
			state: terminalRelaunchState{
				term:    "xterm-256color",
				display: ":1",
			},
			want: false,
		},
		{
			name: "headless session",
			state: terminalRelaunchState{},
			want: false,
		},
		{
			name: "already relaunched",
			state: terminalRelaunchState{
				display:    ":1",
				relaunched: "1",
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := shouldRelaunchInTerminal(test.state); got != test.want {
				t.Fatalf("shouldRelaunchInTerminal(%+v) = %v, want %v", test.state, got, test.want)
			}
		})
	}
}

func TestLaunchLinuxTerminalFallsBackToNextLauncher(t *testing.T) {
	var seenArgs [][]string
	var seenEnv []string

	launched, err := launchLinuxTerminal(
		"/tmp/ootmm-autotracker",
		[]string{"--pj64", "--ws-addr", ":17026"},
		func(name string) (string, error) {
			switch name {
			case "x-terminal-emulator":
				return "/usr/bin/x-terminal-emulator", nil
			case "gnome-terminal":
				return "/usr/bin/gnome-terminal", nil
			default:
				return "", exec.ErrNotFound
			}
		},
		func(cmd *exec.Cmd) error {
			seenArgs = append(seenArgs, append([]string(nil), cmd.Args...))
			seenEnv = append([]string(nil), cmd.Env...)
			if strings.Contains(cmd.Path, "x-terminal-emulator") {
				return errors.New("boom")
			}
			return nil
		},
		[]string{"A=B", terminalRelaunchEnv + "=1"},
	)
	if err != nil {
		t.Fatalf("launchLinuxTerminal returned error: %v", err)
	}
	if !launched {
		t.Fatalf("launchLinuxTerminal returned launched=false")
	}

	if len(seenArgs) != 2 {
		t.Fatalf("saw %d launch attempts, want 2", len(seenArgs))
	}

	wantFirst := []string{"/usr/bin/x-terminal-emulator", "-e", "/tmp/ootmm-autotracker", "--pj64", "--ws-addr", ":17026"}
	if !reflect.DeepEqual(seenArgs[0], wantFirst) {
		t.Fatalf("first launch args = %#v, want %#v", seenArgs[0], wantFirst)
	}

	wantSecond := []string{"/usr/bin/gnome-terminal", "--", "/tmp/ootmm-autotracker", "--pj64", "--ws-addr", ":17026"}
	if !reflect.DeepEqual(seenArgs[1], wantSecond) {
		t.Fatalf("second launch args = %#v, want %#v", seenArgs[1], wantSecond)
	}

	if !reflect.DeepEqual(seenEnv, []string{"A=B", terminalRelaunchEnv + "=1"}) {
		t.Fatalf("launch env = %#v", seenEnv)
	}
}

func TestLaunchLinuxTerminalReturnsErrorWithoutLauncher(t *testing.T) {
	launched, err := launchLinuxTerminal(
		"/tmp/ootmm-autotracker",
		nil,
		func(string) (string, error) {
			return "", exec.ErrNotFound
		},
		func(cmd *exec.Cmd) error {
			return nil
		},
		nil,
	)
	if err == nil {
		t.Fatalf("launchLinuxTerminal returned nil error")
	}
	if launched {
		t.Fatalf("launchLinuxTerminal returned launched=true")
	}
	if !strings.Contains(err.Error(), "no supported terminal emulator found") {
		t.Fatalf("launchLinuxTerminal error = %q", err)
	}
}