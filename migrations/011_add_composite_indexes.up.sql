-- Composite indexes for common query patterns
-- Improves performance for filtered + sorted queries

-- Sessions by experiment with created_at ordering
CREATE INDEX IF NOT EXISTS idx_sessions_experiment_created ON sessions(experiment_id, created_at DESC);

-- Sessions by project with created_at ordering
CREATE INDEX IF NOT EXISTS idx_sessions_project_created ON sessions(project_id, created_at DESC);
