package script

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ResolveHeaderValue 将 Header 值里的占位符替换为运行时生成的值。
// 每次请求调用一次，支持同一个值里出现多个占位符。
//
// 支持的占位符：
//   - {{random_int}}          随机非负整数（0 ~ 9999999999）
//   - {{random_int:N}}        0 ~ N 的随机整数
//   - {{uuid}}                随机 UUID v4（如 550e8400-e29b-41d4-a716-446655440000）
//   - {{timestamp}}           当前 Unix 时间戳（秒）
//   - {{timestamp_ms}}        当前 Unix 时间戳（毫秒）
//   - {{random_string:N}}     N 位随机字母数字字符串
//   - {{pool_key}}            号池分配的 API Key（原始值，不加任何前缀）
func ResolveHeaderValue(val, poolKeyValue string) string {
	if !strings.Contains(val, "{{") {
		return val
	}
	return placeholderRe.ReplaceAllStringFunc(val, func(m string) string {
		inner := m[2 : len(m)-2] // 去掉 {{ 和 }}
		switch {
		case inner == "pool_key":
			return poolKeyValue

		case inner == "uuid":
			return uuid.New().String()

		case inner == "timestamp":
			return strconv.FormatInt(time.Now().Unix(), 10)

		case inner == "timestamp_ms":
			return strconv.FormatInt(time.Now().UnixMilli(), 10)

		case inner == "random_int":
			return strconv.FormatInt(rand.Int63n(10_000_000_000), 10)

		case strings.HasPrefix(inner, "random_int:"):
			nStr := strings.TrimPrefix(inner, "random_int:")
			if n, err := strconv.ParseInt(strings.TrimSpace(nStr), 10, 64); err == nil && n > 0 {
				return strconv.FormatInt(rand.Int63n(n), 10)
			}
			return strconv.FormatInt(rand.Int63n(10_000_000_000), 10)

		case strings.HasPrefix(inner, "random_string:"):
			nStr := strings.TrimPrefix(inner, "random_string:")
			if n, err := strconv.Atoi(strings.TrimSpace(nStr)); err == nil && n > 0 {
				return randomString(n)
			}
			return randomString(16)

		default:
			return m // 未知占位符原样保留
		}
	})
}

// ResolveHeaders 对 Headers map 的所有值执行占位符替换，返回新的 map（不修改原始数据）。
func ResolveHeaders(headers map[string]interface{}, poolKeyValue string) map[string]interface{} {
	if len(headers) == 0 {
		return headers
	}
	out := make(map[string]interface{}, len(headers))
	for k, v := range headers {
		if sv, ok := v.(string); ok {
			out[k] = ResolveHeaderValue(sv, poolKeyValue)
		} else {
			out[k] = v
		}
	}
	return out
}

// placeholderRe 匹配 {{...}} 形式的占位符。
var placeholderRe = regexp.MustCompile(`\{\{[^{}]+\}\}`)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
