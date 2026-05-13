//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

const terminalRelaunchEnv = "OOTMM_AUTOTRACKER_TERMINAL_RELAUNCH"

type terminalRelaunchState struct {
	stdinTTY   bool
	stdoutTTY  bool
	stderrTTY  bool
	term       string
	display    string
	wayland    string
	relaunched string
}

type terminalLauncher struct {
	name      string
	buildArgs func(executable string, args []string) []string
}

var linuxTerminalLaunchers = []terminalLauncher{
	{name: "x-terminal-emulator", buildArgs: func(executable string, args []string) []string {
		return append([]string{"-e", executable}, args...)
	}},
	{name: "gnome-terminal", buildArgs: func(executable string, args []string) []string {
		return append([]string{"--", executable}, args...)
	}},
	{name: "kgx", buildArgs: func(executable string, args []string) []string {
		return append([]string{"--", executable}, args...)
	}},
	{name: "konsole", buildArgs: func(executable string, args []string) []string {
		return append([]string{"-e", executable}, args...)
	}},
	{name: "xfce4-terminal", buildArgs: func(executable string, args []string) []string {
		return append([]string{"-x", executable}, args...)
	}},
	{name: "mate-terminal", buildArgs: func(executable string, args []string) []string {
		return append([]string{"-x", executable}, args...)
	}},
	{name: "alacritty", buildArgs: func(executable string, args []string) []string {
		return append([]string{"-e", executable}, args...)
	}},
	{name: "kitty", buildArgs: func(executable string, args []string) []string {
		return append([]string{executable}, args...)
	}},
	{name: "wezterm", buildArgs: func(executable string, args []string) []string {
		return append([]string{"start", "--always-new-process", "--", executable}, args...)
	}},
	{name: "xterm", buildArgs: func(executable string, args []string) []string {
		return append([]string{"-e", executable}, args...)
	}},
}

func relaunchInTerminalIfNeeded() (bool, error) {
	state := terminalRelaunchState{
		stdinTTY:   isTerminal(os.Stdin),
		stdoutTTY:  isTerminal(os.Stdout),
		stderrTTY:  isTerminal(os.Stderr),
		term:       strings.TrimSpace(os.Getenv("TERM")),
		display:    strings.TrimSpace(os.Getenv("DISPLAY")),
		wayland:    strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")),
		relaunched: strings.TrimSpace(os.Getenv(terminalRelaunchEnv)),
	}
	if !shouldRelaunchInTerminal(state) {
		return false, nil
	}

	executable, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("determine executable path: %w", err)
	}

	childEnv := append(os.Environ(), terminalRelaunchEnv+"=1")
	return launchLinuxTerminal(executable, os.Args[1:], exec.LookPath, func(cmd *exec.Cmd) error {
		return cmd.Start()
	}, childEnv)
}

func shouldRelaunchInTerminal(state terminalRelaunchState) bool {
	if state.relaunched != "" {
		return false
	}
	if state.term != "" {
		return false
	}
	if state.display == "" && state.wayland == "" {
		return false
	}
	if state.stdinTTY || state.stdoutTTY || state.stderrTTY {
		return false
	}
	return true
}

func launchLinuxTerminal(executable string, args []string, lookPath func(string) (string, error), start func(*exec.Cmd) error, env []string) (bool, error) {
	var launchErrors []string

	for _, launcher := range linuxTerminalLaunchers {
		path, err := lookPath(launcher.name)
		if err != nil {
			continue
		}

		cmd := exec.Command(path, launcher.buildArgs(executable, args)...)
		cmd.Env = env
		if err := start(cmd); err != nil {
			launchErrors = append(launchErrors, fmt.Sprintf("%s: %v", launcher.name, err))
			continue
		}

		return true, nil
	}

	if len(launchErrors) > 0 {
		return false, fmt.Errorf("could not start a supported terminal emulator: %s", strings.Join(launchErrors, "; "))
	}

	return false, fmt.Errorf("no supported terminal emulator found")
}

func isTerminal(file *os.File) bool {
	if file == nil {
		return false
	}

	var termios syscall.Termios
	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		file.Fd(),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&termios)),
		0,
		0,
		0,
	)
	return errno == 0
}