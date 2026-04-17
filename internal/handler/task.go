package handler

import (
	"net/http"
	"strconv"

	"fanapi/internal/db"
	"fanapi/internal/model"

	"github.com/gin-gonic/gin"
)

// GetTaskBilling 查询任务计费明细
// @Summary      查询任务计费明细
// @Description  返回指定任务的全部计费流水及汇总（净扣费、是否已退款）。
// @Tags         任务
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      int  true  "任务 ID"
// @Success      200  {object}  object{transactions=[]model.BillingTransaction,total_charged=int,total_refunded=int,net_charged=int,refunded=bool}
// @Failure      404  {object}  object  "任务不存在"
// @Router       /v1/tasks/{id}/billing [get]
func GetTaskBilling(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务 ID 格式错误"})
		return
	}
	userID := c.MustGet("user_id").(int64)

	task := &model.Task{}
	found, err := db.Engine.Where("id = ? AND user_id = ?", id, userID).
		Cols("id", "corr_id", "credits_charged", "status").Get(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败，请稍后重试"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	var txs []model.BillingTransaction
	if task.CorrID != "" {
		if err := db.Engine.Where("user_id = ? AND corr_id = ?", userID, task.CorrID).
			Cols("id", "corr_id", "type", "credits", "balance_after", "metrics", "created_at").
			Asc("id").Find(&txs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败，请稍后重试"})
			return
		}
	}

	var totalCharged, totalRefunded int64
	for _, tx := range txs {
		switch tx.Type {
		case "charge", "hold", "settle":
			totalCharged += tx.Credits
		case "refund":
			totalRefunded += tx.Credits
		}
	}
	netCharged := totalCharged - totalRefunded

	c.JSON(http.StatusOK, gin.H{
		"transactions":   txs,
		"total_charged":  totalCharged,
		"total_refunded": totalRefunded,
		"net_charged":    netCharged,
		"refunded":       totalRefunded > 0,
	})
}

// GetTask 查询任务结果
// @Summary      查询任务结果
// @Description  轮询图片/视频/音频/音乐任务结果。code=150 进行中，200 成功，500 失败。
// @Tags         任务
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      int  true  "任务 ID"
// @Success      200  {object}  model.TaskResult
// @Failure      404  {object}  object  "任务不存在"
// @Router       /v1/tasks/{id} [get]
func GetTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务 ID 格式错误"})
		return
	}
	userID := c.MustGet("user_id").(int64)

	task := &model.Task{}
	found, err := db.Engine.Where("id = ? AND user_id = ?", id, userID).Get(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败，请稍后重试"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, buildTaskResult(task))
}

// GET /admin/tasks
func ListTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	query := db.Engine.Desc("id")
	if taskID := c.Query("task_id"); taskID != "" {
		query = query.Where("id = ?", taskID)
	}
	if userID := c.Query("user_id"); userID != "" {
		query = query.And("user_id = ?", userID)
	}
	if status := c.Query("status"); status != "" {
		query = query.And("status = ?", status)
	}
	if taskType := c.Query("type"); taskType != "" {
		query = query.And("type = ?", taskType)
	}
	if startAt := c.Query("start_at"); startAt != "" {
		query = query.And("created_at >= ?", startAt)
	}
	if endAt := c.Query("end_at"); endAt != "" {
		query = query.And("created_at <= ?", endAt)
	}

	var tasks []model.Task
	total, err := query.Cols("id", "user_id", "channel_id", "api_key_id", "type", "status",
		"progress", "upstream_task_id",
		"error_msg", "credits_charged", "corr_id", "created_at", "updated_at").
		Limit(size, (page-1)*size).FindAndCount(&tasks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks, "total": total})
}

// ListUserTasks 查询当前用户的任务列表
// @Summary      查询任务列表
// @Description  分页获取当前 API Key 对应用户的历史任务。
// @Tags         任务
// @Produce      json
// @Security     ApiKeyAuth
// @Param        page      query     int     false  "页码（默认 1）"
// @Param        size      query     int     false  "每页条数（默认 20，最大 100）"
// @Param        status    query     string  false  "状态过滤：pending/processing/done/failed"
// @Param        type      query     string  false  "任务类型过滤：image/video/audio/music"
// @Param        task_id   query     int     false  "按 task_id 精确查询"
// @Param        start_at  query     string  false  "创建时间起（2006-01-02 15:04:05）"
// @Param        end_at    query     string  false  "创建时间止"
// @Success      200  {object}  object{tasks=[]model.TaskResult,total=int}
// @Router       /v1/tasks [get]
func ListUserTasks(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	query := db.Engine.Where("user_id = ?", userID).Desc("id")
	if status := c.Query("status"); status != "" {
		query = query.And("status = ?", status)
	}
	if taskType := c.Query("type"); taskType != "" {
		query = query.And("type = ?", taskType)
	}
	if taskID := c.Query("task_id"); taskID != "" {
		query = query.And("id = ?", taskID)
	}
	if startAt := c.Query("start_at"); startAt != "" {
		query = query.And("created_at >= ?", startAt)
	}
	if endAt := c.Query("end_at"); endAt != "" {
		query = query.And("created_at <= ?", endAt)
	}

	var tasks []model.Task
	total, err := query.Cols("id", "user_id", "channel_id", "api_key_id", "type", "status",
		"progress", "upstream_task_id",
		"error_msg", "credits_charged", "corr_id", "created_at", "updated_at").
		Limit(size, (page-1)*size).FindAndCount(&tasks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败，请稍后重试"})
		return
	}

	results := make([]model.TaskResult, 0, len(tasks))
	for i := range tasks {
		results = append(results, buildTaskResult(&tasks[i]))
	}
	c.JSON(http.StatusOK, gin.H{"tasks": results, "total": total})
}

// GET /admin/tasks/:id
func GetAdminTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务 ID 格式错误"})
		return
	}
	task := &model.Task{}
	found, err := db.Engine.Where("id = ?", id).Get(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败，请稍后重试"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": task})
}

// buildTaskResult 根据 task 状态组装标准 TaskResult。
// done 状态直接从 task.Result 里读取（response_script 已映射好），
// 其余状态由平台合成，不依赖上游响应。
func buildTaskResult(task *model.Task) model.TaskResult {
	base := model.TaskResult{
		TaskID:         task.ID,
		TaskType:       task.Type,
		ChannelID:      task.ChannelID,
		CreditsCharged: task.CreditsCharged,
		CreatedAt:      task.CreatedAt,
		Request:        task.Request,  // 原始请求参数
		Result:         task.Result,   // 映射后的响应结果
	}
	switch task.Status {
	case "pending":
		base.Code = 150
		base.Status = 0
		base.Msg = "排队中"
		return base

	case "processing":
		base.Code = 150
		base.Status = 1
		base.Msg = "生成中"
		return base

	case "done":
		t := task.UpdatedAt
		base.FinishedAt = &t
		code := 200
		if v, ok := task.Result["code"]; ok {
			if n, ok := toInt(v); ok {
				code = n
			}
		}
		statusVal := 2
		if v, ok := task.Result["status"]; ok {
			if n, ok := toInt(v); ok {
				statusVal = n
			}
		}
		url, _ := task.Result["url"].(string)
		msg, _ := task.Result["msg"].(string)
		base.Code = code
		base.Status = statusVal
		base.URL = url
		base.Msg = msg
		// 多结果任务（如音乐每次生成两首）
		if items, ok := task.Result["items"]; ok {
			if arr, ok := items.([]interface{}); ok {
				base.Items = arr
			}
		}
		return base

	case "failed":
		t := task.UpdatedAt
		base.FinishedAt = &t
		base.Code = 500
		base.Status = 3
		base.Msg = task.ErrorMsg
		return base

	default:
		base.Code = 150
		base.Status = 0
		base.Msg = task.Status
		return base
	}
}

// toInt 将 JSON 数值（float64）或 int 类型安全转换为 int。
func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	}
	return 0, false
}
