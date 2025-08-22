package cmd

import (
	"fmt"

	"secretsnap/internal/api"
	"secretsnap/internal/config"

	"github.com/spf13/cobra"
)

var (
	loginLicense string
	loginAPIURL  string
)

var loginCmd = &cobra.Command{
	Use:   "login --license <KEY>",
	Short: "Login with license key for cloud features",
	Long:  `Login to secretsnap cloud with your license key to enable team sharing and audit features.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if loginLicense == "" {
			return fmt.Errorf("license key is required. Use --license <KEY>")
		}

		if loginAPIURL == "" {
			loginAPIURL = "http://localhost:8080" // Default for local development
		}

		// Create API client
		client := api.NewClient(loginAPIURL, "")

		// Login
		resp, err := client.Login(loginLicense)
		if err != nil {
			return fmt.Errorf("login failed: %v", err)
		}

		// Save token
		if err := config.SaveToken(resp.Token); err != nil {
			return fmt.Errorf("failed to save token: %v", err)
		}

		// Update project config to cloud mode
		projectConfig, err := config.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("failed to load project config: %v", err)
		}

		projectConfig.Mode = "cloud"
		if err := config.SaveProjectConfig(projectConfig); err != nil {
			return fmt.Errorf("failed to save project config: %v", err)
		}

		fmt.Printf("✅ Logged in successfully!\n")
		fmt.Printf("👤 User: %s\n", resp.User.Email)
		fmt.Printf("📋 Plan: %s\n", resp.User.Plan)
		fmt.Printf("🔧 Mode: %s\n", projectConfig.Mode)
		fmt.Printf("🔑 Token saved to: %s\n", config.GetKeysConfigPath())

		return nil
	},
}

func init() {
	loginCmd.Flags().StringVarP(&loginLicense, "license", "l", "", "License key (required)")
	loginCmd.Flags().StringVarP(&loginAPIURL, "api-url", "", "", "API URL (default: http://localhost:8080)")
	loginCmd.MarkFlagRequired("license")
}
