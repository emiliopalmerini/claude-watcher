package web

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/emiliopalmerini/claude-watcher/internal/web/templates"
	sqlc "github.com/emiliopalmerini/claude-watcher/sqlc/generated"
)

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	// Get stats
	startDate := time.Unix(0, 0).Format(time.RFC3339) // All time
	statsRow, _ := queries.GetAggregateStats(ctx, startDate)

	stats := templates.DashboardStats{
		SessionCount:  statsRow.SessionCount,
		TotalTokens:   toInt64(statsRow.TotalTokenInput) + toInt64(statsRow.TotalTokenOutput),
		TotalCost:     toFloat64(statsRow.TotalCostUsd),
		TotalTurns:    toInt64(statsRow.TotalTurns),
		TokenInput:    toInt64(statsRow.TotalTokenInput),
		TokenOutput:   toInt64(statsRow.TotalTokenOutput),
		CacheRead:     toInt64(statsRow.TotalTokenCacheRead),
		CacheWrite:    toInt64(statsRow.TotalTokenCacheWrite),
		TotalErrors:   toInt64(statsRow.TotalErrors),
	}

	// Get active experiment
	activeExp, _ := queries.GetActiveExperiment(ctx)
	if activeExp.Name != "" {
		stats.ActiveExperiment = activeExp.Name
	}

	// Get top tools
	tools, _ := queries.GetTopToolsUsage(ctx, sqlc.GetTopToolsUsageParams{
		CreatedAt: startDate,
		Limit:     5,
	})

	topTools := make([]templates.ToolUsage, 0, len(tools))
	for _, t := range tools {
		if t.TotalInvocations.Valid {
			topTools = append(topTools, templates.ToolUsage{
				Name:  t.ToolName,
				Count: int64(t.TotalInvocations.Float64),
			})
		}
	}
	stats.TopTools = topTools

	// Get recent sessions
	sessions, _ := queries.ListSessions(ctx, 5)
	recentSessions := make([]templates.SessionSummary, 0, len(sessions))
	for _, sess := range sessions {
		summary := templates.SessionSummary{
			ID:         sess.ID,
			CreatedAt:  sess.CreatedAt,
			ExitReason: sess.ExitReason,
		}
		if m, err := queries.GetSessionMetricsBySessionID(ctx, sess.ID); err == nil {
			summary.Turns = m.TurnCount
			summary.Tokens = m.TokenInput + m.TokenOutput
			if m.CostEstimateUsd.Valid {
				summary.Cost = m.CostEstimateUsd.Float64
			}
		}
		recentSessions = append(recentSessions, summary)
	}
	stats.RecentSessions = recentSessions

	templates.Dashboard(stats).Render(ctx, w)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	sessions, _ := queries.ListSessions(ctx, 50)

	sessionList := make([]templates.SessionSummary, 0, len(sessions))
	for _, sess := range sessions {
		summary := templates.SessionSummary{
			ID:         sess.ID,
			CreatedAt:  sess.CreatedAt,
			ExitReason: sess.ExitReason,
			ProjectID:  sess.ProjectID,
		}
		if sess.ExperimentID.Valid {
			summary.ExperimentID = sess.ExperimentID.String
		}
		if m, err := queries.GetSessionMetricsBySessionID(ctx, sess.ID); err == nil {
			summary.Turns = m.TurnCount
			summary.Tokens = m.TokenInput + m.TokenOutput
			if m.CostEstimateUsd.Valid {
				summary.Cost = m.CostEstimateUsd.Float64
			}
		}
		sessionList = append(sessionList, summary)
	}

	templates.Sessions(sessionList).Render(ctx, w)
}

func (s *Server) handleSessionDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	queries := sqlc.New(s.db)

	session, err := queries.GetSessionByID(ctx, id)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	detail := templates.SessionDetail{
		ID:             session.ID,
		ProjectID:      session.ProjectID,
		Cwd:            session.Cwd,
		PermissionMode: session.PermissionMode,
		ExitReason:     session.ExitReason,
		CreatedAt:      session.CreatedAt,
	}

	if session.ExperimentID.Valid {
		detail.ExperimentID = session.ExperimentID.String
	}
	if session.StartedAt.Valid {
		detail.StartedAt = session.StartedAt.String
	}
	if session.EndedAt.Valid {
		detail.EndedAt = session.EndedAt.String
	}
	if session.DurationSeconds.Valid {
		detail.DurationSeconds = session.DurationSeconds.Int64
	}

	// Get metrics
	if m, err := queries.GetSessionMetricsBySessionID(ctx, id); err == nil {
		detail.MessageCountUser = m.MessageCountUser
		detail.MessageCountAssistant = m.MessageCountAssistant
		detail.TurnCount = m.TurnCount
		detail.TokenInput = m.TokenInput
		detail.TokenOutput = m.TokenOutput
		detail.TokenCacheRead = m.TokenCacheRead
		detail.TokenCacheWrite = m.TokenCacheWrite
		detail.ErrorCount = m.ErrorCount
		if m.CostEstimateUsd.Valid {
			detail.CostEstimateUsd = m.CostEstimateUsd.Float64
		}
	}

	// Get tools
	tools, _ := queries.ListSessionToolsBySessionID(ctx, id)
	for _, t := range tools {
		detail.Tools = append(detail.Tools, templates.ToolUsage{
			Name:  t.ToolName,
			Count: t.InvocationCount,
		})
	}

	// Get files
	files, _ := queries.ListSessionFilesBySessionID(ctx, id)
	for _, f := range files {
		detail.Files = append(detail.Files, templates.FileOperation{
			Path:      f.FilePath,
			Operation: f.Operation,
			Count:     f.OperationCount,
		})
	}

	templates.SessionDetailPage(detail).Render(ctx, w)
}

func (s *Server) handleExperiments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	exps, _ := queries.ListExperiments(ctx)

	experiments := make([]templates.Experiment, 0, len(exps))
	for _, e := range exps {
		exp := templates.Experiment{
			ID:        e.ID,
			Name:      e.Name,
			IsActive:  e.IsActive == 1,
			StartedAt: e.StartedAt,
			CreatedAt: e.CreatedAt,
		}
		if e.Description.Valid {
			exp.Description = e.Description.String
		}
		if e.Hypothesis.Valid {
			exp.Hypothesis = e.Hypothesis.String
		}
		if e.EndedAt.Valid {
			exp.EndedAt = e.EndedAt.String
		}
		experiments = append(experiments, exp)
	}

	templates.Experiments(experiments).Render(ctx, w)
}

func (s *Server) handleExperimentDetail(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	pricing, _ := queries.ListModelPricing(ctx)

	models := make([]templates.ModelPricing, 0, len(pricing))
	for _, p := range pricing {
		model := templates.ModelPricing{
			ID:              p.ID,
			DisplayName:     p.DisplayName,
			InputPerMillion: p.InputPerMillion,
			OutputPerMillion: p.OutputPerMillion,
			IsDefault:       p.IsDefault == 1,
		}
		if p.CacheReadPerMillion.Valid {
			model.CacheReadPerMillion = p.CacheReadPerMillion.Float64
		}
		if p.CacheWritePerMillion.Valid {
			model.CacheWritePerMillion = p.CacheWritePerMillion.Float64
		}
		models = append(models, model)
	}

	templates.Settings(models).Render(ctx, w)
}

// API Handlers

func (s *Server) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	period := r.URL.Query().Get("period")
	startDate := getStartDateForPeriod(period)

	statsRow, err := queries.GetAggregateStats(ctx, startDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stats := map[string]interface{}{
		"session_count": statsRow.SessionCount,
		"total_tokens":  toInt64(statsRow.TotalTokenInput) + toInt64(statsRow.TotalTokenOutput),
		"total_cost":    toFloat64(statsRow.TotalCostUsd),
		"token_input":   toInt64(statsRow.TotalTokenInput),
		"token_output":  toInt64(statsRow.TotalTokenOutput),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleAPIChartTokens(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	period := r.URL.Query().Get("period")
	startDate := getStartDateForPeriod(period)
	if period == "" {
		// Default to last 30 days for charts
		startDate = time.Now().AddDate(0, 0, -30).Format(time.RFC3339)
	}

	stats, _ := queries.GetDailyStats(ctx, sqlc.GetDailyStatsParams{
		CreatedAt: startDate,
		Limit:     30,
	})

	labels := make([]string, len(stats))
	tokens := make([]int64, len(stats))
	sessions := make([]int64, len(stats))

	for i, stat := range stats {
		labels[i] = stat.Date.(string)
		tokens[i] = toInt64(stat.TotalTokens)
		sessions[i] = stat.SessionCount
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"labels":   labels,
		"tokens":   tokens,
		"sessions": sessions,
	})
}

func (s *Server) handleAPIChartCost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	period := r.URL.Query().Get("period")
	startDate := getStartDateForPeriod(period)
	if period == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format(time.RFC3339)
	}

	stats, _ := queries.GetDailyStats(ctx, sqlc.GetDailyStatsParams{
		CreatedAt: startDate,
		Limit:     30,
	})

	labels := make([]string, len(stats))
	costs := make([]float64, len(stats))

	for i, stat := range stats {
		labels[i] = stat.Date.(string)
		costs[i] = toFloat64(stat.TotalCost)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"labels": labels,
		"costs":  costs,
	})
}

func (s *Server) handleAPICreateExperiment(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) handleAPIActivateExperiment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	queries := sqlc.New(s.db)

	queries.DeactivateAllExperiments(ctx)
	if err := queries.ActivateExperiment(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/experiments")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAPIDeactivateExperiment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	queries := sqlc.New(s.db)

	if err := queries.DeactivateExperiment(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/experiments")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAPIDeleteExperiment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	queries := sqlc.New(s.db)

	if err := queries.DeleteExperiment(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/experiments")
	w.WriteHeader(http.StatusOK)
}

// Helpers

func toInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return n
	case float64:
		return int64(n)
	default:
		return 0
	}
}

func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	default:
		return 0
	}
}

func getStartDateForPeriod(period string) string {
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
