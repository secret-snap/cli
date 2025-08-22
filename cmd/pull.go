package cmd

import (
	"encoding/base64"
	"fmt"
	"os"

	"secretsnap/internal/api"
	"secretsnap/internal/config"
	"secretsnap/internal/crypto"

	"github.com/spf13/cobra"
)

var (
	pullOutFile string
	pullProject string
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull latest bundle from cloud",
	Long:  `Download and decrypt the latest bundle from the cloud project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config and token
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}

		token, err := config.LoadToken()
		if err != nil {
			return fmt.Errorf("failed to load token: %v", err)
		}

		if token == "" {
			return fmt.Errorf("not logged in. Run 'secretsnap login' first")
		}

		// Use project from config if not specified
		if pullProject == "" {
			pullProject = cfg.Project
		}

		if pullProject == "" {
			return fmt.Errorf("no project specified. Use --project or set in config")
		}

		// Create API client
		client := api.NewClient("http://localhost:8080", token)

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

		// Write output file
		if err := os.WriteFile(pullOutFile, decryptedData, 0600); err != nil {
			return fmt.Errorf("failed to write output file: %v", err)
		}

		fmt.Printf("✅ Pulled bundle v%d to %s\n", resp.Version, pullOutFile)
		return nil
	},
}

func init() {
	pullCmd.Flags().StringVarP(&pullOutFile, "out", "o", ".env", "Output file path")
	pullCmd.Flags().StringVarP(&pullProject, "project", "", "", "Project ID or name")
}
