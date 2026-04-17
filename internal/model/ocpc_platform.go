package model

import "time"

// OcpcPlatform 存储广告推广平台账户配置。
// 支持多个百度 / 360 账户，每行代表一个独立的广告账户。
// 落地页 URL 应携带 ?ocpc_id=<ID> 参数，以便将用户注册与具体账户关联。
type OcpcPlatform struct {
	ID       int64  `xorm:"pk autoincr 'id'" json:"id"`
	Platform string `xorm:"varchar(16) notnull 'platform'" json:"platform"` // "baidu" | "360"
	Name     string `xorm:"varchar(64) notnull 'name'" json:"name"`
	Enabled  bool   `xorm:"notnull default(true) 'enabled'" json:"enabled"`

	// 百度 OCPC 字段（platform == "baidu" 时使用）
	BaiduToken   string `xorm:"varchar(512) 'baidu_token'" json:"baidu_token"`
	BaiduPageURL string `xorm:"varchar(512) 'baidu_page_url'" json:"baidu_page_url"`

	// 百度转化类型（可选，0 表示使用默认值：注册=68，购买=10）
	BaiduRegType   int `xorm:"default(68) 'baidu_reg_type'" json:"baidu_reg_type"`
	BaiduOrderType int `xorm:"default(10) 'baidu_order_type'" json:"baidu_order_type"`

	// 360 OCPC 字段（platform == "360" 时使用）
	E360Key    string `xorm:"varchar(256) 'e360_key'" json:"e360_key"`
	E360Secret string `xorm:"varchar(256) 'e360_secret'" json:"e360_secret"`
	E360Jzqs   string `xorm:"varchar(128) 'e360_jzqs'" json:"e360_jzqs"`
	// E360SoType: "1" = 搜索推广, "2" = 展示广告
	E360SoType string `xorm:"varchar(4) default('1') 'e360_so_type'" json:"e360_so_type"`
	// 360 转化事件（可选，留空使用默认值：注册=REGISTERED，购买=ORDER）
	E360RegEvent   string `xorm:"varchar(32) 'e360_reg_event'" json:"e360_reg_event"`
	E360OrderEvent string `xorm:"varchar(32) 'e360_order_event'" json:"e360_order_event"`

	CreatedAt time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (*OcpcPlatform) TableName() string { return "ocpc_platforms" }
