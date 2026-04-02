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

// buildTaskResult 根据 task 状态组装标准 TaskResult。
// done 状态直接从 task.Result 里读取（response_script 已映射好），
// 其余状态由平台合成，不依赖上游响应。
func buildTaskResult(task *model.Task) model.TaskResult {
switch task.Status {
case "pending":
return model.TaskResult{Code: 150, Status: 0, Msg: "排队中"}

case "processing":
return model.TaskResult{Code: 150, Status: 1, Msg: "生成中"}

case "done":
// task.Result 由 worker 的 response_script 写入，已经是标准格式
// 从 JSON 字段中提取 code / url / status / msg
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
return model.TaskResult{Code: code, Status: statusVal, URL: url, Msg: msg}

case "failed":
return model.TaskResult{Code: 500, Status: 3, Msg: task.ErrorMsg}

default:
return model.TaskResult{Code: 150, Status: 0, Msg: task.Status}
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
