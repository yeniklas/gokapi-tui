package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadFrom_Valid(t *testing.T) {
	path := writeConfig(t, `
server_url: "http://gokapi.example.com"
api_key: "secret123"
defaults:
  allowed_downloads: 5
  expiry_days: 30
  password: "hunter2"
`)
	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ServerURL != "http://gokapi.example.com" {
		t.Errorf("ServerURL = %q, want %q", cfg.ServerURL, "http://gokapi.example.com")
	}
	if cfg.APIKey != "secret123" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "secret123")
	}
	if cfg.Defaults.AllowedDownloads != 5 {
		t.Errorf("AllowedDownloads = %d, want 5", cfg.Defaults.AllowedDownloads)
	}
	if cfg.Defaults.ExpiryDays != 30 {
		t.Errorf("ExpiryDays = %d, want 30", cfg.Defaults.ExpiryDays)
	}
	if cfg.Defaults.Password != "hunter2" {
		t.Errorf("Password = %q, want %q", cfg.Defaults.Password, "hunter2")
	}
}

func TestLoadFrom_DefaultsFilled(t *testing.T) {
	path := writeConfig(t, `
server_url: "http://gokapi.example.com"
api_key: "secret123"
`)
	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Defaults.AllowedDownloads != 1 {
		t.Errorf("AllowedDownloads = %d, want 1 (default)", cfg.Defaults.AllowedDownloads)
	}
	if cfg.Defaults.ExpiryDays != 7 {
		t.Errorf("ExpiryDays = %d, want 7 (default)", cfg.Defaults.ExpiryDays)
	}
}

func TestLoadFrom_MissingServerURL(t *testing.T) {
	path := writeConfig(t, `api_key: "secret123"`)
	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for missing server_url, got nil")
	}
}

func TestLoadFrom_MissingAPIKey(t *testing.T) {
	path := writeConfig(t, `server_url: "http://gokapi.example.com"`)
	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for missing api_key, got nil")
	}
}

func TestLoadFrom_InvalidYAML(t *testing.T) {
	path := writeConfig(t, `{this is not: yaml: [`)
	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoadFrom_FileNotFound(t *testing.T) {
	_, err := LoadFrom("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
