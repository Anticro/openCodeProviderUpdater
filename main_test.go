package main

import (
	"bufio"
	"os"
	"testing"
)

func TestUIFlagDetection(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"opencode-provider-updater"}
	if isUIMode() {
		t.Error("expected false with no args")
	}

	os.Args = []string{"opencode-provider-updater", "-ui"}
	if !isUIMode() {
		t.Error("expected true with -ui")
	}

	os.Args = []string{"opencode-provider-updater", "--help"}
	if isUIMode() {
		t.Error("expected false with --help")
	}
}

func TestExistingFunctionsCompile(t *testing.T) {
	var (
		_ func(map[string]any, string, *bufio.Scanner)
		_ func(map[string]any, string, *bufio.Scanner)
		_ func(string) (map[string]any, error)
		_ func(string) string
		_ func(map[string]any)
		_ func(map[string]any, string)
	)
	_, _, _, _, _, _ = updateProvider, addProvider, fetchModels, sanitizeKey, listProviders, saveConfig
	_ = t
}
