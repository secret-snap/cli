package cmd

import (
	"fmt"
	"os"

	"secretsnap/internal/api"
	"secretsnap/internal/config"
	"secretsnap/internal/utils"

	"github.com/spf13/cobra"
)

var (
	shareProject string
	shareUser    string
	shareRole    string
)

var shareCmd = &cobra.Command{
	Use:   "share --user <email> --role <read|write>",
	Short: "Share project with team member",
	Long:  `Share a project with another user by email address.`,
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
		if shareProject == "" {
			shareProject = projectConfig.ProjectID
		}

		if shareProject == "" {
			return fmt.Errorf("no project specified. Use --project or run 'secretsnap project create <name>' first")
		}

		if shareUser == "" {
			return fmt.Errorf("user email is required. Use --user <email>")
		}

		if shareRole == "" {
			shareRole = "read" // Default role
		}

		// Validate role
		if shareRole != "read" && shareRole != "write" {
			return fmt.Errorf("role must be 'read' or 'write', got '%s'", shareRole)
		}

		// Create API client
		client := api.NewClient("http://localhost:8080", token)

		// Share project
		err = client.Share(shareProject, shareUser, shareRole)
		if err != nil {
			return fmt.Errorf("failed to share project: %v", err)
		}

		fmt.Printf("✅ Invited %s\n", shareUser)
		fmt.Printf("🔑 Role: %s\n", shareRole)
		fmt.Printf("📦 Project: %s\n", projectConfig.ProjectName)

		// Show feature-specific upsell for team sharing
		if err := utils.ShowFeatureUpsell("team"); err != nil {
			// Don't fail the command if upsell fails
			fmt.Fprintf(os.Stderr, "Warning: failed to show upsell: %v\n", err)
		}

		return nil
	},
}

func init() {
	shareCmd.Flags().StringVarP(&shareProject, "project", "", "", "Project ID or name")
	shareCmd.Flags().StringVarP(&shareUser, "user", "u", "", "User email (required)")
	shareCmd.Flags().StringVarP(&shareRole, "role", "r", "read", "Role (read|write)")
	shareCmd.MarkFlagRequired("user")
}
