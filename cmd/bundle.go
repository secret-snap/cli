package cmd

import (
	"fmt"
	"os"

	"secretsnap/internal/config"
	"secretsnap/internal/crypto"
	"secretsnap/internal/utils"

	"github.com/spf13/cobra"
)

var (
	bundleOutFile  string
	bundlePass     string
	bundlePassFile string
	bundlePassMode bool
	bundlePush     bool
	bundleProject  string
	bundleForce    bool
	bundleExpire   string
	bundleVersion  int
)

var bundleCmd = &cobra.Command{
	Use:   "bundle [path-to-.env]",
	Short: "Encrypt a .env file into a bundle",
	Long:  `Encrypt a .env file using age encryption. Supports local mode (cached key), passphrase mode, and cloud mode.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]

		// Validate input file exists and is not empty
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			return fmt.Errorf("input file '%s' does not exist", inputFile)
		}

		data, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %v", err)
		}

		if len(data) == 0 {
			return fmt.Errorf("input file '%s' is empty", inputFile)
		}

		// Load project config
		projectConfig, err := config.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("failed to load project config: %v", err)
		}

		// Determine mode based on flags and config
		mode := determineMode(projectConfig, bundlePass, bundlePassFile, bundlePassMode, bundlePush)

		var encryptedData []byte

		switch mode {
		case "passphrase":
			// Passphrase mode
			passphrase, err := utils.GetPassphrase(bundlePass, bundlePassFile)
			if err != nil {
				return fmt.Errorf("failed to get passphrase: %v", err)
			}

			encryptedData, err = crypto.EncryptWithPassphrase(data, passphrase)
			if err != nil {
				return fmt.Errorf("failed to encrypt: %v", err)
			}

		case "cloud":
			// Cloud mode (paid)
			if !bundlePush {
				return fmt.Errorf("cloud mode requires --push flag")
			}

			// Check if user is logged in
			token, err := config.LoadToken()
			if err != nil {
				return fmt.Errorf("failed to load token: %v", err)
			}
			if token == "" {
				return fmt.Errorf("cloud sync is Pro. Run `secretsnap login --license …` or use local mode (no `--push`)")
			}

			// Generate fresh data key for this bundle
			dataKey, err := crypto.GenerateDataKey()
			if err != nil {
				return fmt.Errorf("failed to generate data key: %v", err)
			}

			encryptedData, err = crypto.EncryptWithKey(data, dataKey)
			if err != nil {
				return fmt.Errorf("failed to encrypt: %v", err)
			}

			// TODO: Implement cloud push logic
			return fmt.Errorf("cloud push not yet implemented")

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

			encryptedData, err = crypto.EncryptWithKey(data, keyBytes)
			if err != nil {
				return fmt.Errorf("failed to encrypt: %v", err)
			}
		}

		// Check if output file exists and handle --force
		if _, err := os.Stat(bundleOutFile); err == nil && !bundleForce {
			return fmt.Errorf("refusing to overwrite %s. Use `--force`", bundleOutFile)
		}

		// Write output file
		if err := os.WriteFile(bundleOutFile, encryptedData, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %v", err)
		}

		fmt.Printf("✅ Encrypted %s to %s\n", inputFile, bundleOutFile)
		return nil
	},
}

func init() {
	bundleCmd.Flags().StringVarP(&bundleOutFile, "out", "o", "secrets.envsnap", "Output file path")
	bundleCmd.Flags().StringVarP(&bundlePass, "pass", "p", "", "Passphrase (prompted if not provided)")
	bundleCmd.Flags().StringVarP(&bundlePassFile, "pass-file", "", "", "Read passphrase from file")
	bundleCmd.Flags().BoolVarP(&bundlePassMode, "pass-mode", "", false, "Use passphrase mode (prompt for passphrase)")
	bundleCmd.Flags().BoolVarP(&bundlePush, "push", "", false, "Push to cloud (cloud mode only)")
	bundleCmd.Flags().StringVarP(&bundleProject, "project", "", "", "Project ID or name (cloud mode only)")
	bundleCmd.Flags().BoolVarP(&bundleForce, "force", "f", false, "Overwrite output file if it exists")
	bundleCmd.Flags().StringVarP(&bundleExpire, "expire", "", "", "Expiration time (e.g., 24h)")
	bundleCmd.Flags().IntVarP(&bundleVersion, "version", "", 0, "Version number")
}

// determineMode determines the encryption mode based on flags and config
func determineMode(projectConfig *config.ProjectConfig, pass, passFile string, passMode, push bool) string {
	if pass != "" || passFile != "" || passMode {
		return "passphrase"
	}
	if push {
		return "cloud"
	}
	return "local"
}
