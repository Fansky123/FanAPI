-- 为 ocpc_platforms 表新增可配置的转化类型字段
ALTER TABLE ocpc_platforms
    ADD COLUMN IF NOT EXISTS baidu_reg_type   INT         NOT NULL DEFAULT 68  COMMENT '百度注册转化 newType（默认68）',
    ADD COLUMN IF NOT EXISTS baidu_order_type INT         NOT NULL DEFAULT 10  COMMENT '百度购买转化 newType（默认10）',
    ADD COLUMN IF NOT EXISTS e360_reg_event   VARCHAR(32) NOT NULL DEFAULT ''  COMMENT '360 注册转化事件（默认 REGISTERED）',
    ADD COLUMN IF NOT EXISTS e360_order_event VARCHAR(32) NOT NULL DEFAULT ''  COMMENT '360 购买转化事件（默认 ORDER）';
