package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ProjectConfig represents the local project configuration
type ProjectConfig struct {
	ProjectName string `json:"project_name"`
	ProjectID   string `json:"project_id"`
	Mode        string `json:"mode"` // "local", "passphrase", "cloud"
	BundlePath  string `json:"bundle_path"`
}

// ProjectKey represents a cached project key
type ProjectKey struct {
	KeyID     string    `json:"key_id"`
	Algorithm string    `json:"alg"`
	KeyB64    string    `json:"key_b64"`
	CreatedAt time.Time `json:"created_at"`
}

// KeysConfig represents the global keys configuration
type KeysConfig struct {
	Projects map[string]ProjectKey `json:"projects"`
}

// GlobalConfig represents global configuration
type GlobalConfig struct {
	Token string `json:"token"` // JWT token for cloud mode
}

// UsageStats tracks usage for upsell messages
type UsageStats struct {
	FreeRuns     int       `json:"free_runs"`
	LastUpsell   time.Time `json:"last_upsell"`
	UpsellShown  bool      `json:"upsell_shown"`
}

var (
	configDir     string
	projectFile   string
	keysFile      string
	globalDir     string
	tokenFile     string
	gitignoreFile string
	usageFile     string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get home directory: %v", err))
	}

	configDir = filepath.Join(home, ".secretsnap")
	projectFile = ".secretsnap.json"
	keysFile = filepath.Join(configDir, "keys.json")
	globalDir = configDir
	tokenFile = filepath.Join(globalDir, "token")
	gitignoreFile = ".gitignore"
	usageFile = filepath.Join(globalDir, "usage.json")
}

// EnsureConfigDir creates the global config directory with proper permissions
func EnsureConfigDir() error {
	return os.MkdirAll(configDir, 0700)
}

// LoadProjectConfig loads the project configuration from the current directory
func LoadProjectConfig() (*ProjectConfig, error) {
	if _, err := os.Stat(projectFile); os.IsNotExist(err) {
		// Create default project config
		projectName := filepath.Base(getCurrentDir())
		config := &ProjectConfig{
			ProjectName: projectName,
			ProjectID:   "local",
			Mode:        "local",
			BundlePath:  "secrets.envsnap",
		}
		if err := SaveProjectConfig(config); err != nil {
			return nil, err
		}
		return config, nil
	}

	data, err := os.ReadFile(projectFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config file: %v", err)
	}

	var config ProjectConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse project config file: %v", err)
	}

	return &config, nil
}

// SaveProjectConfig saves the project configuration to the current directory
func SaveProjectConfig(config *ProjectConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %v", err)
	}

	if err := os.WriteFile(projectFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write project config file: %v", err)
	}

	return nil
}

// LoadKeysConfig loads the global keys configuration
func LoadKeysConfig() (*KeysConfig, error) {
	if err := EnsureConfigDir(); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %v", err)
	}

	if _, err := os.Stat(keysFile); os.IsNotExist(err) {
		// Create default keys config
		config := &KeysConfig{
			Projects: make(map[string]ProjectKey),
		}
		if err := SaveKeysConfig(config); err != nil {
			return nil, err
		}
		return config, nil
	}

	data, err := os.ReadFile(keysFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read keys file: %v", err)
	}

	var config KeysConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse keys file: %v", err)
	}

	return &config, nil
}

// SaveKeysConfig saves the global keys configuration
func SaveKeysConfig(config *KeysConfig) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keys config: %v", err)
	}

	// Atomic write: write to temp file first, then rename
	tempFile := keysFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write temp keys file: %v", err)
	}

	if err := os.Rename(tempFile, keysFile); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to rename keys file: %v", err)
	}

	return nil
}

// GetProjectKey retrieves the cached key for a project
func GetProjectKey(projectName string) (*ProjectKey, error) {
	keys, err := LoadKeysConfig()
	if err != nil {
		return nil, err
	}

	key, exists := keys.Projects[projectName]
	if !exists {
		return nil, fmt.Errorf("no key found for project '%s'", projectName)
	}

	return &key, nil
}

// SaveProjectKey saves a project key to the cache
func SaveProjectKey(projectName string, key *ProjectKey) error {
	keys, err := LoadKeysConfig()
	if err != nil {
		return err
	}

	if keys.Projects == nil {
		keys.Projects = make(map[string]ProjectKey)
	}

	keys.Projects[projectName] = *key
	return SaveKeysConfig(keys)
}

// LoadToken loads the JWT token for cloud mode
func LoadToken() (string, error) {
	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		return "", nil
	}

	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", fmt.Errorf("failed to read token file: %v", err)
	}

	return string(data), nil
}

// SaveToken saves the JWT token for cloud mode
func SaveToken(token string) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	if err := os.WriteFile(tokenFile, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to write token file: %v", err)
	}

	return nil
}

// EnsureGitignoreEntries ensures the necessary entries are in .gitignore
func EnsureGitignoreEntries() error {
	entries := []string{
		".secretsnap.key",
		"secrets.envsnap.key",
		".secretsnap/",
	}

	// Read existing .gitignore
	var existingContent []byte
	if _, err := os.Stat(gitignoreFile); err == nil {
		existingContent, err = os.ReadFile(gitignoreFile)
		if err != nil {
			return fmt.Errorf("failed to read .gitignore: %v", err)
		}
	}

	// Check which entries are missing
	content := string(existingContent)
	missingEntries := []string{}
	for _, entry := range entries {
		if !containsLine(content, entry) {
			missingEntries = append(missingEntries, entry)
		}
	}

	// Add missing entries
	if len(missingEntries) > 0 {
		if len(content) > 0 && !endsWithNewline(content) {
			content += "\n"
		}
		for _, entry := range missingEntries {
			content += entry + "\n"
		}

		if err := os.WriteFile(gitignoreFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write .gitignore: %v", err)
		}
	}

	return nil
}

// Helper functions
func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

func containsLine(content, line string) bool {
	// Simple check - in a real implementation you might want more sophisticated parsing
	return contains(content, line)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		contains(s[1:len(s)-1], substr))))
}

func endsWithNewline(s string) bool {
	return len(s) > 0 && s[len(s)-1] == '\n'
}

// GetProjectConfigPath returns the path to the project config file
func GetProjectConfigPath() string {
	return projectFile
}

// GetKeysConfigPath returns the path to the keys config file
func GetKeysConfigPath() string {
	return keysFile
}

// LoadUsageStats loads usage statistics
func LoadUsageStats() (*UsageStats, error) {
	if err := EnsureConfigDir(); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %v", err)
	}

	if _, err := os.Stat(usageFile); os.IsNotExist(err) {
		// Create default usage stats
		stats := &UsageStats{
			FreeRuns:    0,
			LastUpsell:  time.Time{},
			UpsellShown: false,
		}
		if err := SaveUsageStats(stats); err != nil {
			return nil, err
		}
		return stats, nil
	}

	data, err := os.ReadFile(usageFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read usage file: %v", err)
	}

	var stats UsageStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse usage file: %v", err)
	}

	return &stats, nil
}

// SaveUsageStats saves usage statistics
func SaveUsageStats(stats *UsageStats) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal usage stats: %v", err)
	}

	// Atomic write: write to temp file first, then rename
	tempFile := usageFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write temp usage file: %v", err)
	}

	if err := os.Rename(tempFile, usageFile); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to rename usage file: %v", err)
	}

	return nil
}

// IncrementFreeRun increments the free run counter
func IncrementFreeRun() error {
	stats, err := LoadUsageStats()
	if err != nil {
		return err
	}

	stats.FreeRuns++
	return SaveUsageStats(stats)
}

// ShouldShowUpsell checks if we should show an upsell message
func ShouldShowUpsell() (bool, error) {
	stats, err := LoadUsageStats()
	if err != nil {
		return false, err
	}

	// Show upsell after 3rd run and not more than once per day
	if stats.FreeRuns >= 3 && !stats.UpsellShown {
		// Check if we've shown upsell in the last 24 hours
		if time.Since(stats.LastUpsell) > 24*time.Hour {
			return true, nil
		}
	}

	return false, nil
}

// MarkUpsellShown marks that we've shown an upsell message
func MarkUpsellShown() error {
	stats, err := LoadUsageStats()
	if err != nil {
		return err
	}

	stats.UpsellShown = true
	stats.LastUpsell = time.Now()
	return SaveUsageStats(stats)
}
