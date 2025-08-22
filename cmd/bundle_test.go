package cmd

import (
	"testing"

	"secretsnap/internal/config"
)

func TestDetermineMode(t *testing.T) {
	tests := []struct {
		name           string
		projectConfig  *config.ProjectConfig
		pass           string
		passFile       string
		passMode       bool
		push           bool
		expectedMode   string
		expectedReason string
	}{
		{
			name: "Cloud mode - explicit push flag",
			projectConfig: &config.ProjectConfig{
				ProjectName: "test-project",
				ProjectID:   "local",
				Mode:        "local",
			},
			push:           true,
			expectedMode:   "cloud",
			expectedReason: "explicit --push flag should always use cloud mode",
		},
		{
			name: "Cloud mode - project configured for cloud",
			projectConfig: &config.ProjectConfig{
				ProjectName: "cloud-project",
				ProjectID:   "proj-123",
				Mode:        "cloud",
			},
			push:           false,
			expectedMode:   "cloud",
			expectedReason: "cloud project should use cloud mode even without --push",
		},
		{
			name: "Cloud mode - project configured for cloud with push flag",
			projectConfig: &config.ProjectConfig{
				ProjectName: "cloud-project",
				ProjectID:   "proj-123",
				Mode:        "cloud",
			},
			push:           true,
			expectedMode:   "cloud",
			expectedReason: "cloud project with --push should use cloud mode",
		},
		{
			name: "Local mode - default project",
			projectConfig: &config.ProjectConfig{
				ProjectName: "test-project",
				ProjectID:   "local",
				Mode:        "local",
			},
			push:           false,
			expectedMode:   "local",
			expectedReason: "local project should use local mode",
		},
		{
			name: "Local mode - empty mode",
			projectConfig: &config.ProjectConfig{
				ProjectName: "test-project",
				ProjectID:   "local",
				Mode:        "",
			},
			push:           false,
			expectedMode:   "local",
			expectedReason: "empty mode should default to local",
		},
		{
			name:           "Local mode - nil project config",
			projectConfig:  nil,
			push:           false,
			expectedMode:   "local",
			expectedReason: "nil config should default to local",
		},
		{
			name: "Cloud mode - explicit pass flag with push (push wins)",
			projectConfig: &config.ProjectConfig{
				ProjectName: "test-project",
				ProjectID:   "local",
				Mode:        "local",
			},
			pass:           "mypass",
			push:           true, // Push flag has highest priority
			expectedMode:   "cloud",
			expectedReason: "--push flag should have highest priority over passphrase",
		},
		{
			name: "Cloud mode - pass file with push (push wins)",
			projectConfig: &config.ProjectConfig{
				ProjectName: "test-project",
				ProjectID:   "local",
				Mode:        "local",
			},
			passFile:       "pass.txt",
			push:           true, // Push flag has highest priority
			expectedMode:   "cloud",
			expectedReason: "--push flag should have highest priority over passphrase",
		},
		{
			name: "Cloud mode - pass mode flag with push (push wins)",
			projectConfig: &config.ProjectConfig{
				ProjectName: "test-project",
				ProjectID:   "local",
				Mode:        "local",
			},
			passMode:       true,
			push:           true, // Push flag has highest priority
			expectedMode:   "cloud",
			expectedReason: "--push flag should have highest priority over passphrase",
		},
		{
			name: "Cloud mode - cloud project with passphrase flags (should still be cloud)",
			projectConfig: &config.ProjectConfig{
				ProjectName: "cloud-project",
				ProjectID:   "proj-123",
				Mode:        "cloud",
			},
			pass:           "mypass",
			push:           false,
			expectedMode:   "cloud",
			expectedReason: "cloud project should prioritize cloud mode over passphrase",
		},
		{
			name: "Passphrase mode - explicit pass flag without push",
			projectConfig: &config.ProjectConfig{
				ProjectName: "test-project",
				ProjectID:   "local",
				Mode:        "local",
			},
			pass:           "mypass",
			push:           false,
			expectedMode:   "passphrase",
			expectedReason: "passphrase should be used when no push flag",
		},
		{
			name: "Passphrase mode - pass file without push",
			projectConfig: &config.ProjectConfig{
				ProjectName: "test-project",
				ProjectID:   "local",
				Mode:        "local",
			},
			passFile:       "pass.txt",
			push:           false,
			expectedMode:   "passphrase",
			expectedReason: "passphrase should be used when no push flag",
		},
		{
			name: "Passphrase mode - pass mode flag without push",
			projectConfig: &config.ProjectConfig{
				ProjectName: "test-project",
				ProjectID:   "local",
				Mode:        "local",
			},
			passMode:       true,
			push:           false,
			expectedMode:   "passphrase",
			expectedReason: "passphrase should be used when no push flag",
		},
		{
			name: "Cloud mode - cloud project with push flag (highest priority)",
			projectConfig: &config.ProjectConfig{
				ProjectName: "cloud-project",
				ProjectID:   "proj-123",
				Mode:        "cloud",
			},
			pass:           "mypass",
			push:           true,
			expectedMode:   "cloud",
			expectedReason: "--push flag should have highest priority",
		},
		{
			name: "Local mode - cloud project with invalid project ID",
			projectConfig: &config.ProjectConfig{
				ProjectName: "test-project",
				ProjectID:   "local", // Invalid for cloud
				Mode:        "cloud",
			},
			push:           false,
			expectedMode:   "local",
			expectedReason: "cloud mode with invalid project ID should fall back to local",
		},
		{
			name: "Local mode - cloud project with empty project ID",
			projectConfig: &config.ProjectConfig{
				ProjectName: "test-project",
				ProjectID:   "",
				Mode:        "cloud",
			},
			push:           false,
			expectedMode:   "local",
			expectedReason: "cloud mode with empty project ID should fall back to local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineMode(tt.projectConfig, tt.pass, tt.passFile, tt.passMode, tt.push)

			if result != tt.expectedMode {
				t.Errorf("determineMode() = %v, want %v", result, tt.expectedMode)
				t.Logf("Reason: %s", tt.expectedReason)
				t.Logf("Input: projectConfig=%+v, pass=%q, passFile=%q, passMode=%v, push=%v",
					tt.projectConfig, tt.pass, tt.passFile, tt.passMode, tt.push)
			}
		})
	}
}

func TestDetermineModePriorityOrder(t *testing.T) {
	// Test that the priority order is correct: cloud → local → passphrase

	cloudProject := &config.ProjectConfig{
		ProjectName: "cloud-project",
		ProjectID:   "proj-123",
		Mode:        "cloud",
	}

	localProject := &config.ProjectConfig{
		ProjectName: "local-project",
		ProjectID:   "local",
		Mode:        "local",
	}

	tests := []struct {
		name          string
		projectConfig *config.ProjectConfig
		pass          string
		push          bool
		expectedMode  string
		description   string
	}{
		{
			name:          "Cloud project without flags",
			projectConfig: cloudProject,
			expectedMode:  "cloud",
			description:   "Cloud project should default to cloud mode (revenue priority)",
		},
		{
			name:          "Cloud project with push flag",
			projectConfig: cloudProject,
			push:          true,
			expectedMode:  "cloud",
			description:   "Cloud project with push should use cloud mode",
		},
		{
			name:          "Local project without flags",
			projectConfig: localProject,
			expectedMode:  "local",
			description:   "Local project should default to local mode",
		},
		{
			name:          "Local project with push flag",
			projectConfig: localProject,
			push:          true,
			expectedMode:  "cloud",
			description:   "Local project with push should use cloud mode (flag overrides config)",
		},
		{
			name:          "Local project with passphrase",
			projectConfig: localProject,
			pass:          "mypass",
			expectedMode:  "passphrase",
			description:   "Local project with passphrase should use passphrase mode (fallback)",
		},
		{
			name:          "Cloud project with passphrase (should still be cloud)",
			projectConfig: cloudProject,
			pass:          "mypass",
			expectedMode:  "cloud",
			description:   "Cloud project should prioritize cloud over passphrase (revenue priority)",
		},
		{
			name:          "Cloud project with push and passphrase",
			projectConfig: cloudProject,
			pass:          "mypass",
			push:          true,
			expectedMode:  "cloud",
			description:   "Push flag should have highest priority",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineMode(tt.projectConfig, tt.pass, "", false, tt.push)

			if result != tt.expectedMode {
				t.Errorf("determineMode() = %v, want %v", result, tt.expectedMode)
				t.Logf("Description: %s", tt.description)
				t.Logf("Input: projectConfig=%+v, pass=%q, push=%v",
					tt.projectConfig, tt.pass, tt.push)
			}
		})
	}
}

func TestDetermineModeEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		projectConfig *config.ProjectConfig
		pass          string
		push          bool
		expectedMode  string
		description   string
	}{
		{
			name:          "Nil project config",
			projectConfig: nil,
			expectedMode:  "local",
			description:   "Nil config should default to local mode",
		},
		{
			name: "Empty project config",
			projectConfig: &config.ProjectConfig{
				ProjectName: "",
				ProjectID:   "",
				Mode:        "",
			},
			expectedMode: "local",
			description:  "Empty config should default to local mode",
		},
		{
			name: "Invalid cloud project (empty ID)",
			projectConfig: &config.ProjectConfig{
				ProjectName: "test",
				ProjectID:   "",
				Mode:        "cloud",
			},
			expectedMode: "local",
			description:  "Cloud mode with empty project ID should fall back to local",
		},
		{
			name: "Invalid cloud project (local ID)",
			projectConfig: &config.ProjectConfig{
				ProjectName: "test",
				ProjectID:   "local",
				Mode:        "cloud",
			},
			expectedMode: "local",
			description:  "Cloud mode with 'local' project ID should fall back to local",
		},
		{
			name: "Unknown mode",
			projectConfig: &config.ProjectConfig{
				ProjectName: "test",
				ProjectID:   "local",
				Mode:        "unknown",
			},
			expectedMode: "local",
			description:  "Unknown mode should default to local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineMode(tt.projectConfig, tt.pass, "", false, tt.push)

			if result != tt.expectedMode {
				t.Errorf("determineMode() = %v, want %v", result, tt.expectedMode)
				t.Logf("Description: %s", tt.description)
				t.Logf("Input: projectConfig=%+v, pass=%q, push=%v",
					tt.projectConfig, tt.pass, tt.push)
			}
		})
	}
}

// Benchmark the determineMode function
func BenchmarkDetermineMode(b *testing.B) {
	cloudProject := &config.ProjectConfig{
		ProjectName: "cloud-project",
		ProjectID:   "proj-123",
		Mode:        "cloud",
	}

	localProject := &config.ProjectConfig{
		ProjectName: "local-project",
		ProjectID:   "local",
		Mode:        "local",
	}

	benchmarks := []struct {
		name          string
		projectConfig *config.ProjectConfig
		pass          string
		push          bool
	}{
		{"CloudMode", cloudProject, "", false},
		{"LocalMode", localProject, "", false},
		{"PassphraseMode", localProject, "mypass", false},
		{"PushFlag", localProject, "", true},
		{"ComplexCase", cloudProject, "mypass", true},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				determineMode(bm.projectConfig, bm.pass, "", false, bm.push)
			}
		})
	}
}
