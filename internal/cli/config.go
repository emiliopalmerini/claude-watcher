package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/domain"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `View and update configuration settings.`,
}

var configModelCmd = &cobra.Command{
	Use:   "model [name]",
	Short: "Get or set the default model",
	Long: `Get or set the default model for cost calculations.

Without arguments, shows the current default model.
With an argument, sets the default model.

Accepts short names: opus, sonnet, haiku
Or partial matches: "opus 4.5", "Claude Sonnet"

Examples:
  mclaude config model           # Show current default
  mclaude config model opus      # Set Opus 4.5 as default
  mclaude config model sonnet    # Set Sonnet 4.5 as default
  mclaude config model haiku     # Set Haiku 3.5 as default`,
	Args: cobra.MaximumNArgs(1),
	RunE: runConfigModel,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configModelCmd)
}

func runConfigModel(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if len(args) == 0 {
		return showDefaultModel(ctx)
	}
	return setDefaultModel(ctx, args[0])
}

func showDefaultModel(ctx context.Context) error {
	defaultModel, err := app.PricingRepo.GetDefault(ctx)
	if err != nil || defaultModel == nil {
		fmt.Println("No default model configured")
		fmt.Println("\nUse 'mclaude config model <name>' to set one")
		fmt.Println("Available: opus, sonnet, haiku")
		return nil
	}

	fmt.Printf("Default model: %s\n", defaultModel.DisplayName)
	fmt.Printf("  ID: %s\n", defaultModel.ID)
	fmt.Printf("  Input:  $%.2f / 1M tokens\n", defaultModel.InputPerMillion)
	fmt.Printf("  Output: $%.2f / 1M tokens\n", defaultModel.OutputPerMillion)
	if defaultModel.CacheReadPerMillion != nil {
		fmt.Printf("  Cache Read:  $%.2f / 1M tokens\n", *defaultModel.CacheReadPerMillion)
	}
	if defaultModel.CacheWritePerMillion != nil {
		fmt.Printf("  Cache Write: $%.2f / 1M tokens\n", *defaultModel.CacheWritePerMillion)
	}

	return nil
}

func setDefaultModel(ctx context.Context, name string) error {
	models, err := app.PricingRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if len(models) == 0 {
		return fmt.Errorf("no models configured. Run migrations to add default models")
	}

	nameLower := strings.ToLower(name)
	var match *domain.ModelPricing

	for _, m := range models {
		idLower := strings.ToLower(m.ID)
		displayLower := strings.ToLower(m.DisplayName)

		switch nameLower {
		case "opus":
			if strings.Contains(idLower, "opus") {
				match = m
			}
		case "sonnet":
			if strings.Contains(idLower, "sonnet") {
				match = m
			}
		case "haiku":
			if strings.Contains(idLower, "haiku") {
				match = m
			}
		}

		if match == nil {
			if strings.Contains(idLower, nameLower) || strings.Contains(displayLower, nameLower) {
				match = m
			}
		}

		if match != nil {
			break
		}
	}

	if match == nil {
		fmt.Printf("No model matching %q found\n\n", name)
		fmt.Println("Available models:")
		for _, m := range models {
			fmt.Printf("  - %s (%s)\n", m.DisplayName, m.ID)
		}
		return fmt.Errorf("model not found")
	}

	if err := app.PricingRepo.SetDefault(ctx, match.ID); err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	fmt.Printf("Default model set to: %s\n", match.DisplayName)
	return nil
}
