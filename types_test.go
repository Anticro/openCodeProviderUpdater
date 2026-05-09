package main

import (
	"encoding/json"
	"testing"
)

func TestProviderInfoJSONRoundTrip(t *testing.T) {
	p := ProviderInfo{
		Key:     "ollama",
		Name:    "Ollama (local)",
		BaseURL: "http://localhost:11434/v1",
		Models:  []string{"gemma4:26b", "llama3.2:3b"},
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var p2 ProviderInfo
	if err := json.Unmarshal(data, &p2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if p2.Key != "ollama" {
		t.Errorf("Key = %q, want %q", p2.Key, "ollama")
	}
	if p2.Name != "Ollama (local)" {
		t.Errorf("Name = %q, want %q", p2.Name, "Ollama (local)")
	}
	if p2.BaseURL != "http://localhost:11434/v1" {
		t.Errorf("BaseURL = %q, want %q", p2.BaseURL, "http://localhost:11434/v1")
	}
	if len(p2.Models) != 2 || p2.Models[0] != "gemma4:26b" {
		t.Errorf("Models = %v, want [gemma4:26b llama3.2:3b]", p2.Models)
	}
}
