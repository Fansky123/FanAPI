package handler

import (
	"net/http"
	"strconv"
	"time"

	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func parseDateTime(value string, endOfDay bool) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			if layout == "2006-01-02" && endOfDay {
				return t.Add(24*time.Hour - time.Nanosecond), nil
			}
			return t, nil
		}
	}
	return time.Time{}, strconv.ErrSyntax
}

// POST /admin/channels
func CreateChannel(c *gin.Context) {
	var ch model.Channel
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.CreateChannel(c.Request.Context(), &ch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ch)
}

// GET /admin/channels
func ListChannels(c *gin.Context) {
	channels, err := service.ListChannels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"channels": channels})
}

// PUT /admin/channels/:id
func UpdateChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var ch model.Channel
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ch.ID = id
	if err := service.UpdateChannel(c.Request.Context(), &ch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ch)
}

// DELETE /admin/channels/:id
func DeleteChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := service.DeleteChannel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "channel disabled"})
}

// PUT /admin/users/:id/password — 管理员重置任意用户密码
func ResetUserPassword(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var req struct {
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash error"})
		return
	}
	affected, err := db.Engine.ID(targetID).Cols("password_hash").Update(&model.User{PasswordHash: string(hash)})
	if err != nil || affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

// POST /admin/users/:id/recharge — 为用户手动充值（直接填写 credits 数量）
func Recharge(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	adminID := c.MustGet("user_id").(int64)

	var req struct {
		Amount int64 `json:"amount" binding:"required,gt=0"` // credits 数量
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.Recharge(c.Request.Context(), targetID, adminID, req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"credited_credits": req.Amount,
		"credited_cny":     float64(req.Amount) / 1_000_000,
	})
}

// GET /admin/users — 用户列表（分页）
func ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}

	var users []model.User
	total, err := db.Engine.Cols("id", "email", "role", "balance", "created_at").
		Limit(size, (page-1)*size).FindAndCount(&users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users, "total": total})
}

// GET /admin/transactions — 全局账单流水（分页）
func ListAllTransactions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	startAt, err := parseDateTime(c.Query("start_at"), false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_at"})
		return
	}
	endAt, err := parseDateTime(c.Query("end_at"), true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_at"})
		return
	}

	var txs []model.BillingTransaction
	query := db.Engine.Desc("id")
	if !startAt.IsZero() {
		query = query.Where("created_at >= ?", startAt)
	}
	if !endAt.IsZero() {
		query = query.And("created_at <= ?", endAt)
	}
	total, err := query.Limit(size, (page-1)*size).FindAndCount(&txs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type summaryRow struct {
		Revenue int64 `xorm:"'revenue'"`
		Cost    int64 `xorm:"'cost'"`
		Profit  int64 `xorm:"'profit'"`
		Count   int64 `xorm:"'count'"`
	}
	where := "WHERE 1=1"
	args := make([]interface{}, 0, 2)
	if !startAt.IsZero() {
		where += " AND created_at >= ?"
		args = append(args, startAt)
	}
	if !endAt.IsZero() {
		where += " AND created_at <= ?"
		args = append(args, endAt)
	}
	summary := summaryRow{}
	sql := `SELECT
		COALESCE(SUM(CASE
			WHEN type IN ('charge','settle') THEN credits
			WHEN type = 'refund' THEN -credits
			ELSE 0 END), 0) AS revenue,
		COALESCE(SUM(CASE WHEN type IN ('charge','settle') THEN cost ELSE 0 END), 0) AS cost,
		COALESCE(SUM(CASE
			WHEN type IN ('charge','settle') THEN credits - cost
			WHEN type = 'refund' THEN -credits
			ELSE 0 END), 0) AS profit,
		COUNT(*) AS count
	FROM billing_transactions ` + where
	if _, err := db.Engine.SQL(sql, args...).Get(&summary); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": txs,
		"total":        total,
		"summary": gin.H{
			"revenue":           summary.Revenue,
			"cost":              summary.Cost,
			"profit":            summary.Profit,
			"transaction_count": summary.Count,
		},
	})
}
