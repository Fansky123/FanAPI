package handler

import (
	"fmt"
	"sync"

	"fanapi/docs"
	"fanapi/internal/db"
	"fanapi/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag"
)

var swaggerMu sync.Mutex

// SwaggerJSON 动态将 swagger host 替换为实际请求域名后返回 JSON spec。
func SwaggerJSON(c *gin.Context) {
	host := c.Request.Host
	if fwd := c.GetHeader("X-Forwarded-Host"); fwd != "" {
		host = fwd
	}
	schemes := []string{"http"}
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		schemes = []string{"https"}
	}

	swaggerMu.Lock()
	docs.SwaggerInfo.Host = host
	docs.SwaggerInfo.Schemes = schemes
	doc, err := swag.ReadDoc()
	swaggerMu.Unlock()

	if err != nil {
		c.JSON(500, gin.H{"error": "swagger doc error"})
		return
	}
	c.Data(200, "application/json; charset=utf-8", []byte(doc))
}

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
  data-url="/openapi.json"
  data-configuration='{"theme":"default","darkMode":false,"layout":"sidebar","hideDarkModeToggle":true}'
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
