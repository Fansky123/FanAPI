package billing

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"fanapi/internal/model"
)

// Calc 根据请求参数计算预扣费用。
//
// 返回值：
//   - inputHold：输入部分的预扣金额（credits）
//   - outputHold：输出部分的预扣金额（credits）
//
// 对于 token 类型：
//   - 若 billing_config.input_from_response = true，则输入费由响应 usage 字段结算，
//     inputHold 为基于消息内容长度的估算值（防止余额耗尽）；
//   - 否则从请求中精确计算输入 token 数。
//
// 对于其他类型（image/video/audio/count/custom）：
//   - 在请求时即可精确计算全部费用，outputHold 始终为 0，无需结算退差。
func Calc(ch *model.Channel, req map[string]interface{}) (inputHold int64, outputHold int64, err error) {
	return CalcForUser(ch, req, "")
}

// CalcForUser 与 Calc 相同，但支持用户分组定价。
// userGroup 为空时使用渠道默认价。billing_config 中可通过 "pricing_groups" 字段覆盖：
//
//	{
//	  "input_price_per_1m_tokens": 15,
//	  "output_price_per_1m_tokens": 60,
//	  "pricing_groups": {
//	    "vip": {"input_price_per_1m_tokens": 8, "output_price_per_1m_tokens": 32}
//	  }
//	}
func CalcForUser(ch *model.Channel, req map[string]interface{}, userGroup string) (inputHold int64, outputHold int64, err error) {
	cfg := applyGroupPricing(map[string]interface{}(ch.BillingConfig), userGroup)
	data := map[string]map[string]interface{}{"request": req}

	switch ch.BillingType {
	case "token":
		return calcToken(cfg, data)
	case "image":
		cost, e := calcImage(cfg, data)
		return cost, 0, e
	case "video":
		cost, e := calcVideo(cfg, data)
		return cost, 0, e
	case "audio":
		cost, e := calcAudio(cfg, data)
		return cost, 0, e
	case "count":
		cost, e := calcCount(cfg)
		return cost, 0, e
	case "custom":
		cost, e := calcCustom(ch.BillingScript, req)
		return cost, 0, e
	default:
		return 0, 0, fmt.Errorf("未知计费类型: %s", ch.BillingType)
	}
}

// applyGroupPricing merges group-specific pricing fields on top of the base cfg.
func applyGroupPricing(cfg map[string]interface{}, group string) map[string]interface{} {
	if group == "" {
		return cfg
	}
	pgs, ok := cfg["pricing_groups"].(map[string]interface{})
	if !ok {
		return cfg
	}
	overrides, ok := pgs[group].(map[string]interface{})
	if !ok {
		return cfg
	}
	// shallow merge: override only the pricing keys present in the group config
	merged := make(map[string]interface{}, len(cfg))
	for k, v := range cfg {
		merged[k] = v
	}
	for k, v := range overrides {
		merged[k] = v
	}
	return merged
}

// CalcActualCost 根据请求 + SSE 响应中的实际用量计算真实总费用（仅用于 LLM token 类型结算）。
//
// 无论 input_from_response 如何，结算值始终包含输入 + 输出两部分：
//   - input_from_response=false：输入 token 数从请求中计算，与 hold 保持一致
//   - input_from_response=true：输入 token 数从响应 usage 字段读取（更精确）
func CalcActualCost(ch *model.Channel, req, resp map[string]interface{}) (int64, error) {
	return CalcActualCostForUser(ch, req, resp, "")
}

func CalcActualCostForUser(ch *model.Channel, req, resp map[string]interface{}, userGroup string) (int64, error) {
	if ch.BillingType != "token" {
		return 0, nil
	}
	cfg := applyGroupPricing(map[string]interface{}(ch.BillingConfig), userGroup)
	data := map[string]map[string]interface{}{"request": req, "response": resp}

	outputPricePer1m := getInt64Val(cfg, "output_price_per_1m_tokens")
	inputPricePer1m := getInt64Val(cfg, "input_price_per_1m_tokens")

	// 从响应获取实际输出 token 数
	outputPath := getStr(cfg, "metric_paths.output_tokens", "response.usage.completion_tokens")
	outputTokens, _ := getInt64FromData(data, outputPath)
	outputCost := int64(math.Ceil(float64(outputTokens) * float64(outputPricePer1m) / 1000000))

	// 缓存计费说明：
	//
	// OpenAI / Gemini 协议：
	//   - prompt_tokens（或 promptTokenCount）**包含**缓存命中的 token（cached 是其子集）。
	//   - 正确做法：先从 inputTokens 中扣除 cacheReadTokens，再按常规价计算；
	//     cacheReadTokens 单独按折扣价（默认 0.5x）计算。
	//   - 若不扣除，缓存 token 会被按"1.0x + 0.5x = 1.5x"收费，严重超收。
	//
	// Claude 协议：
	//   - input_tokens **不含**缓存 token（cache_creation / cache_read 是独立字段）。
	//   - 正确做法：inputTokens 按常规价，cacheCreateTokens 按 1.25x，cacheReadTokens 按 0.1x，
	//     三者互不重叠，无需扣除。
	//
	// 结论：只有 OpenAI / Gemini 协议需要先扣除 cacheReadTokens 再计算基础输入费用。
	proto := ch.Protocol
	if proto == "" {
		proto = "openai"
	}
	// openaiStyleCache 表示 prompt_tokens 已将缓存 token 计入其中（OpenAI、Gemini 均如此）
	openaiStyleCache := proto == "openai" || proto == "gemini"

	cacheCreatePricePer1m := getInt64Val(cfg, "cache_creation_price_per_1m_tokens")
	if cacheCreatePricePer1m == 0 && inputPricePer1m > 0 {
		// Claude cache write = 1.25x input；OpenAI/Gemini 无写入缓存计费，此处默认也用 1.25x
		cacheCreatePricePer1m = int64(math.Ceil(float64(inputPricePer1m) * 1.25))
	}
	cacheReadPricePer1m := getInt64Val(cfg, "cache_read_price_per_1m_tokens")
	if cacheReadPricePer1m == 0 && inputPricePer1m > 0 {
		// 各协议缓存读取默认倍率不同，不设置时按协议取合理默认值：
		//   Claude  : 0.10x（$0.30/$3，缓存读取大幅折扣）
		//   Gemini  : 0.25x（Context Caching 官方折扣）
		//   OpenAI  : 0.50x（Prompt Caching 官方折扣，如 gpt-4o）
		var cacheReadRatio float64
		switch proto {
		case "claude":
			cacheReadRatio = 0.10
		case "gemini":
			cacheReadRatio = 0.25
		default: // openai
			cacheReadRatio = 0.50
		}
		cacheReadPricePer1m = int64(math.Ceil(float64(inputPricePer1m) * cacheReadRatio))
	}
	cacheCreateTokens, _ := getInt64FromData(data, "response.usage.cache_creation_tokens")
	cacheReadTokens, _ := getInt64FromData(data, "response.usage.cache_read_tokens")
	cacheCost := int64(math.Ceil(float64(cacheCreateTokens)*float64(cacheCreatePricePer1m)/1000000)) +
		int64(math.Ceil(float64(cacheReadTokens)*float64(cacheReadPricePer1m)/1000000))

	// calcInputCost 计算基础输入费用：
	// OpenAI/Gemini 协议下，inputTokens 已包含缓存 token，需先扣除再按正常价计费。
	calcInputCost := func(inputTokens int64) int64 {
		base := inputTokens
		if openaiStyleCache && cacheReadTokens > 0 {
			base -= cacheReadTokens
			if base < 0 {
				base = 0
			}
		}
		return int64(math.Ceil(float64(base) * float64(inputPricePer1m) / 1000000))
	}

	if !getBool(cfg, "input_from_response") {
		// 输入 token 数从请求中计算（与 hold 阶段保持一致），结算 = 输入 + 实际输出 + 缓存
		inputPath := getStr(cfg, "metric_paths.input_tokens", "request.input_tokens")
		inputTokens, err := getInt64FromData(data, inputPath)
		if err != nil {
			inputTokens = estimateTokensFromMessages(req)
		}
		return calcInputCost(inputTokens) + outputCost + cacheCost, nil
	}

	// input_from_response=true：从响应 usage 中获取实际输入 token 数
	inputPath := getStr(cfg, "metric_paths.input_tokens", "response.usage.prompt_tokens")
	inputTokens, _ := getInt64FromData(data, inputPath)
	return calcInputCost(inputTokens) + outputCost + cacheCost, nil
}

// ---- 各计费类型内部计算函数 ----

// calcToken 计算 LLM token 类型的预扣费用。
// 预扣仅包含输入费用；输出费在结算时按实际 token 数计算。
func calcToken(cfg map[string]interface{}, data map[string]map[string]interface{}) (int64, int64, error) {
	inputPricePer1m := getInt64Val(cfg, "input_price_per_1m_tokens")

	if getBool(cfg, "input_from_response") {
		// 输入费延迟到响应结算，预扣时用消息内容长度估算，避免余额不足风险
		// 输出费不预扣，结算时根据实际 token 数计算
		inputEst := estimateTokensFromMessages(data["request"])
		inputHold := int64(math.Ceil(float64(inputEst) * float64(inputPricePer1m) / 1000000))
		return inputHold, 0, nil
	}

	// 从请求字段精确获取输入 token 数；输出费不预扣，结算时按实际用量扣除
	inputPath := getStr(cfg, "metric_paths.input_tokens", "request.input_tokens")
	inputTokens, err := getInt64FromData(data, inputPath)
	if err != nil {
		// 路径不存在时降级为消息估算
		inputTokens = estimateTokensFromMessages(data["request"])
	}
	inputCost := int64(math.Ceil(float64(inputTokens) * float64(inputPricePer1m) / 1000000))
	return inputCost, 0, nil
}

// calcImage 根据请求中的 size 档位、宽高比、数量计算图片生成费用。
//
// 支持两种定价模式（billing_config 示例见下）：
//
// 模式一：size_prices（按档位字符串直接定价，推荐用于各档位成本明确的场景）
//
//	{
//	  "size_prices": { "1k": 5000, "2k": 15000, "4k": 50000 },
//	  "default_size_price": 50000,   // size 不在映射中时的兜底价
//	  "metric_paths": { "size": "request.size", "count": "request.n" }
//	}
//
// 模式二：base_price + resolution_tiers（按像素总数分档乘以倍率）
//
//	{
//	  "base_price": 10000,
//	  "resolution_tiers": [{"max_pixels":1048576,"multiplier":1.0}, ...],
//	  "metric_paths": { "size": "request.size", "aspect_ratio": "request.aspect_ratio", "count": "request.n" }
//	}
func calcImage(cfg map[string]interface{}, data map[string]map[string]interface{}) (int64, error) {
	sizePath := getStr(cfg, "metric_paths.size", "request.size")
	countPath := getStr(cfg, "metric_paths.count", "request.n")

	sizeStr := getStrFromData(data, sizePath)
	count, err := getInt64FromData(data, countPath)
	if err != nil || count == 0 {
		count = 1
	}

	// 模式一：size_prices 映射表（按 size 字符串直接定价）
	if sizePricesRaw, ok := cfg["size_prices"]; ok {
		b, _ := json.Marshal(sizePricesRaw)
		var sizePrices map[string]int64
		if json.Unmarshal(b, &sizePrices) == nil {
			sizeKey := strings.ToLower(strings.TrimSpace(sizeStr))
			if price, found := sizePrices[sizeKey]; found {
				return price * count, nil
			}
			// 兜底：default_size_price
			if def := getInt64Val(cfg, "default_size_price"); def > 0 {
				return def * count, nil
			}
			// 取映射中最大的价格作为最终兜底
			var maxPrice int64
			for _, p := range sizePrices {
				if p > maxPrice {
					maxPrice = p
				}
			}
			return maxPrice * count, nil
		}
	}

	// 模式二：base_price + resolution_tiers（原有逻辑，按像素分档乘倍率）
	ratioPath := getStr(cfg, "metric_paths.aspect_ratio", "request.aspect_ratio")
	ratioStr := getStrFromData(data, ratioPath)
	pixels := ParseSizeToPixels(sizeStr, ratioStr)
	multiplier := resolutionMultiplier(cfg, pixels)
	basePrice := getInt64Val(cfg, "base_price")
	return int64(float64(basePrice) * multiplier * float64(count)), nil
}

// calcVideo 根据请求中的 size 档位、宽高比、时长计算视频生成费用。
// size（"720p"/"1080p"/"2k"/"4k"）与 aspect_ratio（如 "9:16"）共同决定实际像素数，乘以时长和倍率。
func calcVideo(cfg map[string]interface{}, data map[string]map[string]interface{}) (int64, error) {
	sizePath := getStr(cfg, "metric_paths.size", "request.size")
	ratioPath := getStr(cfg, "metric_paths.aspect_ratio", "request.aspect_ratio")
	durPath := getStr(cfg, "metric_paths.duration", "request.duration")

	sizeStr := getStrFromData(data, sizePath)
	ratioStr := getStrFromData(data, ratioPath)
	duration, _ := getInt64FromData(data, durPath)

	pixels := ParseSizeToPixels(sizeStr, ratioStr)
	multiplier := resolutionMultiplier(cfg, pixels)
	pricePerSec := getInt64Val(cfg, "price_per_second")
	return int64(float64(pricePerSec) * float64(duration) * multiplier), nil
}

// calcAudio 根据请求中的时长计算音频生成费用。
func calcAudio(cfg map[string]interface{}, data map[string]map[string]interface{}) (int64, error) {
	durPath := getStr(cfg, "metric_paths.duration", "request.duration")
	duration, _ := getInt64FromData(data, durPath)
	pricePerSec := getInt64Val(cfg, "price_per_second")
	return pricePerSec * duration, nil
}

// calcCount 按次固定收费。
func calcCount(cfg map[string]interface{}) (int64, error) {
	return getInt64Val(cfg, "price_per_call"), nil
}

// calcCustom 调用 JS 自定义计费脚本（goja 运行时）。
func calcCustom(script string, req map[string]interface{}) (int64, error) {
	return RunBillingScript(script, req, nil)
}

// ---- 辅助函数 ----

// resolutionMultiplier 根据像素数从分辨率分档配置中匹配倍率。
func resolutionMultiplier(cfg map[string]interface{}, pixels int64) float64 {
	tiersRaw, ok := cfg["resolution_tiers"]
	if !ok {
		return 1.0
	}
	b, _ := json.Marshal(tiersRaw)
	var tiers []struct {
		MaxPixels  int64   `json:"max_pixels"`
		Multiplier float64 `json:"multiplier"`
	}
	if err := json.Unmarshal(b, &tiers); err != nil {
		return 1.0
	}
	for _, t := range tiers {
		if pixels <= t.MaxPixels {
			return t.Multiplier
		}
	}
	if len(tiers) > 0 {
		return tiers[len(tiers)-1].Multiplier
	}
	return 1.0
}

// estimateTokensFromMessages 通过遍历请求 messages（OpenAI/Claude 格式）或 contents（Gemini 格式）
// 字段的字符总长度估算 token 数（约 4 字符 = 1 token）。
// 当无法从请求直接获取 input_tokens 时作为备用估算。
func estimateTokensFromMessages(req map[string]interface{}) int64 {
	if req == nil {
		return 0
	}
	// 优先读 messages（OpenAI / Claude 格式），没有时尝试 contents（Gemini 原生格式）
	var payload interface{}
	if msgs, ok := req["messages"]; ok {
		payload = msgs
	} else if contents, ok := req["contents"]; ok {
		payload = contents
	} else {
		return 0
	}
	totalChars := countStringLen(payload)
	if totalChars == 0 {
		return 0
	}
	// 4 字符估算为 1 token，并乘以 1.2 留出余量
	return int64(math.Ceil(float64(totalChars) / 4.0 * 1.2))
}

// EstimateTokensFromRequest 是 estimateTokensFromMessages 的公开版本，供 handler 层在
// 用户中断时基于请求内容估算 prompt_tokens。
func EstimateTokensFromRequest(req map[string]interface{}) int64 {
	return estimateTokensFromMessages(req)
}

// countStringLen 递归统计任意 JSON 结构中所有字符串值的字节长度。
func countStringLen(v interface{}) int64 {
	switch val := v.(type) {
	case string:
		return int64(len(val))
	case []interface{}:
		var total int64
		for _, item := range val {
			total += countStringLen(item)
		}
		return total
	case map[string]interface{}:
		var total int64
		for _, item := range val {
			total += countStringLen(item)
		}
		return total
	}
	return 0
}

func getInt64FromData(data map[string]map[string]interface{}, path string) (int64, error) {
	v, err := Extract(data, path)
	if err != nil {
		return 0, err
	}
	return ToInt64(v)
}

// getStrFromData 从 data 的点分隔路径中提取字符串值，路径不存在或类型不符时返回空字符串。
func getStrFromData(data map[string]map[string]interface{}, path string) string {
	v, err := Extract(data, path)
	if err != nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func getInt64Val(cfg map[string]interface{}, key string) int64 {
	v, ok := cfg[key]
	if !ok {
		return 0
	}
	n, _ := ToInt64(v)
	return n
}

// getBool 从 billing_config 中读取布尔值开关。
func getBool(cfg map[string]interface{}, key string) bool {
	v, ok := cfg[key]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

// getStr 从 billing_config 中读取字符串（支持点分隔嵌套路径）。
func getStr(cfg map[string]interface{}, key, fallback string) string {
	parts := splitKey(key)
	cur := cfg
	for i, p := range parts {
		val, ok := cur[p]
		if !ok {
			return fallback
		}
		if i == len(parts)-1 {
			if s, ok := val.(string); ok {
				return s
			}
			return fallback
		}
		sub, ok := val.(map[string]interface{})
		if !ok {
			return fallback
		}
		cur = sub
	}
	return fallback
}

func splitKey(key string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			parts = append(parts, key[start:i])
			start = i + 1
		}
	}
	return append(parts, key[start:])
}

// CalcUpstreamCost 计算本次请求需要支付给上游供应商的进价成本（预估值）。
//
// BillingConfig 中的进价字段（与售价字段同结构，以 _cost_ 代替 _price_）：
//   - token 类型：input_cost_per_1m_tokens、output_cost_per_1m_tokens
//   - image 类型：base_cost（替代 base_price）
//   - video/audio 类型：cost_per_second（替代 price_per_second）
//   - count 类型：cost_per_call（替代 price_per_call）
//   - custom 类型：不支持，返回 0
//
// 若渠道未配置进价字段，则进价默认为 0（即成本未知）。
func CalcUpstreamCost(ch *model.Channel, req map[string]interface{}) (int64, error) {
	cfg := map[string]interface{}(ch.BillingConfig)
	data := map[string]map[string]interface{}{"request": req}

	switch ch.BillingType {
	case "token":
		inputHold, outputHold, err := calcUpstreamToken(cfg, data)
		return inputHold + outputHold, err
	case "image":
		return calcUpstreamImage(cfg, data)
	case "video":
		return calcUpstreamVideo(cfg, data)
	case "audio":
		return calcUpstreamAudio(cfg, data)
	case "count":
		return getInt64Val(cfg, "cost_per_call"), nil
	case "custom":
		return 0, nil
	default:
		return 0, nil
	}
}

// CalcActualUpstreamCost 根据响应中的实际用量计算上游真实进价成本（仅用于 token 类型结算）。
// 与 CalcActualCost 逻辑相同，但使用 *_cost_* 进价字段。
func CalcActualUpstreamCost(ch *model.Channel, req, resp map[string]interface{}) (int64, error) {
	if ch.BillingType != "token" {
		return 0, nil
	}
	cfg := map[string]interface{}(ch.BillingConfig)
	data := map[string]map[string]interface{}{"request": req, "response": resp}

	outputCostPer1m := getInt64Val(cfg, "output_cost_per_1m_tokens")
	inputCostPer1m := getInt64Val(cfg, "input_cost_per_1m_tokens")

	outputPath := getStr(cfg, "metric_paths.output_tokens", "response.usage.completion_tokens")
	outputTokens, _ := getInt64FromData(data, outputPath)
	outputCost := int64(math.Ceil(float64(outputTokens) * float64(outputCostPer1m) / 1000000))

	proto := ch.Protocol
	if proto == "" {
		proto = "openai"
	}
	openaiStyleCache := proto == "openai" || proto == "gemini"

	// 缓存 token 进价（与售价逻辑相同，字段名用 _cost_ 替代 _price_）
	cacheCreateCostPer1m := getInt64Val(cfg, "cache_creation_cost_per_1m_tokens")
	if cacheCreateCostPer1m == 0 && inputCostPer1m > 0 {
		cacheCreateCostPer1m = int64(math.Ceil(float64(inputCostPer1m) * 1.25))
	}
	cacheReadCostPer1m := getInt64Val(cfg, "cache_read_cost_per_1m_tokens")
	if cacheReadCostPer1m == 0 && inputCostPer1m > 0 {
		var cacheReadRatio float64
		switch proto {
		case "claude":
			cacheReadRatio = 0.10
		case "gemini":
			cacheReadRatio = 0.25
		default:
			cacheReadRatio = 0.50
		}
		cacheReadCostPer1m = int64(math.Ceil(float64(inputCostPer1m) * cacheReadRatio))
	}
	cacheCreateTokens, _ := getInt64FromData(data, "response.usage.cache_creation_tokens")
	cacheReadTokens, _ := getInt64FromData(data, "response.usage.cache_read_tokens")
	cacheCost := int64(math.Ceil(float64(cacheCreateTokens)*float64(cacheCreateCostPer1m)/1000000)) +
		int64(math.Ceil(float64(cacheReadTokens)*float64(cacheReadCostPer1m)/1000000))

	calcInputCost := func(inputTokens int64) int64 {
		base := inputTokens
		if openaiStyleCache && cacheReadTokens > 0 {
			base -= cacheReadTokens
			if base < 0 {
				base = 0
			}
		}
		return int64(math.Ceil(float64(base) * float64(inputCostPer1m) / 1000000))
	}

	if !getBool(cfg, "input_from_response") {
		inputPath := getStr(cfg, "metric_paths.input_tokens", "request.input_tokens")
		inputTokens, err := getInt64FromData(data, inputPath)
		if err != nil {
			inputTokens = estimateTokensFromMessages(req)
		}
		return calcInputCost(inputTokens) + outputCost + cacheCost, nil
	}

	inputPath := getStr(cfg, "metric_paths.input_tokens", "response.usage.prompt_tokens")
	inputTokens, _ := getInt64FromData(data, inputPath)
	return calcInputCost(inputTokens) + outputCost + cacheCost, nil
}

func calcUpstreamToken(cfg map[string]interface{}, data map[string]map[string]interface{}) (int64, int64, error) {
	inputCostPer1m := getInt64Val(cfg, "input_cost_per_1m_tokens")

	// 输出进价也不预扣，与 calcToken 保持一致
	if getBool(cfg, "input_from_response") {
		inputEst := estimateTokensFromMessages(data["request"])
		inputHold := int64(math.Ceil(float64(inputEst) * float64(inputCostPer1m) / 1000000))
		return inputHold, 0, nil
	}

	inputPath := getStr(cfg, "metric_paths.input_tokens", "request.input_tokens")
	inputTokens, err := getInt64FromData(data, inputPath)
	if err != nil {
		inputTokens = estimateTokensFromMessages(data["request"])
	}
	inputCost := int64(math.Ceil(float64(inputTokens) * float64(inputCostPer1m) / 1000000))
	return inputCost, 0, nil
}

func calcUpstreamImage(cfg map[string]interface{}, data map[string]map[string]interface{}) (int64, error) {
	sizePath := getStr(cfg, "metric_paths.size", "request.size")
	countPath := getStr(cfg, "metric_paths.count", "request.n")

	sizeStr := getStrFromData(data, sizePath)
	count, err := getInt64FromData(data, countPath)
	if err != nil || count == 0 {
		count = 1
	}

	// 模式一：size_costs 映射表（按档位直接定进价）
	if sizeCostsRaw, ok := cfg["size_costs"]; ok {
		b, _ := json.Marshal(sizeCostsRaw)
		var sizeCosts map[string]int64
		if json.Unmarshal(b, &sizeCosts) == nil {
			sizeKey := strings.ToLower(strings.TrimSpace(sizeStr))
			if cost, found := sizeCosts[sizeKey]; found {
				return cost * count, nil
			}
			if def := getInt64Val(cfg, "default_size_cost"); def > 0 {
				return def * count, nil
			}
			var maxCost int64
			for _, p := range sizeCosts {
				if p > maxCost {
					maxCost = p
				}
			}
			return maxCost * count, nil
		}
	}

	// 模式二：base_cost + resolution_tiers（原有逻辑）
	ratioPath := getStr(cfg, "metric_paths.aspect_ratio", "request.aspect_ratio")
	ratioStr := getStrFromData(data, ratioPath)
	pixels := ParseSizeToPixels(sizeStr, ratioStr)
	multiplier := resolutionMultiplier(cfg, pixels)
	baseCost := getInt64Val(cfg, "base_cost")
	return int64(float64(baseCost) * multiplier * float64(count)), nil
}

func calcUpstreamVideo(cfg map[string]interface{}, data map[string]map[string]interface{}) (int64, error) {
	sizePath := getStr(cfg, "metric_paths.size", "request.size")
	ratioPath := getStr(cfg, "metric_paths.aspect_ratio", "request.aspect_ratio")
	durPath := getStr(cfg, "metric_paths.duration", "request.duration")

	sizeStr := getStrFromData(data, sizePath)
	ratioStr := getStrFromData(data, ratioPath)
	duration, _ := getInt64FromData(data, durPath)

	pixels := ParseSizeToPixels(sizeStr, ratioStr)
	multiplier := resolutionMultiplier(cfg, pixels)
	costPerSec := getInt64Val(cfg, "cost_per_second")
	return int64(float64(costPerSec) * float64(duration) * multiplier), nil
}

func calcUpstreamAudio(cfg map[string]interface{}, data map[string]map[string]interface{}) (int64, error) {
	durPath := getStr(cfg, "metric_paths.duration", "request.duration")
	duration, _ := getInt64FromData(data, durPath)
	costPerSec := getInt64Val(cfg, "cost_per_second")
	return costPerSec * duration, nil
}
