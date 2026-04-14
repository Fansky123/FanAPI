package taskresult

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"fanapi/internal/model"
)

var imageHTTPClient = &http.Client{Timeout: 60 * time.Second}

// convertResultURLs 检查 image 类型任务结果中的 url / urls 字段：
// 若图片地址的主机与渠道 base_url 的主机不同（即第三方存储），
// 则下载图片并替换为 base64 data URI，避免暴露上游地址。
// 同域地址保持原样。
func convertResultURLs(result model.JSON, channelBaseURL string) model.JSON {
	baseRoot := extractRootDomain(channelBaseURL)

	if u, ok := result["url"].(string); ok && u != "" {
		result["url"] = maybeToBase64(u, baseRoot)
	}

	if rawURLs, ok := result["urls"]; ok {
		if arr, ok := rawURLs.([]interface{}); ok {
			converted := make([]interface{}, len(arr))
			for i, v := range arr {
				if s, ok := v.(string); ok {
					converted[i] = maybeToBase64(s, baseRoot)
				} else {
					converted[i] = v
				}
			}
			result["urls"] = converted
		}
	}

	return result
}

// maybeToBase64 若 imgURL 与 baseHost 属于同一根域名（同一供应商），则下载转为 data URI，否则原样返回。
// 例：api.wuyinkeji.com 与 openpt.wuyinkeji.com 根域均为 wuyinkeji.com → 转换。
// 下载失败时保留原 URL（不阻断任务完成）。
func maybeToBase64(imgURL, baseHost string) string {
	if extractRootDomain(imgURL) != baseHost {
		return imgURL
	}
	data, mime, err := downloadImage(imgURL)
	if err != nil {
		return imgURL
	}
	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(data))
}

func extractHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Host)
}

// extractRootDomain 提取 URL 的根域名（最后两段），用于判断是否同一供应商。
// 例：api.wuyinkeji.com → wuyinkeji.com，example.com → example.com。
func extractRootDomain(rawURL string) string {
	host := extractHost(rawURL)
	if host == "" {
		return ""
	}
	parts := strings.Split(host, ".")
	if len(parts) <= 2 {
		return host
	}
	return strings.Join(parts[len(parts)-2:], ".")
}

func downloadImage(imgURL string) ([]byte, string, error) {
	resp, err := imageHTTPClient.Get(imgURL) // #nosec G107 — URL 来自可信的上游 API 响应
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("http %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	mime := resp.Header.Get("Content-Type")
	if mime == "" {
		mime = "image/png"
	}
	if idx := strings.Index(mime, ";"); idx != -1 {
		mime = strings.TrimSpace(mime[:idx])
	}
	return data, mime, nil
}
