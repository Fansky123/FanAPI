ALTER TABLE llm_logs ADD COLUMN IF NOT EXISTS client_request jsonb;
