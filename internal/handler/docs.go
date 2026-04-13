package handler

import (
"net/http"

"github.com/gin-gonic/gin"
)

// APIDocs 将旧 /docs 路径重定向到 Swagger UI
func APIDocs(c *gin.Context) {
c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
}
