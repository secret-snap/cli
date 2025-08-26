package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"secretsnap/internal/config"
	"secretsnap/internal/crypto"
	"secretsnap/internal/utils"

	"github.com/spf13/cobra"
)

var (
	runPass     string
	runPassFile string
	runPassMode bool
)

var runCmd = &cobra.Command{
	Use:   "run [bundle-file] -- [command...]",
	Short: "Run a command with environment variables from a bundle",
	Long:  `Decrypt a bundle to temporary environment variables and run a command. The temporary file is securely deleted after execution.`,
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		bundleFile := args[0]
		commandArgs := args[1:]

		// Validate bundle file exists
		if _, err := os.Stat(bundleFile); os.IsNotExist(err) {
			return fmt.Errorf("bundle file '%s' does not exist", bundleFile)
		}

		encryptedData, err := os.ReadFile(bundleFile)
		if err != nil {
			return fmt.Errorf("failed to read bundle file: %v", err)
		}

		if len(encryptedData) == 0 {
			return fmt.Errorf("bundle file '%s' is empty", bundleFile)
		}

		// Load project config
		projectConfig, err := config.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("failed to load project config: %v", err)
		}

		// Determine mode based on flags
		mode := determineRunMode(runPass, runPassFile, runPassMode)

		var decryptedData []byte

		switch mode {
		case "passphrase":
			// Passphrase mode
			passphrase, err := utils.GetPassphrase(runPass, runPassFile)
			if err != nil {
				return fmt.Errorf("failed to get passphrase: %v", err)
			}

			decryptedData, err = crypto.DecryptWithPassphrase(encryptedData, passphrase)
			if err != nil {
				return fmt.Errorf("failed to decrypt: %v", err)
			}

		default:
			// Local mode (default)
			projectKey, err := config.GetProjectKey(projectConfig.ProjectName)
			if err != nil {
				return fmt.Errorf("no local project key found for '%s'. Fix:\n"+
					"• On teammate's machine: `secretsnap key export --project %s`\n"+
					"• Or use passphrase: `--pass`\n"+
					"• Or use paid pull: `secretsnap login` then `secretsnap pull`",
					projectConfig.ProjectName, projectConfig.ProjectName)
			}

			keyBytes, err := crypto.KeyFromBase64(projectKey.KeyB64)
			if err != nil {
				return fmt.Errorf("failed to decode project key: %v", err)
			}

			decryptedData, err = crypto.DecryptWithKey(encryptedData, keyBytes)
			if err != nil {
				return fmt.Errorf("failed to decrypt: %v", err)
			}
		}

		// Parse environment variables from decrypted data
		envVars, err := parseEnvFile(decryptedData)
		if err != nil {
			return fmt.Errorf("failed to parse environment variables: %v", err)
		}

		// Create command
		command := exec.Command(commandArgs[0], commandArgs[1:]...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.Stdin = os.Stdin

		// Set environment variables
		command.Env = append(os.Environ(), envVars...)

		// Run command
		if err := command.Run(); err != nil {
			return fmt.Errorf("command failed: %v", err)
		}

		// Track usage and show upsell for free users
		if mode == "local" || mode == "passphrase" {
			if err := config.IncrementFreeRun(); err != nil {
				// Don't fail the command if upsell tracking fails
				fmt.Fprintf(os.Stderr, "Warning: failed to track usage: %v\n", err)
			}
			
			// Show contextual upsell
			if err := utils.ShowContextualUpsell("run"); err != nil {
				// Don't fail the command if upsell fails
				fmt.Fprintf(os.Stderr, "Warning: failed to show upsell: %v\n", err)
			}
		}

		return nil
	},
}

func init() {
	runCmd.Flags().StringVarP(&runPass, "pass", "p", "", "Passphrase (prompted if not provided)")
	runCmd.Flags().StringVarP(&runPassFile, "pass-file", "", "", "Read passphrase from file")
	runCmd.Flags().BoolVarP(&runPassMode, "pass-mode", "", false, "Use passphrase mode (prompt for passphrase)")
}

// determineRunMode determines the decryption mode based on flags
func determineRunMode(pass, passFile string, passMode bool) string {
	if pass != "" || passFile != "" || passMode {
		return "passphrase"
	}
	return "local"
}



// parseEnvFile parses environment variables from a .env file format
func parseEnvFile(data []byte) ([]string, error) {
	lines := strings.Split(string(data), "\n")
	var envVars []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check if line contains key=value format
		if strings.Contains(line, "=") {
			envVars = append(envVars, line)
		}
	}

	return envVars, nil
}
