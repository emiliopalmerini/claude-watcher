package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileLogger implements the Logger port for file-based logging
type FileLogger struct {
	logPath string
}

// NewFileLogger creates a new file logger
func NewFileLogger(homeDir string) *FileLogger {
	return &FileLogger{
		logPath: filepath.Join(homeDir, ".claude", "session-tracker.log"),
	}
}

// Debug logs a debug message
func (l *FileLogger) Debug(message string) {
	l.write("DEBUG", message)
}

// Error logs an error message (also writes to stderr)
func (l *FileLogger) Error(message string) {
	fmt.Fprintln(os.Stderr, message)
	l.write("ERROR", message)
}

func (l *FileLogger) write(level, message string) {
	f, err := os.OpenFile(l.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format(time.RFC3339)
	fmt.Fprintf(f, "%s %s: %s\n", timestamp, level, message)
}
