package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigurationOverridesAndFallback(t *testing.T) {
	defaults := defaultConfiguration()
	if defaults.Keys.OpenInEditor != "z" || defaults.EditorDirCommand != "zed --classic || zeditor --classic || code --new-window" || defaults.EditorFileCommand != "zed --new || zeditor --new || code --new-window" {
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
		{"rename conflict", `{"keys":{"open_in_editor":"r"}}`, "z", true},
		{"Finder conflict", `{"keys":{"open_in_editor":"f"}}`, "z", true},
		{"multiple", `{"keys":{"open_in_zed":"zz"}}`, "z", true},
		{"unknown", `{"keys":{"zed":"e"}}`, "z", true},
		{"invalid", "{", "z", true},
		{"null", "null", "z", true},
		{"trailing", "{} {}", "z", true},
		{"bad command", `{"editor_command":"zed ||"}`, "z", true},
		{"shell command", `{"editor_command":"zed; code"}`, "z", true},
		{"command type", `{"editor_command":[]}`, "z", true},
		{"bad dir command", `{"editor_dir_command":"zed ||"}`, "z", true},
		{"bad file command", `{"editor_file_command":"zed ||"}`, "z", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(tt.body), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := readConfiguration(path)
			if (err != nil) != tt.invalid || got.Keys.OpenInEditor != tt.want || got.EditorDirCommand != defaults.EditorDirCommand || got.EditorFileCommand != defaults.EditorFileCommand {
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

func TestEditorCommandOverrides(t *testing.T) {
	defaults := defaultConfiguration()
	for _, tt := range []struct {
		body, dir, file string
	}{
		{`{"editor_command":"code --new-window"}`, "code --new-window", "code --new-window"},
		{`{"editor_command":""}`, "", ""},
		{`{"editor_dir_command":"zed --classic"}`, "zed --classic", defaults.EditorFileCommand},
		{`{"editor_file_command":"code --new-window"}`, defaults.EditorDirCommand, "code --new-window"},
		{`{"editor_dir_command":""}`, "", defaults.EditorFileCommand},
		{`{"editor_file_command":""}`, defaults.EditorDirCommand, ""},
		{`{"editor_command":"legacy","editor_dir_command":"dirs"}`, "dirs", "legacy"},
		{`{"editor_file_command":"files","editor_command":"legacy"}`, "legacy", "files"},
		{`{"editor_dir_command":"dirs","editor_file_command":"files","editor_command":"legacy"}`, "dirs", "files"},
	} {
		path := filepath.Join(t.TempDir(), "config.json")
		if err := os.WriteFile(path, []byte(tt.body), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := readConfiguration(path)
		if err != nil || got.EditorDirCommand != tt.dir || got.EditorFileCommand != tt.file || got.Keys.OpenInEditor != "z" {
			t.Fatalf("config=%+v, error=%v", got, err)
		}
	}
}
