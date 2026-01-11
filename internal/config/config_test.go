package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultsVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pm-assist.yaml")
	if err := os.WriteFile(path, []byte("project:\n  name: test\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Version != CurrentSchemaVersion {
		t.Fatalf("expected version %d, got %d", CurrentSchemaVersion, cfg.Version)
	}
}

func TestLoadRejectsUnsupportedVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pm-assist.yaml")
	if err := os.WriteFile(path, []byte("version: 99\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Fatalf("expected error for unsupported version")
	}
}
