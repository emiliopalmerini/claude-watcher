package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/ports"
	"github.com/emiliopalmerini/mclaude/internal/util"
)

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Manage sessions",
	Long:  `List and manage recorded sessions.`,
}

var sessionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sessions",
	Long: `List recorded sessions with optional filters.

Examples:
  mclaude sessions list                     # Last 10 sessions
  mclaude sessions list --last 20           # Last 20 sessions
  mclaude sessions list --experiment "exp"  # Sessions for experiment
  mclaude sessions list --project <id>      # Sessions for project`,
	RunE: runSessionsList,
}

// Flags
var (
	sessionsLast       int
	sessionsExperiment string
	sessionsProject    string
)

func init() {
	rootCmd.AddCommand(sessionsCmd)
	sessionsCmd.AddCommand(sessionsListCmd)

	sessionsListCmd.Flags().IntVarP(&sessionsLast, "last", "n", 10, "Number of sessions to show")
	sessionsListCmd.Flags().StringVarP(&sessionsExperiment, "experiment", "e", "", "Filter by experiment name")
	sessionsListCmd.Flags().StringVar(&sessionsProject, "project", "", "Filter by project ID")
}

func runSessionsList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	opts := ports.ListSessionsOptions{
		Limit: sessionsLast,
	}

	if sessionsExperiment != "" {
		exp, err := getExperimentByName(ctx, app.ExperimentRepo, sessionsExperiment)
		if err != nil {
			return err
		}
		opts.ExperimentID = &exp.ID
	} else if sessionsProject != "" {
		opts.ProjectID = &sessionsProject
	}

	sessions, err := app.SessionRepo.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDATE\tTURNS\tTOKENS\tCOST\tREASON")
	fmt.Fprintln(w, "--\t----\t-----\t------\t----\t------")

	for _, s := range sessions {
		id := s.ID
		if len(id) > 12 {
			id = id[:12]
		}

		date := util.FormatDateTime(s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

		turns := "-"
		tokens := "-"
		cost := "-"
		m, err := app.MetricsRepo.GetBySessionID(ctx, s.ID)
		if err == nil && m != nil {
			turns = fmt.Sprintf("%d", m.TurnCount)
			totalTokens := m.TokenInput + m.TokenOutput
			tokens = util.FormatNumber(totalTokens)
			if m.CostEstimateUSD != nil {
				cost = fmt.Sprintf("$%.4f", *m.CostEstimateUSD)
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", id, date, turns, tokens, cost, s.ExitReason)
	}

	w.Flush()

	fmt.Printf("\nShowing %d session(s)\n", len(sessions))
	return nil
}
