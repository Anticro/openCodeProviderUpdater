package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeTestConfig() map[string]any {
	return map[string]any{
		"provider": map[string]any{
			"ollama": map[string]any{
				"name": "Ollama (local)",
				"npm":  "@ai-sdk/openai-compatible",
				"options": map[string]any{
					"baseURL": "http://localhost:11434/v1",
				},
				"models": map[string]any{
					"gemma4:26b":  map[string]any{"name": "gemma4:26b"},
					"llama3.2:3b": map[string]any{"name": "llama3.2:3b"},
				},
			},
			"remote_server": map[string]any{
				"name": "Remote Server",
				"npm":  "@ai-sdk/openai-compatible",
				"options": map[string]any{
					"baseURL": "http://192.168.1.50:11434/v1",
				},
				"models": map[string]any{
					"deepseek-r1:7b": map[string]any{"name": "deepseek-r1:7b"},
				},
			},
		},
	}
}

func TestGetProviderList(t *testing.T) {
	cfg := makeTestConfig()
	providers := getProviderList(cfg)

	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}

	var ollama, remote *ProviderInfo
	for i := range providers {
		switch providers[i].Key {
		case "ollama":
			ollama = &providers[i]
		case "remote_server":
			remote = &providers[i]
		}
	}

	if ollama == nil {
		t.Fatal("ollama provider not found")
	}
	if ollama.Name != "Ollama (local)" {
		t.Errorf("ollama name = %q", ollama.Name)
	}
	if ollama.BaseURL != "http://localhost:11434/v1" {
		t.Errorf("ollama baseURL = %q", ollama.BaseURL)
	}
	if len(ollama.Models) != 2 {
		t.Errorf("ollama models count = %d, want 2", len(ollama.Models))
	}

	if remote == nil {
		t.Fatal("remote_server provider not found")
	}
	if len(remote.Models) != 1 || remote.Models[0] != "deepseek-r1:7b" {
		t.Errorf("remote models = %v", remote.Models)
	}
}

func TestGetProviderListEmpty(t *testing.T) {
	providers := getProviderList(map[string]any{})
	if len(providers) != 0 {
		t.Errorf("expected 0 providers, got %d", len(providers))
	}

	providers = getProviderList(map[string]any{
		"provider": map[string]any{},
	})
	if len(providers) != 0 {
		t.Errorf("expected 0 providers for empty section, got %d", len(providers))
	}
}

func TestUpdateProviderModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"models": [
				{"name": "new-model-1"},
				{"name": "new-model-2"},
				{"name": "new-model-3"}
			]
		}`))
	}))
	defer server.Close()

	cfg := makeTestConfig()
	provider := cfg["provider"].(map[string]any)
	ollama := provider["ollama"].(map[string]any)
	options := ollama["options"].(map[string]any)
	options["baseURL"] = server.URL + "/v1"

	updatedCfg, err := updateProviderModels(cfg, "ollama")
	if err != nil {
		t.Fatalf("updateProviderModels failed: %v", err)
	}

	updatedProvider := updatedCfg["provider"].(map[string]any)
	updatedOllama := updatedProvider["ollama"].(map[string]any)
	updatedModels := updatedOllama["models"].(map[string]any)

	if len(updatedModels) != 3 {
		t.Errorf("expected 3 models after update, got %d", len(updatedModels))
	}

	if _, ok := updatedModels["new-model-1"]; !ok {
		t.Error("expected new-model-1 after update")
	}
	if _, ok := updatedModels["gemma4:26b"]; ok {
		t.Error("gemma4:26b should not be in updated models")
	}
}

func TestUpdateProviderModelsNotFound(t *testing.T) {
	cfg := makeTestConfig()
	_, err := updateProviderModels(cfg, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

func TestAddNewProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"models": [
				{"name": "phi4:14b"}
			]
		}`))
	}))
	defer server.Close()

	cfg := makeTestConfig()
	updatedCfg, err := addNewProvider(cfg, server.URL, "Test Server")
	if err != nil {
		t.Fatalf("addNewProvider failed: %v", err)
	}

	providers := getProviderList(updatedCfg)
	if len(providers) != 3 {
		t.Fatalf("expected 3 providers after add, got %d", len(providers))
	}

	var found bool
	for _, p := range providers {
		if p.Name == "Test Server" {
			found = true
			if len(p.Models) != 1 || p.Models[0] != "phi4:14b" {
				t.Errorf("new provider models = %v", p.Models)
			}
		}
	}
	if !found {
		t.Error("new provider not found in list")
	}
}

func TestAddNewProviderEmptyConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"models": [{"name": "test-model"}]}`))
	}))
	defer server.Close()

	cfg := map[string]any{}
	updatedCfg, err := addNewProvider(cfg, server.URL, "Fresh Server")
	if err != nil {
		t.Fatalf("addNewProvider on empty config failed: %v", err)
	}

	providers := getProviderList(updatedCfg)
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}
	if providers[0].Name != "Fresh Server" {
		t.Errorf("name = %q", providers[0].Name)
	}
}
