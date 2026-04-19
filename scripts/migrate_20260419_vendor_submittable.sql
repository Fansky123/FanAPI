-- 号池增加"允许号商自助上传 Key"开关
-- 新部署由 xorm Sync2 自动处理，仅旧版升级需手动执行

ALTER TABLE key_pools ADD COLUMN IF NOT EXISTS vendor_submittable boolean NOT NULL DEFAULT false;
