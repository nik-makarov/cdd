package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigurationOverridesAndFallback(t *testing.T) {
	defaults := defaultConfiguration()
	if defaults.Keys.OpenInZed != "z" {
		t.Fatal("shipped config must bind Zed to z")
	}
	for _, tt := range []struct {
		name, body, want string
		invalid          bool
	}{
		{"defaults", `{}`, "z", false},
		{"custom", `{"keys":{"open_in_zed":"e"}}`, "e", false},
		{"disabled", `{"keys":{"open_in_zed":""}}`, "", false},
		{"conflict", `{"keys":{"open_in_zed":"v"}}`, "z", true},
		{"multiple", `{"keys":{"open_in_zed":"zz"}}`, "z", true},
		{"unknown", `{"keys":{"zed":"e"}}`, "z", true},
		{"invalid", "{", "z", true},
		{"null", "null", "z", true},
		{"trailing", "{} {}", "z", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(tt.body), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := readConfiguration(path)
			if (err != nil) != tt.invalid || got.Keys.OpenInZed != tt.want {
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
