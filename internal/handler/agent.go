package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/service"

	"github.com/gin-gonic/gin"
)

// GET /agent/users — 只返回该客服邀请的用户（按余额升序）
func AgentListUsers(c *gin.Context) {
	agentID := c.MustGet("user_id").(int64)
	page := 1
	size := 50
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	if s := c.Query("size"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 200 {
			size = n
		}
	}

	type userRow struct {
		ID            int64   `json:"id"`
		Username      string  `json:"username"`
		Email         *string `json:"email"`
		Balance       int64   `json:"balance"`
		TotalRecharge int64   `json:"total_recharge"`
		TotalSpend    int64   `json:"total_spend"`
	}

	rows, err := db.Engine.QueryString(`
SELECT
u.id, u.username, u.email, u.balance,
COALESCE((SELECT SUM(credits) FROM billing_transactions WHERE user_id = u.id AND type = 'recharge'), 0) AS total_recharge,
COALESCE((SELECT SUM(credits) FROM billing_transactions WHERE user_id = u.id AND type IN ('charge','hold','settle')), 0) AS total_spend
FROM users u
WHERE u.inviter_id = $1
ORDER BY u.balance ASC
LIMIT $2 OFFSET $3
`, agentID, size, (page-1)*size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	result := make([]userRow, 0, len(rows))
	for _, r := range rows {
		id, _ := strconv.ParseInt(r["id"], 10, 64)
		balance, _ := strconv.ParseInt(r["balance"], 10, 64)
		recharge, _ := strconv.ParseInt(r["total_recharge"], 10, 64)
		spend, _ := strconv.ParseInt(r["total_spend"], 10, 64)
		row := userRow{
			ID:            id,
			Username:      r["username"],
			Balance:       balance,
			TotalRecharge: recharge,
			TotalSpend:    spend,
		}
		if email, ok := r["email"]; ok && email != "" {
			row.Email = &email
		}
		result = append(result, row)
	}

	totalRows, _ := db.Engine.Where("inviter_id = ?", agentID).Count(new(model.User))
	c.JSON(http.StatusOK, gin.H{"users": result, "total": totalRows})
}

// GET /agent/invite — 获取当前客服的邀请码
func AgentGetInvite(c *gin.Context) {
	agentID := c.MustGet("user_id").(int64)
	agent := &model.User{}
	found, err := db.Engine.ID(agentID).Cols("id", "invite_code", "wechat_qr").Get(agent)
	if err != nil || !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	if agent.InviteCode == "" {
		agent.InviteCode = fmt.Sprintf("%d%s", agentID, generateCode6())
		db.Engine.ID(agentID).Cols("invite_code").Update(agent) //nolint:errcheck
	}
	c.JSON(http.StatusOK, gin.H{
		"invite_code": agent.InviteCode,
		"wechat_qr":   agent.WechatQR,
	})
}

// POST /agent/users/:id/recharge — 仅允许为自己邀请的用户充值
func AgentRecharge(c *gin.Context) {
	agentID := c.MustGet("user_id").(int64)
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID 格式错误"})
		return
	}

	target := &model.User{}
	found, err := db.Engine.ID(targetID).Cols("id", "inviter_id").Get(target)
	if err != nil || !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	if target.InviterID == nil || *target.InviterID != agentID {
		c.JSON(http.StatusForbidden, gin.H{"error": "只能为自己邀请的用户充值"})
		return
	}

	var req struct {
		Amount int64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.Recharge(c.Request.Context(), targetID, agentID, req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"credited_credits": req.Amount,
		"credited_cny":     float64(req.Amount) / 1_000_000,
	})
}

// PUT /agent/wechat-qr — 客服更新自己的微信二维码
func AgentUpdateWechatQR(c *gin.Context) {
	agentID := c.MustGet("user_id").(int64)
	var req struct {
		WechatQR string `json:"wechat_qr" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := db.Engine.ID(agentID).Cols("wechat_qr").Update(&model.User{WechatQR: req.WechatQR}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "微信二维码已更新"})
}

// PUT /admin/users/:id/role — 管理员设置用户角色
func SetUserRole(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID 格式错误"})
		return
	}
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	allowed := map[string]bool{"user": true, "agent": true, "admin": true}
	if !allowed[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "角色值无效，允许: user / agent / admin"})
		return
	}
	if _, err := db.Engine.ID(targetID).Cols("role").Update(&model.User{Role: req.Role}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "角色已更新"})
}

func safeMapVal(rows []map[string]string, key string) string {
	if len(rows) > 0 {
		return rows[0][key]
	}
	return "0"
}

func generateCode6() string {
	b := make([]byte, 3)
	rand.Read(b) //nolint:errcheck
	return hex.EncodeToString(b)
}
