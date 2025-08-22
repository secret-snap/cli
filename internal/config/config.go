package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Mode    string `json:"mode"`    // "local" or "cloud"
	Project string `json:"project"` // project name or ID
}

type GlobalConfig struct {
	Token string `json:"token"` // JWT token for cloud mode
}

var (
	configDir  string
	configFile string
	globalDir  string
	globalFile string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get home directory: %v", err))
	}

	configDir = filepath.Join(home, ".secretsnap")
	configFile = filepath.Join(configDir, "config.json")
	globalDir = configDir
	globalFile = filepath.Join(globalDir, "token")
}

func EnsureConfigDir() error {
	return os.MkdirAll(configDir, 0755)
}

func LoadConfig() (*Config, error) {
	if err := EnsureConfigDir(); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %v", err)
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create default config
		config := &Config{
			Mode:    "local",
			Project: "local",
		}
		if err := SaveConfig(config); err != nil {
			return nil, err
		}
		return config, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

func LoadToken() (string, error) {
	if _, err := os.Stat(globalFile); os.IsNotExist(err) {
		return "", nil
	}

	data, err := os.ReadFile(globalFile)
	if err != nil {
		return "", fmt.Errorf("failed to read token file: %v", err)
	}

	return string(data), nil
}

func SaveToken(token string) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	if err := os.WriteFile(globalFile, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to write token file: %v", err)
	}

	return nil
}

func GetConfigPath() string {
	return configFile
}
