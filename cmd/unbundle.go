package cmd

import (
	"fmt"
	"os"

	"secretsnap/internal/config"
	"secretsnap/internal/crypto"

	"github.com/spf13/cobra"
)

var (
	unbundleOutFile  string
	unbundlePass     string
	unbundlePassFile string
	unbundleForce    bool
)

var unbundleCmd = &cobra.Command{
	Use:   "unbundle [path-to-bundle]",
	Short: "Decrypt a bundle back to a .env file",
	Long:  `Decrypt a bundle file back to a .env file. Supports local mode (cached key), passphrase mode, and cloud mode.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]

		// Validate input file exists
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			return fmt.Errorf("input file '%s' does not exist", inputFile)
		}

		encryptedData, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %v", err)
		}

		if len(encryptedData) == 0 {
			return fmt.Errorf("input file '%s' is empty", inputFile)
		}

		// Load project config
		projectConfig, err := config.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("failed to load project config: %v", err)
		}

		// Determine mode based on flags
		mode := determineUnbundleMode(unbundlePass, unbundlePassFile)

		var decryptedData []byte

		switch mode {
		case "passphrase":
			// Passphrase mode
			passphrase, err := getPassphrase(unbundlePass, unbundlePassFile)
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

		// Check if output file exists and handle --force
		if _, err := os.Stat(unbundleOutFile); err == nil && !unbundleForce {
			return fmt.Errorf("refusing to overwrite %s. Use `--force`", unbundleOutFile)
		}

		// Write output file with secure permissions
		if err := os.WriteFile(unbundleOutFile, decryptedData, 0600); err != nil {
			return fmt.Errorf("failed to write output file: %v", err)
		}

		// Check if file permissions are correct and warn if not
		if info, err := os.Stat(unbundleOutFile); err == nil {
			if info.Mode().Perm() != 0600 {
				fmt.Printf("⚠️  Warning: %s has permissions %v, should be 0600\n", unbundleOutFile, info.Mode().Perm())
			}
		}

		fmt.Printf("✅ Decrypted %s to %s\n", inputFile, unbundleOutFile)
		return nil
	},
}

func init() {
	unbundleCmd.Flags().StringVarP(&unbundleOutFile, "out", "o", ".env", "Output file path")
	unbundleCmd.Flags().StringVarP(&unbundlePass, "pass", "p", "", "Passphrase (prompted if not provided)")
	unbundleCmd.Flags().StringVarP(&unbundlePassFile, "pass-file", "", "", "Read passphrase from file")
	unbundleCmd.Flags().BoolVarP(&unbundleForce, "force", "f", false, "Overwrite output file if it exists")
}

// determineUnbundleMode determines the decryption mode based on flags
func determineUnbundleMode(pass, passFile string) string {
	if pass != "" || passFile != "" {
		return "passphrase"
	}
	return "local"
}
