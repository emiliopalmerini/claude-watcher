-- Add new session metadata columns
ALTER TABLE sessions ADD COLUMN thinking_level TEXT;
ALTER TABLE sessions ADD COLUMN claude_summary TEXT;
ALTER TABLE sessions ADD COLUMN subagent_count INTEGER DEFAULT 0;

-- Create tags table with predefined tags
CREATE TABLE IF NOT EXISTS tags (
    name TEXT PRIMARY KEY,
    category TEXT NOT NULL,
    color TEXT NOT NULL
);

-- Create session_tags junction table
CREATE TABLE IF NOT EXISTS session_tags (
    session_id TEXT NOT NULL,
    tag_name TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now')),
    PRIMARY KEY (session_id, tag_name),
    FOREIGN KEY (tag_name) REFERENCES tags(name) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_session_tags_session ON session_tags(session_id);
CREATE INDEX IF NOT EXISTS idx_session_tags_tag ON session_tags(tag_name);

-- Seed predefined tags
-- Task types
INSERT INTO tags (name, category, color) VALUES ('feature', 'task_type', '#22c55e');
INSERT INTO tags (name, category, color) VALUES ('bugfix', 'task_type', '#ef4444');
INSERT INTO tags (name, category, color) VALUES ('refactor', 'task_type', '#3b82f6');
INSERT INTO tags (name, category, color) VALUES ('exploration', 'task_type', '#8b5cf6');
INSERT INTO tags (name, category, color) VALUES ('docs', 'task_type', '#f59e0b');
INSERT INTO tags (name, category, color) VALUES ('test', 'task_type', '#06b6d4');
INSERT INTO tags (name, category, color) VALUES ('config', 'task_type', '#64748b');

-- Architecture patterns
INSERT INTO tags (name, category, color) VALUES ('vertical-slice', 'architecture', '#10b981');
INSERT INTO tags (name, category, color) VALUES ('hexagonal', 'architecture', '#6366f1');
INSERT INTO tags (name, category, color) VALUES ('mvc', 'architecture', '#f97316');
INSERT INTO tags (name, category, color) VALUES ('solid', 'architecture', '#ec4899');
INSERT INTO tags (name, category, color) VALUES ('ddd', 'architecture', '#14b8a6');

-- Prompt strategies
INSERT INTO tags (name, category, color) VALUES ('detailed-upfront', 'prompt_style', '#a855f7');
INSERT INTO tags (name, category, color) VALUES ('iterative', 'prompt_style', '#0ea5e9');
INSERT INTO tags (name, category, color) VALUES ('minimal', 'prompt_style', '#84cc16');

-- Outcomes
INSERT INTO tags (name, category, color) VALUES ('success', 'outcome', '#22c55e');
INSERT INTO tags (name, category, color) VALUES ('partial', 'outcome', '#eab308');
INSERT INTO tags (name, category, color) VALUES ('failed', 'outcome', '#ef4444');
INSERT INTO tags (name, category, color) VALUES ('rework-needed', 'outcome', '#f97316');
