package cmd

import (
	"fmt"

	"secretsnap/internal/config"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize secretsnap configuration",
	Long:  `Initialize secretsnap with local configuration. Creates .secretsnap.json in the current directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}

		fmt.Printf("✅ Secretsnap initialized!\n")
		fmt.Printf("📁 Config file: %s\n", config.GetConfigPath())
		fmt.Printf("🔧 Mode: %s\n", cfg.Mode)
		fmt.Printf("📦 Project: %s\n", cfg.Project)

		return nil
	},
}

// Command is registered in commands.go
