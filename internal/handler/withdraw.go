package handler

import (
	"net/http"
	"strconv"

	"fanapi/internal/db"
	"fanapi/internal/model"

	"github.com/gin-gonic/gin"
)

// ── 用户端 ──────────────────────────────────────────────────────────────────

// SavePaymentQR PUT /user/payment-qr  保存/更新收款码
func SavePaymentQR(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	var req struct {
		WechatQR string `json:"wechat_qr"`
		AlipayQR string `json:"alipay_qr"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 限制 base64 大小（约 300KB 原始图片）
	const maxLen = 400 * 1024
	if len(req.WechatQR) > maxLen || len(req.AlipayQR) > maxLen {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "收款码图片过大，请压缩后上传"})
		return
	}

	user := &model.User{PaymentQRWechat: req.WechatQR, PaymentQRAlipay: req.AlipayQR}
	if _, err := db.Engine.ID(userID).Cols("payment_qr_wechat", "payment_qr_alipay").Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetPaymentQR GET /user/payment-qr  获取当前收款码
func GetPaymentQR(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	var user model.User
	if found, err := db.Engine.ID(userID).Cols("payment_qr_wechat", "payment_qr_alipay").Get(&user); err != nil || !found {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"wechat_qr": user.PaymentQRWechat,
		"alipay_qr": user.PaymentQRAlipay,
	})
}

// SubmitWithdraw POST /user/withdraw  提交提现申请
func SubmitWithdraw(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	var req struct {
		Amount      int64  `json:"amount" binding:"required,min=1"`
		PaymentType string `json:"payment_type" binding:"required,oneof=wechat alipay"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 读取冻结余额和收款码
	var user model.User
	if found, err := db.Engine.ID(userID).
		Cols("frozen_balance", "payment_qr_wechat", "payment_qr_alipay").
		Get(&user); err != nil || !found {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取用户信息失败"})
		return
	}

	if req.Amount > user.FrozenBalance {
		c.JSON(http.StatusBadRequest, gin.H{"error": "提现积分超过冻结余额"})
		return
	}

	// 选取对应收款码快照
	qr := user.PaymentQRWechat
	if req.PaymentType == "alipay" {
		qr = user.PaymentQRAlipay
	}
	if qr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先在收款码设置中上传对应收款码"})
		return
	}

	// 检查是否已有待处理申请
	pending, _ := db.Engine.Where("user_id = ? AND status = 'pending'", userID).Count(&model.WithdrawRequest{})
	if pending > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "您已有待处理的提现申请，请等待处理后再提交"})
		return
	}

	record := &model.WithdrawRequest{
		UserID:      userID,
		Amount:      req.Amount,
		Status:      "pending",
		PaymentType: req.PaymentType,
		PaymentQR:   qr,
	}
	if _, err := db.Engine.Insert(record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": record.ID})
}

// ListWithdrawHistory GET /user/withdraw/history  用户提现记录
func ListWithdrawHistory(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	var records []model.WithdrawRequest
	total, err := db.Engine.Where("user_id = ?", userID).
		OrderBy("created_at DESC").
		Limit(size, (page-1)*size).
		FindAndCount(&records)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	// 不向用户返回收款码图片快照（数据量大）
	for i := range records {
		records[i].PaymentQR = ""
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "records": records})
}

// ── 管理端 ──────────────────────────────────────────────────────────────────

// AdminListWithdrawals GET /admin/withdrawals  列出所有提现申请
func AdminListWithdrawals(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	status := c.Query("status") // 可选过滤：pending/approved/rejected
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	sess := db.Engine.Table("withdraw_requests").
		Select("withdraw_requests.*, users.username").
		Join("LEFT", "users", "users.id = withdraw_requests.user_id").
		OrderBy("withdraw_requests.created_at DESC")
	if status != "" {
		sess = sess.Where("withdraw_requests.status = ?", status)
	}

	var records []model.WithdrawRequest
	total, err := sess.Limit(size, (page-1)*size).FindAndCount(&records)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "records": records})
}

// AdminPendingWithdrawCount GET /admin/withdrawals/pending-count  待处理数量（徽标用）
func AdminPendingWithdrawCount(c *gin.Context) {
	count, _ := db.Engine.Where("status = 'pending'").Count(&model.WithdrawRequest{})
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// AdminApproveWithdrawal POST /admin/withdrawals/:id/approve  审批通过并划扣积分
func AdminApproveWithdrawal(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Remark string `json:"remark"`
	}
	_ = c.ShouldBindJSON(&req)

	var record model.WithdrawRequest
	if found, err := db.Engine.ID(id).Get(&record); err != nil || !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		return
	}
	if record.Status != "pending" {
		c.JSON(http.StatusConflict, gin.H{"error": "申请已处理"})
		return
	}

	// 原子扣减冻结余额
	n, err := db.Engine.Exec(
		"UPDATE users SET frozen_balance = frozen_balance - $1 WHERE id = $2 AND frozen_balance >= $1",
		record.Amount, record.UserID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "扣减积分失败"})
		return
	}
	affected, _ := n.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "用户冻结积分不足，无法划扣"})
		return
	}

	// 更新申请状态
	record.Status = "approved"
	record.AdminRemark = req.Remark
	if _, err := db.Engine.ID(id).Cols("status", "admin_remark").Update(&record); err != nil {
		// 回滚冻结余额
		db.Engine.Exec("UPDATE users SET frozen_balance = frozen_balance + $1 WHERE id = $2", record.Amount, record.UserID) //nolint:errcheck
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// AdminRejectWithdrawal POST /admin/withdrawals/:id/reject  拒绝提现申请
func AdminRejectWithdrawal(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Remark string `json:"remark"`
	}
	_ = c.ShouldBindJSON(&req)

	var record model.WithdrawRequest
	if found, err := db.Engine.ID(id).Get(&record); err != nil || !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		return
	}
	if record.Status != "pending" {
		c.JSON(http.StatusConflict, gin.H{"error": "申请已处理"})
		return
	}

	record.Status = "rejected"
	record.AdminRemark = req.Remark
	if _, err := db.Engine.ID(id).Cols("status", "admin_remark").Update(&record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
