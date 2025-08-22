package cmd

import (
	"fmt"
	"os"

	"secretsnap/internal/config"
	"secretsnap/internal/crypto"

	"github.com/spf13/cobra"
)

var (
	bundleOutFile string
	bundlePass    string
	bundlePush    bool
	bundleProject string
)

var bundleCmd = &cobra.Command{
	Use:   "bundle [path-to-.env]",
	Short: "Encrypt a .env file into a bundle",
	Long:  `Encrypt a .env file using age encryption. In local mode, uses passphrase. In cloud mode, uses symmetric key encryption.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]

		// Read input file
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %v", err)
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}

		var encryptedData []byte

		if cfg.Mode == "local" {
			// Local mode: use passphrase
			if bundlePass == "" {
				fmt.Print("Enter passphrase: ")
				fmt.Scanln(&bundlePass)
			}

			encryptedData, err = crypto.EncryptWithPassphrase(data, bundlePass)
			if err != nil {
				return fmt.Errorf("failed to encrypt: %v", err)
			}
		} else {
			// Cloud mode: use symmetric key
			dataKey, err := crypto.GenerateDataKey()
			if err != nil {
				return fmt.Errorf("failed to generate data key: %v", err)
			}

			encryptedData, err = crypto.EncryptWithKey(data, dataKey)
			if err != nil {
				return fmt.Errorf("failed to encrypt: %v", err)
			}

			if bundlePush {
				// TODO: Implement cloud push logic
				return fmt.Errorf("cloud push not yet implemented")
			}
		}

		// Write output file
		if err := os.WriteFile(bundleOutFile, encryptedData, 0600); err != nil {
			return fmt.Errorf("failed to write output file: %v", err)
		}

		fmt.Printf("✅ Encrypted %s to %s\n", inputFile, bundleOutFile)
		return nil
	},
}

func init() {
	bundleCmd.Flags().StringVarP(&bundleOutFile, "out", "o", "secrets.envsnap", "Output file path")
	bundleCmd.Flags().StringVarP(&bundlePass, "pass", "p", "", "Passphrase (prompted if not provided)")
	bundleCmd.Flags().BoolVarP(&bundlePush, "push", "", false, "Push to cloud (cloud mode only)")
	bundleCmd.Flags().StringVarP(&bundleProject, "project", "", "", "Project ID or name (cloud mode only)")
}
