package cmd

import (
	"fmt"

	"secretsnap/internal/api"
	"secretsnap/internal/config"

	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage cloud projects",
	Long:  `Create and manage projects for team collaboration and secret sharing.`,
}

var projectCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new project",
	Long:  `Create a new project for team collaboration.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

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

		// Create API client
		client := api.NewClient("http://localhost:8080", token)

		// Create project
		project, err := client.CreateProject(projectName)
		if err != nil {
			return fmt.Errorf("failed to create project: %v", err)
		}

		// Update config with project
		cfg.Project = project.ID
		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %v", err)
		}

		fmt.Printf("✅ Project created successfully!\n")
		fmt.Printf("📦 Project ID: %s\n", project.ID)
		fmt.Printf("📝 Name: %s\n", project.Name)

		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectCreateCmd)
}
