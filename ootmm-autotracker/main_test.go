package main

import (
	"testing"
	"time"
)

func TestNoteOoTMMUnavailableTracksElapsedTime(t *testing.T) {
	var unavailableSince time.Time
	start := time.Unix(100, 0)

	if got := noteOoTMMUnavailable(&unavailableSince, start); got != 0 {
		t.Fatalf("first unavailable duration = %s, want 0", got)
	}
	if !unavailableSince.Equal(start) {
		t.Fatalf("unavailableSince = %s, want %s", unavailableSince, start)
	}

	later := start.Add(ootmmLostTimeout + 1500*time.Millisecond)
	if got := noteOoTMMUnavailable(&unavailableSince, later); got != later.Sub(start) {
		t.Fatalf("later unavailable duration = %s, want %s", got, later.Sub(start))
	}
}

func TestParseBackendChoiceMatchesAliases(t *testing.T) {
	options := []*backendOption{
		{kind: backendRetroArch, name: "RetroArch"},
		{kind: backendPJ64, name: "Project64"},
	}

	tests := []struct {
		name  string
		input string
		want  backendKind
	}{
		{name: "retroarch number", input: "1", want: backendRetroArch},
		{name: "retroarch alias", input: "ra", want: backendRetroArch},
		{name: "retroarch name", input: "RetroArch", want: backendRetroArch},
		{name: "pj64 number", input: "2", want: backendPJ64},
		{name: "pj64 alias", input: "pj64", want: backendPJ64},
		{name: "project64 name", input: "Project64", want: backendPJ64},
		{name: "project shortcut", input: "project", want: backendPJ64},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			chosen := parseBackendChoice(test.input, options)
			if chosen == nil {
				t.Fatalf("parseBackendChoice(%q) returned nil", test.input)
			}
			if chosen.kind != test.want {
				t.Fatalf("parseBackendChoice(%q) = %q, want %q", test.input, chosen.kind, test.want)
			}
		})
	}
}

func TestParseBackendChoiceRejectsUnknownInput(t *testing.T) {
	options := []*backendOption{
		{kind: backendRetroArch, name: "RetroArch"},
		{kind: backendPJ64, name: "Project64"},
	}

	if chosen := parseBackendChoice("mupen", options); chosen != nil {
		t.Fatalf("parseBackendChoice returned %q for invalid input", chosen.kind)
	}
}

func TestShortCommitHash(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "whitespace", input: "  ", want: ""},
		{name: "devel", input: "(devel)", want: ""},
		{name: "short", input: "abc1234", want: "abc1234"},
		{name: "full", input: "0123456789abcdef0123456789abcdef01234567", want: "0123456789ab"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := shortCommitHash(test.input); got != test.want {
				t.Fatalf("shortCommitHash(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestStartupCommitHashPrefersInjectedValue(t *testing.T) {
	original := commitHash
	commitHash = "0123456789abcdef0123456789abcdef01234567"
	t.Cleanup(func() {
		commitHash = original
	})

	if got := startupCommitHash(); got != "0123456789ab" {
		t.Fatalf("startupCommitHash() = %q, want %q", got, "0123456789ab")
	}
}
