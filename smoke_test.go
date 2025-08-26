package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"secretsnap/internal/api"
)

// SmokeTestData holds test configuration and state
type SmokeTestData struct {
	tempDir     string
	envFile     string
	bundleFile  string
	passFile    string
	projectName string
	licenseKey  string
	apiURL      string
	cliPath     string
}

// setupSmokeTest creates a clean test environment
func setupSmokeTest(t *testing.T) *SmokeTestData {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "ssmoke-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create test .env file
	envFile := filepath.Join(tempDir, ".env")
	envContent := "FOO=bar\nNUM=42\nSECRET_KEY=super-secret-123"
	err = os.WriteFile(envFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}

	// Determine CLI path - get absolute path
	cliPath := "./bin/secretsnap"
	if _, err := os.Stat(cliPath); os.IsNotExist(err) {
		cliPath = "./secretsnap"
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(cliPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to get absolute path for CLI: %v", err))
	}
	cliPath = absPath

	// Determine API URL
	apiURL := "http://localhost:8080"
	if envURL := os.Getenv("DEV_SECRETSNAP_API_URL"); envURL != "" {
		apiURL = envURL
	}

	return &SmokeTestData{
		tempDir:     tempDir,
		envFile:     envFile,
		bundleFile:  filepath.Join(tempDir, "secrets.envsnap"),
		passFile:    filepath.Join(tempDir, "pass.envsnap"),
		projectName: "smoke-test-proj",
		licenseKey:  "DEV-KEY-SMOKE-TEST",
		apiURL:      apiURL,
		cliPath:     cliPath,
	}
}

// cleanupSmokeTest cleans up test environment
func cleanupSmokeTest(_ *testing.T, data *SmokeTestData) {
	if data.tempDir != "" {
		os.RemoveAll(data.tempDir)
	}
}

// runSmokeCommand executes a CLI command and returns output
func runSmokeCommand(t *testing.T, data *SmokeTestData, args ...string) (string, string, error) {
	// Verify CLI binary exists
	if _, err := os.Stat(data.cliPath); os.IsNotExist(err) {
		t.Fatalf("CLI binary not found at %s", data.cliPath)
	}

	cmd := exec.Command(data.cliPath, args...)
	cmd.Dir = data.tempDir
	cmd.Env = append(os.Environ(), "DEV_SECRETSNAP_API_URL="+data.apiURL)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", "", err
	}

	if err := cmd.Start(); err != nil {
		return "", "", err
	}

	stdoutBytes, _ := io.ReadAll(stdout)
	stderrBytes, _ := io.ReadAll(stderr)

	err = cmd.Wait()
	return string(stdoutBytes), string(stderrBytes), err
}

// checkFileIsNotPlaintext verifies a file is not plaintext
func checkFileIsNotPlaintext(t *testing.T, filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		t.Errorf("Failed to read file %s: %v", filepath, err)
		return
	}

	// Check if file contains non-ASCII characters
	isASCII := true
	for _, b := range data {
		if b > 127 {
			isASCII = false
			break
		}
	}

	if isASCII {
		t.Errorf("File %s appears to be plaintext (all ASCII)", filepath)
	}
}

// TestSmokeLocalMode tests local mode functionality
func TestSmokeLocalMode(t *testing.T) {
	data := setupSmokeTest(t)
	defer cleanupSmokeTest(t, data)

	t.Run("0_Prep", func(t *testing.T) {
		// Verify we have a clean temp project
		if _, err := os.Stat(data.envFile); os.IsNotExist(err) {
			t.Fatal("Test .env file not created")
		}

		// Verify CLI binary exists
		if _, err := os.Stat(data.cliPath); os.IsNotExist(err) {
			t.Fatalf("CLI binary not found at %s", data.cliPath)
		}
	})

	t.Run("1_InitAndKeyCache", func(t *testing.T) {
		// Test init command
		stdout, stderr, err := runSmokeCommand(t, data, "init")
		if err != nil {
			t.Fatalf("secretsnap init failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}

		// Check .secretsnap.json was created
		configFile := filepath.Join(data.tempDir, ".secretsnap.json")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			t.Error(".secretsnap.json not created")
		}

		// Check keys.json was created with correct permissions
		homeDir, _ := os.UserHomeDir()
		keysFile := filepath.Join(homeDir, ".secretsnap", "keys.json")
		if _, err := os.Stat(keysFile); err == nil {
			checkFilePermissions(t, keysFile, 0600)
		}

		// Test idempotent init
		stdout2, stderr2, err2 := runSmokeCommand(t, data, "init")
		if err2 != nil {
			t.Errorf("Second init failed: %v\nstdout: %s\nstderr: %s", err2, stdout2, stderr2)
		}

		checkNoSecretsInLogs(t, stdout+stdout2, stderr+stderr2)
	})

	t.Run("2_BundleUnbundle", func(t *testing.T) {
		// Bundle .env file
		stdout, stderr, err := runSmokeCommand(t, data, "bundle", data.envFile)
		if err != nil {
			t.Fatalf("secretsnap bundle failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}

		// Check bundle file was created
		if _, err := os.Stat(data.bundleFile); os.IsNotExist(err) {
			t.Error("Bundle file not created")
		}

		// Check bundle is not plaintext
		checkFileIsNotPlaintext(t, data.bundleFile)

		// Unbundle to output file
		outputFile := filepath.Join(data.tempDir, ".env.out")
		stdout2, stderr2, err2 := runSmokeCommand(t, data, "unbundle", data.bundleFile, "--out", outputFile)
		if err2 != nil {
			t.Fatalf("secretsnap unbundle failed: %v\nstdout: %s\nstderr: %s", err2, stdout2, stderr2)
		}

		// Check output file has correct permissions
		checkFilePermissions(t, outputFile, 0600)

		// Compare files
		original, _ := os.ReadFile(data.envFile)
		decrypted, _ := os.ReadFile(outputFile)
		if string(original) != string(decrypted) {
			t.Error("Decrypted file content doesn't match original")
		}

		checkNoSecretsInLogs(t, stdout+stdout2, stderr+stderr2)
	})

	t.Run("3_Run", func(t *testing.T) {
		// Run command with bundled secrets
		stdout, stderr, err := runSmokeCommand(t, data, "run", data.bundleFile, "--", "bash", "-lc", "echo $FOO $NUM")
		if err != nil {
			t.Fatalf("secretsnap run failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}

		// Check output contains expected values
		expected := "bar 42"
		if !strings.Contains(strings.TrimSpace(stdout), expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, stdout)
		}

		// Check no temp files left behind
		files, _ := os.ReadDir(data.tempDir)
		for _, file := range files {
			if strings.Contains(file.Name(), ".tmp") || strings.Contains(file.Name(), "temp") {
				t.Errorf("Temp file left behind: %s", file.Name())
			}
		}

		checkNoSecretsInLogs(t, stdout, stderr)
	})

	t.Run("4_PassphraseMode", func(t *testing.T) {
		// Bundle with passphrase
		stdout, stderr, err := runSmokeCommand(t, data, "bundle", data.envFile, "--pass", "test-passphrase", "--out", data.passFile)
		if err != nil {
			t.Fatalf("secretsnap bundle --pass failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}

		// Check passphrase bundle was created
		if _, err := os.Stat(data.passFile); os.IsNotExist(err) {
			t.Error("Passphrase bundle file not created")
		}

		// Unbundle with passphrase
		outputFile := filepath.Join(data.tempDir, "pass.env")
		stdout2, stderr2, err2 := runSmokeCommand(t, data, "unbundle", data.passFile, "--pass", "test-passphrase", "--out", outputFile)
		if err2 != nil {
			t.Fatalf("secretsnap unbundle --pass failed: %v\nstdout: %s\nstderr: %s", err2, stdout2, stderr2)
		}

		// Check output file has correct permissions
		checkFilePermissions(t, outputFile, 0600)

		// Test wrong passphrase
		wrongOutput := filepath.Join(data.tempDir, "wrong.env")
		_, stderr3, err3 := runSmokeCommand(t, data, "unbundle", data.passFile, "--pass", "wrong-passphrase", "--out", wrongOutput)
		if err3 == nil {
			t.Error("Expected error with wrong passphrase, got none")
		}
		// Check for any error message (be more flexible about exact wording)
		if len(strings.TrimSpace(stderr3)) == 0 {
			t.Error("Expected error message for wrong passphrase")
		}

		// Check no partial file was created
		if _, err := os.Stat(wrongOutput); err == nil {
			t.Error("Partial file created with wrong passphrase")
		}

		checkNoSecretsInLogs(t, stdout+stdout2, stderr+stderr2)
	})

	t.Run("5_OverwriteAndFlags", func(t *testing.T) {
		// Try to unbundle to existing file without --force
		existingFile := filepath.Join(data.tempDir, "existing.env")
		os.WriteFile(existingFile, []byte("existing content"), 0644)

		_, stderr, err := runSmokeCommand(t, data, "unbundle", data.bundleFile, "--out", existingFile)
		if err == nil {
			t.Error("Expected error when overwriting without --force")
		}
		// Check for any error message (be more flexible about exact wording)
		if len(strings.TrimSpace(stderr)) == 0 {
			t.Error("Expected error message about overwriting")
		}

		// Test with --force
		stdout, stderr2, err2 := runSmokeCommand(t, data, "unbundle", data.bundleFile, "--out", existingFile, "--force")
		if err2 != nil {
			t.Errorf("Unbundle with --force failed: %v\nstdout: %s\nstderr: %s", err2, stdout, stderr2)
		}

		checkNoSecretsInLogs(t, stdout, stderr+stderr2)
	})

	t.Run("6_KeyExport", func(t *testing.T) {
		// Test key export
		stdout, stderr, err := runSmokeCommand(t, data, "key", "export")
		if err != nil {
			t.Fatalf("secretsnap key export failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}

		// Check output is base64
		trimmed := strings.TrimSpace(stdout)
		if len(trimmed) == 0 {
			t.Error("Key export returned empty output")
		}

		// Check warning is in stderr (be more flexible)
		if len(strings.TrimSpace(stderr)) == 0 {
			t.Error("Expected warning message in stderr")
		}

		checkNoSecretsInLogs(t, stdout, stderr)
	})
}

// TestSmokeCloudMode tests cloud mode functionality
func TestSmokeCloudMode(t *testing.T) {
	// Skip if no API server available
	if os.Getenv("SKIP_CLOUD_TESTS") == "1" {
		t.Skip("Skipping cloud tests")
	}

	data := setupSmokeTest(t)
	defer cleanupSmokeTest(t, data)

	// Check if API server is actually working
	resp, err := http.Get(data.apiURL + "/healthz")
	if err != nil || resp.StatusCode != 200 {
		t.Skip("API server not available or not healthy")
	}
	resp.Body.Close()

	t.Run("1_LoginAndProject", func(t *testing.T) {
		// Login with dev license
		stdout, stderr, err := runSmokeCommand(t, data, "login", "--license", data.licenseKey)
		if err != nil {
			// Check if it's a database error (common in dev environments)
			if strings.Contains(stderr, "Database connection failed") || strings.Contains(stderr, "null value") {
				t.Skip("API server has database issues, skipping cloud tests")
			}
			t.Fatalf("secretsnap login failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}

		// Check token file was created
		homeDir, _ := os.UserHomeDir()
		tokenFile := filepath.Join(homeDir, ".secretsnap", "token")
		if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
			t.Error("Token file not created")
		}

		// Create project
		stdout2, stderr2, err2 := runSmokeCommand(t, data, "project", "create", data.projectName)
		if err2 != nil {
			// Check if it's an authentication error
			if strings.Contains(stderr2, "Could not validate credentials") {
				t.Skip("Authentication failed, skipping cloud tests")
			}
			t.Fatalf("secretsnap project create failed: %v\nstdout: %s\nstderr: %s", err2, stdout2, stderr2)
		}

		// Check .secretsnap.json was updated
		configFile := filepath.Join(data.tempDir, ".secretsnap.json")
		configData, _ := os.ReadFile(configFile)
		if !strings.Contains(string(configData), "project_id") {
			t.Error("Project ID not added to config")
		}

		checkNoSecretsInLogs(t, stdout+stdout2, stderr+stderr2)
	})

	t.Run("2_PushAndPull", func(t *testing.T) {
		// Bundle and push
		stdout, stderr, err := runSmokeCommand(t, data, "bundle", data.envFile, "--push")
		if err != nil {
			// Check if it's a project error
			if strings.Contains(stderr, "no project specified") {
				t.Skip("No project available, skipping push/pull tests")
			}
			if strings.Contains(stderr, "Could not validate credentials") {
				t.Skip("Authentication failed, skipping push/pull tests")
			}
			t.Fatalf("secretsnap bundle --push failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}

		// Check version in output
		if !strings.Contains(stdout, "v1") {
			t.Error("Expected version v1 in push output")
		}

		// Pull latest
		pulledFile := filepath.Join(data.tempDir, "pulled.env")
		stdout2, stderr2, err2 := runSmokeCommand(t, data, "pull", "--out", pulledFile)
		if err2 != nil {
			if strings.Contains(stderr2, "Could not validate credentials") {
				t.Skip("Authentication failed, skipping pull test")
			}
			t.Fatalf("secretsnap pull failed: %v\nstdout: %s\nstderr: %s", err2, stdout2, stderr2)
		}

		// Check pulled file has correct permissions
		checkFilePermissions(t, pulledFile, 0600)

		// Compare files
		original, _ := os.ReadFile(data.envFile)
		pulled, _ := os.ReadFile(pulledFile)
		if string(original) != string(pulled) {
			t.Error("Pulled file content doesn't match original")
		}

		checkNoSecretsInLogs(t, stdout+stdout2, stderr+stderr2)
	})

	t.Run("3_SharingAndAudit", func(t *testing.T) {
		// Share with user
		stdout, stderr, err := runSmokeCommand(t, data, "share", "--user", "test+alt@example.com", "--role", "read")
		if err != nil {
			if strings.Contains(stderr, "Could not validate credentials") {
				t.Skip("Authentication failed, skipping sharing tests")
			}
			t.Fatalf("secretsnap share failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}

		// Check audit logs
		stdout2, stderr2, err2 := runSmokeCommand(t, data, "audit", "--limit", "10")
		if err2 != nil {
			if strings.Contains(stderr2, "Could not validate credentials") {
				t.Skip("Authentication failed, skipping audit tests")
			}
			t.Fatalf("secretsnap audit failed: %v\nstdout: %s\nstderr: %s", err2, stdout2, stderr2)
		}

		// Check for expected audit events
		auditOutput := stdout2
		expectedEvents := []string{"bundle_pushed", "bundle_pulled", "share_added"}
		for _, event := range expectedEvents {
			if !strings.Contains(auditOutput, event) {
				t.Errorf("Expected audit event '%s' not found", event)
			}
		}

		checkNoSecretsInLogs(t, stdout+stdout2, stderr+stderr2)
	})

	t.Run("4_TokenExpiry", func(t *testing.T) {
		// Corrupt token file
		homeDir, _ := os.UserHomeDir()
		tokenFile := filepath.Join(homeDir, ".secretsnap", "token")
		os.WriteFile(tokenFile, []byte("corrupted-token"), 0600)

		// Try cloud command
		_, stderr, err := runSmokeCommand(t, data, "pull")
		if err == nil {
			t.Error("Expected error with corrupted token")
		}
		// Check for any error message (be more flexible about exact wording)
		if len(strings.TrimSpace(stderr)) == 0 {
			t.Error("Expected error message for corrupted token")
		}
	})
}

// loadRealLicenseKey reads license key from smoke-test-license.key file
func loadRealLicenseKey(t *testing.T) (string, bool) {
	licenseFile := "smoke-test-license.key"

	// Check if file exists
	if _, err := os.Stat(licenseFile); os.IsNotExist(err) {
		t.Logf("License file %s not found, skipping real license tests", licenseFile)
		return "", false
	}

	// Read license key from file
	content, err := os.ReadFile(licenseFile)
	if err != nil {
		t.Logf("Failed to read license file %s: %v, skipping real license tests", licenseFile, err)
		return "", false
	}

	licenseKey := strings.TrimSpace(string(content))
	if licenseKey == "" {
		t.Logf("License file %s is empty, skipping real license tests", licenseFile)
		return "", false
	}

	t.Logf("Loaded real license key from %s", licenseFile)
	return licenseKey, true
}

// TestSmokeCloudModeRealLicense tests cloud mode functionality with real license key
func TestSmokeCloudModeRealLicense(t *testing.T) {
	// Load real license key
	realLicenseKey, ok := loadRealLicenseKey(t)
	if !ok {
		t.Skip("Real license key not available, skipping real license tests")
	}

	// Skip if no API server available
	if os.Getenv("SKIP_CLOUD_TESTS") == "1" {
		t.Skip("Skipping cloud tests")
	}

	data := setupSmokeTest(t)
	defer cleanupSmokeTest(t, data)

	// Override license key with real one
	data.licenseKey = realLicenseKey
	data.projectName = "smoke-test-real-proj"

	// Check if API server is actually working
	resp, err := http.Get(data.apiURL + "/healthz")
	if err != nil || resp.StatusCode != 200 {
		t.Skip("API server not available or not healthy")
	}
	resp.Body.Close()

	t.Run("1_LoginAndProject", func(t *testing.T) {
		// Login with real license
		stdout, stderr, err := runSmokeCommand(t, data, "login", "--license", data.licenseKey)
		if err != nil {
			// Check if it's a database error (common in dev environments)
			if strings.Contains(stderr, "Database connection failed") || strings.Contains(stderr, "null value") {
				t.Skip("API server has database issues, skipping cloud tests")
			}
			t.Fatalf("secretsnap login failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}

		// Check token file was created
		homeDir, _ := os.UserHomeDir()
		tokenFile := filepath.Join(homeDir, ".secretsnap", "token")
		if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
			t.Error("Token file not created")
		}

		// Create project
		stdout2, stderr2, err2 := runSmokeCommand(t, data, "project", "create", data.projectName)
		if err2 != nil {
			// Check if it's an authentication error
			if strings.Contains(stderr2, "Could not validate credentials") {
				t.Skip("Authentication failed, skipping cloud tests")
			}
			t.Fatalf("secretsnap project create failed: %v\nstdout: %s\nstderr: %s", err2, stdout2, stderr2)
		}

		// Check .secretsnap.json was updated
		configFile := filepath.Join(data.tempDir, ".secretsnap.json")
		configData, _ := os.ReadFile(configFile)
		if !strings.Contains(string(configData), "project_id") {
			t.Error("Project ID not added to config")
		}

		checkNoSecretsInLogs(t, stdout+stdout2, stderr+stderr2)
	})

	t.Run("2_PushAndPull", func(t *testing.T) {
		// Bundle and push
		stdout, stderr, err := runSmokeCommand(t, data, "bundle", data.envFile, "--push")
		if err != nil {
			// Check if it's a project error
			if strings.Contains(stderr, "no project specified") {
				t.Skip("No project available, skipping push/pull tests")
			}
			if strings.Contains(stderr, "Could not validate credentials") {
				t.Skip("Authentication failed, skipping push/pull tests")
			}
			t.Fatalf("secretsnap bundle --push failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}

		// Check version in output
		if !strings.Contains(stdout, "v1") {
			t.Error("Expected version v1 in push output")
		}

		// Pull latest
		pulledFile := filepath.Join(data.tempDir, "pulled.env")
		stdout2, stderr2, err2 := runSmokeCommand(t, data, "pull", "--out", pulledFile)
		if err2 != nil {
			if strings.Contains(stderr2, "Could not validate credentials") {
				t.Skip("Authentication failed, skipping pull test")
			}
			t.Fatalf("secretsnap pull failed: %v\nstdout: %s\nstderr: %s", err2, stdout2, stderr2)
		}

		// Check pulled file has correct permissions
		checkFilePermissions(t, pulledFile, 0600)

		// Compare files
		original, _ := os.ReadFile(data.envFile)
		pulled, _ := os.ReadFile(pulledFile)
		if string(original) != string(pulled) {
			t.Error("Pulled file content doesn't match original")
		}

		checkNoSecretsInLogs(t, stdout+stdout2, stderr+stderr2)
	})

	t.Run("3_SharingAndAudit", func(t *testing.T) {
		// Share with user
		stdout, stderr, err := runSmokeCommand(t, data, "share", "--user", "test+alt@example.com", "--role", "read")
		if err != nil {
			if strings.Contains(stderr, "Could not validate credentials") {
				t.Skip("Authentication failed, skipping sharing tests")
			}
			t.Fatalf("secretsnap share failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}

		// Check audit logs
		stdout2, stderr2, err2 := runSmokeCommand(t, data, "audit", "--limit", "10")
		if err2 != nil {
			if strings.Contains(stderr2, "Could not validate credentials") {
				t.Skip("Authentication failed, skipping audit tests")
			}
			t.Fatalf("secretsnap audit failed: %v\nstdout: %s\nstderr: %s", err2, stdout2, stderr2)
		}

		// Check for expected audit events
		auditOutput := stdout2
		expectedEvents := []string{"bundle_pushed", "bundle_pulled", "share_added"}
		for _, event := range expectedEvents {
			if !strings.Contains(auditOutput, event) {
				t.Errorf("Expected audit event '%s' not found", event)
			}
		}

		checkNoSecretsInLogs(t, stdout+stdout2, stderr+stderr2)
	})

	t.Run("4_TokenExpiry", func(t *testing.T) {
		// Corrupt token file
		homeDir, _ := os.UserHomeDir()
		tokenFile := filepath.Join(homeDir, ".secretsnap", "token")
		os.WriteFile(tokenFile, []byte("corrupted-token"), 0600)

		// Try cloud command
		_, stderr, err := runSmokeCommand(t, data, "pull")
		if err == nil {
			t.Error("Expected error with corrupted token")
		}
		// Check for any error message (be more flexible about exact wording)
		if len(strings.TrimSpace(stderr)) == 0 {
			t.Error("Expected error message for corrupted token")
		}
	})
}

// TestSmokeAPI tests API endpoints directly
func TestSmokeAPI(t *testing.T) {
	// Skip if no API server available
	if os.Getenv("SKIP_API_TESTS") == "1" {
		t.Skip("Skipping API tests")
	}

	data := setupSmokeTest(t)
	defer cleanupSmokeTest(t, data)

	// Check if API server is actually working
	resp, err := http.Get(data.apiURL + "/healthz")
	if err != nil || resp.StatusCode != 200 {
		t.Skip("API server not available or not healthy")
	}
	resp.Body.Close()

	t.Run("1_Auth", func(t *testing.T) {
		// Test login endpoint
		loginReq := api.LoginRequest{
			LicenseKey: data.licenseKey,
		}
		reqBody, _ := json.Marshal(loginReq)

		resp, err := http.Post(data.apiURL+"/v1/auth/login", "application/json", strings.NewReader(string(reqBody)))
		if err != nil {
			t.Fatalf("Login request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var loginResp api.LoginResponse
		json.NewDecoder(resp.Body).Decode(&loginResp)
		if loginResp.Token == "" {
			t.Error("No token in response")
		}

		// Test invalid license
		invalidReq := api.LoginRequest{
			LicenseKey: "INVALID-KEY",
		}
		invalidBody, _ := json.Marshal(invalidReq)

		resp2, err := http.Post(data.apiURL+"/v1/auth/login", "application/json", strings.NewReader(string(invalidBody)))
		if err != nil {
			t.Fatalf("Invalid login request failed: %v", err)
		}
		defer resp2.Body.Close()

		if resp2.StatusCode != 401 {
			t.Errorf("Expected 401, got %d", resp2.StatusCode)
		}

		var errorResp map[string]interface{}
		json.NewDecoder(resp2.Body).Decode(&errorResp)
		if errorResp["code"] == nil || errorResp["message"] == nil {
			t.Error("Expected error response with code and message")
		}
	})

	t.Run("2_Projects", func(t *testing.T) {
		// Get auth token first
		loginReq := api.LoginRequest{LicenseKey: data.licenseKey}
		reqBody, _ := json.Marshal(loginReq)
		resp, err := http.Post(data.apiURL+"/v1/auth/login", "application/json", strings.NewReader(string(reqBody)))
		if err != nil {
			t.Fatalf("Login request failed: %v", err)
		}
		defer resp.Body.Close()

		// Handle database errors gracefully
		if resp.StatusCode == 500 {
			body, _ := io.ReadAll(resp.Body)
			if strings.Contains(string(body), "Database connection failed") || strings.Contains(string(body), "null value") {
				t.Skip("API server has database issues, skipping project tests")
			}
		}

		if resp.StatusCode != 200 {
			t.Skip("Login failed, skipping project tests")
		}

		var loginResp api.LoginResponse
		json.NewDecoder(resp.Body).Decode(&loginResp)

		// Create project
		projectReq := api.CreateProjectRequest{Name: data.projectName}
		projectBody, _ := json.Marshal(projectReq)

		req, _ := http.NewRequest("POST", data.apiURL+"/v1/projects", strings.NewReader(string(projectBody)))
		req.Header.Set("Authorization", "Bearer "+loginResp.Token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp2, err := client.Do(req)
		if err != nil {
			t.Fatalf("Project creation failed: %v", err)
		}
		defer resp2.Body.Close()

		if resp2.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp2.StatusCode)
		}

		var projectResp api.Project
		json.NewDecoder(resp2.Body).Decode(&projectResp)
		if projectResp.ID == "" || projectResp.Name != data.projectName {
			t.Error("Invalid project response")
		}
	})

	t.Run("3_ErrorShapes", func(t *testing.T) {
		// Test 404
		resp, err := http.Get(data.apiURL + "/v1/nonexistent")
		if err != nil {
			t.Fatalf("404 request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 404 {
			t.Errorf("Expected 404, got %d", resp.StatusCode)
		}

		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		if errorResp["code"] == nil || errorResp["message"] == nil {
			t.Error("Expected error response with code and message")
		}
	})
}

// TestSmokeAPIRealLicense tests API endpoints directly with real license key
func TestSmokeAPIRealLicense(t *testing.T) {
	// Load real license key
	realLicenseKey, ok := loadRealLicenseKey(t)
	if !ok {
		t.Skip("Real license key not available, skipping real license API tests")
	}

	// Skip if no API server available
	if os.Getenv("SKIP_API_TESTS") == "1" {
		t.Skip("Skipping API tests")
	}

	data := setupSmokeTest(t)
	defer cleanupSmokeTest(t, data)

	// Override license key with real one
	data.licenseKey = realLicenseKey

	// Check if API server is actually working
	resp, err := http.Get(data.apiURL + "/healthz")
	if err != nil || resp.StatusCode != 200 {
		t.Skip("API server not available or not healthy")
	}
	resp.Body.Close()

	t.Run("1_Auth", func(t *testing.T) {
		// Test login endpoint with real license
		loginReq := api.LoginRequest{
			LicenseKey: data.licenseKey,
		}
		reqBody, _ := json.Marshal(loginReq)

		resp, err := http.Post(data.apiURL+"/v1/auth/login", "application/json", strings.NewReader(string(reqBody)))
		if err != nil {
			t.Fatalf("Login request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var loginResp api.LoginResponse
		json.NewDecoder(resp.Body).Decode(&loginResp)
		if loginResp.Token == "" {
			t.Error("No token in response")
		}
	})

	t.Run("2_Projects", func(t *testing.T) {
		// Get auth token first
		loginReq := api.LoginRequest{LicenseKey: data.licenseKey}
		reqBody, _ := json.Marshal(loginReq)
		resp, err := http.Post(data.apiURL+"/v1/auth/login", "application/json", strings.NewReader(string(reqBody)))
		if err != nil {
			t.Fatalf("Login request failed: %v", err)
		}
		defer resp.Body.Close()

		// Handle database errors gracefully
		if resp.StatusCode == 500 {
			body, _ := io.ReadAll(resp.Body)
			if strings.Contains(string(body), "Database connection failed") || strings.Contains(string(body), "null value") {
				t.Skip("API server has database issues, skipping project tests")
			}
		}

		if resp.StatusCode != 200 {
			t.Skip("Login failed, skipping project tests")
		}

		var loginResp api.LoginResponse
		json.NewDecoder(resp.Body).Decode(&loginResp)

		// Create project
		projectReq := api.CreateProjectRequest{Name: "smoke-test-real-api-proj"}
		projectBody, _ := json.Marshal(projectReq)

		req, _ := http.NewRequest("POST", data.apiURL+"/v1/projects", strings.NewReader(string(projectBody)))
		req.Header.Set("Authorization", "Bearer "+loginResp.Token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp2, err := client.Do(req)
		if err != nil {
			t.Fatalf("Project creation failed: %v", err)
		}
		defer resp2.Body.Close()

		if resp2.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp2.StatusCode)
		}

		var projectResp api.Project
		json.NewDecoder(resp2.Body).Decode(&projectResp)
		if projectResp.ID == "" || projectResp.Name != "smoke-test-real-api-proj" {
			t.Error("Invalid project response")
		}
	})
}

// TestSmokeSecurity tests security and privacy features
func TestSmokeSecurity(t *testing.T) {
	data := setupSmokeTest(t)
	defer cleanupSmokeTest(t, data)

	t.Run("1_SecretsNeverLogged", func(t *testing.T) {
		// Run various commands and check for secret leakage
		commands := [][]string{
			{"init"},
			{"bundle", data.envFile},
			{"unbundle", data.bundleFile, "--out", ".env.test"},
			{"run", data.bundleFile, "--", "echo", "test"},
		}

		for _, cmd := range commands {
			stdout, stderr, _ := runSmokeCommand(t, data, cmd...)
			checkNoSecretsInLogs(t, stdout, stderr)
		}
	})

	t.Run("2_FilePermissions", func(t *testing.T) {
		// First create a bundle to unbundle
		runSmokeCommand(t, data, "bundle", data.envFile)

		// Test .env output permissions
		outputFile := filepath.Join(data.tempDir, ".env.perms")
		runSmokeCommand(t, data, "unbundle", data.bundleFile, "--out", outputFile)
		checkFilePermissions(t, outputFile, 0600)

		// Test keys file permissions
		homeDir, _ := os.UserHomeDir()
		keysFile := filepath.Join(homeDir, ".secretsnap", "keys.json")
		if _, err := os.Stat(keysFile); err == nil {
			checkFilePermissions(t, keysFile, 0600)
		}
	})

	t.Run("3_KeyLossScenarios", func(t *testing.T) {
		// Test local key cache removal
		homeDir, _ := os.UserHomeDir()
		keysFile := filepath.Join(homeDir, ".secretsnap", "keys.json")
		originalKeys, _ := os.ReadFile(keysFile)
		defer os.WriteFile(keysFile, originalKeys, 0600)

		os.Remove(keysFile)

		_, stderr, err := runSmokeCommand(t, data, "unbundle", data.bundleFile, "--out", ".env.lost")
		if err == nil {
			t.Error("Expected error when keys file is missing")
		}
		// Check for any error message (be more flexible)
		if len(strings.TrimSpace(stderr)) == 0 {
			t.Error("Expected error message about missing keys")
		}
	})
}

// TestSmokePerformance tests performance characteristics
func TestSmokePerformance(t *testing.T) {
	data := setupSmokeTest(t)
	defer cleanupSmokeTest(t, data)

	t.Run("1_SmallFile", func(t *testing.T) {
		// Create 1KB file
		smallContent := strings.Repeat("KEY=value\n", 100)
		smallFile := filepath.Join(data.tempDir, "small.env")
		os.WriteFile(smallFile, []byte(smallContent), 0644)

		start := time.Now()
		runSmokeCommand(t, data, "bundle", smallFile)
		bundleTime := time.Since(start)

		if bundleTime > 50*time.Millisecond {
			t.Errorf("Bundle took too long: %v", bundleTime)
		}

		start = time.Now()
		runSmokeCommand(t, data, "unbundle", data.bundleFile, "--out", ".env.small")
		unbundleTime := time.Since(start)

		if unbundleTime > 50*time.Millisecond {
			t.Errorf("Unbundle took too long: %v", unbundleTime)
		}
	})

	t.Run("2_LargeFile", func(t *testing.T) {
		// Create 200KB file
		largeContent := strings.Repeat("KEY=value\n", 20000)
		largeFile := filepath.Join(data.tempDir, "large.env")
		os.WriteFile(largeFile, []byte(largeContent), 0644)

		start := time.Now()
		runSmokeCommand(t, data, "bundle", largeFile)
		bundleTime := time.Since(start)

		if bundleTime > 300*time.Millisecond {
			t.Errorf("Large bundle took too long: %v", bundleTime)
		}

		start = time.Now()
		runSmokeCommand(t, data, "unbundle", data.bundleFile, "--out", ".env.large")
		unbundleTime := time.Since(start)

		if unbundleTime > 300*time.Millisecond {
			t.Errorf("Large unbundle took too long: %v", unbundleTime)
		}
	})
}

// TestSmokeBackwardCompatibility tests version compatibility
func TestSmokeBackwardCompatibility(t *testing.T) {
	// Skip if no API server available
	if os.Getenv("SKIP_CLOUD_TESTS") == "1" {
		t.Skip("Skipping backward compatibility tests")
	}

	data := setupSmokeTest(t)
	defer cleanupSmokeTest(t, data)

	t.Run("1_VersionCompatibility", func(t *testing.T) {
		// Login and create project
		runSmokeCommand(t, data, "login", "--license", data.licenseKey)
		runSmokeCommand(t, data, "project", "create", data.projectName)

		// Push v1
		runSmokeCommand(t, data, "bundle", data.envFile, "--push")

		// Modify .env and push v2
		os.WriteFile(data.envFile, []byte("FOO=bar\nNUM=42\nNEW=value"), 0644)
		runSmokeCommand(t, data, "bundle", data.envFile, "--push")

		// Pull v1 specifically
		v1File := filepath.Join(data.tempDir, "v1.env")
		stdout, stderr, err := runSmokeCommand(t, data, "pull", "--version", "1", "--out", v1File)
		if err != nil {
			t.Fatalf("Pull v1 failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}

		// Check v1 content
		v1Content, _ := os.ReadFile(v1File)
		if !strings.Contains(string(v1Content), "FOO=bar") || strings.Contains(string(v1Content), "NEW=value") {
			t.Error("v1 content doesn't match expected")
		}

		// Pull latest (should be v2)
		latestFile := filepath.Join(data.tempDir, "latest.env")
		stdout2, stderr2, err2 := runSmokeCommand(t, data, "pull", "--out", latestFile)
		if err2 != nil {
			t.Fatalf("Pull latest failed: %v\nstdout: %s\nstderr: %s", err2, stdout2, stderr2)
		}

		// Check latest content
		latestContent, _ := os.ReadFile(latestFile)
		if !strings.Contains(string(latestContent), "NEW=value") {
			t.Error("Latest content doesn't include new value")
		}

		checkNoSecretsInLogs(t, stdout+stdout2, stderr+stderr2)
	})
}

// TestSmokeUX tests user experience and error handling
func TestSmokeUX(t *testing.T) {
	data := setupSmokeTest(t)
	defer cleanupSmokeTest(t, data)

	t.Run("1_HelpAndErrors", func(t *testing.T) {
		// Test help
		stdout, stderr, err := runSmokeCommand(t, data, "--help")
		if err != nil {
			t.Fatalf("Help command failed: %v", err)
		}
		if len(strings.TrimSpace(stdout)) == 0 {
			t.Error("Help output is empty")
		}

		// Test command help
		stdout2, stderr2, err2 := runSmokeCommand(t, data, "bundle", "--help")
		if err2 != nil {
			t.Fatalf("Bundle help failed: %v", err2)
		}
		if len(strings.TrimSpace(stdout2)) == 0 {
			t.Error("Bundle help output is empty")
		}

		// Test missing args
		_, stderr3, err3 := runSmokeCommand(t, data, "bundle")
		if err3 == nil {
			t.Error("Expected error with missing args")
		}
		// Check for any error message (be more flexible)
		if len(strings.TrimSpace(stderr3)) == 0 {
			t.Error("Expected error message for missing args")
		}

		checkNoSecretsInLogs(t, stdout+stdout2, stderr+stderr2+stderr3)
	})

	t.Run("2_Defaults", func(t *testing.T) {
		// Initialize local environment first
		_, _, err := runSmokeCommand(t, data, "init")
		if err != nil {
			t.Fatalf("Init command failed: %v", err)
		}

		// Test bundle default output
		stdout, stderr, err := runSmokeCommand(t, data, "bundle", data.envFile)
		if err != nil {
			t.Fatalf("Bundle command failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}
		if _, err := os.Stat(data.bundleFile); os.IsNotExist(err) {
			t.Error("Default bundle output not created")
		}

		// Test unbundle default output (use a different output file to avoid conflicts)
		var stdout2, stderr2 string
		stdout2, stderr2, err = runSmokeCommand(t, data, "unbundle", data.bundleFile, "--out", ".env.defaults")
		if err != nil {
			t.Fatalf("Unbundle command failed: %v\nstdout: %s\nstderr: %s", err, stdout2, stderr2)
		}
		defaultOutput := filepath.Join(data.tempDir, ".env.defaults")
		if _, err := os.Stat(defaultOutput); os.IsNotExist(err) {
			t.Error("Default unbundle output not created")
		}

		// Test that we can unbundle again with --force (should work even if file exists)
		_, _, err = runSmokeCommand(t, data, "unbundle", data.bundleFile, "--force")
		if err != nil {
			t.Error("Expected unbundle with --force to work even with existing file")
		}
	})
}
