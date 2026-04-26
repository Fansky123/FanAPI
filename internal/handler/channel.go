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

const creditsPerCNY = 1_000_000.0

func creditsToCNY(credits int64) float64 {
	return float64(credits) / creditsPerCNY
}

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID 格式错误"})
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

// PATCH /admin/channels/:id/active — 仅更新渠道启用状态，不影响其他字段
func PatchChannelActive(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID 格式错误"})
		return
	}
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.PatchChannelActive(c.Request.Context(), id, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /admin/channels/:id
func DeleteChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID 格式错误"})
		return
	}
	if err := service.DeleteChannel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "渠道已删除"})
}

// PUT /admin/users/:id/password — 管理员重置任意用户密码
func ResetUserPassword(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID 格式错误"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	affected, err := db.Engine.ID(targetID).Cols("password_hash").Update(&model.User{PasswordHash: string(hash)})
	if err != nil || affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "密码已重置"})
}

// POST /admin/users/:id/recharge — 为用户手动充值（直接填写 credits 数量）
func Recharge(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID 格式错误"})
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
	total, err := db.Engine.Cols("id", "username", "email", "role", "group", "balance", "is_active", "created_at").
		Limit(size, (page-1)*size).FindAndCount(&users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users, "total": total})
}

// PUT /admin/users/:id/group — 设置用户定价分组
func SetUserGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID 格式错误"})
		return
	}
	var req struct {
		Group string `json:"group"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := db.Engine.ID(id).Cols("group").Update(&model.User{Group: req.Group}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "group updated"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_at 时间格式错误"})
		return
	}
	endAt, err := parseDateTime(c.Query("end_at"), true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_at 时间格式错误"})
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
			WHEN type IN ('charge','settle','hold') THEN credits
			WHEN type = 'refund' THEN -credits
			ELSE 0 END), 0) AS revenue,
		COALESCE(SUM(CASE
			WHEN type IN ('charge','settle','hold') THEN cost
			WHEN type = 'refund' THEN -cost
			ELSE 0 END), 0) AS cost,
		COALESCE(SUM(CASE
			WHEN type IN ('charge','settle','hold') THEN credits - cost
			WHEN type = 'refund' THEN -(credits - cost)
			ELSE 0 END), 0) AS profit,
		COUNT(*) AS count
	FROM billing_transactions ` + where
	if _, err := db.Engine.SQL(sql, args...).Get(&summary); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type transactionView struct {
		ID           int64      `json:"id"`
		UserID       int64      `json:"user_id"`
		ChannelID    int64      `json:"channel_id"`
		APIKeyID     int64      `json:"api_key_id"`
		PoolKeyID    int64      `json:"pool_key_id"`
		CorrID       string     `json:"corr_id"`
		Type         string     `json:"type"`
		Amount       float64    `json:"amount"`
		Cost         float64    `json:"cost"`
		Profit       float64    `json:"profit"`
		BalanceAfter int64      `json:"balance_after"`
		Metrics      model.JSON `json:"metrics"`
		CreatedAt    time.Time  `json:"created_at"`
	}

	views := make([]transactionView, len(txs))
	for i, tx := range txs {
		profitCredits := int64(0)
		if tx.Type == "refund" {
			profitCredits = -(tx.Credits - tx.Cost)
		} else if tx.Type == "charge" || tx.Type == "settle" || tx.Type == "hold" {
			profitCredits = tx.Credits - tx.Cost
		}

		views[i] = transactionView{
			ID:           tx.ID,
			UserID:       tx.UserID,
			ChannelID:    tx.ChannelID,
			APIKeyID:     tx.APIKeyID,
			PoolKeyID:    tx.PoolKeyID,
			CorrID:       tx.CorrID,
			Type:         tx.Type,
			Amount:       creditsToCNY(tx.Credits),
			Cost:         creditsToCNY(tx.Cost),
			Profit:       creditsToCNY(profitCredits),
			BalanceAfter: tx.BalanceAfter,
			Metrics:      tx.Metrics,
			CreatedAt:    tx.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": views,
		"total":        total,
		"summary": gin.H{
			"revenue":           creditsToCNY(summary.Revenue),
			"cost":              creditsToCNY(summary.Cost),
			"profit":            creditsToCNY(summary.Profit),
			"transaction_count": summary.Count,
		},
	})
}

// GetAdminStats GET /admin/stats
func GetAdminStats(c *gin.Context) {
	totalChannels, _ := db.Engine.Count(new(model.Channel))
	activeChannels, _ := db.Engine.Where("is_active = true").Count(new(model.Channel))
	totalUsers, _ := db.Engine.Where("role = 'user'").Count(new(model.User))

	type sumRow struct {
		Revenue int64
		Cost    int64
		Count   int64
	}

	var todayRow, totalRow sumRow

	today := time.Now().Truncate(24 * time.Hour)
	// revenue = charge(图片/视频/音频一次性扣费) + settle(LLM实际结算) - refund(退款)
	// cost    = 对应类型的上游成本（refund 抄销对应的预写成本）
	db.Engine.SQL(`SELECT
		COALESCE(SUM(CASE
			WHEN type IN ('charge','settle','hold') THEN credits
			WHEN type = 'refund' THEN -credits
			ELSE 0 END),0) AS revenue,
		COALESCE(SUM(CASE
			WHEN type IN ('charge','settle','hold') THEN cost
			WHEN type = 'refund' THEN -cost
			ELSE 0 END),0) AS cost,
		COUNT(*) AS count
	FROM billing_transactions
	WHERE type IN ('charge','settle','hold','refund') AND created_at >= ?`, today).Get(&todayRow)

	db.Engine.SQL(`SELECT
		COALESCE(SUM(CASE
			WHEN type IN ('charge','settle','hold') THEN credits
			WHEN type = 'refund' THEN -credits
			ELSE 0 END),0) AS revenue,
		COALESCE(SUM(CASE
			WHEN type IN ('charge','settle','hold') THEN cost
			WHEN type = 'refund' THEN -cost
			ELSE 0 END),0) AS cost,
		COUNT(*) AS count
	FROM billing_transactions
	WHERE type IN ('charge','settle','hold','refund')`).Get(&totalRow)

	c.JSON(http.StatusOK, gin.H{
		"channels":        totalChannels,
		"active_channels": activeChannels,
		"users":           totalUsers,
		"today": gin.H{
			"revenue": todayRow.Revenue,
			"cost":    todayRow.Cost,
			"profit":  todayRow.Revenue - todayRow.Cost,
			"count":   todayRow.Count,
		},
		"total": gin.H{
			"revenue": totalRow.Revenue,
			"cost":    totalRow.Cost,
			"profit":  totalRow.Revenue - totalRow.Cost,
			"count":   totalRow.Count,
		},
	})
}
