package main

import "testing"

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
