package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// configFilePath returns the absolute path to the OpenCode config file.
func configFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get home directory: %v\n", err)
		os.Exit(1)
	}
	return filepath.Join(home, ".config", "opencode", "opencode.json")
}

// readConfigFile reads and parses a JSON config from path.
func readConfigFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// readConfig reads the config from the default location.
func readConfig() (map[string]any, error) {
	return readConfigFile(configFilePath())
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// writeConfigFile saves cfg as JSON to path, creating a backup first.
func writeConfigFile(path string, cfg map[string]any) error {
	if fileExists(path) {
		backupPath := fmt.Sprintf("%s_%d", path, time.Now().Unix())
		if err := os.Rename(path, backupPath); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}
	data, err := json.MarshalIndent(cfg, "", "   ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// writeConfig saves cfg to the default config location.
func writeConfig(cfg map[string]any) error {
	return writeConfigFile(configFilePath(), cfg)
}
