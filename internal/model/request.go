package model

// ImageRequest 图片生成接口的标准入参（平台统一格式）。
//
// 固定字段：
//   - Model：模型名称，如 "flux-pro"
//   - Prompt：图片描述词
//   - Size：分辨率档位，"1k"/"2k"/"3k"/"4k" 分别对应长边像素 1024/2048/3072/4096。
//     与 AspectRatio 配合决定最终宽高："2k" + "16:9" = 2048×1152 的横版图，
//     "2k" + "9:16" = 1152×2048 的竖版图，留空默认 "1k"
//   - AspectRatio：宽高比，如 "16:9"、"9:16"、"1:1"；与 Size 共同决定实际像素；留空默认 1:1 方图
//   - ReferImages：参考图片 URL 列表（图生图场景）
//   - N：生成数量，默认 1
//
// 模型特有的额外参数通过 Extra 透传到 JS 映射脚本，不做强制校验。
type ImageRequest struct {
	Model       string                 `json:"model" binding:"required"`
	Prompt      string                 `json:"prompt" binding:"required"`
	Size        string                 `json:"size"`         // 分辨率档位："1k"/"2k"/"3k"/"4k"，与 AspectRatio 共同决定最终宽高
	AspectRatio string                 `json:"aspect_ratio"` // 宽高比，如 "16:9"、"9:16"、"1:1"，留空默认 1:1
	ReferImages []string               `json:"refer_images"`
	N           int                    `json:"n"`
	Extra       map[string]interface{} `json:"-"` // 模型自定义扩展字段（从原始 JSON 提取）
}

// VideoRequest 视频生成接口的标准入参（平台统一格式）。
//
// 固定字段：
//   - Model：模型名称，如 "grok-imagine-video"
//   - Prompt：视频描述词
//   - Size：分辨率档位，"720p"/"1080p" 对应高度像素（标准视频），"2k"/"4k" 对应长边像素。
//     与 AspectRatio 配合决定最终宽高："720p" + "16:9" = 1280×720（横屏），
//     "720p" + "9:16" = 720×1280（竖屏），留空默认 "720p"
//   - AspectRatio：宽高比，如 "16:9"（横屏）、"9:16"（竖屏）、"1:1"；留空默认 16:9
//   - Duration：时长秒数字符串，如 "5"、"10"、"15"
//   - ReferImages：参考图片 URL 列表（图生视频场景）
//
// 额外参数通过 Extra 透传到 JS 脚本。
type VideoRequest struct {
	Model       string                 `json:"model" binding:"required"`
	Prompt      string                 `json:"prompt" binding:"required"`
	Size        string                 `json:"size"`         // 分辨率档位："720p"/"1080p"/"2k"/"4k"，与 AspectRatio 共同决定宽高
	AspectRatio string                 `json:"aspect_ratio"` // 宽高比，如 "16:9"（横屏）、"9:16"（竖屏），留空默认 16:9
	Duration    string                 `json:"duration"`     // 时长秒数字符串，如 "5"
	ReferImages []string               `json:"refer_images"`
	Extra       map[string]interface{} `json:"-"`
}

// AudioRequest 音频生成/语音合成接口的标准入参。
//
// 固定字段：
//   - Model：模型名称
//   - Input：文本内容（TTS 场景）或音频 URL（ASR 场景）
//   - Voice：发音人/音色，如 "alloy"
//   - Duration：目标时长（秒），用于计费预扣
type AudioRequest struct {
	Model    string                 `json:"model" binding:"required"`
	Input    string                 `json:"input"`
	Voice    string                 `json:"voice"`
	Duration int                    `json:"duration"` // 秒，计费用
	Extra    map[string]interface{} `json:"-"`
}

// ToMap 将 ImageRequest 序列化为 map（保留 Extra 扩展字段），供 billing 和 JS 脚本使用。
func (r *ImageRequest) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"model":        r.Model,
		"prompt":       r.Prompt,
		"size":         r.Size,
		"aspect_ratio": r.AspectRatio,
		"refer_images": r.ReferImages,
		"n":            r.N,
	}
	for k, v := range r.Extra {
		if _, exists := m[k]; !exists {
			m[k] = v
		}
	}
	return m
}

// ToMap 将 VideoRequest 序列化为 map，供 billing 和 JS 脚本使用。
func (r *VideoRequest) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"model":        r.Model,
		"prompt":       r.Prompt,
		"size":         r.Size,
		"aspect_ratio": r.AspectRatio,
		"duration":     r.Duration,
		"refer_images": r.ReferImages,
	}
	for k, v := range r.Extra {
		if _, exists := m[k]; !exists {
			m[k] = v
		}
	}
	return m
}

// ToMap 将 AudioRequest 序列化为 map。
func (r *AudioRequest) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"model":    r.Model,
		"input":    r.Input,
		"voice":    r.Voice,
		"duration": r.Duration,
	}
	for k, v := range r.Extra {
		if _, exists := m[k]; !exists {
			m[k] = v
		}
	}
	return m
}
