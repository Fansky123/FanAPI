-- 为 api_keys 表新增 key_type 字段（low_price | stable）
-- xorm Sync2 会自动执行此变更，这里仅作手动迁移备份
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS key_type VARCHAR(20) NOT NULL DEFAULT 'low_price';
