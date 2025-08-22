package utils

import (
	"fmt"
	"math/rand"
	"time"

	"secretsnap/internal/config"
)

// UpsellMessage represents a single upsell message with its category
type UpsellMessage struct {
	Message  string
	Category string // "team", "ci", "security", "convenience"
}

// UpsellMessages contains all the variations of upsell messages
var UpsellMessages = []UpsellMessage{
	// Team collaboration messages
	{
		Message:  "🚀 Ready to share secrets with your team? Upgrade to Pro for cloud storage and team collaboration!",
		Category: "team",
	},
	{
		Message:  "👥 Working with a team? Pro lets you share encrypted bundles securely with your teammates.",
		Category: "team",
	},
	{
		Message:  "🤝 Tired of manually sharing keys? Pro handles team access automatically.",
		Category: "team",
	},

	// CI/CD messages
	{
		Message:  "⚡ Deploying to production? Pro integrates seamlessly with GitHub Actions, CircleCI, and more!",
		Category: "ci",
	},
	{
		Message:  "🔧 Automate your deployments with Pro - pull secrets directly in your CI/CD pipeline.",
		Category: "ci",
	},
	{
		Message:  "🏗️ Building with CI/CD? Pro makes secret management in pipelines effortless.",
		Category: "ci",
	},

	// Security messages
	{
		Message:  "🔒 Need enterprise-grade security? Pro includes audit logs and advanced access controls.",
		Category: "security",
	},
	{
		Message:  "📊 Track who accessed what and when with Pro's comprehensive audit logging.",
		Category: "security",
	},
	{
		Message:  "🛡️ Pro adds enterprise security features like access controls and audit trails.",
		Category: "security",
	},

	// Convenience messages
	{
		Message:  "💡 Pro tip: Upgrade to Pro for cloud storage, team sharing, and audit logs!",
		Category: "convenience",
	},
	{
		Message:  "✨ Unlock the full potential of Secretsnap with Pro - cloud storage, team collaboration, and more!",
		Category: "convenience",
	},
	{
		Message:  "🎯 Take your secret management to the next level with Pro features!",
		Category: "convenience",
	},
}

// UpsellCallToAction contains variations of the call-to-action
var UpsellCallToAction = []string{
	"💳 Upgrade now: https://secretsnap.dev/pricing",
	"🔗 Learn more: https://secretsnap.dev/features",
	"📧 Questions? support@secretsnap.dev",
	"💬 Join our Discord: https://discord.gg/secretsnap",
}

// ShowUpsell displays a randomized upsell message if appropriate
func ShowUpsell() error {
	// Check if user is already on paid plan
	token, err := config.LoadToken()
	if err == nil && token != "" {
		return nil // Already paid user
	}

	// Check if we should show upsell
	shouldShow, err := config.ShouldShowUpsell()
	if err != nil {
		return err
	}

	if !shouldShow {
		return nil
	}

	// Mark that we've shown the upsell
	if err := config.MarkUpsellShown(); err != nil {
		return err
	}

	// Show the upsell message
	showRandomUpsell()
	return nil
}

// showRandomUpsell displays a randomized upsell message
func showRandomUpsell() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Select random message
	message := UpsellMessages[rand.Intn(len(UpsellMessages))]
	
	// Select random call-to-action
	cta := UpsellCallToAction[rand.Intn(len(UpsellCallToAction))]

	// Display the upsell
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("💡 %s\n", message.Message)
	fmt.Printf("   %s\n", cta)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// ShowContextualUpsell shows a contextual upsell based on the command being used
func ShowContextualUpsell(command string) error {
	// Check if user is already on paid plan
	token, err := config.LoadToken()
	if err == nil && token != "" {
		return nil // Already paid user
	}

	// Check if we should show upsell
	shouldShow, err := config.ShouldShowUpsell()
	if err != nil {
		return err
	}

	if !shouldShow {
		return nil
	}

	// Mark that we've shown the upsell
	if err := config.MarkUpsellShown(); err != nil {
		return err
	}

	// Show contextual message based on command
	showContextualUpsell(command)
	return nil
}

// showContextualUpsell displays a contextual upsell message
func showContextualUpsell(command string) {
	rand.Seed(time.Now().UnixNano())

	var contextualMessages []string

	switch command {
	case "bundle":
		contextualMessages = []string{
			"🚀 Ready to push your bundle to the cloud? Upgrade to Pro for secure cloud storage!",
			"☁️ Store your bundles securely in the cloud with Pro - no more manual sharing!",
			"📦 Pro tip: Use `--push` to store bundles in the cloud with Pro!",
		}
	case "run":
		contextualMessages = []string{
			"⚡ Deploying to production? Pro integrates with CI/CD for seamless secret injection!",
			"🔧 Automate your deployments with Pro - pull secrets directly in your pipeline!",
			"🏗️ Building with CI/CD? Pro makes secret management in pipelines effortless!",
		}
	case "unbundle":
		contextualMessages = []string{
			"👥 Working with a team? Pro lets you pull the latest bundles from the cloud!",
			"🔄 Tired of manual updates? Pro automatically syncs the latest bundle versions!",
			"📥 Pro tip: Use `secretsnap pull` to get the latest bundle from your team!",
		}
	default:
		// Fall back to general upsell
		showRandomUpsell()
		return
	}

	// Select random contextual message
	message := contextualMessages[rand.Intn(len(contextualMessages))]
	
	// Select random call-to-action
	cta := UpsellCallToAction[rand.Intn(len(UpsellCallToAction))]

	// Display the contextual upsell
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("💡 %s\n", message)
	fmt.Printf("   %s\n", cta)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// ShowFeatureUpsell shows a specific feature-focused upsell
func ShowFeatureUpsell(feature string) error {
	// Check if user is already on paid plan
	token, err := config.LoadToken()
	if err == nil && token != "" {
		return nil // Already paid user
	}

	// Check if we should show upsell
	shouldShow, err := config.ShouldShowUpsell()
	if err != nil {
		return err
	}

	if !shouldShow {
		return nil
	}

	// Mark that we've shown the upsell
	if err := config.MarkUpsellShown(); err != nil {
		return err
	}

	// Show feature-specific message
	showFeatureUpsell(feature)
	return nil
}

// showFeatureUpsell displays a feature-specific upsell message
func showFeatureUpsell(feature string) {
	rand.Seed(time.Now().UnixNano())

	var featureMessages map[string][]string = map[string][]string{
		"cloud": {
			"☁️ Cloud storage is a Pro feature! Store your bundles securely in AWS S3.",
			"🚀 Ready for cloud storage? Upgrade to Pro for secure bundle hosting!",
			"📦 Pro tip: Cloud storage keeps your bundles safe and accessible!",
		},
		"team": {
			"👥 Team sharing is a Pro feature! Collaborate securely with your team.",
			"🤝 Ready to share with your team? Upgrade to Pro for team collaboration!",
			"👨‍💻 Pro tip: Team sharing makes collaboration effortless!",
		},
		"audit": {
			"📊 Audit logs are a Pro feature! Track who accessed what and when.",
			"🔍 Need visibility? Upgrade to Pro for comprehensive audit logging!",
			"📈 Pro tip: Audit logs provide complete visibility into secret access!",
		},
		"ci": {
			"⚡ CI/CD integration is a Pro feature! Deploy secrets automatically.",
			"🏗️ Building with CI/CD? Upgrade to Pro for seamless integration!",
			"🔧 Pro tip: CI/CD integration automates your deployment workflow!",
		},
	}

	messages, exists := featureMessages[feature]
	if !exists {
		// Fall back to general upsell
		showRandomUpsell()
		return
	}

	// Select random feature message
	message := messages[rand.Intn(len(messages))]
	
	// Select random call-to-action
	cta := UpsellCallToAction[rand.Intn(len(UpsellCallToAction))]

	// Display the feature-specific upsell
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("💡 %s\n", message)
	fmt.Printf("   %s\n", cta)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}
