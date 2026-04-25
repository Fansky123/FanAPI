package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var uploadImageCategories = map[string]string{
	"reference":    "reference",
	"channel-icon": "channel-icons",
	"site-setting": "site-settings",
	"payment-qr":   "payment-qr",
}

func saveUploadedImage(c *gin.Context, category string) {
	userID := c.MustGet("user_id").(int64)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的图片"})
		return
	}
	if file.Size <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "上传文件不能为空"})
		return
	}
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "图片不能超过 10MB"})
		return
	}

	contentType := file.Header.Get("Content-Type")
	if contentType == "" || !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持上传图片文件"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		extensions, _ := mime.ExtensionsByType(contentType)
		if len(extensions) > 0 {
			ext = strings.ToLower(extensions[0])
		}
	}
	if ext == "" {
		ext = ".png"
	}

	subdir := filepath.Join("uploads", category)
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}

	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成文件名失败"})
		return
	}
	filename := fmt.Sprintf("%d_%d_%s%s", userID, time.Now().Unix(), hex.EncodeToString(randomBytes), ext)
	fullPath := filepath.Join(subdir, filename)
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片失败"})
		return
	}

	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"url": fmt.Sprintf("%s://%s/uploads/%s/%s", scheme, c.Request.Host, category, filename),
	})
}

// UploadImage POST /upload/image 通用图片上传，返回可公开访问的 URL。
func UploadImage(c *gin.Context) {
	categoryKey := c.PostForm("category")
	category, ok := uploadImageCategories[categoryKey]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的上传分类"})
		return
	}
	saveUploadedImage(c, category)
}

// UploadReferenceImage POST /user/reference-images 兼容旧调用，默认上传到参考图目录。
func UploadReferenceImage(c *gin.Context) {
	saveUploadedImage(c, uploadImageCategories["reference"])
}
