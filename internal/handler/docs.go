package handler

import (
	"fmt"

	"fanapi/internal/db"
	"fanapi/internal/model"

	"github.com/gin-gonic/gin"
)

const scalarHTMLTpl = `<!doctype html>
<html lang="zh-CN">
<head>
<title>%s 接口文档</title>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<style>body{margin:0}</style>
</head>
<body>
<script
  id="api-reference"
  data-url="swagger/doc.json"
  data-configuration='{"theme":"purple","darkMode":true,"layout":"sidebar"}'
><\/script>
<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"><\/script>
</body>
</html>`

// APIDocs 返回 Scalar API 文档页面，标题跟随 site_name 系统配置
func APIDocs(c *gin.Context) {
	siteName := "FanAPI"
	var s model.SystemSetting
	if found, err := db.Engine.Where("key = ?", "site_name").Get(&s); err == nil && found && s.Value != "" {
		siteName = s.Value
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, fmt.Sprintf(scalarHTMLTpl, siteName))
}
