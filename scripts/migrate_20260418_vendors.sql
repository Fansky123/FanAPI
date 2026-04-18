-- 号商系统：创建号商表，并在号池 Key 上关联号商
CREATE TABLE IF NOT EXISTS vendors (
    id            bigserial PRIMARY KEY,
    username      varchar(64) UNIQUE NOT NULL,
    password_hash text NOT NULL,
    email         varchar(255),
    is_active     boolean NOT NULL DEFAULT true,
    balance       bigint NOT NULL DEFAULT 0,   -- 可提现余额（credits，不含平台手续费）
    commission_ratio float8 DEFAULT NULL,       -- 个人手续费比例（NULL 时使用系统默认值 default_vendor_commission）
    invite_code   varchar(32) UNIQUE NOT NULL,  -- 号商唯一注册码（注册时自动生成）
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- 号池 Key 关联号商
ALTER TABLE pool_keys ADD COLUMN IF NOT EXISTS vendor_id bigint DEFAULT NULL REFERENCES vendors(id) ON DELETE SET NULL;

COMMENT ON TABLE vendors IS '号商：向平台提供 API Key 的供应商';
COMMENT ON COLUMN vendors.balance IS '可提现余额（平台扣除手续费后的净额，单位 credits）';
COMMENT ON COLUMN vendors.commission_ratio IS '平台手续费比例（0.02 = 平台抽 2%，号商到手 98%）';
COMMENT ON COLUMN pool_keys.vendor_id IS '所属号商 ID（NULL 表示非号商提供的 Key）';
