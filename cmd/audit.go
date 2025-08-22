package cmd

import (
	"fmt"
	"time"

	"secretsnap/internal/api"
	"secretsnap/internal/config"

	"github.com/spf13/cobra"
)

var (
	auditProject string
	auditLimit   int
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "View project audit logs",
	Long:  `View recent audit logs for a project to track access and changes.`,
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
		if auditProject == "" {
			auditProject = cfg.Project
		}

		if auditProject == "" {
			return fmt.Errorf("no project specified. Use --project or set in config")
		}

		// Create API client
		client := api.NewClient("http://localhost:8080", token)

		// Get audit logs
		logs, err := client.GetAuditLogs(auditProject, auditLimit)
		if err != nil {
			return fmt.Errorf("failed to get audit logs: %v", err)
		}

		if len(logs) == 0 {
			fmt.Println("No audit logs found.")
			return nil
		}

		fmt.Printf("📋 Audit logs for project %s:\n\n", auditProject)
		for _, log := range logs {
			// Parse timestamp
			t, err := time.Parse(time.RFC3339, log.CreatedAt)
			if err != nil {
				t = time.Now() // Fallback
			}

			fmt.Printf("🕐 %s\n", t.Format("2006-01-02 15:04:05"))
			fmt.Printf("📝 Action: %s\n", log.Action)
			if log.Details != nil && len(log.Details) > 0 {
				fmt.Printf("📄 Details: %v\n", log.Details)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	auditCmd.Flags().StringVarP(&auditProject, "project", "", "", "Project ID or name")
	auditCmd.Flags().IntVarP(&auditLimit, "limit", "l", 50, "Number of logs to show")
}
