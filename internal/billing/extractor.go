package billing

import (
	"fmt"
	"strconv"
	"strings"
)

// Extract 通过点分隔路径从嵌套 map 中提取值。
// 路径格式：第一段为顶层 key（"request" 或 "response"），其余段为嵌套字段。
// 示例："response.usage.completion_tokens"
func Extract(data map[string]map[string]interface{}, path string) (interface{}, error) {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid path: %s", path)
	}
	root, rest := parts[0], parts[1]
	src, ok := data[root]
	if !ok {
		return nil, fmt.Errorf("root key %q not found", root)
	}
	return extractNested(src, rest)
}

func extractNested(m map[string]interface{}, path string) (interface{}, error) {
	parts := strings.SplitN(path, ".", 2)
	val, ok := m[parts[0]]
	if !ok {
		return nil, fmt.Errorf("key %q not found", parts[0])
	}
	if len(parts) == 1 {
		return val, nil
	}
	sub, ok := val.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("key %q is not an object", parts[0])
	}
	return extractNested(sub, parts[1])
}

// ToInt64 将 interface{} 强制转换为 int64，支持 JSON 解码后的常见数值类型。
func ToInt64(v interface{}) (int64, error) {
	switch n := v.(type) {
	case float64:
		return int64(n), nil
	case int64:
		return n, nil
	case int:
		return int64(n), nil
	case int32:
		return int64(n), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}

// ParseSizeToPixels 根据分辨率档位（size）和宽高比（aspectRatio）计算总像素数，用于计费分档。
//
// size 取值规则：
//   - 图片档位："1k"/"2k"/"3k"/"4k" — 长边像素分别为 1024/2048/3072/4096
//   - 视频档位："720p"/"1080p" — 高度像素分别为 720/1080；"2k"/"4k" — 长边 2048/4096
//
// aspectRatio 格式为 "W:H"（如 "16:9"、"9:16"、"1:1"）；留空时：
//   - 图片默认 1:1（方图）
//   - 视频默认 16:9（横屏）
//
// 示例：
//   - "2k" + "16:9"  → 2048×1152 = 2,359,296 像素（横版图）
//   - "2k" + "9:16"  → 1152×2048 = 2,359,296 像素（竖版图，像素总数不变）
//   - "720p" + "9:16" → 720×1280  = 921,600 像素（竖版短视频）
func ParseSizeToPixels(size, aspectRatio string) int64 {
	// 1. 解析 size 档位 → longEdge（长边像素）和模式标志
	//    isShortEdge=true 表示 size 编码的是高度（720p/1080p 视频规格）
	var longEdge int64
	var isShortEdge bool // 视频 720p/1080p：数字代表高度而非长边
	switch strings.ToLower(strings.TrimSpace(size)) {
	case "1k":
		longEdge = 1024
	case "2k":
		longEdge = 2048
	case "3k":
		longEdge = 3072
	case "4k":
		longEdge = 4096
	case "720p":
		longEdge = 720
		isShortEdge = true
	case "1080p":
		longEdge = 1080
		isShortEdge = true
	default:
		longEdge = 1024 // 未识别时默认 1k
	}

	// 2. 解析宽高比 "W:H"
	var ratioW, ratioH int64 = 1, 1 // 默认 1:1
	ar := strings.TrimSpace(aspectRatio)
	if ar == "" {
		// 视频默认横屏 16:9
		if isShortEdge {
			ratioW, ratioH = 16, 9
		}
	} else {
		parts := strings.SplitN(ar, ":", 2)
		if len(parts) == 2 {
			w, errW := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
			h, errH := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
			if errW == nil && errH == nil && w > 0 && h > 0 {
				ratioW, ratioH = w, h
			}
		}
	}

	// 3. 根据模式计算实际宽高
	//    isShortEdge=true：longEdge 是短边（高度用于 720p/1080p）
	//    isShortEdge=false：longEdge 是长边
	var width, height int64
	if isShortEdge {
		// shortEdge 固定，通过宽高比求另一边
		shortEdge := longEdge
		if ratioW >= ratioH {
			// 横屏：高度 = shortEdge，宽度 = shortEdge × W/H
			height = shortEdge
			width = shortEdge * ratioW / ratioH
		} else {
			// 竖屏：宽度 = shortEdge，高度 = shortEdge × H/W
			width = shortEdge
			height = shortEdge * ratioH / ratioW
		}
	} else {
		// longEdge 固定，通过宽高比求另一边
		if ratioW >= ratioH {
			// 横屏或方图：宽度 = longEdge，高度 = longEdge × H/W
			width = longEdge
			height = longEdge * ratioH / ratioW
		} else {
			// 竖屏：高度 = longEdge，宽度 = longEdge × W/H
			height = longEdge
			width = longEdge * ratioW / ratioH
		}
	}

	return width * height
}
