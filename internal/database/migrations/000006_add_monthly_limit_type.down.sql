-- Revert to original CHECK constraint (will fail if monthly records exist)

CREATE TABLE limit_events_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    limit_type TEXT NOT NULL CHECK(limit_type IN ('daily', 'weekly')),
    reset_time TEXT,

    sessions_count INTEGER DEFAULT 0,
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    thinking_tokens INTEGER DEFAULT 0,
    total_cost_usd REAL DEFAULT 0.0
);

INSERT INTO limit_events_old SELECT * FROM limit_events WHERE limit_type != 'monthly';

DROP TABLE limit_events;

ALTER TABLE limit_events_old RENAME TO limit_events;

CREATE INDEX IF NOT EXISTS idx_limit_events_timestamp ON limit_events(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_limit_events_type ON limit_events(limit_type);
