// Package cmd provides CLI command implementations for Chief.
// This includes new, edit, status, and list commands that can be
// run from the command line without launching the full TUI.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/minicodemonkey/chief/embed"
	"github.com/minicodemonkey/chief/internal/config"
	"github.com/minicodemonkey/chief/internal/prd"
)

// NewOptions contains configuration for the new command.
type NewOptions struct {
	Name    string // PRD name (default: "main")
	Context string // Optional context to pass to Claude
	BaseDir string // Base directory for .chief/prds/ (default: current directory)
}

// RunNew creates a new PRD by launching an interactive agent session.
func RunNew(opts NewOptions) error {
	// Set defaults
	if opts.Name == "" {
		opts.Name = "main"
	}
	if opts.BaseDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		opts.BaseDir = cwd
	}

	// Validate name (alphanumeric, -, _)
	if !isValidPRDName(opts.Name) {
		return fmt.Errorf("invalid PRD name %q: must contain only letters, numbers, hyphens, and underscores", opts.Name)
	}

	// Load project config to determine agent binary
	cfg, err := config.Load(opts.BaseDir)
	if err != nil {
		cfg = config.Default()
	}
	agentBin := cfg.AgentBinary()

	// Create directory structure: .chief/prds/<name>/
	prdDir := filepath.Join(opts.BaseDir, ".chief", "prds", opts.Name)
	if err := os.MkdirAll(prdDir, 0755); err != nil {
		return fmt.Errorf("failed to create PRD directory: %w", err)
	}

	// Check if prd.md already exists
	prdMdPath := filepath.Join(prdDir, "prd.md")
	if _, err := os.Stat(prdMdPath); err == nil {
		return fmt.Errorf("PRD already exists at %s. Use 'chief edit %s' to modify it", prdMdPath, opts.Name)
	}

	// Get the init prompt with the PRD directory path
	prompt := embed.GetInitPrompt(prdDir, opts.Context)

	// Launch interactive agent session
	fmt.Printf("Creating PRD in %s...\n", prdDir)
	fmt.Printf("Launching %s to help you create your PRD...\n", agentBin)
	fmt.Println()

	if err := runInteractiveAgent(opts.BaseDir, prompt, agentBin); err != nil {
		return fmt.Errorf("%s session failed: %w", agentBin, err)
	}

	// Check if prd.md was created
	if _, err := os.Stat(prdMdPath); os.IsNotExist(err) {
		fmt.Println("\nNo prd.md was created. Run 'chief new' again to try again.")
		return nil
	}

	fmt.Println("\nPRD created successfully!")

	// Run conversion from prd.md to prd.json
	if err := RunConvertWithOptions(ConvertOptions{PRDDir: prdDir, AgentBin: agentBin}); err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	fmt.Printf("\nYour PRD is ready! Run 'chief' or 'chief %s' to start working on it.\n", opts.Name)
	return nil
}

// runInteractiveAgent launches an interactive agent session in the specified directory.
func runInteractiveAgent(workDir, prompt, agentBin string) error {
	// Pass prompt as argument (not -p which is print mode / non-interactive)
	cmd := exec.Command(agentBin, prompt)
	cmd.Dir = workDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ConvertOptions contains configuration for the conversion command.
type ConvertOptions struct {
	PRDDir   string // PRD directory containing prd.md
	Merge    bool   // Auto-merge without prompting on conversion conflicts
	Force    bool   // Auto-overwrite without prompting on conversion conflicts
	AgentBin string // CLI binary to use for conversion (default: "claude")
}

// RunConvert converts prd.md to prd.json using the default agent.
func RunConvert(prdDir string) error {
	return RunConvertWithOptions(ConvertOptions{PRDDir: prdDir})
}

// RunConvertWithOptions converts prd.md to prd.json using the configured agent.
func RunConvertWithOptions(opts ConvertOptions) error {
	return prd.Convert(prd.ConvertOptions{
		PRDDir:   opts.PRDDir,
		Merge:    opts.Merge,
		Force:    opts.Force,
		AgentBin: opts.AgentBin,
	})
}

// isValidPRDName checks if the name contains only valid characters.
func isValidPRDName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}
