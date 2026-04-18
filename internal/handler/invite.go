package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"fanapi/internal/billing"
	"fanapi/internal/db"
	"fanapi/internal/model"

	"github.com/gin-gonic/gin"
)

// GetInviteInfo 返回当前用户的邀请码、已邀请人数、冻结积分余额。
//
// @Summary      查询邀请信息
// @Description  返回邀请码、邀请人数及冻结积分（待解冻返佣）
// @Tags         邀请
// @Security     BearerAuth
// @Success      200  {object}  object{invite_code=string,invite_count=int,frozen_balance=int}
// @Router       /user/invite [get]
func GetInviteInfo(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	var user model.User
	if found, err := db.Engine.ID(userID).Cols("invite_code", "frozen_balance").Get(&user); err != nil || !found {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取用户信息失败"})
		return
	}

	// 兼容旧账号：若邀请码为空则自动生成并持久化
	if user.InviteCode == "" {
		code := generateInviteCode()
		if n, err := db.Engine.Exec(
			"UPDATE users SET invite_code = $1 WHERE id = $2 AND (invite_code IS NULL OR invite_code = '')",
			code, userID,
		); err == nil {
			if rows, _ := n.RowsAffected(); rows > 0 {
				user.InviteCode = code
			}
		}
		// 若并发导致 UPDATE 未命中（其他实例已写入），重新读一次
		if user.InviteCode == "" {
			db.Engine.ID(userID).Cols("invite_code").Get(&user) //nolint:errcheck
		}
	}

	count, _ := db.Engine.Where("inviter_id = ?", userID).Count(&model.User{})

	c.JSON(http.StatusOK, gin.H{
		"invite_code":    user.InviteCode,
		"invite_count":   count,
		"frozen_balance": user.FrozenBalance,
	})
}

// ConvertFrozenBalance 将冻结积分转为可用积分。
//
// @Summary      解冻积分
// @Description  将指定数量的冻结返佣积分转换为可用余额
// @Tags         邀请
// @Security     BearerAuth
// @Param        body  body      object{amount=int}  true  "解冻数量（单位：积分，0 表示全部）"
// @Success      200   {object}  object{converted=int,available_balance=int}
// @Router       /user/invite/convert [post]
func ConvertFrozenBalance(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	var req struct {
		Amount int64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	if found, err := db.Engine.ID(userID).Cols("frozen_balance").Get(&user); err != nil || !found {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取用户信息失败"})
		return
	}

	toConvert := req.Amount
	if toConvert <= 0 || toConvert > user.FrozenBalance {
		toConvert = user.FrozenBalance
	}
	if toConvert <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无可用冻结积分"})
		return
	}

	// 原子操作：减少 frozen_balance，增加可用余额（Redis + DB）
	n, err := db.Engine.Exec(
		"UPDATE users SET frozen_balance = frozen_balance - $1 WHERE id = $2 AND frozen_balance >= $1",
		toConvert, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解冻失败，请稍后重试"})
		return
	}
	affected, _ := n.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "冻结积分不足"})
		return
	}

	// 加入 Redis 余额（Refund 实际是向余额增加积分）
	if err := billing.Refund(c.Request.Context(), userID, toConvert); err != nil {
		// Redis 失败回滚 DB
		db.Engine.Exec("UPDATE users SET frozen_balance = frozen_balance + $1 WHERE id = $2", toConvert, userID) //nolint:errcheck
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解冻失败（余额更新），请稍后重试"})
		return
	}

	newBalance, _ := billing.GetBalance(c.Request.Context(), userID)
	c.JSON(http.StatusOK, gin.H{
		"converted":         toConvert,
		"available_balance": newBalance,
	})
}

// generateInviteCode 生成 16 位十六进制邀请码（本包内使用）。
func generateInviteCode() string {
	b := make([]byte, 8)
	rand.Read(b) //nolint:errcheck
	return hex.EncodeToString(b)
}
