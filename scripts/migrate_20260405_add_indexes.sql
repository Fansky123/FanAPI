-- 性能优化索引：高并发场景下的关键查询加速
-- 使用 CONCURRENTLY 创建，不锁表，可在生产环境不停服执行。
-- 执行方式：psql -U <user> -d <db> -f scripts/migrate_20260405_add_indexes.sql

-- tasks 表：poller 轮询 pending 异步任务的核心索引
-- 对应查询：WHERE status = 'processing' AND upstream_task_id != ''
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_processing_upstream
    ON tasks (status, upstream_task_id)
    WHERE status = 'processing' AND upstream_task_id != '';

-- tasks 表：用户任务列表查询（GET /v1/tasks）
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_user_id_created
    ON tasks (user_id, id DESC);

-- tasks 表：管理后台任务查询（可按 type / status / 时间段筛选）
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_type_status_created
    ON tasks (type, status, created_at DESC);

-- billing_transactions 表：用户账单列表（GET /user/transactions）
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_billing_tx_user_created
    ON billing_transactions (user_id, created_at DESC);

-- billing_transactions 表：管理后台全量账单查询（可按 corr_id 追溯）
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_billing_tx_corr_id
    ON billing_transactions (corr_id)
    WHERE corr_id != '';
