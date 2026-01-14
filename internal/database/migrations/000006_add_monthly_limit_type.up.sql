-- SQLite doesn't support ALTER TABLE to modify constraints
-- Recreate the table with the updated CHECK constraint

CREATE TABLE limit_events_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    limit_type TEXT NOT NULL CHECK(limit_type IN ('daily', 'weekly', 'monthly')),
    reset_time TEXT,

    -- Usage since last limit event (deltas)
    sessions_count INTEGER DEFAULT 0,
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    thinking_tokens INTEGER DEFAULT 0,
    total_cost_usd REAL DEFAULT 0.0
);

INSERT INTO limit_events_new SELECT * FROM limit_events;

DROP TABLE limit_events;

ALTER TABLE limit_events_new RENAME TO limit_events;

CREATE INDEX IF NOT EXISTS idx_limit_events_timestamp ON limit_events(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_limit_events_type ON limit_events(limit_type);
