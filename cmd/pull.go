package cmd

import (
	"encoding/base64"
	"fmt"
	"os"

	"secretsnap/internal/api"
	"secretsnap/internal/config"
	"secretsnap/internal/crypto"
	"secretsnap/internal/utils"

	"github.com/spf13/cobra"
)

var (
	pullOutFile string
	pullProject string
	pullVersion int
	pullForce   bool
)

var pullCmd = &cobra.Command{
	Use:   "pull [--out .env] [--version N]",
	Short: "Pull latest bundle from cloud",
	Long:  `Download and decrypt the latest bundle from the cloud project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load project config and token
		projectConfig, err := config.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("failed to load project config: %v", err)
		}

		token, err := config.LoadToken()
		if err != nil {
			return fmt.Errorf("failed to load token: %v", err)
		}

		if token == "" {
			return fmt.Errorf("not logged in. Run 'secretsnap login --license <KEY>' first")
		}

		// Use project from config if not specified
		if pullProject == "" {
			pullProject = projectConfig.ProjectID
		}

		if pullProject == "" {
			return fmt.Errorf("no project specified. Use --project or run 'secretsnap project create <name>' first")
		}

		// Create API client
		client := api.NewClient(utils.GetAPIURL(), token)

		// Pull bundle
		resp, err := client.BundlePull(pullProject)
		if err != nil {
			return fmt.Errorf("failed to pull bundle: %v", err)
		}

		// Download encrypted data
		encryptedData, err := client.DownloadFromS3(resp.DownloadURL)
		if err != nil {
			return fmt.Errorf("failed to download bundle: %v", err)
		}

		// Decode data key
		dataKey, err := base64.StdEncoding.DecodeString(resp.DataKey)
		if err != nil {
			return fmt.Errorf("failed to decode data key: %v", err)
		}

		// Decrypt data
		decryptedData, err := crypto.DecryptWithKey(encryptedData, dataKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt bundle: %v", err)
		}

		// Check if output file exists and handle --force
		if _, err := os.Stat(pullOutFile); err == nil && !pullForce {
			return fmt.Errorf("refusing to overwrite %s. Use `--force`", pullOutFile)
		}

		// Write output file with secure permissions
		if err := os.WriteFile(pullOutFile, decryptedData, 0600); err != nil {
			return fmt.Errorf("failed to write output file: %v", err)
		}

		// Check if file permissions are correct and warn if not
		if info, err := os.Stat(pullOutFile); err == nil {
			if info.Mode().Perm() != 0600 {
				fmt.Printf("⚠️  Warning: %s has permissions %v, should be 0600\n", pullOutFile, info.Mode().Perm())
			}
		}

		fmt.Printf("✅ Pulled version %d to %s\n", resp.Version, pullOutFile)

		// Show feature-specific upsell for cloud features
		if err := utils.ShowFeatureUpsell("cloud"); err != nil {
			// Don't fail the command if upsell fails
			fmt.Fprintf(os.Stderr, "Warning: failed to show upsell: %v\n", err)
		}

		return nil
	},
}

func init() {
	pullCmd.Flags().StringVarP(&pullOutFile, "out", "o", ".env", "Output file path")
	pullCmd.Flags().StringVarP(&pullProject, "project", "", "", "Project ID or name")
	pullCmd.Flags().IntVarP(&pullVersion, "version", "", 0, "Specific version to pull")
	pullCmd.Flags().BoolVarP(&pullForce, "force", "f", false, "Overwrite output file if it exists")
}
