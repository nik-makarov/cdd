package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigurationOverridesAndFallback(t *testing.T) {
	defaults := defaultConfiguration()
	if defaults.Keys.OpenInEditor != "z" || defaults.EditorCommand != "zed --new || zeditor --new || code --new-window" {
		t.Fatal("shipped config must bind the editor fallback command to z")
	}
	for _, tt := range []struct {
		name, body, want string
		invalid          bool
	}{
		{"defaults", `{}`, "z", false},
		{"custom", `{"keys":{"open_in_zed":"e"}}`, "e", false},
		{"editor key", `{"keys":{"open_in_editor":"e"}}`, "e", false},
		{"editor key wins", `{"keys":{"open_in_editor":"e","open_in_zed":"x"}}`, "e", false},
		{"editor disabled", `{"keys":{"open_in_editor":""}}`, "", false},
		{"disabled", `{"keys":{"open_in_zed":""}}`, "", false},
		{"conflict", `{"keys":{"open_in_zed":"v"}}`, "z", true},
		{"new file conflict", `{"keys":{"open_in_zed":"N"}}`, "z", true},
		{"multiple", `{"keys":{"open_in_zed":"zz"}}`, "z", true},
		{"unknown", `{"keys":{"zed":"e"}}`, "z", true},
		{"invalid", "{", "z", true},
		{"null", "null", "z", true},
		{"trailing", "{} {}", "z", true},
		{"bad command", `{"editor_command":"zed ||"}`, "z", true},
		{"shell command", `{"editor_command":"zed; code"}`, "z", true},
		{"command type", `{"editor_command":[]}`, "z", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(tt.body), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := readConfiguration(path)
			if (err != nil) != tt.invalid || got.Keys.OpenInEditor != tt.want || got.EditorCommand != defaults.EditorCommand {
				t.Fatalf("config=%+v, error=%v", got, err)
			}
		})
	}
	for _, path := range []string{"", filepath.Join(t.TempDir(), "missing")} {
		got, err := readConfiguration(path)
		if err != nil || got != defaults {
			t.Fatal("missing config must use shipped defaults")
		}
	}
}

func TestEditorCommandOverride(t *testing.T) {
	for _, command := range []string{"code --new-window", ""} {
		path := filepath.Join(t.TempDir(), "config.json")
		if err := os.WriteFile(path, []byte(`{"editor_command":"`+command+`"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := readConfiguration(path)
		if err != nil || got.EditorCommand != command || got.Keys.OpenInEditor != "z" {
			t.Fatalf("config=%+v, error=%v", got, err)
		}
	}
}
