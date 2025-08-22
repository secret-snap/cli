package cmd

import (
	"github.com/spf13/cobra"
)

// InitCommands registers all commands with the root command
func InitCommands(rootCmd *cobra.Command) {
	// Free commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(bundleCmd)
	rootCmd.AddCommand(unbundleCmd)
	rootCmd.AddCommand(runCmd)

	// Paid commands
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(projectCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(shareCmd)
	rootCmd.AddCommand(auditCmd)
}
