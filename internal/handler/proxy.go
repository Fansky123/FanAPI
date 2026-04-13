package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"fanapi/internal/billing"
	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/mq"
	"fanapi/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// bindImageRequest 将请求 body 解析为 ImageRequest。
// 先按结构体绑定固定字段（做必填校验），再将原始 JSON 中其余字段写入 Extra，
// Extra 字段经 ToMap() 合并后透传给 JS 映射脚本。
func bindImageRequest(bodyBytes []byte) (*model.ImageRequest, error) {
	var req model.ImageRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return nil, err
	}
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}
	if req.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	var raw map[string]interface{}
	_ = json.Unmarshal(bodyBytes, &raw)
	known := map[string]bool{"model": true, "prompt": true, "size": true, "aspect_ratio": true, "refer_images": true, "n": true}
	req.Extra = make(map[string]interface{})
	for k, v := range raw {
		if !known[k] {
			req.Extra[k] = v
		}
	}
	return &req, nil
}

// bindVideoRequest 将请求 body 解析为 VideoRequest 并合并 Extra 字段。
func bindVideoRequest(bodyBytes []byte) (*model.VideoRequest, error) {
	var req model.VideoRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return nil, err
	}
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}
	if req.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	var raw map[string]interface{}
	_ = json.Unmarshal(bodyBytes, &raw)
	known := map[string]bool{"model": true, "prompt": true, "size": true, "aspect_ratio": true, "duration": true, "refer_images": true}
	req.Extra = make(map[string]interface{})
	for k, v := range raw {
		if !known[k] {
			req.Extra[k] = v
		}
	}
	return &req, nil
}

// bindAudioRequest 将请求 body 解析为 AudioRequest 并合并 Extra 字段。
func bindAudioRequest(bodyBytes []byte) (*model.AudioRequest, error) {
	var req model.AudioRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return nil, err
	}
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}
	var raw map[string]interface{}
	_ = json.Unmarshal(bodyBytes, &raw)
	known := map[string]bool{"model": true, "input": true, "voice": true, "duration": true}
	req.Extra = make(map[string]interface{})
	for k, v := range raw {
		if !known[k] {
			req.Extra[k] = v
		}
	}
	return &req, nil
}

// createTask 是图片/视频/音频任务的通用创建逻辑。
// reqData 是平台标准格式的 map（由 bind* + ToMap() 产生），包含 size、aspect_ratio 等字段。
// 计费在此精确完成（size+aspect_ratio 已知）；worker 内执行 request_script 再转为 vendor 格式。
func createTask(c *gin.Context, taskType string, reqData map[string]interface{}) {
	userID := c.MustGet("user_id").(int64)
	apiKeyID, _ := c.Get("api_key_id")
	var apiKeyIDVal int64
	if apiKeyID != nil {
		apiKeyIDVal = apiKeyID.(int64)
	}

	// 渠道解析：优先 channel_id 查询参数（兼容旧客户端），否则用 reqData["model"] 按渠道名路由。
	var ch *model.Channel
	if channelIDStr := c.Query("channel_id"); channelIDStr != "" {
		channelID, parseErr := strconv.ParseInt(channelIDStr, 10, 64)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id 格式错误"})
			return
		}
		var chErr error
		ch, chErr = service.GetChannel(c.Request.Context(), channelID)
		if chErr != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": chErr.Error()})
			return
		}
	} else {
		routingModel, _ := reqData["model"].(string)
		if routingModel == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请指定 model 或 channel_id"})
			return
		}
		var chErr error
		ch, chErr = service.GetChannelByName(c.Request.Context(), routingModel)
		if chErr != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "渠道不存在: " + routingModel})
			return
		}
	}
	channelID := ch.ID

	// 用渠道配置的真实模型名覆盖用户传入的路由键。
	if ch.Model != "" {
		reqData["model"] = ch.Model
	}

	// 精确计费：图片/视频/音频在请求时参数已全部已知，无需两阶段结算
	cost, _, calcErr := billing.Calc(ch, reqData)
	if calcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "计费计算失败，请稍后重试"})
		return
	}
	// 计算上游进价成本（用于记录利润，不影响用户扣费）
	upstreamCost, _ := billing.CalcUpstreamCost(ch, reqData)

	if cost > 0 {
		if chargeErr := billing.Charge(c.Request.Context(), userID, cost); chargeErr != nil {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": chargeErr.Error()})
			return
		}
	}

	// 将平台标准格式原样存入 DB，方便排障；vendor 格式只在 worker 内转换
	reqJSON := model.JSON{}
	for k, v := range reqData {
		reqJSON[k] = v
	}
	corrID := uuid.New().String()
	task := &model.Task{
		UserID:         userID,
		ChannelID:      channelID,
		APIKeyID:       apiKeyIDVal,
		Type:           taskType,
		Status:         "pending",
		Request:        reqJSON,
		CreditsCharged: cost,
		CorrID:         corrID,
	}
	if _, err := db.Engine.Insert(task); err != nil {
		_ = billing.Refund(c.Request.Context(), userID, cost)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败，请稍后重试"})
		return
	}

	// 写计费流水
	_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, corrID, "charge", cost, upstreamCost, model.JSON{
		"task_id": task.ID,
		"type":    taskType,
	})

	// 解析号池 Key（原来在 worker 内做，现在移到这里，worker 无需访问 Redis）
	var poolKeyID int64
	var poolKeyValue string
	if ch.KeyPoolID > 0 {
		pk, pkErr := service.GetOrAssignPoolKey(c.Request.Context(), ch.KeyPoolID, userID)
		if pkErr != nil {
			db.Engine.Where("id = ?", task.ID).Cols("status", "error_msg").Update(&model.Task{Status: "failed", ErrorMsg: "key pool error"})
			if cost > 0 {
				_ = billing.Refund(c.Request.Context(), userID, cost)
				_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, corrID, "refund", cost, upstreamCost, model.JSON{
					"task_id": task.ID,
					"reason":  "key pool error",
				})
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "号池分配失败: " + pkErr.Error()})
			return
		}
		poolKeyID = pk.ID
		poolKeyValue = pk.Value
	}

	// 发布到 NATS；消息携带渠道完整配置，worker 只需 NATS 连接
	natSubject := fmt.Sprintf("task.%s.%d", taskType, channelID)
	job := &model.TaskJob{
		TaskID:         task.ID,
		TaskType:       taskType,
		UserID:         userID,
		APIKeyID:       apiKeyIDVal,
		CorrID:         corrID,
		CreditsCharged: cost,
		ChannelID:      channelID,
		BaseURL:        ch.BaseURL,
		Method:         ch.Method,
		Headers:        ch.Headers,
		TimeoutMs:      ch.TimeoutMs,
		QueryTimeoutMs: ch.QueryTimeoutMs,
		RequestScript:  ch.RequestScript,
		ResponseScript: ch.ResponseScript,
		ErrorScript:    ch.ErrorScript,
		QueryURL:       ch.QueryURL,
		QueryMethod:    ch.QueryMethod,
		QueryScript:    ch.QueryScript,
		PoolKeyID:      poolKeyID,
		PoolKeyValue:   poolKeyValue,
		Payload:        reqData,
	}
	msgBytes, _ := json.Marshal(job)
	if pubErr := mq.Publish(natSubject, msgBytes); pubErr != nil {
		db.Engine.Where("id = ?", task.ID).Cols("status", "error_msg").Update(&model.Task{Status: "failed", ErrorMsg: "publish error"})
		if cost > 0 {
			_ = billing.Refund(c.Request.Context(), userID, cost)
			_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, corrID, "refund", cost, upstreamCost, model.JSON{
				"task_id": task.ID,
				"reason":  "publish error",
			})
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "任务投递失败，请稍后重试"})
		return
	}

	// 标记为处理中（worker 收到后即开始执行）
	db.Engine.Where("id = ?", task.ID).Cols("status").Update(&model.Task{Status: "processing"})

	c.JSON(http.StatusAccepted, gin.H{"task_id": task.ID})
}

// POST /v1/image
func CreateImageTask(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return
	}
	req, err := bindImageRequest(bodyBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createTask(c, "image", req.ToMap())
}

// POST /v1/video
func CreateVideoTask(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return
	}
	req, err := bindVideoRequest(bodyBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createTask(c, "video", req.ToMap())
}

// POST /v1/audio
func CreateAudioTask(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return
	}
	req, err := bindAudioRequest(bodyBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createTask(c, "audio", req.ToMap())
}

// bindMusicRequest 将请求 body 解析为 MusicRequest 并合并 Extra 字段。
func bindMusicRequest(bodyBytes []byte) (*model.MusicRequest, error) {
	var req model.MusicRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return nil, err
	}
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}
	var raw map[string]interface{}
	_ = json.Unmarshal(bodyBytes, &raw)
	known := map[string]bool{
		"model": true, "input_type": true, "mv_version": true, "make_instrumental": true,
		"gpt_description_prompt": true, "prompt": true, "tags": true, "title": true,
		"continue_clip_id": true, "continue_at": true, "cover_clip_id": true,
		"task": true, "metadata_params": true, "callback_url": true,
	}
	req.Extra = make(map[string]interface{})
	for k, v := range raw {
		if !known[k] {
			req.Extra[k] = v
		}
	}
	return &req, nil
}

// POST /v1/music
func CreateMusicTask(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return
	}
	req, err := bindMusicRequest(bodyBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createTask(c, "music", req.ToMap())
}
