// Package cli provides a helper for resolving the AI CLI command.
// The command can be overridden with the CHIEF_CLI environment variable.
package cli

import "os"

// Command returns the AI CLI binary to invoke.
// It reads the CHIEF_CLI environment variable, defaulting to "claude".
func Command() string {
	if cmd := os.Getenv("CHIEF_CLI"); cmd != "" {
		return cmd
	}
	return "claude"
}
