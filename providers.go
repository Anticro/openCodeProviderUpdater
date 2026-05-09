package main

import (
	"fmt"
	"net/url"
)

// getProviderList extracts a structured list of all providers from the config map.
func getProviderList(cfg map[string]any) []ProviderInfo {
	providerRaw, ok := cfg["provider"]
	if !ok {
		return []ProviderInfo{}
	}

	provider, ok := providerRaw.(map[string]any)
	if !ok || len(provider) == 0 {
		return []ProviderInfo{}
	}

	var result []ProviderInfo
	for key, provData := range provider {
		prov, ok := provData.(map[string]any)
		if !ok {
			continue
		}

		name, _ := prov["name"].(string)

		baseURL := ""
		if optionsRaw, ok := prov["options"]; ok {
			if options, ok := optionsRaw.(map[string]any); ok {
				baseURL, _ = options["baseURL"].(string)
			}
		}

		var models []string
		if modelsRaw, ok := prov["models"]; ok {
			if modelMap, ok := modelsRaw.(map[string]any); ok {
				for m := range modelMap {
					models = append(models, m)
				}
			}
		}

		result = append(result, ProviderInfo{
			Key:     key,
			Name:    name,
			BaseURL: baseURL,
			Models:  models,
		})
	}

	return result
}

// updateProviderModels fetches fresh models for a provider and updates the config.
func updateProviderModels(cfg map[string]any, key string) (map[string]any, error) {
	providerRaw, ok := cfg["provider"]
	if !ok {
		return nil, fmt.Errorf("no provider section")
	}

	provider, ok := providerRaw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid provider section")
	}

	prov, ok := provider[key].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", key)
	}

	optionsRaw, ok := prov["options"]
	if !ok {
		return nil, fmt.Errorf("no options")
	}

	options, ok := optionsRaw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid options")
	}

	baseURL, ok := options["baseURL"].(string)
	if !ok {
		return nil, fmt.Errorf("no baseURL")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse baseURL: %w", err)
	}

	apiURL := fmt.Sprintf("%s://%s/api/tags", u.Scheme, u.Host)
	models, err := fetchProviderModels(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}

	prov["models"] = models
	provider[key] = prov
	cfg["provider"] = provider
	return cfg, nil
}

// addNewProvider adds a new provider entry to the config and returns the updated config.
func addNewProvider(cfg map[string]any, baseURL, name string) (map[string]any, error) {
	apiURL := baseURL + "/api/tags"
	models, err := fetchProviderModels(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse baseURL: %w", err)
	}

	key := sanitizeProviderKey(u.Host)

	providerRaw, ok := cfg["provider"]
	var provider map[string]any
	if ok {
		provider, _ = providerRaw.(map[string]any)
	}
	if provider == nil {
		provider = make(map[string]any)
	}

	provider[key] = map[string]any{
		"models": models,
		"name":   name,
		"npm":    "@ai-sdk/openai-compatible",
		"options": map[string]any{
			"baseURL": baseURL + "/v1",
		},
	}

	cfg["provider"] = provider
	return cfg, nil
}
