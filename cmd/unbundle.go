package cmd

import (
	"fmt"
	"os"

	"secretsnap/internal/config"
	"secretsnap/internal/crypto"

	"github.com/spf13/cobra"
)

var (
	unbundleOutFile string
	unbundlePass    string
)

var unbundleCmd = &cobra.Command{
	Use:   "unbundle [bundle-file]",
	Short: "Decrypt a bundle to a .env file",
	Long:  `Decrypt a bundle file back to a .env file. In local mode, requires passphrase.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]

		// Read encrypted file
		encryptedData, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read bundle file: %v", err)
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}

		var decryptedData []byte

		if cfg.Mode == "local" {
			// Local mode: use passphrase
			if unbundlePass == "" {
				fmt.Print("Enter passphrase: ")
				fmt.Scanln(&unbundlePass)
			}

			decryptedData, err = crypto.DecryptWithPassphrase(encryptedData, unbundlePass)
			if err != nil {
				return fmt.Errorf("failed to decrypt: %v", err)
			}
		} else {
			// Cloud mode: TODO implement
			return fmt.Errorf("cloud mode not yet implemented")
		}

		// Write output file
		if err := os.WriteFile(unbundleOutFile, decryptedData, 0600); err != nil {
			return fmt.Errorf("failed to write output file: %v", err)
		}

		fmt.Printf("✅ Decrypted %s to %s\n", inputFile, unbundleOutFile)
		return nil
	},
}

func init() {
	unbundleCmd.Flags().StringVarP(&unbundleOutFile, "out", "o", ".env", "Output file path")
	unbundleCmd.Flags().StringVarP(&unbundlePass, "pass", "p", "", "Passphrase (prompted if not provided)")
}
