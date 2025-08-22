package utils

import (
	"fmt"
	"os"
)

// GetPassphrase retrieves the passphrase from flags or prompts user
func GetPassphrase(pass, passFile string) (string, error) {
	if pass != "" {
		return pass, nil
	}

	if passFile != "" {
		data, err := os.ReadFile(passFile)
		if err != nil {
			return "", fmt.Errorf("failed to read passphrase file: %v", err)
		}
		// Remove trailing newline if present
		if len(data) > 0 && data[len(data)-1] == '\n' {
			data = data[:len(data)-1]
		}
		return string(data), nil
	}

	// Prompt user for passphrase
	fmt.Print("Enter passphrase: ")
	var passphrase string
	fmt.Scanln(&passphrase)
	return passphrase, nil
}
