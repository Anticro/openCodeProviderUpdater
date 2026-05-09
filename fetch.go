package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// fetchProviderModels calls an Ollama-compatible /api/tags endpoint
// and returns a map of model name -> {"name": modelName}.
func fetchProviderModels(apiURL string) (map[string]any, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	models := make(map[string]any, len(result.Models))
	for _, m := range result.Models {
		models[m.Name] = map[string]any{
			"name": m.Name,
		}
	}

	return models, nil
}
