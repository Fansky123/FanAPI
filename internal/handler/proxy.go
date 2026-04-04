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

// asyncTaskMsg 是通过 NATS 发送给 script worker 的消息结构。
// Payload 存储平台标准格式（size="4k", aspect_ratio="9:16"），
// worker 在执行时调用 channel.RequestScript（JS 脚本）将其转换为 vendor 格式。
type asyncTaskMsg struct {
	TaskID    int64                  `json:"task_id"`
	ChannelID int64                  `json:"channel_id"`
	UserID    int64                  `json:"user_id"`
	Payload   map[string]interface{} `json:"payload"`
}

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

	channelIDStr := c.Query("channel_id")
	if channelIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id required"})
		return
	}
	channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	ch, err := service.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 精确计费：图片/视频/音频在请求时参数已全部已知，无需两阶段结算
	cost, _, calcErr := billing.Calc(ch, reqData)
	if calcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "billing error: " + calcErr.Error()})
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
	task := &model.Task{
		UserID:         userID,
		ChannelID:      channelID,
		APIKeyID:       apiKeyIDVal,
		Type:           taskType,
		Status:         "pending",
		Request:        reqJSON,
		CreditsCharged: cost,
	}
	if _, err := db.Engine.Insert(task); err != nil {
		_ = billing.Refund(c.Request.Context(), userID, cost)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	// 写计费流水
	corrID := uuid.New().String()
	_ = service.WriteTx(c.Request.Context(), userID, channelID, apiKeyIDVal, corrID, "charge", cost, upstreamCost, model.JSON{
		"task_id": task.ID,
		"type":    taskType,
	})

	// 发布到 NATS；subject 格式：task.{image|video|audio}.{channel_id}
	natSubject := fmt.Sprintf("task.%s.%d", taskType, channelID)
	msg := asyncTaskMsg{
		TaskID:    task.ID,
		ChannelID: channelID,
		UserID:    userID,
		Payload:   reqData,
	}
	msgBytes, _ := json.Marshal(msg)
	if pubErr := mq.Publish(natSubject, msgBytes); pubErr != nil {
		db.Engine.Where("id = ?", task.ID).Update(&model.Task{Status: "failed", ErrorMsg: "publish error"})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue task"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"task_id": task.ID})
}

// POST /v1/image
func CreateImageTask(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}
	req, err := bindAudioRequest(bodyBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createTask(c, "audio", req.ToMap())
}
