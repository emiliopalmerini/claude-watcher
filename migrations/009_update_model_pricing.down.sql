-- Revert to original pricing
DELETE FROM model_pricing;

INSERT INTO model_pricing (id, display_name, input_per_million, output_per_million, cache_read_per_million, cache_write_per_million, is_default)
VALUES
    ('claude-sonnet-4-20250514', 'Claude Sonnet 4', 3.00, 15.00, 0.30, 3.75, 1),
    ('claude-opus-4-20250514', 'Claude Opus 4', 15.00, 75.00, 1.50, 18.75, 0),
    ('claude-haiku-3-5-20241022', 'Claude Haiku 3.5', 0.80, 4.00, 0.08, 1.00, 0);
