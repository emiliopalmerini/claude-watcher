package cli

import (
	"context"
	"fmt"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/ports"
)

// getExperimentByName looks up an experiment by name via the repository.
// Returns a descriptive error if not found or if the lookup fails.
func getExperimentByName(ctx context.Context, repo ports.ExperimentRepository, name string) (*domain.Experiment, error) {
	exp, err := repo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get experiment: %w", err)
	}
	if exp == nil {
		return nil, fmt.Errorf("experiment %q not found", name)
	}
	return exp, nil
}

// truncate shortens a string to maxLen characters, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
