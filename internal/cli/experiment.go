package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/util"
)

var experimentCmd = &cobra.Command{
	Use:   "experiment",
	Short: "Manage experiments",
	Long:  `Create, list, activate, and manage experiments for A/B testing Claude usage styles.`,
}

var experimentCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new experiment",
	Long: `Create a new experiment and automatically activate it.

Examples:
  mclaude experiment create "minimal-prompts" --description "Testing shorter prompts" --hypothesis "Reduces token usage"`,
	Args: cobra.ExactArgs(1),
	RunE: runExperimentCreate,
}

var experimentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all experiments",
	RunE:  runExperimentList,
}

var experimentActivateCmd = &cobra.Command{
	Use:   "activate <name>",
	Short: "Activate an experiment",
	Long:  `Activate an experiment. Only one experiment can be active at a time.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runExperimentActivate,
}

var experimentDeactivateCmd = &cobra.Command{
	Use:   "deactivate [name]",
	Short: "Deactivate an experiment",
	Long:  `Deactivate an experiment. If no name is provided, deactivates the currently active experiment.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runExperimentDeactivate,
}

var experimentEndCmd = &cobra.Command{
	Use:   "end <name>",
	Short: "End an experiment",
	Long:  `End an experiment by setting its end date and deactivating it.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runExperimentEnd,
}

var experimentDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete an experiment",
	Long:  `Delete an experiment. Sessions linked to this experiment will have their experiment_id set to NULL.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runExperimentDelete,
}

var experimentStatsCmd = &cobra.Command{
	Use:   "stats <name>",
	Short: "Show statistics for an experiment",
	Long: `Show detailed statistics for a specific experiment.

Examples:
  mclaude experiment stats "baseline"`,
	Args: cobra.ExactArgs(1),
	RunE: runExperimentStats,
}

var experimentCompareCmd = &cobra.Command{
	Use:   "compare <exp1> <exp2> [exp3...]",
	Short: "Compare statistics between experiments",
	Long: `Compare statistics side-by-side between two or more experiments.

Examples:
  mclaude experiment compare "baseline" "minimal-prompts"
  mclaude experiment compare "exp1" "exp2" "exp3"`,
	Args: cobra.MinimumNArgs(2),
	RunE: runExperimentCompare,
}

// Flags
var (
	expDescription string
	expHypothesis  string
)

// expData holds aggregated stats for an experiment (used in compare)
type expData struct {
	name         string
	sessions     int64
	turns        int64
	userMsgs     int64
	assistMsgs   int64
	tokenInput   int64
	tokenOutput  int64
	cacheRead    int64
	cacheWrite   int64
	cost         float64
	errors       int64
	totalTokens  int64
	tokensPerSes int64
	costPerSes   float64
}

func init() {
	rootCmd.AddCommand(experimentCmd)

	experimentCmd.AddCommand(experimentCreateCmd)
	experimentCmd.AddCommand(experimentListCmd)
	experimentCmd.AddCommand(experimentActivateCmd)
	experimentCmd.AddCommand(experimentDeactivateCmd)
	experimentCmd.AddCommand(experimentEndCmd)
	experimentCmd.AddCommand(experimentDeleteCmd)
	experimentCmd.AddCommand(experimentStatsCmd)
	experimentCmd.AddCommand(experimentCompareCmd)

	// Flags for create command
	experimentCreateCmd.Flags().StringVarP(&expDescription, "description", "d", "", "Description of the experiment")
	experimentCreateCmd.Flags().StringVarP(&expHypothesis, "hypothesis", "H", "", "Hypothesis to test")
}

func runExperimentCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name := args[0]

	existing, err := app.ExperimentRepo.GetByName(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check experiment: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("experiment with name %q already exists", name)
	}

	if err := app.ExperimentRepo.DeactivateAll(ctx); err != nil {
		return fmt.Errorf("failed to deactivate experiments: %w", err)
	}

	now := time.Now().UTC()
	exp := &domain.Experiment{
		ID:        uuid.New().String(),
		Name:      name,
		StartedAt: now,
		IsActive:  true,
		CreatedAt: now,
	}
	if expDescription != "" {
		exp.Description = &expDescription
	}
	if expHypothesis != "" {
		exp.Hypothesis = &expHypothesis
	}

	if err := app.ExperimentRepo.Create(ctx, exp); err != nil {
		return fmt.Errorf("failed to create experiment: %w", err)
	}

	fmt.Printf("Created and activated experiment: %s\n", name)
	return nil
}

func runExperimentList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	expStats, err := app.StatsRepo.GetAllExperimentStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get experiment stats: %w", err)
	}

	experiments, err := app.ExperimentRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list experiments: %w", err)
	}

	if len(experiments) == 0 {
		fmt.Println("No experiments found")
		return nil
	}

	statsMap := make(map[string]domain.ExperimentStats)
	for _, es := range expStats {
		statsMap[es.ExperimentID] = es
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tSESSIONS\tTOKENS\tCOST\tSTARTED\tENDED")
	fmt.Fprintln(w, "----\t------\t--------\t------\t----\t-------\t-----")

	for _, exp := range experiments {
		status := "inactive"
		if exp.IsActive {
			status = "ACTIVE"
		} else if exp.EndedAt != nil {
			status = "ended"
		}

		started := exp.StartedAt.Format("2006-01-02")
		ended := "-"
		if exp.EndedAt != nil {
			ended = exp.EndedAt.Format("2006-01-02")
		}

		sessions := int64(0)
		tokens := int64(0)
		cost := 0.0
		if es, ok := statsMap[exp.ID]; ok {
			sessions = es.SessionCount
			tokens = es.TotalTokenInput + es.TotalTokenOutput
			cost = es.TotalCostUsd
		}

		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t$%.2f\t%s\t%s\n",
			exp.Name, status, sessions, util.FormatNumber(tokens), cost, started, ended)
	}

	w.Flush()
	return nil
}

func runExperimentActivate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name := args[0]

	exp, err := getExperimentByName(ctx, app.ExperimentRepo, name)
	if err != nil {
		return err
	}

	if exp.IsActive {
		fmt.Printf("Experiment %q is already active\n", name)
		return nil
	}

	if err := app.ExperimentRepo.DeactivateAll(ctx); err != nil {
		return fmt.Errorf("failed to deactivate experiments: %w", err)
	}

	if err := app.ExperimentRepo.Activate(ctx, exp.ID); err != nil {
		return fmt.Errorf("failed to activate experiment: %w", err)
	}

	fmt.Printf("Activated experiment: %s\n", name)
	return nil
}

func runExperimentDeactivate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if len(args) == 0 {
		active, err := app.ExperimentRepo.GetActive(ctx)
		if err != nil {
			return fmt.Errorf("failed to get active experiment: %w", err)
		}
		if active == nil {
			fmt.Println("No active experiment to deactivate")
			return nil
		}

		if err := app.ExperimentRepo.Deactivate(ctx, active.ID); err != nil {
			return fmt.Errorf("failed to deactivate experiment: %w", err)
		}

		fmt.Printf("Deactivated experiment: %s\n", active.Name)
		return nil
	}

	name := args[0]
	exp, err := getExperimentByName(ctx, app.ExperimentRepo, name)
	if err != nil {
		return err
	}

	if !exp.IsActive {
		fmt.Printf("Experiment %q is already inactive\n", name)
		return nil
	}

	if err := app.ExperimentRepo.Deactivate(ctx, exp.ID); err != nil {
		return fmt.Errorf("failed to deactivate experiment: %w", err)
	}

	fmt.Printf("Deactivated experiment: %s\n", name)
	return nil
}

func runExperimentEnd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name := args[0]

	exp, err := getExperimentByName(ctx, app.ExperimentRepo, name)
	if err != nil {
		return err
	}

	if exp.EndedAt != nil {
		return fmt.Errorf("experiment %q has already ended", name)
	}

	now := time.Now().UTC()
	exp.EndedAt = &now
	exp.IsActive = false

	if err := app.ExperimentRepo.Update(ctx, exp); err != nil {
		return fmt.Errorf("failed to end experiment: %w", err)
	}

	fmt.Printf("Ended experiment: %s\n", name)
	return nil
}

func runExperimentDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name := args[0]

	exp, err := getExperimentByName(ctx, app.ExperimentRepo, name)
	if err != nil {
		return err
	}

	if err := app.ExperimentRepo.Delete(ctx, exp.ID); err != nil {
		return fmt.Errorf("failed to delete experiment: %w", err)
	}

	fmt.Printf("Deleted experiment: %s\n", name)
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func runExperimentStats(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name := args[0]

	exp, err := getExperimentByName(ctx, app.ExperimentRepo, name)
	if err != nil {
		return err
	}

	stats, err := app.StatsRepo.GetAggregateByExperiment(ctx, exp.ID, "1970-01-01T00:00:00Z")
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	fmt.Println()
	fmt.Printf("  Experiment: %s\n", exp.Name)
	fmt.Printf("  ==============%s\n", repeatChar('=', len(exp.Name)))
	fmt.Println()

	if exp.Description != nil && *exp.Description != "" {
		fmt.Printf("  Description:  %s\n", *exp.Description)
	}
	if exp.Hypothesis != nil && *exp.Hypothesis != "" {
		fmt.Printf("  Hypothesis:   %s\n", *exp.Hypothesis)
	}

	status := "inactive"
	if exp.IsActive {
		status = "ACTIVE"
	} else if exp.EndedAt != nil {
		status = "ended"
	}
	fmt.Printf("  Status:       %s\n", status)
	fmt.Printf("  Started:      %s\n", exp.StartedAt.Format("2006-01-02"))
	if exp.EndedAt != nil {
		fmt.Printf("  Ended:        %s\n", exp.EndedAt.Format("2006-01-02"))
	}
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

	if stats.SessionCount > 0 {
		fmt.Printf("  Efficiency\n")
		fmt.Printf("  ----------\n")
		fmt.Printf("  Tokens/session:    %s\n", util.FormatNumber(totalTokens/stats.SessionCount))
		fmt.Printf("  Cost/session:      $%.4f\n", stats.TotalCostUsd/float64(stats.SessionCount))
		fmt.Println()
	}

	return nil
}

func runExperimentCompare(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	var experiments []expData

	for _, name := range args {
		exp, err := app.ExperimentRepo.GetByName(ctx, name)
		if err != nil {
			return fmt.Errorf("failed to get experiment: %w", err)
		}
		if exp == nil {
			return fmt.Errorf("experiment %q not found", name)
		}

		stats, err := app.StatsRepo.GetAggregateByExperiment(ctx, exp.ID, "1970-01-01T00:00:00Z")
		if err != nil {
			return fmt.Errorf("failed to get stats for %q: %w", name, err)
		}

		totalTokens := stats.TotalTokenInput + stats.TotalTokenOutput
		tokensPerSes := int64(0)
		costPerSes := 0.0
		if stats.SessionCount > 0 {
			tokensPerSes = totalTokens / stats.SessionCount
			costPerSes = stats.TotalCostUsd / float64(stats.SessionCount)
		}

		experiments = append(experiments, expData{
			name:         name,
			sessions:     stats.SessionCount,
			turns:        stats.TotalTurns,
			userMsgs:     stats.TotalUserMessages,
			assistMsgs:   stats.TotalAssistantMessages,
			tokenInput:   stats.TotalTokenInput,
			tokenOutput:  stats.TotalTokenOutput,
			cacheRead:    stats.TotalTokenCacheRead,
			cacheWrite:   stats.TotalTokenCacheWrite,
			cost:         stats.TotalCostUsd,
			errors:       stats.TotalErrors,
			totalTokens:  totalTokens,
			tokensPerSes: tokensPerSes,
			costPerSes:   costPerSes,
		})
	}

	fmt.Println()
	fmt.Printf("  Experiment Comparison\n")
	fmt.Printf("  =====================\n")
	fmt.Println()

	maxNameLen := 18
	for _, e := range experiments {
		if len(e.name) > maxNameLen {
			maxNameLen = len(e.name)
		}
	}
	colWidth := maxNameLen + 2
	if colWidth < 14 {
		colWidth = max(colWidth, 14)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  METRIC\t")
	for _, e := range experiments {
		fmt.Fprintf(w, "%s\t", e.name)
	}
	fmt.Fprintln(w)

	fmt.Fprintf(w, "  ------\t")
	for range experiments {
		fmt.Fprintf(w, "------\t")
	}
	fmt.Fprintln(w)

	printCompareRow(w, "Sessions", experiments, func(e expData) string { return fmt.Sprintf("%d", e.sessions) })
	printCompareRow(w, "Turns", experiments, func(e expData) string { return util.FormatNumber(e.turns) })
	printCompareRow(w, "User messages", experiments, func(e expData) string { return util.FormatNumber(e.userMsgs) })
	printCompareRow(w, "Assistant msgs", experiments, func(e expData) string { return util.FormatNumber(e.assistMsgs) })
	printCompareRow(w, "Errors", experiments, func(e expData) string { return fmt.Sprintf("%d", e.errors) })
	fmt.Fprintln(w)
	printCompareRow(w, "Token input", experiments, func(e expData) string { return util.FormatNumber(e.tokenInput) })
	printCompareRow(w, "Token output", experiments, func(e expData) string { return util.FormatNumber(e.tokenOutput) })
	printCompareRow(w, "Cache read", experiments, func(e expData) string { return util.FormatNumber(e.cacheRead) })
	printCompareRow(w, "Cache write", experiments, func(e expData) string { return util.FormatNumber(e.cacheWrite) })
	printCompareRow(w, "Total tokens", experiments, func(e expData) string { return util.FormatNumber(e.totalTokens) })
	fmt.Fprintln(w)
	printCompareRow(w, "Cost", experiments, func(e expData) string { return fmt.Sprintf("$%.2f", e.cost) })
	printCompareRow(w, "Tokens/session", experiments, func(e expData) string { return util.FormatNumber(e.tokensPerSes) })
	printCompareRow(w, "Cost/session", experiments, func(e expData) string { return fmt.Sprintf("$%.4f", e.costPerSes) })

	w.Flush()
	fmt.Println()

	return nil
}

func printCompareRow(w *tabwriter.Writer, label string, experiments []expData, getValue func(expData) string) {
	fmt.Fprintf(w, "  %s\t", label)
	for _, e := range experiments {
		fmt.Fprintf(w, "%s\t", getValue(e))
	}
	fmt.Fprintln(w)
}

func repeatChar(c rune, n int) string {
	result := make([]rune, n)
	for i := range result {
		result[i] = c
	}
	return string(result)
}
