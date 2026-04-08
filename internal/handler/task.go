package handler

import (
	"net/http"
	"strconv"

	"fanapi/internal/db"
	"fanapi/internal/model"

	"github.com/gin-gonic/gin"
)

// GET /v1/tasks/:id
// 返回任务的标准响应格式 TaskResult：
//   - pending    → code=150, status=0, msg="排队中"
//   - processing → code=150, status=1, msg="生成中"
//   - done       → 直接返回 task.Result（已由 response_script 映射为标准格式）
//   - failed     → code=500, status=3, msg=task.ErrorMsg
func GetTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	userID := c.MustGet("user_id").(int64)

	task := &model.Task{}
	found, err := db.Engine.Where("id = ? AND user_id = ?", id, userID).Get(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
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
	total, err := query.Limit(size, (page-1)*size).FindAndCount(&tasks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks, "total": total})
}

// GET /v1/tasks  (用户查看自己的任务列表，需 API Key)
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

	var tasks []model.Task
	total, err := query.Limit(size, (page-1)*size).FindAndCount(&tasks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	task := &model.Task{}
	found, err := db.Engine.Where("id = ?", id).Get(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
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
		UpstreamTaskID: task.UpstreamTaskID,
		CreditsCharged: task.CreditsCharged,
		Request:        task.Request,
		Result:         task.Result,
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
