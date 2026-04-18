package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/service"

	"github.com/gin-gonic/gin"
)

// POST /admin/cards/generate
// Body: { "count": 10, "credits": 10000000, "note": "批次A" }
func GenerateCards(c *gin.Context) {
	var req struct {
		Count   int    `json:"count" binding:"required,min=1,max=500"`
		Credits int64  `json:"credits" binding:"required,min=1"`
		Note    string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID := c.GetInt64("user_id")
	cards := make([]*model.Card, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		code, err := genCode()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成卡密失败"})
			return
		}
		cards = append(cards, &model.Card{
			Code:      code,
			Credits:   req.Credits,
			Status:    "unused",
			Note:      req.Note,
			CreatedBy: adminID,
		})
	}
	if _, err := db.Engine.Insert(&cards); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cards": cards, "count": len(cards)})
}

// GET /admin/cards?status=unused&page=1&size=20
func ListCards(c *gin.Context) {
	status := c.Query("status") // "", "unused", "used"
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 200 {
		size = 20
	}

	sess := db.Engine.NewSession().OrderBy("id DESC")
	if status != "" {
		sess.Where("status = ?", status)
	}

	var cards []model.Card
	total, err := sess.Limit(size, (page-1)*size).FindAndCount(&cards)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cards": cards, "total": total})
}

// DELETE /admin/cards/:id  — 删除未使用的卡密
func DeleteCard(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var card model.Card
	if found, _ := db.Engine.ID(id).Get(&card); !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if card.Status != "unused" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能删除未使用的卡密"})
		return
	}
	db.Engine.ID(id).Delete(new(model.Card))
	c.JSON(http.StatusOK, gin.H{"message": "卡密已删除"})
}

// POST /user/cards/redeem
// Body: { "code": "FANAPI-XXXXXXXXXXXXXXXX" }
func RedeemCard(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	code := strings.TrimSpace(strings.ToUpper(req.Code))
	userID := c.GetInt64("user_id")

	// 使用数据库乐观锁：先查再 CAS 更新
	var card model.Card
	if found, _ := db.Engine.Where("code = ?", code).Get(&card); !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "卡密不存在"})
		return
	}
	if card.Status != "unused" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "卡密已被使用"})
		return
	}

	now := time.Now()
	updated, err := db.Engine.Where("id = ? AND status = 'unused'", card.ID).
		Cols("status", "used_by", "used_at").
		Update(&model.Card{Status: "used", UsedBy: userID, UsedAt: &now})
	if err != nil || updated == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "兑换失败，请重试"})
		return
	}

	// 写充值流水并加余额
	corrID := fmt.Sprintf("card-%d-%d", card.ID, userID)
	if err := service.WriteTx(c.Request.Context(), userID, 0, 0, 0, corrID, "recharge", card.Credits, 0, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新余额失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"credits": card.Credits,
		"message": fmt.Sprintf("兑换成功，已充值 ¥%.4f", float64(card.Credits)/1e6),
	})
}

// GET /user/cards/redeem-history?page=1&size=20
// 查询当前用户的兑换记录
func GetRedeemHistory(c *gin.Context) {
	userID := c.GetInt64("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 200 {
		size = 20
	}

	var records []model.Card
	total, err := db.Engine.
		Where("used_by = ? AND status = 'used'", userID).
		OrderBy("used_at DESC").
		Limit(size, (page-1)*size).
		FindAndCount(&records)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"records": records, "total": total})
}

// genCode 生成随机卡密，格式：FANAPI-XXXXXXXXXXXXXXXX（16位大写hex）
func genCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "FANAPI-" + strings.ToUpper(hex.EncodeToString(b)), nil
}
