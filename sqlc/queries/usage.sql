-- name: CreateUsageLimit :exec
INSERT INTO usage_limits (id, limit_value, warn_threshold, enabled, updated_at)
VALUES (?, ?, ?, ?, datetime('now'))
ON CONFLICT (id) DO UPDATE SET
    limit_value = excluded.limit_value,
    warn_threshold = excluded.warn_threshold,
    enabled = excluded.enabled,
    updated_at = datetime('now');

-- name: GetUsageLimit :one
SELECT * FROM usage_limits WHERE id = ?;

-- name: ListUsageLimits :many
SELECT * FROM usage_limits ORDER BY id;

-- name: DeleteUsageLimit :exec
DELETE FROM usage_limits WHERE id = ?;

-- name: GetRollingWindowUsage :one
SELECT
    CAST(COALESCE(SUM(m.token_cache_read + m.token_cache_write), 0) AS REAL) as total_tokens,
    CAST(COALESCE(SUM(m.cost_estimate_usd), 0) AS REAL) as total_cost
FROM sessions s
JOIN session_metrics m ON s.id = m.session_id
WHERE s.started_at >= datetime('now', ? || ' hours');

-- name: UpsertPlanConfig :exec
INSERT INTO plan_config (id, plan_type, window_hours, learned_token_limit, learned_at, updated_at)
VALUES (1, ?, ?, ?, ?, datetime('now'))
ON CONFLICT (id) DO UPDATE SET
    plan_type = excluded.plan_type,
    window_hours = excluded.window_hours,
    learned_token_limit = excluded.learned_token_limit,
    learned_at = excluded.learned_at,
    updated_at = datetime('now');

-- name: GetPlanConfig :one
SELECT * FROM plan_config WHERE id = 1;

-- name: UpdateLearnedLimit :exec
UPDATE plan_config
SET learned_token_limit = ?, learned_at = datetime('now'), updated_at = datetime('now')
WHERE id = 1;
