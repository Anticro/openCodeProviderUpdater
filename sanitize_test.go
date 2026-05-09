package main

import "testing"

func TestSanitizeProviderKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"localhost:11434", "localhost_11434"},
		{"192.168.1.50:11434", "192_168_1_50_11434"},
		{"example.com", "example_com"},
		{"valid-host_123", "valid-host_123"},
		{"", ""},
		{"!!!@#$", "______"},
		{"127.0.0.1:8080", "127_0_0_1_8080"},
		{"Ollama-Server_1", "Ollama-Server_1"},
	}

	for _, tt := range tests {
		got := sanitizeProviderKey(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeProviderKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
