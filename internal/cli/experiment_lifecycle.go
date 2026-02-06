package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

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
