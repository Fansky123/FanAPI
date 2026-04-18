-- 邀请返佣功能：用户表新增冻结余额和返佣比例
ALTER TABLE users ADD COLUMN IF NOT EXISTS frozen_balance bigint NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS rebate_ratio float8 DEFAULT NULL;

-- 计费流水新增号池 Key ID（用于号商收益统计）
ALTER TABLE billing_transactions ADD COLUMN IF NOT EXISTS pool_key_id bigint NOT NULL DEFAULT 0;

COMMENT ON COLUMN users.frozen_balance IS '冻结余额（邀请返佣所得，需手动解冻才可使用）';
COMMENT ON COLUMN users.rebate_ratio IS '个人返佣比例（NULL 时使用系统默认值 default_rebate_ratio）';
COMMENT ON COLUMN billing_transactions.pool_key_id IS '号池 Key ID（0 表示未使用号池）';
