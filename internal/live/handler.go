package live

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	apperrors "claude-watcher/internal/shared/errors"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	liveSessions, err := h.repo.GetLiveSessions(ctx)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	recentSessions, err := h.repo.GetRecentSessions(ctx)
	if err != nil {
		log.Printf("error fetching recent sessions: %v", err)
	}

	// Convert to view models
	active := make([]LiveSession, 0, len(liveSessions))
	for _, s := range liveSessions {
		ts := parseTimestamp(s.Timestamp)
		session := LiveSession{
			SessionID:       truncateID(s.SessionID),
			Hostname:        s.Hostname,
			WorkingDir:      shortPath(nullString(s.WorkingDirectory)),
			GitBranch:       nullString(s.GitBranch),
			Duration:        toInt64(s.DurationSeconds),
			Tokens:          toInt64(s.InputTokens) + toInt64(s.OutputTokens) + toInt64(s.ThinkingTokens),
			TokensPerMinute: float64(s.TokensPerMinute),
			Timestamp:       ts,
			LastActivity:    relativeTime(ts),
			Status:          sessionStatus(ts),
		}
		active = append(active, session)
	}

	recent := make([]LiveSession, 0, len(recentSessions))
	for _, s := range recentSessions {
		ts := parseTimestamp(s.Timestamp)
		session := LiveSession{
			SessionID:    truncateID(s.SessionID),
			Hostname:     s.Hostname,
			WorkingDir:   shortPath(nullString(s.WorkingDirectory)),
			GitBranch:    nullString(s.GitBranch),
			Duration:     toInt64(s.DurationSeconds),
			Timestamp:    ts,
			LastActivity: relativeTime(ts),
			Status:       "ended",
		}
		recent = append(recent, session)
	}

	data := LiveData{
		ActiveSessions: active,
		RecentSessions: recent,
	}

	Live(data).Render(ctx, w)
}

// SessionsPartial returns just the sessions grid for HTMX polling
func (h *Handler) SessionsPartial(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	liveSessions, err := h.repo.GetLiveSessions(ctx)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	active := make([]LiveSession, 0, len(liveSessions))
	for _, s := range liveSessions {
		ts := parseTimestamp(s.Timestamp)
		session := LiveSession{
			SessionID:       truncateID(s.SessionID),
			Hostname:        s.Hostname,
			WorkingDir:      shortPath(nullString(s.WorkingDirectory)),
			GitBranch:       nullString(s.GitBranch),
			Duration:        toInt64(s.DurationSeconds),
			Tokens:          toInt64(s.InputTokens) + toInt64(s.OutputTokens) + toInt64(s.ThinkingTokens),
			TokensPerMinute: float64(s.TokensPerMinute),
			Timestamp:       ts,
			LastActivity:    relativeTime(ts),
			Status:          sessionStatus(ts),
		}
		active = append(active, session)
	}

	LiveSessionsGrid(active).Render(ctx, w)
}

func truncateID(id string) string {
	if len(id) > 8 {
		return id[:8] + "..."
	}
	return id
}

func shortPath(path string) string {
	if path == "" {
		return "-"
	}
	return filepath.Base(path)
}

func nullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func parseTimestamp(ts interface{}) time.Time {
	switch v := ts.(type) {
	case time.Time:
		return v
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			t, err = time.Parse("2006-01-02 15:04:05", v)
			if err != nil {
				return time.Time{}
			}
		}
		return t
	default:
		return time.Time{}
	}
}

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	diff := time.Since(t)
	if diff < time.Minute {
		return fmt.Sprintf("%ds ago", int(diff.Seconds()))
	}
	if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	}
	return fmt.Sprintf("%dh ago", int(diff.Hours()))
}

func sessionStatus(t time.Time) string {
	if t.IsZero() {
		return "stale"
	}
	diff := time.Since(t)
	if diff < 30*time.Second {
		return "active"
	}
	if diff < 5*time.Minute {
		return "idle"
	}
	return "stale"
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	default:
		return 0
	}
}
