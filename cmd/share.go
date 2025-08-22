package cmd

import (
	"fmt"

	"secretsnap/internal/api"
	"secretsnap/internal/config"

	"github.com/spf13/cobra"
)

var (
	shareProject string
	shareUser    string
	shareRole    string
)

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Share project with team member",
	Long:  `Share a project with another user by email address.`,
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
		if shareProject == "" {
			shareProject = cfg.Project
		}

		if shareProject == "" {
			return fmt.Errorf("no project specified. Use --project or set in config")
		}

		if shareUser == "" {
			return fmt.Errorf("user email is required. Use --user")
		}

		if shareRole == "" {
			shareRole = "member" // Default role
		}

		// Create API client
		client := api.NewClient("http://localhost:8080", token)

		// Share project
		err = client.Share(shareProject, shareUser, shareRole)
		if err != nil {
			return fmt.Errorf("failed to share project: %v", err)
		}

		fmt.Printf("✅ Project shared successfully!\n")
		fmt.Printf("👤 User: %s\n", shareUser)
		fmt.Printf("🔑 Role: %s\n", shareRole)

		return nil
	},
}

func init() {
	shareCmd.Flags().StringVarP(&shareProject, "project", "", "", "Project ID or name")
	shareCmd.Flags().StringVarP(&shareUser, "user", "u", "", "User email (required)")
	shareCmd.Flags().StringVarP(&shareRole, "role", "r", "member", "Role (member, owner)")
	shareCmd.MarkFlagRequired("user")
}
