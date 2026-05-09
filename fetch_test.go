package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchProviderModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/tags" {
			t.Errorf("expected /api/tags, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"models": [
				{"name": "gemma4:26b"},
				{"name": "llama3.2:3b"}
			]
		}`))
	}))
	defer server.Close()

	models, err := fetchProviderModels(server.URL + "/api/tags")
	if err != nil {
		t.Fatalf("fetchProviderModels failed: %v", err)
	}

	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}

	m1, ok := models["gemma4:26b"]
	if !ok {
		t.Fatal("expected model gemma4:26b")
	}
	m1map, ok := m1.(map[string]any)
	if !ok {
		t.Fatal("model value should be map[string]any")
	}
	if m1map["name"] != "gemma4:26b" {
		t.Errorf("model name = %v, want gemma4:26b", m1map["name"])
	}

	m2, ok := models["llama3.2:3b"]
	if !ok {
		t.Fatal("expected model llama3.2:3b")
	}
	m2map, ok := m2.(map[string]any)
	if !ok {
		t.Fatal("model value should be map[string]any")
	}
	if m2map["name"] != "llama3.2:3b" {
		t.Errorf("model name = %v, want llama3.2:3b", m2map["name"])
	}
}

func TestFetchProviderModelsHTTPError(t *testing.T) {
	_, err := fetchProviderModels("http://127.0.0.1:1/api/tags")
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

func TestFetchProviderModelsInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`this is not json`))
	}))
	defer server.Close()

	_, err := fetchProviderModels(server.URL + "/api/tags")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
