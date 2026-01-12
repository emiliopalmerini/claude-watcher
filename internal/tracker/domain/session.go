package domain

import "time"

// Session represents a Claude Code session with all its metadata and statistics
type Session struct {
	SessionID      string
	InstanceID     string
	Hostname       string
	Timestamp      time.Time
	ExitReason     string
	PermissionMode string
	WorkingDir     string
	Statistics     Statistics
}

// NewSession creates a new session with the given parameters
func NewSession(
	sessionID, instanceID, hostname, exitReason, permissionMode, workingDir string,
	stats Statistics,
) Session {
	return Session{
		SessionID:      sessionID,
		InstanceID:     instanceID,
		Hostname:       hostname,
		Timestamp:      time.Now().UTC(),
		ExitReason:     exitReason,
		PermissionMode: permissionMode,
		WorkingDir:     workingDir,
		Statistics:     stats,
	}
}
