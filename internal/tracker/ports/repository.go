package ports

import "claude-watcher/internal/tracker/domain"

// SessionRepository defines the interface for persisting sessions
type SessionRepository interface {
	Save(session domain.Session) error
}
