package handler

import (
	"net/http"
	"strconv"

	"fanapi/internal/db"
	"fanapi/internal/model"

	"github.com/gin-gonic/gin"
	"xorm.io/xorm"
)

// GET /admin/llm-logs
// Query params: user_id, channel_id, status, corr_id, model, start_at, end_at, page, page_size
func AdminListLLMLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	type filterSet struct {
		userID    string
		channelID string
		status    string
		corrID    string
		model     string
		startAt   string
		endAt     string
	}
	f := filterSet{
		userID:    c.Query("user_id"),
		channelID: c.Query("channel_id"),
		status:    c.Query("status"),
		corrID:    c.Query("corr_id"),
		model:     c.Query("model"),
		startAt:   c.Query("start_at"),
		endAt:     c.Query("end_at"),
	}

	applyFilters := func() *xorm.Session {
		s := db.Engine.NewSession()
		if f.userID != "" {
			s.And("user_id = ?", f.userID)
		}
		if f.channelID != "" {
			s.And("channel_id = ?", f.channelID)
		}
		if f.status != "" {
			s.And("status = ?", f.status)
		}
		if f.corrID != "" {
			s.And("corr_id = ?", f.corrID)
		}
		if f.model != "" {
			s.And("model = ?", f.model)
		}
		if f.startAt != "" {
			if t, err := parseDateTime(f.startAt, false); err == nil {
				s.And("created_at >= ?", t)
			}
		}
		if f.endAt != "" {
			if t, err := parseDateTime(f.endAt, true); err == nil {
				s.And("created_at <= ?", t)
			}
		}
		return s
	}

	countSess := applyFilters()
	defer countSess.Close()
	total, err := countSess.Count(new(model.LLMLog))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	listSess := applyFilters()
	defer listSess.Close()
	var logs []model.LLMLog
	err = listSess.OrderBy("id DESC").Limit(pageSize, offset).Find(&logs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 聚合每条日志对应的净扣费积分
	creditsMap := map[string]int64{}
	if len(logs) > 0 {
		type txRow struct {
			CorrID  string `xorm:"corr_id"`
			Credits int64  `xorm:"credits"`
		}
		inList := "'" + logs[0].CorrID + "'"
		for _, l := range logs[1:] {
			inList += ",'" + l.CorrID + "'"
		}
		sqlStr := `SELECT corr_id,
			COALESCE(SUM(CASE WHEN type IN ('hold','charge','settle') THEN credits WHEN type='refund' THEN -credits ELSE 0 END),0) AS credits
			FROM billing_transactions WHERE corr_id IN (` + inList + `) GROUP BY corr_id`
		var rows []txRow
		_ = db.Engine.SQL(sqlStr).Find(&rows)
		for _, r := range rows {
			creditsMap[r.CorrID] = r.Credits
		}
	}

	type logWithCredits struct {
		model.LLMLog
		CreditsCharged int64 `json:"credits_charged"`
	}
	result := make([]logWithCredits, len(logs))
	for i, l := range logs {
		result[i] = logWithCredits{LLMLog: l, CreditsCharged: creditsMap[l.CorrID]}
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":      result,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GET /admin/llm-logs/:id
func AdminGetLLMLog(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var log model.LLMLog
	has, err := db.Engine.ID(id).Get(&log)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, log)
}

// GET /v1/llm-logs  (用户查自己的日志，不含 upstream_request 详情)
func UserListLLMLogs(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	sess := db.Engine.Where("user_id = ?", userID)

	if v := c.Query("status"); v != "" {
		sess.And("status = ?", v)
	}
	if v := c.Query("corr_id"); v != "" {
		sess.And("corr_id = ?", v)
	}
	if v := c.Query("model"); v != "" {
		sess.And("model = ?", v)
	}
	if v := c.Query("channel_id"); v != "" {
		sess.And("channel_id = ?", v)
	}
	if v := c.Query("start_at"); v != "" {
		if t, err := parseDateTime(v, false); err == nil {
			sess.And("created_at >= ?", t)
		}
	}
	if v := c.Query("end_at"); v != "" {
		if t, err := parseDateTime(v, true); err == nil {
			sess.And("created_at <= ?", t)
		}
	}

	var total int64
	countSess := *sess
	total, err := countSess.Count(new(model.LLMLog))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var logs []model.LLMLog
	// 用户列表不返回 upstream_request / upstream_response / upstream_url 等上游信息
	err = sess.Cols("id", "corr_id", "model", "is_stream",
		"upstream_status", "usage", "status", "error_msg", "created_at").
		OrderBy("id DESC").Limit(pageSize, offset).Find(&logs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查询每条日志对应的净扣费积分（hold/charge/settle 扣除 refund 后的实际消耗）
	creditsMap := map[string]int64{}
	if len(logs) > 0 {
		corrIDs := make([]interface{}, len(logs))
		placeholders := make([]string, len(logs))
		for i, l := range logs {
			corrIDs[i] = l.CorrID
			placeholders[i] = "?" // xorm 用 ? 占位
		}
		type txRow struct {
			CorrID  string `xorm:"corr_id"`
			Credits int64  `xorm:"credits"`
		}
		var rows []txRow
		inList := "'" + logs[0].CorrID + "'"
		for _, l := range logs[1:] {
			inList += ",'" + l.CorrID + "'"
		}
		sqlStr := `SELECT corr_id,
			COALESCE(SUM(CASE WHEN type IN ('hold','charge','settle') THEN credits WHEN type='refund' THEN -credits ELSE 0 END),0) AS credits
			FROM billing_transactions WHERE corr_id IN (` + inList + `) GROUP BY corr_id`
		_ = db.Engine.SQL(sqlStr).Find(&rows)
		for _, r := range rows {
			creditsMap[r.CorrID] = r.Credits
		}
	}

	type logWithCredits struct {
		model.LLMLog
		CreditsCharged int64 `json:"credits_charged"`
	}
	result := make([]logWithCredits, len(logs))
	for i, l := range logs {
		result[i] = logWithCredits{LLMLog: l, CreditsCharged: creditsMap[l.CorrID]}
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":      result,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GET /v1/llm-logs/:id  （用户查自己某条日志的完整详情，含 upstream_request）
func UserGetLLMLog(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var log model.LLMLog
	has, err := db.Engine.ID(id).Where("user_id = ?", userID).Get(&log)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, log)
}
