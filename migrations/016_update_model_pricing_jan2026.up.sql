-- Update model pricing to latest API values (January 2026)
-- Source: https://platform.claude.com/docs/en/about-claude/pricing

-- Clear existing pricing
DELETE FROM model_pricing;

-- Insert updated pricing for all current models
INSERT INTO model_pricing (id, display_name, input_per_million, output_per_million, cache_read_per_million, cache_write_per_million, is_default)
VALUES
    -- Claude 4.5 family
    ('claude-opus-4-5-20251101', 'Claude Opus 4.5', 5.00, 25.00, 0.50, 6.25, 0),
    ('claude-sonnet-4-5-20241022', 'Claude Sonnet 4.5', 3.00, 15.00, 0.30, 3.75, 0),
    ('claude-haiku-4-5-20250101', 'Claude Haiku 4.5', 1.00, 5.00, 0.10, 1.25, 0),

    -- Claude 4.1 family
    ('claude-opus-4-1-20250414', 'Claude Opus 4.1', 15.00, 75.00, 1.50, 18.75, 0),

    -- Claude 4 family
    ('claude-opus-4-20250514', 'Claude Opus 4', 15.00, 75.00, 1.50, 18.75, 0),
    ('claude-sonnet-4-20250514', 'Claude Sonnet 4', 3.00, 15.00, 0.30, 3.75, 1),

    -- Claude 3.7 (deprecated)
    ('claude-3-7-sonnet-20250219', 'Claude Sonnet 3.7', 3.00, 15.00, 0.30, 3.75, 0),

    -- Claude 3.5 family
    ('claude-3-5-haiku-20241022', 'Claude Haiku 3.5', 0.80, 4.00, 0.08, 1.00, 0),

    -- Legacy aliases for compatibility
    ('claude-haiku-3-5-20241022', 'Claude Haiku 3.5 (alias)', 0.80, 4.00, 0.08, 1.00, 0),

    -- Claude 3 family (deprecated/legacy)
    ('claude-3-opus-20240229', 'Claude Opus 3', 15.00, 75.00, 1.50, 18.75, 0),
    ('claude-3-haiku-20240307', 'Claude Haiku 3', 0.25, 1.25, 0.03, 0.30, 0);
