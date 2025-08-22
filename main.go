package main

import (
	"fmt"
	"os"

	"secretsnap/cmd"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "secretsnap",
	Short: "Secrets Snapshot - Encrypt .env files into bundles",
	Long: `Secrets Snapshot is a tool for encrypting .env files into secure bundles
that can be shared and deployed safely across machines and CI environments.

Free mode: Local encryption with passphrase
Paid mode: Cloud storage with team sharing and audit logs`,
}

func main() {
	// Initialize commands
	cmd.InitCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
