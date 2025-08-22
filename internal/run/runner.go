package run

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// Runner executes commands with environment variables from a decrypted .env file
type Runner struct {
	envData []byte
}

// NewRunner creates a new runner with the decrypted environment data
func NewRunner(envData []byte) *Runner {
	return &Runner{
		envData: envData,
	}
}

// Run executes the command with the environment variables loaded
func (r *Runner) Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	// Create a temporary .env file
	tempDir, err := os.MkdirTemp("", "secretsnap-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempEnvFile := filepath.Join(tempDir, ".env")
	if err := os.WriteFile(tempEnvFile, r.envData, 0600); err != nil {
		return fmt.Errorf("failed to write temp env file: %v", err)
	}

	// Parse environment variables
	envVars, err := r.parseEnvFile(r.envData)
	if err != nil {
		return fmt.Errorf("failed to parse env file: %v", err)
	}

	// Create command
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment variables
	cmd.Env = append(os.Environ(), envVars...)

	// Execute command
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		return fmt.Errorf("command failed: %v", err)
	}

	return nil
}

// parseEnvFile parses a .env file and returns environment variable strings
func (r *Runner) parseEnvFile(data []byte) ([]string, error) {
	var envVars []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 && (value[0] == '"' && value[len(value)-1] == '"') {
			value = value[1 : len(value)-1]
		}

		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan env file: %v", err)
	}

	return envVars, nil
}
