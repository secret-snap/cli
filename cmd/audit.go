package cmd

import (
	"fmt"
	"os"
	"time"

	"secretsnap/internal/api"
	"secretsnap/internal/config"
	"secretsnap/internal/utils"

	"github.com/spf13/cobra"
)

var (
	auditProject string
	auditLimit   int
)

var auditCmd = &cobra.Command{
	Use:   "audit [--limit 50]",
	Short: "View project audit logs",
	Long:  `View recent audit logs for a project to track access and changes.`,
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
		if auditProject == "" {
			auditProject = projectConfig.ProjectID
		}

		if auditProject == "" {
			return fmt.Errorf("no project specified. Use --project or run 'secretsnap project create <name>' first")
		}

		// Create API client
		client := api.NewClient(utils.GetAPIURL(), token)

		// Get audit logs
		logs, err := client.GetAuditLogs(auditProject, auditLimit)
		if err != nil {
			return fmt.Errorf("failed to get audit logs: %v", err)
		}

		if len(logs) == 0 {
			fmt.Println("No audit logs found.")
			return nil
		}

		fmt.Printf("📋 Audit logs for project %s:\n\n", projectConfig.ProjectName)
		for _, log := range logs {
			// Parse timestamp
			t, err := time.Parse(time.RFC3339, log.CreatedAt)
			if err != nil {
				t = time.Now() // Fallback
			}

			fmt.Printf("🕐 %s\n", t.Format("2006-01-02 15:04:05"))
			fmt.Printf("📝 Action: %s\n", log.Action)
			if len(log.Details) > 0 {
				fmt.Printf("📄 Details: %v\n", log.Details)
			}
			fmt.Println()
		}

		// Show feature-specific upsell for audit logs
		if err := utils.ShowFeatureUpsell("audit"); err != nil {
			// Don't fail the command if upsell fails
			fmt.Fprintf(os.Stderr, "Warning: failed to show upsell: %v\n", err)
		}

		return nil
	},
}

func init() {
	auditCmd.Flags().StringVarP(&auditProject, "project", "", "", "Project ID or name")
	auditCmd.Flags().IntVarP(&auditLimit, "limit", "l", 50, "Number of logs to show")
}
