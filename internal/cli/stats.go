package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/util"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show usage statistics",
	Long: `Show summary statistics for Claude Code usage.

Examples:
  mclaude stats                          # All-time stats
  mclaude stats --period today           # Today's stats
  mclaude stats --period week            # This week's stats
  mclaude stats --experiment "baseline"  # Stats for an experiment
  mclaude stats --project <id>           # Stats for a project`,
	RunE: runStats,
}

// Flags
var (
	statsPeriod     string
	statsExperiment string
	statsProject    string
)

func init() {
	rootCmd.AddCommand(statsCmd)

	statsCmd.Flags().StringVarP(&statsPeriod, "period", "p", "all", "Time period: today, week, month, all")
	statsCmd.Flags().StringVarP(&statsExperiment, "experiment", "e", "", "Filter by experiment name")
	statsCmd.Flags().StringVar(&statsProject, "project", "", "Filter by project ID")
}

func runStats(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	startDate := getStartDate(statsPeriod)

	var stats *domain.AggregateStats
	var filterLabel string

	if statsExperiment != "" {
		exp, err := app.ExperimentRepo.GetByName(ctx, statsExperiment)
		if err != nil {
			return fmt.Errorf("failed to get experiment: %w", err)
		}
		if exp == nil {
			return fmt.Errorf("experiment %q not found", statsExperiment)
		}

		stats, err = app.StatsRepo.GetAggregateByExperiment(ctx, exp.ID, startDate)
		if err != nil {
			return fmt.Errorf("failed to get stats: %w", err)
		}
		filterLabel = fmt.Sprintf("Experiment: %s", statsExperiment)
	} else if statsProject != "" {
		var err error
		stats, err = app.StatsRepo.GetAggregateByProject(ctx, statsProject, startDate)
		if err != nil {
			return fmt.Errorf("failed to get stats: %w", err)
		}
		filterLabel = fmt.Sprintf("Project: %s", truncate(statsProject, 16))
	} else {
		var err error
		stats, err = app.StatsRepo.GetAggregate(ctx, startDate)
		if err != nil {
			return fmt.Errorf("failed to get stats: %w", err)
		}
		filterLabel = "All sessions"
	}

	// Get active experiment
	activeExpName := "-"
	activeExp, _ := app.ExperimentRepo.GetActive(ctx)
	if activeExp != nil {
		activeExpName = activeExp.Name
	}

	// Get top tools
	tools, _ := app.StatsRepo.GetTopTools(ctx, startDate, 5)

	printStats(stats, filterLabel, statsPeriod, activeExpName, tools)

	return nil
}

func getStartDate(period string) string {
	now := time.Now().UTC()
	var start time.Time

	switch period {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, time.UTC)
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	default:
		start = time.Unix(0, 0)
	}

	return start.Format(time.RFC3339)
}

func printStats(stats *domain.AggregateStats, filterLabel, period, activeExp string, tools []domain.ToolUsageStats) {
	periodLabel := "All time"
	switch period {
	case "today":
		periodLabel = "Today"
	case "week":
		periodLabel = "This week"
	case "month":
		periodLabel = "This month"
	}

	fmt.Println()
	fmt.Printf("  mclaude Stats\n")
	fmt.Printf("  =====================\n")
	fmt.Println()

	fmt.Printf("  Period:            %s\n", periodLabel)
	fmt.Printf("  Filter:            %s\n", filterLabel)
	fmt.Printf("  Active experiment: %s\n", activeExp)
	fmt.Println()

	fmt.Printf("  Sessions\n")
	fmt.Printf("  --------\n")
	fmt.Printf("  Total:             %d\n", stats.SessionCount)
	fmt.Printf("  Turns:             %s\n", util.FormatNumber(stats.TotalTurns))
	fmt.Printf("  User messages:     %s\n", util.FormatNumber(stats.TotalUserMessages))
	fmt.Printf("  Assistant msgs:    %s\n", util.FormatNumber(stats.TotalAssistantMessages))
	fmt.Printf("  Errors:            %d\n", stats.TotalErrors)
	fmt.Println()

	fmt.Printf("  Tokens\n")
	fmt.Printf("  ------\n")
	fmt.Printf("  Input:             %s\n", util.FormatNumber(stats.TotalTokenInput))
	fmt.Printf("  Output:            %s\n", util.FormatNumber(stats.TotalTokenOutput))
	fmt.Printf("  Cache read:        %s\n", util.FormatNumber(stats.TotalTokenCacheRead))
	fmt.Printf("  Cache write:       %s\n", util.FormatNumber(stats.TotalTokenCacheWrite))
	totalTokens := stats.TotalTokenInput + stats.TotalTokenOutput
	fmt.Printf("  Total:             %s\n", util.FormatNumber(totalTokens))
	fmt.Println()

	fmt.Printf("  Cost\n")
	fmt.Printf("  ----\n")
	fmt.Printf("  Estimated:         $%.4f\n", stats.TotalCostUsd)
	fmt.Println()

	if len(tools) > 0 {
		fmt.Printf("  Top Tools\n")
		fmt.Printf("  ---------\n")
		for _, tool := range tools {
			fmt.Printf("  %-18s %s calls\n", tool.ToolName, util.FormatNumber(tool.TotalInvocations))
		}
		fmt.Println()
	}
}
