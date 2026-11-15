package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//go:embed config/default.json
var defaultConfigJSON []byte

type configuration struct {
	EditorCommand string `json:"editor_command"`
	Keys          struct {
		OpenInEditor string `json:"open_in_editor"`
	} `json:"keys"`
}

func defaultConfiguration() configuration {
	var config configuration
	if err := json.Unmarshal(defaultConfigJSON, &config); err != nil {
		panic("invalid embedded default configuration: " + err.Error())
	}
	return config
}

func configurationPath(name string) string {
	dir, err := os.UserConfigDir()
	if err != nil || !filepath.IsAbs(dir) {
		return ""
	}
	return filepath.Join(dir, "cdd", name)
}

func readConfiguration(path string) (configuration, error) {
	defaults := defaultConfiguration()
	if path == "" {
		return defaults, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return defaults, nil
	}
	if err != nil {
		return defaults, err
	}
	if !strings.HasPrefix(strings.TrimSpace(string(data)), "{") {
		return defaults, errors.New("Config must be a JSON object")
	}
	var overrides struct {
		EditorCommand *string `json:"editor_command"`
		Keys          struct {
			OpenInEditor *string `json:"open_in_editor"`
			OpenInZed    *string `json:"open_in_zed"`
		} `json:"keys"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&overrides); err != nil {
		return defaults, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return defaults, errors.New("Config must contain one JSON object")
	}
	config := defaults
	if overrides.EditorCommand != nil {
		config.EditorCommand = *overrides.EditorCommand
	}
	if overrides.Keys.OpenInZed != nil {
		config.Keys.OpenInEditor = *overrides.Keys.OpenInZed
	}
	if overrides.Keys.OpenInEditor != nil {
		config.Keys.OpenInEditor = *overrides.Keys.OpenInEditor
	}
	if _, err := parseEditorCommand(config.EditorCommand); err != nil {
		return defaults, fmt.Errorf("editor_command: %w", err)
	}
	key := config.Keys.OpenInEditor
	if key != "" {
		if len(key) != 1 || !((key[0] >= 'a' && key[0] <= 'z') || (key[0] >= 'A' && key[0] <= 'Z')) {
			return defaults, errors.New("keys.open_in_editor must be one ASCII letter, or empty to disable")
		}
		if strings.Contains("hjklqnNv", key) {
			return defaults, fmt.Errorf("keys.open_in_editor conflicts with built-in key %q", key)
		}
	}
	return config, nil
}
