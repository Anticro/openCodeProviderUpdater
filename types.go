package main

// ProviderInfo holds a provider's display information for the web UI.
// Fields:
//
//	Key     - The map key used in the config JSON's "provider" object
//	          (e.g. "ollama", "localhost_11434"). Used by the API to
//	          reference which provider to update.
//	Name    - The display name from the provider config (e.g. "Ollama (local)").
//	          Comes from the "name" field inside the provider object.
//	BaseURL - The "options.baseURL" field (e.g. "http://localhost:11434/v1").
//	          This is the OpenAI-compatible endpoint address.
//	Models  - Slice of installed model name strings
//	          (e.g. ["gemma4:26b", "llama3.2:3b"]).
//	          Extracted from the keys of the "models" map.
type ProviderInfo struct {
	Key     string   `json:"key"`
	Name    string   `json:"name"`
	BaseURL string   `json:"baseURL"`
	Models  []string `json:"models"`
}
