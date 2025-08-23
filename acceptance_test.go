package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"secretsnap/internal/api"
	"secretsnap/internal/crypto"
)

// TestServer mocks the API server for testing
type TestServer struct {
	*httptest.Server
	auditLogs []api.AuditLog
}

// TestData holds test configuration and state
type TestData struct {
	tempDir     string
	envFile     string
	bundleFile  string
	passFile    string
	projectName string
	licenseKey  string
	token       string
}

func setupTestServer() *TestServer {
	server := &TestServer{
		auditLogs: make([]api.AuditLog, 0),
	}

	mux := http.NewServeMux()

	// Mock login endpoint
	mux.HandleFunc("/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req api.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Accept any license key for testing
		resp := api.LoginResponse{
			Token: "fake-jwt-token-" + req.LicenseKey,
			User: api.User{
				ID:    "user-123",
				Email: "test@example.com",
				Plan:  "pro",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Mock bundle push endpoint
	mux.HandleFunc("/v1/bundles/push", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req api.BundlePushRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		resp := api.BundlePushResponse{
			UploadURL: server.URL + "/upload/test-bundle",
			BundleID:  "bundle-123",
			S3Key:     "test-bundle-key",
		}

		// Record audit log
		server.auditLogs = append(server.auditLogs, api.AuditLog{
			ID:        "audit-123",
			Action:    "bundle_push",
			Details:   map[string]interface{}{"project_id": req.ProjectID, "size_bytes": req.SizeBytes},
			CreatedAt: time.Now().Format(time.RFC3339),
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Mock bundle finalize endpoint
	mux.HandleFunc("/v1/bundles/finalize", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	// Mock bundle pull endpoint
	mux.HandleFunc("/v1/bundles/pull", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		resp := api.BundlePullResponse{
			DownloadURL: server.URL + "/download/test-bundle",
			DataKey:     "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA=", // base64 encoded 32-byte key
			Version:     1,
		}

		// Record audit log
		server.auditLogs = append(server.auditLogs, api.AuditLog{
			ID:        "audit-124",
			Action:    "bundle_pull",
			Details:   map[string]interface{}{"project_id": r.URL.Query().Get("project_id")},
			CreatedAt: time.Now().Format(time.RFC3339),
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Mock S3 upload endpoint
	mux.HandleFunc("/upload/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Mock S3 download endpoint
	mux.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Return properly encrypted test data using the same key
		testData := []byte("FOO=bar\nSECRET=mysecret123\nAPI_KEY=sk-1234567890abcdef")
		key, _ := base64.StdEncoding.DecodeString("AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA=")

		// Encrypt the test data with the key
		encryptedData, err := crypto.EncryptWithKey(testData, key)
		if err != nil {
			http.Error(w, "Failed to encrypt test data", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(encryptedData)
	})

	// Mock project creation endpoint
	mux.HandleFunc("/v1/projects", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req api.CreateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		resp := api.Project{
			ID:   "proj-" + req.Name,
			Name: req.Name,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Mock share endpoint
	mux.HandleFunc("/v1/shares", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req api.ShareRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Record audit log
		server.auditLogs = append(server.auditLogs, api.AuditLog{
			ID:        "audit-125",
			Action:    "project_share",
			Details:   map[string]interface{}{"project_id": req.ProjectID, "user_email": req.UserEmail, "role": req.Role},
			CreatedAt: time.Now().Format(time.RFC3339),
		})

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	// Mock audit logs endpoint
	mux.HandleFunc("/v1/audit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(server.auditLogs)
	})

	server.Server = httptest.NewServer(mux)
	return server
}

func setupTestData(t *testing.T) *TestData {
	tempDir, err := os.MkdirTemp("", "secretsnap-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create test .env file
	envFile := filepath.Join(tempDir, ".env")
	envContent := "FOO=bar\nSECRET=mysecret123\nAPI_KEY=sk-1234567890abcdef"
	if err := os.WriteFile(envFile, []byte(envContent), 0600); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	// Create test passphrase file
	passFile := filepath.Join(tempDir, "passphrase.txt")
	if err := os.WriteFile(passFile, []byte("test-passphrase"), 0600); err != nil {
		t.Fatalf("Failed to create test passphrase file: %v", err)
	}

	return &TestData{
		tempDir:     tempDir,
		envFile:     envFile,
		bundleFile:  filepath.Join(tempDir, "secrets.envsnap"),
		passFile:    passFile,
		projectName: "test-project",
		licenseKey:  "test-license-key-123",
		token:       "fake-jwt-token",
	}
}

func cleanupTestData(t *testing.T, data *TestData) {
	if err := os.RemoveAll(data.tempDir); err != nil {
		t.Logf("Warning: failed to cleanup temp directory: %v", err)
	}
}

// ensureNoToken ensures there's no token file for license enforcement tests
func ensureNoToken(t *testing.T) {
	tokenFile := filepath.Join(os.Getenv("HOME"), ".secretsnap", "token")
	if _, err := os.Stat(tokenFile); err == nil {
		if err := os.Remove(tokenFile); err != nil {
			t.Logf("Warning: failed to remove token file: %v", err)
		}
	}
}

func runCommand(t *testing.T, data *TestData, args ...string) (string, string, error) {
	// Set environment variables for testing
	env := os.Environ()
	env = append(env, "ALLOW_ALL_LICENSES=1") // Allow fake licenses for testing

	// Set API URL if provided in environment
	if apiURL := os.Getenv("SECRETSNAP_API_URL"); apiURL != "" {
		env = append(env, "SECRETSNAP_API_URL="+apiURL)
	}

	// Get current directory for building
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Build the CLI binary from current directory
	cmd := exec.Command("go", "build", "-o", "secretsnap", ".")
	cmd.Dir = currentDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Run the command from temp directory
	cliPath := filepath.Join(currentDir, "secretsnap")
	cmd = exec.Command(cliPath, args...)
	cmd.Env = env
	cmd.Dir = data.tempDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	stdoutBytes, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("Failed to read stdout: %v", err)
	}

	stderrBytes, err := io.ReadAll(stderr)
	if err != nil {
		t.Fatalf("Failed to read stderr: %v", err)
	}

	err = cmd.Wait()
	return string(stdoutBytes), string(stderrBytes), err
}

func checkFilePermissions(t *testing.T, filepath string, expectedPerm os.FileMode) {
	info, err := os.Stat(filepath)
	if err != nil {
		t.Errorf("Failed to stat file %s: %v", filepath, err)
		return
	}

	actualPerm := info.Mode().Perm()
	if actualPerm != expectedPerm {
		t.Errorf("File %s has permissions %v, expected %v", filepath, actualPerm, expectedPerm)
	}
}

func checkNoSecretsInLogs(t *testing.T, stdout, stderr string) {
	secrets := []string{
		"mysecret123",
		"sk-1234567890abcdef",
		"test-passphrase",
	}

	for _, secret := range secrets {
		if strings.Contains(stdout, secret) {
			t.Errorf("Secret '%s' found in stdout logs", secret)
		}
		if strings.Contains(stderr, secret) {
			t.Errorf("Secret '%s' found in stderr logs", secret)
		}
	}
}

// TestFreeModeCommands tests that free mode commands work without prompts
func TestFreeModeCommands(t *testing.T) {
	data := setupTestData(t)
	defer cleanupTestData(t, data)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "init command",
			args:    []string{"init"},
			wantErr: false,
		},
		{
			name:    "bundle .env command",
			args:    []string{"bundle", data.envFile},
			wantErr: false,
		},
		{
			name:    "unbundle command",
			args:    []string{"unbundle", data.bundleFile, "--out", ".env.decrypted"},
			wantErr: false,
		},
		{
			name:    "run echo $FOO command",
			args:    []string{"run", data.bundleFile, "--", "echo", "$FOO"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCommand(t, data, tt.args...)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
			}

			// Check that no secrets are logged
			checkNoSecretsInLogs(t, stdout, stderr)
		})
	}

	// Check file permissions for decrypted files
	decryptedFile := filepath.Join(data.tempDir, ".env.decrypted")
	if _, err := os.Stat(decryptedFile); err == nil {
		checkFilePermissions(t, decryptedFile, 0600)
	}
}

// TestPassphraseMode tests passphrase mode commands
func TestPassphraseMode(t *testing.T) {
	data := setupTestData(t)
	defer cleanupTestData(t, data)

	// First initialize
	_, _, err := runCommand(t, data, "init")
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "bundle with passphrase",
			args:    []string{"bundle", data.envFile, "--pass", "test-passphrase"},
			wantErr: false,
		},
		{
			name:    "unbundle with passphrase",
			args:    []string{"unbundle", data.bundleFile, "--pass", "test-passphrase", "--out", ".env.pass"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCommand(t, data, tt.args...)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
			}

			// Check that no secrets are logged
			checkNoSecretsInLogs(t, stdout, stderr)
		})
	}

	// Check file permissions for decrypted files
	decryptedFile := filepath.Join(data.tempDir, ".env.pass")
	if _, err := os.Stat(decryptedFile); err == nil {
		checkFilePermissions(t, decryptedFile, 0600)
	}
}

// TestPaidMode tests paid mode commands with fake license
func TestPaidMode(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	data := setupTestData(t)
	defer cleanupTestData(t, data)

	// Set API URL to test server
	os.Setenv("SECRETSNAP_API_URL", server.URL)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "login with fake license",
			args:    []string{"login", "--license", data.licenseKey},
			wantErr: false,
		},
		{
			name:    "bundle with push",
			args:    []string{"bundle", data.envFile, "--push", "--project", data.projectName},
			wantErr: false,
		},
		{
			name:    "pull with output",
			args:    []string{"pull", "--out", ".env.pulled"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCommand(t, data, tt.args...)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
			}

			// Check that no secrets are logged
			checkNoSecretsInLogs(t, stdout, stderr)
		})
	}

	// Check file permissions for pulled files
	pulledFile := filepath.Join(data.tempDir, ".env.pulled")
	if _, err := os.Stat(pulledFile); err == nil {
		checkFilePermissions(t, pulledFile, 0600)
	}
}

// TestAuditLogs tests that audit logs record push/pull/share actions
func TestAuditLogs(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	data := setupTestData(t)
	defer cleanupTestData(t, data)

	// Set API URL to test server
	os.Setenv("SECRETSNAP_API_URL", server.URL)

	// Login first
	_, _, err := runCommand(t, data, "login", "--license", data.licenseKey)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	// Create project first
	_, _, err = runCommand(t, data, "project", "create", data.projectName)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Perform actions that should generate audit logs
	actions := []struct {
		name string
		args []string
	}{
		{
			name: "bundle push",
			args: []string{"bundle", data.envFile, "--push", "--project", data.projectName},
		},
		{
			name: "pull",
			args: []string{"pull", "--project", data.projectName, "--out", ".env.pulled", "--force"},
		},
	}

	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			stdout, stderr, err := runCommand(t, data, action.args...)
			if err != nil {
				t.Errorf("Failed to execute %s: %v\nstdout: %s\nstderr: %s", action.name, err, stdout, stderr)
			}
		})
	}

	// Check audit logs
	if len(server.auditLogs) < 2 {
		t.Errorf("Expected at least 2 audit logs, got %d", len(server.auditLogs))
	}

	// Verify specific audit log entries
	actionsFound := make(map[string]bool)
	for _, log := range server.auditLogs {
		actionsFound[log.Action] = true
	}

	expectedActions := []string{"bundle_push", "bundle_pull"}
	for _, action := range expectedActions {
		if !actionsFound[action] {
			t.Errorf("Expected audit log action '%s' not found", action)
		}
	}
}

// TestFilePermissions tests that files are created with 0600 permissions
func TestFilePermissions(t *testing.T) {
	data := setupTestData(t)
	defer cleanupTestData(t, data)

	// Initialize
	_, _, err := runCommand(t, data, "init")
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Bundle with passphrase
	_, _, err = runCommand(t, data, "bundle", data.envFile, "--pass", "test-passphrase")
	if err != nil {
		t.Fatalf("Failed to bundle: %v", err)
	}

	// Unbundle to different output files
	outputFiles := []string{
		".env.decrypted1",
		".env.decrypted2",
		".env.decrypted3",
	}

	for _, outputFile := range outputFiles {
		_, _, err := runCommand(t, data, "unbundle", data.bundleFile, "--pass", "test-passphrase", "--out", outputFile)
		if err != nil {
			t.Fatalf("Failed to unbundle to %s: %v", outputFile, err)
		}

		// Check file permissions
		filePath := filepath.Join(data.tempDir, outputFile)
		checkFilePermissions(t, filePath, 0600)
	}
}

// TestNoSecretsInLogs tests that secrets are not leaked in logs
func TestNoSecretsInLogs(t *testing.T) {
	data := setupTestData(t)
	defer cleanupTestData(t, data)

	// Test various commands that might log secrets
	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "init command",
			args: []string{"init"},
		},
		{
			name: "bundle with passphrase",
			args: []string{"bundle", data.envFile, "--pass", "test-passphrase"},
		},
		{
			name: "unbundle with passphrase",
			args: []string{"unbundle", data.bundleFile, "--pass", "test-passphrase", "--out", ".env.test"},
		},
		{
			name: "run command",
			args: []string{"run", data.bundleFile, "--pass", "test-passphrase", "--", "echo", "test"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runCommand(t, data, tc.args...)

			// Don't fail on command errors, just check logs
			if err != nil {
				t.Logf("Command failed (expected for some tests): %v", err)
			}

			// Check that no secrets are logged
			checkNoSecretsInLogs(t, stdout, stderr)
		})
	}
}

// TestLicenseEnforcement tests that paid features require valid license
func TestLicenseEnforcement(t *testing.T) {
	data := setupTestData(t)
	defer cleanupTestData(t, data)

	// Ensure no token file exists for license enforcement tests
	ensureNoToken(t)

	// Initialize project first
	_, _, err := runCommand(t, data, "init")
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Test that paid commands fail without login
	// Note: These commands will fail with connection errors because they try to connect to localhost:8080
	// The license enforcement happens in the CLI logic before the API call
	tests := []struct {
		name          string
		args          []string
		wantErr       bool
		errorContains string
	}{
		{
			name:          "bundle push without login",
			args:          []string{"bundle", data.envFile, "--push", "--project", data.projectName},
			wantErr:       true,
			errorContains: "cloud sync is Pro",
		},
		{
			name:          "pull without login",
			args:          []string{"pull", "--project", data.projectName},
			wantErr:       true,
			errorContains: "not logged in",
		},
		{
			name:          "share without login",
			args:          []string{"share", "--user", "test@example.com", "--role", "read"},
			wantErr:       true,
			errorContains: "not logged in",
		},
		{
			name:          "audit without login",
			args:          []string{"audit", "--project", data.projectName},
			wantErr:       true,
			errorContains: "not logged in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCommand(t, data, tt.args...)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
			}

			// Check that the error message contains the expected text
			if tt.wantErr && err != nil {
				errorOutput := stdout + stderr
				if !strings.Contains(errorOutput, tt.errorContains) {
					t.Errorf("Error output does not contain expected text '%s'. Got: %s", tt.errorContains, errorOutput)
				}
			}

			// Check that no secrets are logged
			checkNoSecretsInLogs(t, stdout, stderr)
		})
	}

	// Test that commands work after login
	server := setupTestServer()
	defer server.Close()

	// Set API URL to test server
	os.Setenv("SECRETSNAP_API_URL", server.URL)

	// Login first
	_, _, err = runCommand(t, data, "login", "--license", data.licenseKey)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	// Create project
	_, _, err = runCommand(t, data, "project", "create", data.projectName)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Test that paid commands work after login
	successTests := []struct {
		name string
		args []string
	}{
		{
			name: "bundle push with login",
			args: []string{"bundle", data.envFile, "--push", "--project", data.projectName},
		},
		{
			name: "pull with login",
			args: []string{"pull", "--project", data.projectName, "--out", ".env.pulled", "--force"},
		},
		{
			name: "share with login",
			args: []string{"share", "--user", "test@example.com", "--role", "read"},
		},
		{
			name: "audit with login",
			args: []string{"audit", "--project", data.projectName},
		},
	}

	for _, tt := range successTests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runCommand(t, data, tt.args...)

			if err != nil {
				t.Errorf("Command failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
			}

			// Check that no secrets are logged
			checkNoSecretsInLogs(t, stdout, stderr)
		})
	}
}

// TestIntegration tests a complete workflow
func TestIntegration(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	data := setupTestData(t)
	defer cleanupTestData(t, data)

	// Set API URL to test server
	os.Setenv("SECRETSNAP_API_URL", server.URL)

	// Complete workflow test
	steps := []struct {
		name string
		args []string
	}{
		{"init", []string{"init"}},
		{"bundle local", []string{"bundle", data.envFile}},
		{"unbundle local", []string{"unbundle", data.bundleFile, "--out", ".env.local"}},
		{"bundle passphrase", []string{"bundle", data.envFile, "--pass", "test-passphrase", "--out", "secrets.pass"}},
		{"unbundle passphrase", []string{"unbundle", "secrets.pass", "--pass", "test-passphrase", "--out", ".env.pass"}},
		{"login", []string{"login", "--license", data.licenseKey}},
		{"bundle push", []string{"bundle", data.envFile, "--push", "--project", data.projectName}},
		{"pull", []string{"pull", "--out", ".env.pulled"}},
		{"run command", []string{"run", data.bundleFile, "--", "echo", "$FOO"}},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			stdout, stderr, err := runCommand(t, data, step.args...)

			if err != nil {
				t.Errorf("Step '%s' failed: %v\nstdout: %s\nstderr: %s", step.name, err, stdout, stderr)
			}

			// Check that no secrets are logged
			checkNoSecretsInLogs(t, stdout, stderr)
		})
	}

	// Verify all output files have correct permissions
	outputFiles := []string{
		".env.local",
		".env.pass",
		".env.pulled",
	}

	for _, outputFile := range outputFiles {
		filePath := filepath.Join(data.tempDir, outputFile)
		if _, err := os.Stat(filePath); err == nil {
			checkFilePermissions(t, filePath, 0600)
		}
	}

	// Verify audit logs were created
	if len(server.auditLogs) < 2 {
		t.Errorf("Expected audit logs for push/pull actions, got %d", len(server.auditLogs))
	}
}
