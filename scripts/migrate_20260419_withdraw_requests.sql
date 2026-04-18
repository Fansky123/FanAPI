-- 提现申请系统 + 用户收款码

-- 用户表新增收款码字段
ALTER TABLE users ADD COLUMN IF NOT EXISTS payment_qr_wechat TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS payment_qr_alipay TEXT NOT NULL DEFAULT '';

-- 提现申请表
CREATE TABLE IF NOT EXISTS withdraw_requests (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id),
    amount       BIGINT NOT NULL,                            -- 提现积分数（微单位）
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',    -- pending / approved / rejected
    payment_type VARCHAR(20) NOT NULL DEFAULT '',           -- wechat / alipay
    payment_qr   TEXT NOT NULL DEFAULT '',                  -- 快照：申请时的收款码 base64
    admin_remark TEXT NOT NULL DEFAULT '',                  -- 管理员备注（拒绝理由等）
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS withdraw_requests_user_id_idx ON withdraw_requests(user_id);
CREATE INDEX IF NOT EXISTS withdraw_requests_status_idx  ON withdraw_requests(status);
