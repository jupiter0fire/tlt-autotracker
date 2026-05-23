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

func TestNoteOoTMMUnavailableAfterValidSaveIgnoresStartupInvalidSaves(t *testing.T) {
	var unavailableSince time.Time
	start := time.Unix(100, 0)

	if got := noteOoTMMUnavailableAfterValidSave(false, &unavailableSince, start); got != 0 {
		t.Fatalf("startup invalid-save duration = %s, want 0", got)
	}
	if !unavailableSince.IsZero() {
		t.Fatalf("startup invalid-save timestamp = %s, want zero", unavailableSince)
	}

	if got := noteOoTMMUnavailableAfterValidSave(true, &unavailableSince, start); got != 0 {
		t.Fatalf("first post-valid invalid-save duration = %s, want 0", got)
	}
	if !unavailableSince.Equal(start) {
		t.Fatalf("post-valid invalid-save timestamp = %s, want %s", unavailableSince, start)
	}

	later := start.Add(ootmmLostTimeout + 1500*time.Millisecond)
	if got := noteOoTMMUnavailableAfterValidSave(true, &unavailableSince, later); got != later.Sub(start) {
		t.Fatalf("later post-valid invalid-save duration = %s, want %s", got, later.Sub(start))
	}
}

func TestShouldRestartSessionOnReadFailure(t *testing.T) {
	tests := []struct {
		name            string
		hasReadValidSave bool
		backendConnected bool
		want            bool
	}{
		{
			name:             "startup read failure waits quietly",
			hasReadValidSave: false,
			backendConnected: true,
			want:             false,
		},
		{
			name:             "tracked session read failure restarts",
			hasReadValidSave: true,
			backendConnected: true,
			want:             true,
		},
		{
			name:             "disconnected backend restarts immediately",
			hasReadValidSave: false,
			backendConnected: false,
			want:             true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := shouldRestartSessionOnReadFailure(test.hasReadValidSave, test.backendConnected); got != test.want {
				t.Fatalf(
					"shouldRestartSessionOnReadFailure(validSave=%t, connected=%t) = %t, want %t",
					test.hasReadValidSave,
					test.backendConnected,
					got,
					test.want,
				)
			}
		})
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

func TestSplitCommaSeparatedListTrimsAndSkipsEmptyEntries(t *testing.T) {
	got := splitCommaSeparatedList(" http://localhost:5173, ,https://www.thelasttracker.org/ ")
	want := []string{"http://localhost:5173", "https://www.thelasttracker.org/"}
	if len(got) != len(want) {
		t.Fatalf("splitCommaSeparatedList length = %d, want %d (%#v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("splitCommaSeparatedList[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
