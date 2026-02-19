package tui

import (
	"strings"
	"testing"

	"github.com/minicodemonkey/chief/internal/config"
)

func TestSettingsOverlay_LoadFromConfig(t *testing.T) {
	s := NewSettingsOverlay()
	cfg := &config.Config{
		Worktree: config.WorktreeConfig{
			Setup: "npm install",
		},
		OnComplete: config.OnCompleteConfig{
			Push:     true,
			CreatePR: false,
		},
	}
	s.LoadFromConfig(cfg)

	if len(s.items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(s.items))
	}
	if s.items[0].Key != "agent.provider" {
		t.Errorf("agent.provider item: got key=%s", s.items[0].Key)
	}
	if s.items[1].Key != "worktree.setup" || s.items[1].StringVal != "npm install" {
		t.Errorf("worktree.setup item: got key=%s val=%s", s.items[1].Key, s.items[1].StringVal)
	}
	if s.items[2].Key != "onComplete.push" || !s.items[2].BoolVal {
		t.Errorf("onComplete.push item: got key=%s val=%v", s.items[2].Key, s.items[2].BoolVal)
	}
	if s.items[3].Key != "onComplete.createPR" || s.items[3].BoolVal {
		t.Errorf("onComplete.createPR item: got key=%s val=%v", s.items[3].Key, s.items[3].BoolVal)
	}
	if s.selectedIndex != 0 {
		t.Errorf("expected selectedIndex=0, got %d", s.selectedIndex)
	}
}

func TestSettingsOverlay_ApplyToConfig(t *testing.T) {
	s := NewSettingsOverlay()
	cfg := config.Default()
	s.LoadFromConfig(cfg)

	// Modify items (index 0=agent.provider, 1=worktree.setup, 2=push, 3=createPR)
	s.items[1].StringVal = "go mod download"
	s.items[2].BoolVal = true
	s.items[3].BoolVal = true

	resultCfg := config.Default()
	s.ApplyToConfig(resultCfg)

	if resultCfg.Worktree.Setup != "go mod download" {
		t.Errorf("expected setup='go mod download', got '%s'", resultCfg.Worktree.Setup)
	}
	if !resultCfg.OnComplete.Push {
		t.Error("expected push=true")
	}
	if !resultCfg.OnComplete.CreatePR {
		t.Error("expected createPR=true")
	}
}

func TestSettingsOverlay_Navigation(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())

	if s.selectedIndex != 0 {
		t.Fatalf("expected initial index=0, got %d", s.selectedIndex)
	}

	s.MoveDown()
	if s.selectedIndex != 1 {
		t.Errorf("expected index=1 after MoveDown, got %d", s.selectedIndex)
	}

	s.MoveDown()
	if s.selectedIndex != 2 {
		t.Errorf("expected index=2 after second MoveDown, got %d", s.selectedIndex)
	}

	s.MoveDown()
	if s.selectedIndex != 3 {
		t.Errorf("expected index=3 after third MoveDown, got %d", s.selectedIndex)
	}

	// Can't go beyond last item (4 items total, last index=3)
	s.MoveDown()
	if s.selectedIndex != 3 {
		t.Errorf("expected index=3 (clamped), got %d", s.selectedIndex)
	}

	s.MoveUp()
	if s.selectedIndex != 2 {
		t.Errorf("expected index=2 after MoveUp, got %d", s.selectedIndex)
	}

	// Can't go before first item
	s.MoveUp()
	s.MoveUp()
	s.MoveUp()
	if s.selectedIndex != 0 {
		t.Errorf("expected index=0 (clamped), got %d", s.selectedIndex)
	}
}

func TestSettingsOverlay_ToggleBool(t *testing.T) {
	s := NewSettingsOverlay()
	cfg := &config.Config{
		OnComplete: config.OnCompleteConfig{Push: false},
	}
	s.LoadFromConfig(cfg)

	// Select "Push to remote" (index 2: agent.provider, worktree.setup, onComplete.push)
	s.MoveDown()
	s.MoveDown()

	key, val := s.ToggleBool()
	if key != "onComplete.push" {
		t.Errorf("expected key='onComplete.push', got '%s'", key)
	}
	if !val {
		t.Error("expected val=true after toggle")
	}

	// Toggle back
	key, val = s.ToggleBool()
	if val {
		t.Error("expected val=false after second toggle")
	}
	_ = key
}

func TestSettingsOverlay_ToggleBool_OnStringItem(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())

	// Selected item is "Setup command" (string type)
	key, _ := s.ToggleBool()
	if key != "" {
		t.Errorf("expected empty key for string item toggle, got '%s'", key)
	}
}

func TestSettingsOverlay_RevertToggle(t *testing.T) {
	s := NewSettingsOverlay()
	cfg := &config.Config{
		OnComplete: config.OnCompleteConfig{Push: false},
	}
	s.LoadFromConfig(cfg)

	s.MoveDown()
	s.MoveDown() // Select "Push to remote" (index 2)
	s.ToggleBool()
	if !s.items[2].BoolVal {
		t.Fatal("expected true after toggle")
	}

	s.RevertToggle()
	if s.items[2].BoolVal {
		t.Error("expected false after revert")
	}
}

func TestSettingsOverlay_StringEditing(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())

	// Selected item is "Provider" (agent.provider, index 0)
	if s.IsEditing() {
		t.Fatal("should not be editing initially")
	}

	s.StartEditing()
	if !s.IsEditing() {
		t.Fatal("should be editing after StartEditing")
	}
	if s.editBuffer != "" {
		t.Errorf("expected empty edit buffer, got '%s'", s.editBuffer)
	}

	s.AddEditChar('k')
	s.AddEditChar('i')
	s.AddEditChar('m')
	s.AddEditChar('i')
	if s.editBuffer != "kimi" {
		t.Errorf("expected 'kimi', got '%s'", s.editBuffer)
	}

	s.DeleteEditChar()
	if s.editBuffer != "kim" {
		t.Errorf("expected 'kim' after delete, got '%s'", s.editBuffer)
	}

	s.ConfirmEdit()
	if s.IsEditing() {
		t.Fatal("should not be editing after ConfirmEdit")
	}
	if s.items[0].StringVal != "kim" {
		t.Errorf("expected StringVal='kim', got '%s'", s.items[0].StringVal)
	}
}

func TestSettingsOverlay_CancelEdit(t *testing.T) {
	s := NewSettingsOverlay()
	cfg := &config.Config{
		Worktree: config.WorktreeConfig{Setup: "original"},
	}
	s.LoadFromConfig(cfg)

	// worktree.setup is at index 1; move down to select it
	s.MoveDown()
	s.StartEditing()
	s.AddEditChar('x')
	s.CancelEdit()

	if s.IsEditing() {
		t.Fatal("should not be editing after CancelEdit")
	}
	if s.items[1].StringVal != "original" {
		t.Errorf("expected 'original' preserved, got '%s'", s.items[1].StringVal)
	}
}

func TestSettingsOverlay_StartEditingOnBoolItem(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	s.MoveDown()
	s.MoveDown() // Select "Push to remote" (bool, index 2)

	s.StartEditing()
	if s.IsEditing() {
		t.Error("should not start editing on a bool item")
	}
}

func TestSettingsOverlay_GHError(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())

	if s.HasGHError() {
		t.Fatal("should not have GH error initially")
	}

	s.SetGHError("gh not found")
	if !s.HasGHError() {
		t.Fatal("should have GH error after SetGHError")
	}

	s.DismissGHError()
	if s.HasGHError() {
		t.Fatal("should not have GH error after dismiss")
	}
}

func TestSettingsOverlay_Render(t *testing.T) {
	s := NewSettingsOverlay()
	cfg := &config.Config{
		Worktree: config.WorktreeConfig{Setup: "npm install"},
		OnComplete: config.OnCompleteConfig{
			Push:     true,
			CreatePR: false,
		},
	}
	s.LoadFromConfig(cfg)
	s.SetSize(80, 24)

	rendered := s.Render()

	// Check header
	if !strings.Contains(rendered, "Settings") {
		t.Error("expected 'Settings' in header")
	}
	if !strings.Contains(rendered, ".chief/config.yaml") {
		t.Error("expected config path in header")
	}

	// Check section headers
	if !strings.Contains(rendered, "Agent") {
		t.Error("expected 'Agent' section")
	}
	if !strings.Contains(rendered, "Worktree") {
		t.Error("expected 'Worktree' section")
	}
	if !strings.Contains(rendered, "On Complete") {
		t.Error("expected 'On Complete' section")
	}

	// Check values
	if !strings.Contains(rendered, "npm install") {
		t.Error("expected 'npm install' value")
	}
	if !strings.Contains(rendered, "Yes") {
		t.Error("expected 'Yes' for push")
	}
	if !strings.Contains(rendered, "No") {
		t.Error("expected 'No' for createPR")
	}

	// Check footer
	if !strings.Contains(rendered, "Esc: close") {
		t.Error("expected 'Esc: close' in footer")
	}
}

func TestSettingsOverlay_RenderGHError(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	s.SetSize(80, 24)

	s.SetGHError("gh not found")
	rendered := s.Render()

	if !strings.Contains(rendered, "GitHub CLI Error") {
		t.Error("expected 'GitHub CLI Error' in rendered output")
	}
	if !strings.Contains(rendered, "gh not found") {
		t.Error("expected error message in rendered output")
	}
	if !strings.Contains(rendered, "Press any key to dismiss") {
		t.Error("expected dismiss hint in footer")
	}
}

func TestSettingsOverlay_RenderEditing(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	s.SetSize(80, 24)

	s.StartEditing()
	rendered := s.Render()

	if !strings.Contains(rendered, "Enter: save") {
		t.Error("expected 'Enter: save' in footer during editing")
	}
}

func TestSettingsOverlay_RenderSelectedIndicator(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	s.SetSize(80, 24)

	rendered := s.Render()

	// The selected item should have a ">" indicator
	if !strings.Contains(rendered, ">") {
		t.Error("expected '>' cursor indicator for selected item")
	}
}

func TestSettingsOverlay_RenderEmptyStringValue(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	s.SetSize(80, 24)

	rendered := s.Render()

	if !strings.Contains(rendered, "(not set)") {
		t.Error("expected '(not set)' for empty setup command")
	}
}

func TestSettingsOverlay_GetSelectedItem(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())

	item := s.GetSelectedItem()
	if item == nil {
		t.Fatal("expected non-nil selected item")
	}
	if item.Key != "agent.provider" {
		t.Errorf("expected first item key='agent.provider', got '%s'", item.Key)
	}

	s.MoveDown()
	item = s.GetSelectedItem()
	if item.Key != "worktree.setup" {
		t.Errorf("expected second item key='worktree.setup', got '%s'", item.Key)
	}
}
