package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigFilePath(t *testing.T) {
	path := configFilePath()
	if path == "" {
		t.Fatal("configFilePath returned empty string")
	}
	if filepath.Base(path) != "opencode.json" {
		t.Errorf("expected opencode.json, got %s", filepath.Base(path))
	}
}

func TestReadWriteConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.json")

	initial := map[string]any{
		"test":   "value",
		"number": float64(42),
	}

	if err := writeConfigFile(path, initial); err != nil {
		t.Fatalf("writeConfigFile failed: %v", err)
	}

	cfg, err := readConfigFile(path)
	if err != nil {
		t.Fatalf("readConfigFile failed: %v", err)
	}

	if cfg["test"] != "value" {
		t.Errorf("test field = %v, want 'value'", cfg["test"])
	}
	if cfg["number"] != float64(42) {
		t.Errorf("number field = %v, want 42", cfg["number"])
	}

	updated := map[string]any{
		"new": "data",
	}
	if err := writeConfigFile(path, updated); err != nil {
		t.Fatalf("second writeConfigFile failed: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	foundBackup := false
	for _, e := range entries {
		if e.Name() != "opencode.json" && e.Name() != "opencode.json_" {
			foundBackup = true
		}
	}
	if !foundBackup {
		t.Error("no backup file found after second write")
	}

	cfg, err = readConfigFile(path)
	if err != nil {
		t.Fatalf("readConfigFile after update failed: %v", err)
	}
	if cfg["new"] != "data" {
		t.Errorf("expected 'data', got %v", cfg["new"])
	}
}

func TestReadConfigFileNotFound(t *testing.T) {
	_, err := readConfigFile("/nonexistent/path/config.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
