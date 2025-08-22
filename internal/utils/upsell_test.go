package utils

import (
	"testing"
	"time"

	"secretsnap/internal/config"
)

func TestUpsellMessageVariations(t *testing.T) {
	// Test that we have multiple message variations
	if len(UpsellMessages) < 3 {
		t.Errorf("Expected at least 3 upsell message variations, got %d", len(UpsellMessages))
	}

	// Test that we have multiple call-to-action variations
	if len(UpsellCallToAction) < 2 {
		t.Errorf("Expected at least 2 call-to-action variations, got %d", len(UpsellCallToAction))
	}

	// Test that messages have categories
	for i, msg := range UpsellMessages {
		if msg.Category == "" {
			t.Errorf("Message %d has no category", i)
		}
		if msg.Message == "" {
			t.Errorf("Message %d has no content", i)
		}
	}
}

func TestShouldShowUpsell(t *testing.T) {
	// Test initial state (should not show upsell)
	shouldShow, err := config.ShouldShowUpsell()
	if err != nil {
		t.Fatalf("ShouldShowUpsell failed: %v", err)
	}
	if shouldShow {
		t.Error("Should not show upsell initially")
	}

	// Increment free runs to trigger upsell
	for i := 0; i < 3; i++ {
		if err := config.IncrementFreeRun(); err != nil {
			t.Fatalf("IncrementFreeRun failed: %v", err)
		}
	}

	// Now should show upsell
	shouldShow, err = config.ShouldShowUpsell()
	if err != nil {
		t.Fatalf("ShouldShowUpsell failed: %v", err)
	}
	if !shouldShow {
		t.Error("Should show upsell after 3 runs")
	}

	// Mark upsell as shown
	if err := config.MarkUpsellShown(); err != nil {
		t.Fatalf("MarkUpsellShown failed: %v", err)
	}

	// Should not show again immediately
	shouldShow, err = config.ShouldShowUpsell()
	if err != nil {
		t.Fatalf("ShouldShowUpsell failed: %v", err)
	}
	if shouldShow {
		t.Error("Should not show upsell immediately after marking as shown")
	}
}

func TestContextualUpsell(t *testing.T) {
	// Test that contextual upsell has messages for each command
	commands := []string{"bundle", "run", "unbundle"}
	
	for _, cmd := range commands {
		// This is a basic test - in a real test we'd mock the config
		// and verify the right messages are shown
		t.Logf("Testing contextual upsell for command: %s", cmd)
	}
}

func TestFeatureUpsell(t *testing.T) {
	// Test that feature upsell has messages for each feature
	features := []string{"cloud", "team", "audit", "ci"}
	
	for _, feature := range features {
		// This is a basic test - in a real test we'd mock the config
		// and verify the right messages are shown
		t.Logf("Testing feature upsell for feature: %s", feature)
	}
}

func TestUpsellRateLimiting(t *testing.T) {
	// Test that upsell is rate limited to once per day
	
	// Reset usage stats
	stats := &config.UsageStats{
		FreeRuns:    5,
		LastUpsell:  time.Now().Add(-25 * time.Hour), // More than 24 hours ago
		UpsellShown: false,
	}
	
	if err := config.SaveUsageStats(stats); err != nil {
		t.Fatalf("SaveUsageStats failed: %v", err)
	}

	// Should show upsell after 24 hours
	shouldShow, err := config.ShouldShowUpsell()
	if err != nil {
		t.Fatalf("ShouldShowUpsell failed: %v", err)
	}
	if !shouldShow {
		t.Error("Should show upsell after 24 hours")
	}

	// Mark as shown
	if err := config.MarkUpsellShown(); err != nil {
		t.Fatalf("MarkUpsellShown failed: %v", err)
	}

	// Should not show again immediately
	shouldShow, err = config.ShouldShowUpsell()
	if err != nil {
		t.Fatalf("ShouldShowUpsell failed: %v", err)
	}
	if shouldShow {
		t.Error("Should not show upsell immediately after marking as shown")
	}
}
