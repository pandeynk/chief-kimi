package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Worktree.Setup != "" {
		t.Errorf("expected empty setup, got %q", cfg.Worktree.Setup)
	}
	if cfg.OnComplete.Push {
		t.Error("expected Push to be false")
	}
	if cfg.OnComplete.CreatePR {
		t.Error("expected CreatePR to be false")
	}
}

func TestLoadNonExistent(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Worktree.Setup != "" {
		t.Errorf("expected empty setup, got %q", cfg.Worktree.Setup)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	cfg := &Config{
		Worktree: WorktreeConfig{
			Setup: "npm install",
		},
		OnComplete: OnCompleteConfig{
			Push:     true,
			CreatePR: true,
		},
	}

	if err := Save(dir, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Worktree.Setup != "npm install" {
		t.Errorf("expected setup %q, got %q", "npm install", loaded.Worktree.Setup)
	}
	if !loaded.OnComplete.Push {
		t.Error("expected Push to be true")
	}
	if !loaded.OnComplete.CreatePR {
		t.Error("expected CreatePR to be true")
	}
}

func TestAgentBinary_Default(t *testing.T) {
	cfg := Default()
	if cfg.AgentBinary() != "claude" {
		t.Errorf("expected default AgentBinary='claude', got %q", cfg.AgentBinary())
	}
}

func TestAgentBinary_Kimi(t *testing.T) {
	cfg := Default()
	cfg.Agent.Provider = "kimi"
	if cfg.AgentBinary() != "kimi" {
		t.Errorf("expected AgentBinary='kimi' when provider='kimi', got %q", cfg.AgentBinary())
	}
}

func TestAgentBinary_EmptyProviderIsDefault(t *testing.T) {
	cfg := Default()
	cfg.Agent.Provider = ""
	if cfg.AgentBinary() != "claude" {
		t.Errorf("expected AgentBinary='claude' for empty provider, got %q", cfg.AgentBinary())
	}
}

func TestSaveAndLoad_WithAgent(t *testing.T) {
	dir := t.TempDir()

	cfg := &Config{
		Agent: AgentConfig{Provider: "kimi"},
	}

	if err := Save(dir, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Agent.Provider != "kimi" {
		t.Errorf("expected Agent.Provider='kimi', got %q", loaded.Agent.Provider)
	}
	if loaded.AgentBinary() != "kimi" {
		t.Errorf("expected AgentBinary='kimi', got %q", loaded.AgentBinary())
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()

	if Exists(dir) {
		t.Error("expected Exists to return false for missing config")
	}

	// Create the config
	chiefDir := filepath.Join(dir, ".chief")
	if err := os.MkdirAll(chiefDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(chiefDir, "config.yaml"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	if !Exists(dir) {
		t.Error("expected Exists to return true for existing config")
	}
}
