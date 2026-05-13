//go:build !linux

package main

func relaunchInTerminalIfNeeded() (bool, error) {
	return false, nil
}