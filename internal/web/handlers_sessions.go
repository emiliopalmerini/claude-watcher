package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/parser"
	"github.com/emiliopalmerini/mclaude/internal/util"
	"github.com/emiliopalmerini/mclaude/internal/web/templates"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	// Get sessions with metrics in single query
	sessions, _ := queries.ListSessionsWithMetrics(ctx, 50)

	// Build quality lookup map
	qualityMap := make(map[string]sqlc.ListSessionQualitiesForSessionsRow)
	if qualities, err := queries.ListSessionQualitiesForSessions(ctx); err == nil {
		for _, q := range qualities {
			qualityMap[q.SessionID] = q
		}
	}

	sessionList := make([]templates.SessionSummary, 0, len(sessions))
	for _, sess := range sessions {
		summary := templates.SessionSummary{
			ID:         sess.ID,
			CreatedAt:  sess.CreatedAt,
			ExitReason: sess.ExitReason,
			ProjectID:  sess.ProjectID,
			Turns:      sess.TurnCount,
			Tokens:     sess.TotalTokens,
		}
		if sess.ExperimentID.Valid {
			summary.ExperimentID = sess.ExperimentID.String
		}
		if sess.CostEstimateUsd.Valid {
			summary.Cost = sess.CostEstimateUsd.Float64
		}
		// Add quality data
		if q, ok := qualityMap[sess.ID]; ok {
			summary.IsReviewed = true
			if q.OverallRating.Valid {
				summary.OverallRating = int(q.OverallRating.Int64)
			}
			if q.IsSuccess.Valid {
				isSuccess := q.IsSuccess.Int64 == 1
				summary.IsSuccess = &isSuccess
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

	var metrics *sqlc.SessionMetric
	if m, err := queries.GetSessionMetricsBySessionID(ctx, id); err == nil {
		metrics = &m
	}
	detail := buildSessionDetail(session, metrics)

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

	// Get sub-agents (aggregated by type+kind)
	subagentStats, _ := queries.GetSubagentStatsBySession(ctx, id)
	for _, sa := range subagentStats {
		usage := templates.SubagentUsage{
			AgentType: sa.AgentType,
			AgentKind: sa.AgentKind,
			Count:     sa.InvocationCount,
			Tokens:    util.ToInt64(sa.TotalTokens),
			Cost:      util.ToFloat64(sa.TotalCost),
		}
		detail.Subagents = append(detail.Subagents, usage)
	}

	// Get quality
	if q, err := queries.GetSessionQualityBySessionID(ctx, id); err == nil {
		quality := templates.SessionQuality{
			SessionID: q.SessionID,
		}
		if q.OverallRating.Valid {
			quality.OverallRating = int(q.OverallRating.Int64)
		}
		if q.IsSuccess.Valid {
			isSuccess := q.IsSuccess.Int64 == 1
			quality.IsSuccess = &isSuccess
		}
		if q.AccuracyRating.Valid {
			quality.AccuracyRating = int(q.AccuracyRating.Int64)
		}
		if q.HelpfulnessRating.Valid {
			quality.HelpfulnessRating = int(q.HelpfulnessRating.Int64)
		}
		if q.EfficiencyRating.Valid {
			quality.EfficiencyRating = int(q.EfficiencyRating.Int64)
		}
		if q.Notes.Valid {
			quality.Notes = q.Notes.String
		}
		if q.ReviewedAt.Valid {
			quality.ReviewedAt = q.ReviewedAt.String
		}
		detail.Quality = &quality
	}

	templates.SessionDetailPage(detail).Render(ctx, w)
}

func (s *Server) handleSessionReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	queries := sqlc.New(s.db)

	// Get session
	session, err := queries.GetSessionByID(ctx, id)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	var metrics *sqlc.SessionMetric
	if m, err := queries.GetSessionMetricsBySessionID(ctx, id); err == nil {
		metrics = &m
	}
	detail := buildSessionDetail(session, metrics)

	// Get existing quality review
	var quality templates.SessionQuality
	if q, err := s.qualityRepo.GetBySessionID(ctx, id); err == nil && q != nil {
		quality = convertDomainQualityToTemplate(q)
	}

	// Get transcript
	var transcriptMessages []templates.TranscriptMessage
	if s.transcriptStorage != nil {
		data, err := s.transcriptStorage.Get(ctx, id)
		if err == nil {
			messages, _ := parser.ParseTranscriptForViewer(data)
			transcriptMessages = convertViewerMessagesToTemplate(messages)
		}
	}

	viewData := templates.SessionReviewData{
		SessionDetail: detail,
		Quality:       quality,
		Transcript:    transcriptMessages,
	}

	templates.SessionReviewPage(viewData).Render(ctx, w)
}

func (s *Server) handleAPISaveQuality(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	quality := &domain.SessionQuality{
		SessionID: id,
	}

	// Parse overall rating
	if v := r.FormValue("overall_rating"); v != "" && v != "0" {
		if rating, err := strconv.Atoi(v); err == nil && rating >= 1 && rating <= 5 {
			quality.OverallRating = &rating
		}
	}

	// Parse is_success
	if v := r.FormValue("is_success"); v != "" {
		success := v == "1"
		quality.IsSuccess = &success
	}

	// Parse dimension ratings
	if v := r.FormValue("accuracy_rating"); v != "" && v != "0" {
		if rating, err := strconv.Atoi(v); err == nil && rating >= 1 && rating <= 5 {
			quality.AccuracyRating = &rating
		}
	}
	if v := r.FormValue("helpfulness_rating"); v != "" && v != "0" {
		if rating, err := strconv.Atoi(v); err == nil && rating >= 1 && rating <= 5 {
			quality.HelpfulnessRating = &rating
		}
	}
	if v := r.FormValue("efficiency_rating"); v != "" && v != "0" {
		if rating, err := strconv.Atoi(v); err == nil && rating >= 1 && rating <= 5 {
			quality.EfficiencyRating = &rating
		}
	}

	// Parse notes
	if v := r.FormValue("notes"); v != "" {
		quality.Notes = &v
	}

	// Set reviewed_at if any rating is provided
	if quality.OverallRating != nil || quality.IsSuccess != nil {
		now := time.Now()
		quality.ReviewedAt = &now
	}

	// Save to database
	if err := s.qualityRepo.Upsert(ctx, quality); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success indicator for HTMX
	templates.QualitySavedIndicator().Render(ctx, w)
}

func convertDomainQualityToTemplate(q *domain.SessionQuality) templates.SessionQuality {
	tq := templates.SessionQuality{
		SessionID: q.SessionID,
		IsSuccess: q.IsSuccess,
	}
	if q.OverallRating != nil {
		tq.OverallRating = *q.OverallRating
	}
	if q.AccuracyRating != nil {
		tq.AccuracyRating = *q.AccuracyRating
	}
	if q.HelpfulnessRating != nil {
		tq.HelpfulnessRating = *q.HelpfulnessRating
	}
	if q.EfficiencyRating != nil {
		tq.EfficiencyRating = *q.EfficiencyRating
	}
	if q.Notes != nil {
		tq.Notes = *q.Notes
	}
	if q.ReviewedAt != nil {
		tq.ReviewedAt = q.ReviewedAt.Format(time.RFC3339)
	}
	return tq
}

func convertViewerMessagesToTemplate(messages []parser.ViewerMessage) []templates.TranscriptMessage {
	result := make([]templates.TranscriptMessage, len(messages))
	for i, m := range messages {
		result[i] = templates.TranscriptMessage{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
		}
		for _, t := range m.Tools {
			result[i].Tools = append(result[i].Tools, templates.TranscriptToolUse{
				Name:  t.Name,
				Input: t.Input,
			})
		}
	}
	return result
}
