package cmd

import (
	"fmt"
	"os"

	"secretsnap/internal/config"
	"secretsnap/internal/crypto"
	"secretsnap/internal/run"

	"github.com/spf13/cobra"
)

var (
	runPass string
)

var runCmd = &cobra.Command{
	Use:   "run [bundle-file] -- [command...]",
	Short: "Run a command with environment variables from a bundle",
	Long:  `Decrypt a bundle and run a command with the environment variables loaded.`,
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		bundleFile := args[0]
		commandArgs := args[1:]

		// Read encrypted file
		encryptedData, err := os.ReadFile(bundleFile)
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
			if runPass == "" {
				fmt.Print("Enter passphrase: ")
				fmt.Scanln(&runPass)
			}

			decryptedData, err = crypto.DecryptWithPassphrase(encryptedData, runPass)
			if err != nil {
				return fmt.Errorf("failed to decrypt: %v", err)
			}
		} else {
			// Cloud mode: TODO implement
			return fmt.Errorf("cloud mode not yet implemented")
		}

		// Create runner and execute command
		runner := run.NewRunner(decryptedData)
		return runner.Run(commandArgs)
	},
}

func init() {
	runCmd.Flags().StringVarP(&runPass, "pass", "p", "", "Passphrase (prompted if not provided)")
}
