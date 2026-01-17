-- Update model pricing to latest API values (January 2025)
-- Delete old entries
DELETE FROM model_pricing;

-- Insert updated pricing
INSERT INTO model_pricing (id, display_name, input_per_million, output_per_million, cache_read_per_million, cache_write_per_million, is_default)
VALUES
    ('claude-opus-4-5-20251101', 'Claude Opus 4.5', 5.00, 25.00, 0.50, 6.25, 0),
    ('claude-sonnet-4-5-20241022', 'Claude Sonnet 4.5', 3.00, 15.00, 0.30, 3.75, 1),
    ('claude-haiku-3-5-20241022', 'Claude Haiku 3.5', 1.00, 5.00, 0.10, 1.25, 0);
