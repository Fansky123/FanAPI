ALTER TABLE llm_logs ADD COLUMN IF NOT EXISTS upstream_headers jsonb;
