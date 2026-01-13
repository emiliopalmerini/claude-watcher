-- Drop session_tags table
DROP INDEX IF EXISTS idx_session_tags_tag;
DROP INDEX IF EXISTS idx_session_tags_session;
DROP TABLE IF EXISTS session_tags;

-- Drop tags table
DROP TABLE IF EXISTS tags;

-- SQLite doesn't support DROP COLUMN directly
-- We need to recreate the table without the new columns
CREATE TABLE sessions_backup AS SELECT
    id,
    session_id,
    instance_id,
    hostname,
    timestamp,
    exit_reason,
    permission_mode,
    working_directory,
    git_branch,
    claude_version,
    duration_seconds,
    user_prompts,
    assistant_responses,
    tool_calls,
    tools_breakdown,
    files_accessed,
    files_modified,
    input_tokens,
    output_tokens,
    thinking_tokens,
    cache_read_tokens,
    cache_write_tokens,
    estimated_cost_usd,
    errors_count,
    model,
    summary
FROM sessions;

DROP TABLE sessions;

CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    instance_id TEXT NOT NULL,
    hostname TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    exit_reason TEXT,
    permission_mode TEXT,
    working_directory TEXT,
    git_branch TEXT,
    claude_version TEXT,
    duration_seconds INTEGER,
    user_prompts INTEGER DEFAULT 0,
    assistant_responses INTEGER DEFAULT 0,
    tool_calls INTEGER DEFAULT 0,
    tools_breakdown TEXT,
    files_accessed TEXT,
    files_modified TEXT,
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    thinking_tokens INTEGER DEFAULT 0,
    cache_read_tokens INTEGER DEFAULT 0,
    cache_write_tokens INTEGER DEFAULT 0,
    estimated_cost_usd REAL DEFAULT 0.0,
    errors_count INTEGER DEFAULT 0,
    model TEXT,
    summary TEXT
);

INSERT INTO sessions SELECT * FROM sessions_backup;
DROP TABLE sessions_backup;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_sessions_timestamp ON sessions(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_sessions_hostname ON sessions(hostname);
CREATE INDEX IF NOT EXISTS idx_sessions_model ON sessions(model);
