-- Migration: 2026-04-05
-- 1. channels 表新增 error_script 字段
--    用于配置每个渠道的自定义错误检测 JS 脚本。
--    函数约定：checkError(response) → 返回非空字符串表示错误，null/false 表示正常。
--
-- 2. tasks 表新增 corr_id 字段
--    关联该任务对应的计费流水（billing_transactions.corr_id），
--    便于用户/管理员追查哪笔扣费来自哪个任务。

-- ---- channels ----
ALTER TABLE channels
    ADD COLUMN IF NOT EXISTS error_script TEXT NOT NULL DEFAULT '';

-- ---- tasks ----
ALTER TABLE tasks
    ADD COLUMN IF NOT EXISTS corr_id TEXT NOT NULL DEFAULT '';

-- （可选）为高频查询建立索引
-- CREATE INDEX IF NOT EXISTS idx_billing_transactions_corr_id ON billing_transactions(corr_id);
-- CREATE INDEX IF NOT EXISTS idx_tasks_corr_id ON tasks(corr_id);
