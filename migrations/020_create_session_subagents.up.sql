CREATE TABLE session_subagents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    agent_type TEXT NOT NULL,
    agent_kind TEXT NOT NULL,
    description TEXT,
    model TEXT,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    token_input INTEGER NOT NULL DEFAULT 0,
    token_output INTEGER NOT NULL DEFAULT 0,
    token_cache_read INTEGER NOT NULL DEFAULT 0,
    token_cache_write INTEGER NOT NULL DEFAULT 0,
    total_duration_ms INTEGER,
    tool_use_count INTEGER NOT NULL DEFAULT 0,
    cost_estimate_usd REAL
);

CREATE INDEX idx_session_subagents_session_id ON session_subagents(session_id);
CREATE INDEX idx_session_subagents_agent_type ON session_subagents(agent_type);
