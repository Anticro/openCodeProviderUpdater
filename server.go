package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
)

//go:embed ui.html
var uiHTML string

type uiServer struct {
	configPath string
	server     *http.Server
}

func startUIServer(cfgPath string) error {
	s := &uiServer{configPath: cfgPath}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleUI)
	mux.HandleFunc("/api/providers", s.handleListProviders)
	mux.HandleFunc("/api/providers/update", s.handleUpdateProvider)
	mux.HandleFunc("/api/providers/add", s.handleAddProvider)
	mux.HandleFunc("/api/shutdown", s.handleShutdown)

	openBrowser("http://localhost:8099")

	s.server = &http.Server{
		Addr:    ":8099",
		Handler: mux,
	}

	fmt.Println("Server listening on :8099")
	return s.server.ListenAndServe()
}

func (s *uiServer) handleUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, uiHTML)
}

func (s *uiServer) handleListProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg, err := readConfigFile(s.configPath)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	providers := getProviderList(cfg)
	writeJSON(w, providers)
}

func (s *uiServer) handleUpdateProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Key == "" {
		writeError(w, "key is required", http.StatusBadRequest)
		return
	}

	cfg, err := readConfigFile(s.configPath)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedCfg, err := updateProviderModels(cfg, body.Key)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := writeConfigFile(s.configPath, updatedCfg); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	providers := getProviderList(updatedCfg)
	for _, p := range providers {
		if p.Key == body.Key {
			writeJSON(w, p)
			return
		}
	}

	writeError(w, "provider not found after update", http.StatusInternalServerError)
}

func (s *uiServer) handleAddProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		BaseURL string `json:"baseURL"`
		Name    string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.BaseURL == "" {
		writeError(w, "baseURL is required", http.StatusBadRequest)
		return
	}
	if body.Name == "" {
		writeError(w, "name is required", http.StatusBadRequest)
		return
	}

	cfg, err := readConfigFile(s.configPath)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedCfg, err := addNewProvider(cfg, body.BaseURL, body.Name)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := writeConfigFile(s.configPath, updatedCfg); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	providers := getProviderList(updatedCfg)
	key := sanitizeProviderKey(newURLHost(body.BaseURL))
	for _, p := range providers {
		if p.Key == key {
			writeJSON(w, p)
			return
		}
	}

	writeError(w, "new provider not found after add", http.StatusInternalServerError)
}

func newURLHost(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	return u.Host
}

func writeJSON(w http.ResponseWriter, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (s *uiServer) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"shutting down"}`))
	go func() {
		s.server.Shutdown(context.Background())
	}()
}

func openBrowser(url string) {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	default:
		return
	}
	exec.Command(cmd, url).Start()
}
