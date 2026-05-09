package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleListProviders(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "opencode.json")

	initialCfg := map[string]any{
		"provider": map[string]any{
			"ollama": map[string]any{
				"name": "Ollama (local)",
				"npm":  "@ai-sdk/openai-compatible",
				"options": map[string]any{
					"baseURL": "http://localhost:11434/v1",
				},
				"models": map[string]any{
					"gemma4:26b": map[string]any{"name": "gemma4:26b"},
				},
			},
		},
	}
	if err := writeConfigFile(cfgPath, initialCfg); err != nil {
		t.Fatal(err)
	}

	s := &uiServer{configPath: cfgPath}
	req := httptest.NewRequest("GET", "/api/providers", nil)
	rec := httptest.NewRecorder()
	s.handleListProviders(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var providers []ProviderInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &providers); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}
	if providers[0].Key != "ollama" {
		t.Errorf("key = %q", providers[0].Key)
	}
}

func TestHandleListProvidersEmpty(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "opencode.json")

	if err := writeConfigFile(cfgPath, map[string]any{}); err != nil {
		t.Fatal(err)
	}

	s := &uiServer{configPath: cfgPath}
	req := httptest.NewRequest("GET", "/api/providers", nil)
	rec := httptest.NewRecorder()
	s.handleListProviders(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var providers []ProviderInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &providers); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(providers) != 0 {
		t.Errorf("expected 0 providers, got %d", len(providers))
	}
}

func TestHandleUpdateProvider(t *testing.T) {
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"models": [{"name": "updated-model"}]}`))
	}))
	defer mockOllama.Close()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "opencode.json")

	initialCfg := map[string]any{
		"provider": map[string]any{
			"ollama": map[string]any{
				"name": "Ollama (local)",
				"npm":  "@ai-sdk/openai-compatible",
				"options": map[string]any{
					"baseURL": mockOllama.URL + "/v1",
				},
				"models": map[string]any{
					"gemma4:26b": map[string]any{"name": "gemma4:26b"},
				},
			},
		},
	}
	if err := writeConfigFile(cfgPath, initialCfg); err != nil {
		t.Fatal(err)
	}

	s := &uiServer{configPath: cfgPath}
	body := strings.NewReader(`{"key": "ollama"}`)
	req := httptest.NewRequest("POST", "/api/providers/update", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.handleUpdateProvider(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var updated ProviderInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if updated.Key != "ollama" {
		t.Errorf("key = %q", updated.Key)
	}
	if len(updated.Models) != 1 || updated.Models[0] != "updated-model" {
		t.Errorf("models = %v, want [updated-model]", updated.Models)
	}

	cfg, err := readConfigFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	providers := getProviderList(cfg)
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider in saved file, got %d", len(providers))
	}
	if len(providers[0].Models) != 1 || providers[0].Models[0] != "updated-model" {
		t.Errorf("saved models = %v", providers[0].Models)
	}
}

func TestHandleUpdateProviderInvalidMethod(t *testing.T) {
	s := &uiServer{configPath: "/tmp/test.json"}
	req := httptest.NewRequest("GET", "/api/providers/update", nil)
	rec := httptest.NewRecorder()
	s.handleUpdateProvider(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleAddProvider(t *testing.T) {
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"models": [{"name": "new-model"}]}`))
	}))
	defer mockOllama.Close()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "opencode.json")

	if err := writeConfigFile(cfgPath, map[string]any{}); err != nil {
		t.Fatal(err)
	}

	s := &uiServer{configPath: cfgPath}
	body := strings.NewReader(`{"baseURL": "` + mockOllama.URL + `", "name": "Test Server"}`)
	req := httptest.NewRequest("POST", "/api/providers/add", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.handleAddProvider(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var added ProviderInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &added); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if added.Name != "Test Server" {
		t.Errorf("name = %q", added.Name)
	}
	if len(added.Models) != 1 || added.Models[0] != "new-model" {
		t.Errorf("models = %v", added.Models)
	}

	cfg, err := readConfigFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	providers := getProviderList(cfg)
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider in saved file, got %d", len(providers))
	}
}

func TestHandleUI(t *testing.T) {
	if uiHTML == "" {
		t.Fatal("uiHTML is empty — ui.html may not be present")
	}

	s := &uiServer{configPath: "/tmp/test.json"}
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	s.handleUI(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("expected text/html content type")
	}
	if !strings.Contains(rec.Body.String(), "OpenCode Provider Updater") {
		t.Error("HTML should contain the title")
	}
}

func TestWriteJSONHelper(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"hello": "world"}
	writeJSON(rec, data)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json, got %s", rec.Header().Get("Content-Type"))
	}

	var result map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["hello"] != "world" {
		t.Errorf("got %v", result)
	}
}

func TestWriteErrorHelper(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, "something went wrong", http.StatusBadRequest)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json")
	}

	var result map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["error"] != "something went wrong" {
		t.Errorf("got error = %q", result["error"])
	}
}
