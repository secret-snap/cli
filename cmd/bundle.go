package cmd

import (
	"fmt"
	"os"

	"secretsnap/internal/api"
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

			// Load project config to get project ID
			projectConfig, err := config.LoadProjectConfig()
			if err != nil {
				return fmt.Errorf("failed to load project config: %v", err)
			}

			// Use project from flag or config
			projectID := bundleProject
			if projectID == "" {
				projectID = projectConfig.ProjectID
			}

			if projectID == "" || projectID == "local" {
				return fmt.Errorf("no project specified. Use --project or run 'secretsnap project create <name>' first")
			}

			// Generate fresh data key for this bundle
			dataKey, err := crypto.GenerateDataKey()
			if err != nil {
				return fmt.Errorf("failed to generate data key: %v", err)
			}

			// Encrypt data with the data key
			encryptedData, err = crypto.EncryptWithKey(data, dataKey)
			if err != nil {
				return fmt.Errorf("failed to encrypt: %v", err)
			}

			// Create API client
			client := api.NewClient("http://localhost:8080", token)

			// Step 1: Get upload URL from API
			fmt.Printf("📤 Starting cloud upload...\n")
			pushResp, err := client.BundlePush(projectID, len(encryptedData))
			if err != nil {
				return fmt.Errorf("failed to get upload URL: %v", err)
			}

			// Step 2: Upload encrypted data to S3
			fmt.Printf("☁️ Uploading to cloud storage...\n")
			if err := client.UploadToS3(pushResp.UploadURL, encryptedData); err != nil {
				return fmt.Errorf("failed to upload to cloud: %v", err)
			}

			// Step 3: Finalize bundle (API will handle KMS wrapping)
			fmt.Printf("🔐 Securing with KMS...\n")
			if err := client.BundleFinalize(pushResp.BundleID, pushResp.S3Key, dataKey); err != nil {
				return fmt.Errorf("failed to finalize bundle: %v", err)
			}

			fmt.Printf("✅ Successfully pushed to cloud!\n")
			fmt.Printf("📦 Bundle ID: %s\n", pushResp.BundleID)
			fmt.Printf("📁 Project: %s\n", projectConfig.ProjectName)

			// Also save local copy if requested
			if bundleOutFile != "secrets.envsnap" {
				if err := os.WriteFile(bundleOutFile, encryptedData, 0644); err != nil {
					return fmt.Errorf("failed to write local copy: %v", err)
				}
				fmt.Printf("💾 Local copy saved to: %s\n", bundleOutFile)
			}

			return nil

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

		// Track usage and show upsell for free users
		if mode == "local" || mode == "passphrase" {
			config.IncrementFreeRun()
			utils.ShowContextualUpsell("bundle")
		}

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
