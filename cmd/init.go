package cmd

import (
	"fmt"
	"os"
	"time"

	"secretsnap/internal/config"
	"secretsnap/internal/crypto"
	"secretsnap/internal/utils"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize secretsnap configuration",
	Long:  `Initialize secretsnap with local configuration. Creates .secretsnap.json in the current directory and generates a project key.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load or create project config
		projectConfig, err := config.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("failed to load project config: %v", err)
		}

		// Check if project key already exists
		existingKey, err := config.GetProjectKey(projectConfig.ProjectName)
		if err == nil && existingKey != nil {
			fmt.Printf("✅ Project '%s' already initialized!\n", projectConfig.ProjectName)
			fmt.Printf("📁 Config file: %s\n", config.GetProjectConfigPath())
			fmt.Printf("🔧 Mode: %s\n", projectConfig.Mode)
			fmt.Printf("🔑 Key ID: %s\n", existingKey.KeyID)
			return nil
		}

		// Generate new project key
		keyBytes, err := crypto.GenerateProjectKey()
		if err != nil {
			return fmt.Errorf("failed to generate project key: %v", err)
		}

		keyID, err := crypto.GenerateKeyID()
		if err != nil {
			return fmt.Errorf("failed to generate key ID: %v", err)
		}

		// Create project key
		projectKey := &config.ProjectKey{
			KeyID:     keyID,
			Algorithm: "age-symmetric-v1",
			KeyB64:    crypto.KeyToBase64(keyBytes),
			CreatedAt: time.Now(),
		}

		// Save project key to cache
		if err := config.SaveProjectKey(projectConfig.ProjectName, projectKey); err != nil {
			return fmt.Errorf("failed to save project key: %v", err)
		}

		// Ensure gitignore entries
		if err := config.EnsureGitignoreEntries(); err != nil {
			return fmt.Errorf("failed to update .gitignore: %v", err)
		}

		fmt.Printf("✅ Secretsnap initialized!\n")
		fmt.Printf("📁 Config file: %s\n", config.GetProjectConfigPath())
		fmt.Printf("🔧 Mode: %s\n", projectConfig.Mode)
		fmt.Printf("📦 Project: %s\n", projectConfig.ProjectName)
		fmt.Printf("🔑 Key ID: %s\n", projectKey.KeyID)
		fmt.Printf("🔒 Key cached at: %s\n", config.GetKeysConfigPath())

		// Show general upsell for new users
		if err := utils.ShowUpsell(); err != nil {
			// Don't fail the command if upsell fails
			fmt.Fprintf(os.Stderr, "Warning: failed to show upsell: %v\n", err)
		}

		return nil
	},
}

// Command is registered in commands.go
