package cli

import (
	"testing"

	"github.com/emiliopalmerini/mclaude/internal/ports"
)

func TestAppContextFieldTypes(t *testing.T) {
	// Compile-time verification that AppContext uses port interfaces.
	var a AppContext
	var _ ports.SessionRepository = a.SessionRepo
	var _ ports.SessionMetricsRepository = a.MetricsRepo
	var _ ports.SessionToolRepository = a.ToolRepo
	var _ ports.SessionFileRepository = a.FileRepo
	var _ ports.SessionCommandRepository = a.CommandRepo
	var _ ports.SessionSubagentRepository = a.SubagentRepo
	var _ ports.ExperimentRepository = a.ExperimentRepo
	var _ ports.ProjectRepository = a.ProjectRepo
	var _ ports.PricingRepository = a.PricingRepo
	var _ ports.SessionQualityRepository = a.QualityRepo
	var _ ports.PlanConfigRepository = a.PlanConfigRepo
	var _ ports.TranscriptStorage = a.TranscriptStorage
	var _ ports.PrometheusClient = a.PrometheusClient
}

func TestAppContextClose_NilDB(t *testing.T) {
	a := &AppContext{}
	if err := a.Close(); err != nil {
		t.Errorf("Close() on nil DB should not error, got: %v", err)
	}
}
