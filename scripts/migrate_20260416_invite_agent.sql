-- 邀请系统 + 客服角色支持
-- 为 users 表新增邀请码、上级客服 ID、微信二维码、微信 OpenID 字段

ALTER TABLE users ADD COLUMN IF NOT EXISTS invite_code VARCHAR(32) NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS inviter_id BIGINT DEFAULT NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS wechat_qr TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS wechat_openid VARCHAR(64) NOT NULL DEFAULT '';

-- 邀请码部分唯一索引（忽略空值，避免多行空字符串冲突）
CREATE UNIQUE INDEX IF NOT EXISTS users_invite_code_idx ON users (invite_code) WHERE invite_code != '';
-- 微信 OpenID 部分唯一索引
CREATE UNIQUE INDEX IF NOT EXISTS users_wechat_openid_idx ON users (wechat_openid) WHERE wechat_openid != '';
