package handler

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"fanapi/docs"
	"fanapi/internal/db"
	"fanapi/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag"
)

var swaggerMu sync.Mutex

var allowedUserDocOps = map[string]map[string]bool{
	"/user/balance": {
		"get": true,
	},
	"/v1/tasks/{id}": {
		"get": true,
	},
}

var allowedUserDocTags = map[string]bool{
	"媒体生成": true,
	"LLM":  true,
}

func readSwaggerDoc(c *gin.Context) ([]byte, error) {
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
		return nil, err
	}
	return []byte(doc), nil
}

func shouldKeepUserDocOperation(path, method string, op map[string]any) bool {
	if methods, ok := allowedUserDocOps[path]; ok && methods[strings.ToLower(method)] {
		return true
	}

	tags, _ := op["tags"].([]any)
	for _, rawTag := range tags {
		tag, _ := rawTag.(string)
		if allowedUserDocTags[tag] {
			return true
		}
	}

	return false
}

func buildUserSwaggerDoc(doc []byte) ([]byte, error) {
	var spec map[string]any
	if err := json.Unmarshal(doc, &spec); err != nil {
		return nil, err
	}

	paths, _ := spec["paths"].(map[string]any)
	filteredPaths := make(map[string]any)
	for path, rawMethods := range paths {
		methods, ok := rawMethods.(map[string]any)
		if !ok {
			continue
		}

		keptMethods := make(map[string]any)
		for method, rawOperation := range methods {
			op, ok := rawOperation.(map[string]any)
			if !ok {
				continue
			}
			if shouldKeepUserDocOperation(path, method, op) {
				keptMethods[method] = rawOperation
			}
		}

		if len(keptMethods) > 0 {
			filteredPaths[path] = keptMethods
		}
	}
	spec["paths"] = filteredPaths

	if info, ok := spec["info"].(map[string]any); ok {
		info["description"] = "LLM 对话 · 媒体生成 · 查询任务结果 · 查询账户余额"
	}

	return json.Marshal(spec)
}

// SwaggerJSON 动态将 swagger host 替换为实际请求域名后返回 JSON spec。
func SwaggerJSON(c *gin.Context) {
	doc, err := readSwaggerDoc(c)
	if err != nil {
		c.JSON(500, gin.H{"error": "swagger doc error"})
		return
	}
	c.Data(200, "application/json; charset=utf-8", doc)
}

// UserSwaggerJSON 返回用户端精简版 OpenAPI spec。
func UserSwaggerJSON(c *gin.Context) {
	doc, err := readSwaggerDoc(c)
	if err != nil {
		c.JSON(500, gin.H{"error": "swagger doc error"})
		return
	}

	filteredDoc, err := buildUserSwaggerDoc(doc)
	if err != nil {
		c.JSON(500, gin.H{"error": "swagger doc error"})
		return
	}

	c.Data(200, "application/json; charset=utf-8", filteredDoc)
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
<div
	id="api-reference"
	data-url="/openapi-user.json"
  data-configuration='{"theme":"default","darkMode":false,"layout":"sidebar","hideDarkModeToggle":true}'
</div>
<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
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
