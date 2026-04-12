-- 为 payment_orders 表新增中台支付所需字段
-- 执行方式：psql -U <user> -d <db> -f scripts/migrate_20260412_payment_order_apply_fields.sql

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS pay_flat INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS pay_from VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS pro_name VARCHAR(128) NOT NULL DEFAULT '';
