package cmd

import (
	"fmt"
	"os"

	"secretsnap/internal/config"

	"github.com/spf13/cobra"
)

var (
	keyExportProject string
	keyExportAccept  bool
)

var keyExportCmd = &cobra.Command{
	Use:   "key export",
	Short: "Export project key for sharing",
	Long:  `Export the current project's key in base64 format for sharing with teammates. Only available in local mode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load project config
		projectConfig, err := config.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("failed to load project config: %v", err)
		}

		// Determine which project to export
		projectName := keyExportProject
		if projectName == "" {
			projectName = projectConfig.ProjectName
		}

		// Check if this is a cloud project
		if projectConfig.Mode == "cloud" && !keyExportAccept {
			return fmt.Errorf("exporting keys from cloud projects is disabled by default for security.\n" +
				"Use --i-accept-risk if you understand the implications.")
		}

		// Get project key
		projectKey, err := config.GetProjectKey(projectName)
		if err != nil {
			return fmt.Errorf("no key found for project '%s'", projectName)
		}

		// Print warning
		fmt.Fprintf(os.Stderr, "⚠️  WARNING: This will expose your project key!\n")
		fmt.Fprintf(os.Stderr, "   Only share this with trusted teammates.\n")
		fmt.Fprintf(os.Stderr, "   Project: %s\n", projectName)
		fmt.Fprintf(os.Stderr, "   Key ID: %s\n\n", projectKey.KeyID)

		// Output the key to stdout
		fmt.Print(projectKey.KeyB64)

		return nil
	},
}

func init() {
	keyExportCmd.Flags().StringVarP(&keyExportProject, "project", "", "", "Project name (defaults to current project)")
	keyExportCmd.Flags().BoolVarP(&keyExportAccept, "i-accept-risk", "", false, "Accept the risk of exporting cloud project keys")
}
