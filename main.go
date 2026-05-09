package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func isUIMode() bool {
	return len(os.Args) > 1 && os.Args[1] == "-ui"
}

func main() {
	if isUIMode() {
		cfgPath := configFilePath()
		fmt.Println("Starting web UI at http://localhost:8099")
		if err := startUIServer(cfgPath); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	configPath := filepath.Join(home, ".config", "opencode", "opencode.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", configPath, err)
		os.Exit(1)
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("1. Update an existing provider")
	fmt.Println("2. Add a new provider")
	fmt.Println("3. List all providers and models")
	fmt.Print("Choose an option (1-3): ")

	scanner.Scan()
	choice, _ := strconv.Atoi(scanner.Text())

	switch choice {
	case 1:
		updateProvider(config, configPath, scanner)
	case 2:
		addProvider(config, configPath, scanner)
	case 3:
		listProviders(config)
	default:
		fmt.Println("Invalid choice")
		os.Exit(1)
	}
}

func updateProvider(config map[string]any, configPath string, scanner *bufio.Scanner) {
	providerRaw, ok := config["provider"]
	if !ok {
		fmt.Println("No provider object found in config")
		os.Exit(1)
	}

	provider, ok := providerRaw.(map[string]any)
	if !ok || len(provider) == 0 {
		fmt.Println("No providers found")
		os.Exit(1)
	}

	keys := make([]string, 0, len(provider))
	for k := range provider {
		keys = append(keys, k)
	}

	fmt.Println("Available providers:")
	for i, k := range keys {
		fmt.Printf("%d. %s\n", i+1, k)
	}
	fmt.Print("Choose a provider (1-based, 0 to abort): ")

	scanner.Scan()
	choice, _ := strconv.Atoi(scanner.Text())

	if choice == 0 {
		return
	}
	if choice < 1 || choice > len(keys) {
		fmt.Println("Invalid choice")
		os.Exit(1)
	}

	selectedKey := keys[choice-1]
	selectedProvider, ok := provider[selectedKey].(map[string]any)
	if !ok {
		fmt.Println("Invalid provider data")
		os.Exit(1)
	}

	optionsRaw, ok := selectedProvider["options"]
	if !ok {
		fmt.Println("No options found for provider")
		os.Exit(1)
	}
	options, ok := optionsRaw.(map[string]any)
	if !ok {
		fmt.Println("Invalid options data")
		os.Exit(1)
	}

	baseURL, ok := options["baseURL"].(string)
	if !ok {
		fmt.Println("No baseURL in options")
		os.Exit(1)
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing baseURL: %v\n", err)
		os.Exit(1)
	}
	apiURL := fmt.Sprintf("%s://%s/api/tags", u.Scheme, u.Host)

	models, err := fetchModels(apiURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching models from %s: %v\n", apiURL, err)
		os.Exit(1)
	}

	selectedProvider["models"] = models

	saveConfig(config, configPath)
}

func addProvider(config map[string]any, configPath string, scanner *bufio.Scanner) {
	fmt.Print("Enter base URL (e.g. http://localhost:11434): ")
	scanner.Scan()
	baseURL := strings.TrimRight(scanner.Text(), "/")

	fmt.Print("Enter provider name (e.g. Ollama local): ")
	scanner.Scan()
	name := scanner.Text()

	apiURL := baseURL + "/api/tags"

	models, err := fetchModels(apiURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching models from %s: %v\n", apiURL, err)
		os.Exit(1)
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing base URL: %v\n", err)
		os.Exit(1)
	}
	key := sanitizeKey(u.Host)

	providerRaw, ok := config["provider"]
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

	config["provider"] = provider

	saveConfig(config, configPath)
}

func fetchModels(apiURL string) (map[string]any, error) {
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

func sanitizeKey(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

func listProviders(config map[string]any) {
	providerRaw, ok := config["provider"]
	if !ok {
		fmt.Println("No provider object found in config")
		return
	}

	provider, ok := providerRaw.(map[string]any)
	if !ok || len(provider) == 0 {
		fmt.Println("No providers found")
		return
	}

	fmt.Println("Providers and models:")
	for key, provData := range provider {
		prov, ok := provData.(map[string]any)
		if !ok {
			fmt.Printf("  %s: invalid provider data\n", key)
			continue
		}

		name, _ := prov["name"].(string)
		fmt.Printf("\n  Provider: %s\n", key)
		if name != "" {
			fmt.Printf("  Name: %s\n", name)
		}

		modelsRaw, ok := prov["models"]
		if !ok {
			fmt.Println("  Models: none")
			continue
		}

		models, ok := modelsRaw.(map[string]any)
		if !ok {
			fmt.Println("  Models: invalid data")
			continue
		}

		if len(models) == 0 {
			fmt.Println("  Models: none")
			continue
		}

		modelNames := make([]string, 0, len(models))
		for m := range models {
			modelNames = append(modelNames, m)
		}

		fmt.Printf("  Models (%d):\n", len(models))
		for _, m := range modelNames {
			fmt.Printf("    - %s\n", m)
		}
	}
}

func saveConfig(config map[string]any, configPath string) {
	backupPath := configPath + "_" + strconv.FormatInt(time.Now().Unix(), 10)

	input, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config for backup: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(backupPath, input, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing backup: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Backup saved to %s\n", backupPath)

	data, err := json.MarshalIndent(config, "", "   ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling config: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(data))
}
