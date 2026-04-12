package handler

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/service"

	"github.com/gin-gonic/gin"
)

const creditsPerYuan = 1_000_000 // 1 元 = 1,000,000 credits

// getSettingValue retrieves a single system setting value by key.
func getSettingValue(key string) string {
	s := &model.SystemSetting{}
	found, _ := db.Engine.Where("key = ?", key).Get(s)
	if !found {
		return ""
	}
	return s.Value
}

type epayCreateReq struct {
	Amount float64 `json:"amount" binding:"required,min=0.01"` // 充值金额（元），最低 0.01
	Type   string  `json:"type" binding:"required"`            // alipay / wxpay
}

// CreateEpayOrder creates a payment order and returns the payment redirect URL.
// POST /pay/epay/create  （需要 JWT 认证）
func CreateEpayOrder(c *gin.Context) {
	var req epayCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	epayURL := getSettingValue("epay_url")
	epayPid := getSettingValue("epay_pid")
	epayKey := getSettingValue("epay_key")
	notifyURL := getSettingValue("epay_notify_url")
	returnURL := getSettingValue("epay_return_url")
	siteName := getSettingValue("site_name")
	if siteName == "" {
		siteName = "FanAPI"
	}

	if epayURL == "" || epayPid == "" || epayKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "支付功能未配置，请联系管理员"})
		return
	}

	userID := c.MustGet("user_id").(int64)
	outTradeNo := fmt.Sprintf("FAN%d%d", userID, time.Now().UnixMilli())
	credits := int64(req.Amount * creditsPerYuan)
	moneyStr := fmt.Sprintf("%.2f", req.Amount)

	// 写入待支付订单
	order := &model.PaymentOrder{
		UserID:     userID,
		OutTradeNo: outTradeNo,
		Amount:     req.Amount,
		Credits:    credits,
		Status:     "pending",
	}
	if _, err := db.Engine.Insert(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建订单失败: " + err.Error()})
		return
	}

	params := map[string]string{
		"pid":          epayPid,
		"type":         req.Type,
		"notify_url":   notifyURL,
		"return_url":   returnURL,
		"name":         siteName + " 余额充值",
		"money":        moneyStr,
		"out_trade_no": outTradeNo,
	}
	params["sign"] = epaySign(params, epayKey)
	params["sign_type"] = "MD5"

	payURL := strings.TrimRight(epayURL, "/") + "/submit.php?" + epayBuildQuery(params)
	c.JSON(http.StatusOK, gin.H{
		"pay_url":      payURL,
		"out_trade_no": outTradeNo,
		"amount":       req.Amount,
		"credits":      credits,
	})
}

// EpayCallback handles asynchronous payment notifications from Epay.
// GET /pay/epay/callback  （Epay 回调，无需用户认证）
func EpayCallback(c *gin.Context) {
	params := make(map[string]string)
	for k, vs := range c.Request.URL.Query() {
		if len(vs) > 0 {
			params[k] = vs[0]
		}
	}

	epayKey := getSettingValue("epay_key")

	// 验证签名
	receivedSign := params["sign"]
	delete(params, "sign")
	delete(params, "sign_type")
	if epaySign(params, epayKey) != receivedSign {
		c.String(http.StatusOK, "fail")
		return
	}

	if params["trade_status"] != "TRADE_SUCCESS" {
		c.String(http.StatusOK, "success") // 非成功状态忽略，不重试
		return
	}

	outTradeNo := params["out_trade_no"]
	tradeNo := params["trade_no"]

	// 幂等：查找订单
	order := &model.PaymentOrder{}
	found, err := db.Engine.Where("out_trade_no = ?", outTradeNo).Get(order)
	if err != nil || !found {
		c.String(http.StatusOK, "fail")
		return
	}
	if order.Status == "paid" {
		c.String(http.StatusOK, "success") // 已处理，幂等返回
		return
	}

	// 更新订单状态
	now := time.Now()
	order.Status = "paid"
	order.TradeNo = tradeNo
	order.PaidAt = &now
	if _, err := db.Engine.ID(order.ID).Cols("status", "trade_no", "paid_at").Update(order); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	// 给用户充值
	ctx := context.Background()
	if err := service.Recharge(ctx, order.UserID, 0, order.Credits); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	c.String(http.StatusOK, "success")
}

// GetUserPaymentOrders returns the authenticated user's payment orders (paginated).
// GET /user/payment-orders
func GetUserPaymentOrders(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	var page, size int
	if err := c.ShouldBindQuery(&struct {
		Page int `form:"page"`
		Size int `form:"size"`
	}{}); err != nil {
		page, size = 1, 20
	} else {
		page = 1
		size = 20
	}
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if s := c.Query("size"); s != "" {
		fmt.Sscanf(s, "%d", &size)
	}
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	var orders []model.PaymentOrder
	total, err := db.Engine.
		Where("user_id = ?", userID).
		OrderBy("created_at DESC").
		Limit(size, (page-1)*size).
		FindAndCount(&orders)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
		"total":  total,
	})
}

// epaySign generates the MD5 signature for Epay parameters.
func epaySign(params map[string]string, key string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := params[k]
		if v == "" {
			continue
		}
		parts = append(parts, k+"="+v)
	}

	raw := strings.Join(parts, "&") + key
	sum := md5.Sum([]byte(raw))
	return fmt.Sprintf("%x", sum)
}

// epayBuildQuery assembles a URL query string preserving original param values.
func epayBuildQuery(params map[string]string) string {
	v := url.Values{}
	for k, val := range params {
		v.Set(k, val)
	}
	return v.Encode()
}

// ─── 中台支付（PayApply）接口 ───────────────────────────────────────────────

type payApplyCreateReq struct {
	Amount  float64 `json:"amount" binding:"required,min=0.01"`    // 充值金额（元）
	PayFlat int     `json:"pay_flat" binding:"required,oneof=1 2"` // 1=微信 2=支付宝
	PayFrom string  `json:"pay_from"`                              // 支付终端：pc / wap / wapwx 等
	ProName string  `json:"pro_name"`                              // 商品名称（可选，默认"余额充值"）
}

// CreatePayApplyOrder 创建中台支付订单并返回支付链接。
// POST /pay/apply/create （需要 JWT 认证）
func CreatePayApplyOrder(c *gin.Context) {
	var req payApplyCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	applyURLRoot := getSettingValue("pay_apply_urlroot")
	applyKey := getSettingValue("pay_apply_key")
	if applyURLRoot == "" || applyKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "支付功能未配置，请联系管理员"})
		return
	}

	proName := req.ProName
	if proName == "" {
		proName = "余额充值"
	}

	userID := c.MustGet("user_id").(int64)

	// 生成本系统订单号（对齐 Python：时间戳 + 4位随机）
	tradeNo := fmt.Sprintf("FAN%s%04d",
		time.Now().Format("20060102150405"),
		rand.Intn(10000),
	)
	payMoneyFen := int64(req.Amount * 100) // 转换为分
	credits := int64(req.Amount * creditsPerYuan)

	// 今日0点时间戳（幂等去重：同用户同金额同产品已有 pending 订单则复用）
	now := time.Now()
	zeroTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	existOrder := &model.PaymentOrder{}
	found, _ := db.Engine.
		Where("user_id = ? AND amount = ? AND pro_name = ? AND pay_flat = ? AND status = 'pending' AND created_at >= ?",
			userID, req.Amount, proName, req.PayFlat, zeroTime).
		Get(existOrder)

	var outTradeNo string
	var orderID int64
	if found {
		outTradeNo = existOrder.OutTradeNo
		orderID = existOrder.ID
		// 更新 pay_from
		db.Engine.ID(orderID).Cols("pay_from").Update(&model.PaymentOrder{PayFrom: req.PayFrom}) //nolint
	} else {
		order := &model.PaymentOrder{
			UserID:     userID,
			OutTradeNo: tradeNo,
			Amount:     req.Amount,
			Credits:    credits,
			Status:     "pending",
			PayFlat:    req.PayFlat,
			PayFrom:    req.PayFrom,
			ProName:    proName,
		}
		id, err := db.Engine.Insert(order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建订单失败: " + err.Error()})
			return
		}
		outTradeNo = tradeNo
		orderID = id
	}

	// 获取客户端 IP
	ip := c.GetHeader("X-Forwarded-For")
	if idx := strings.Index(ip, ","); idx != -1 {
		ip = ip[:idx]
	}
	ip = strings.TrimSpace(ip)
	if ip == "" {
		ip = c.ClientIP()
	}

	// 调用中台获取支付链接
	applyURL := strings.TrimRight(applyURLRoot, "/") + "/api/pay/apply/"
	payload := map[string]interface{}{
		"pro_key":     applyKey,
		"trade_no":    outTradeNo,
		"pro_name":    proName,
		"pay_money":   payMoneyFen,
		"pay_flat":    req.PayFlat,
		"pro_user_id": userID,
		"ip":          ip,
		"pay_from":    req.PayFrom,
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(applyURL, "application/json", bytes.NewReader(body)) //nolint
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "调用支付中台失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var ret map[string]interface{}
	if err := json.Unmarshal(respBody, &ret); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "支付中台响应解析失败"})
		return
	}

	payURL := ""
	if data, ok := ret["data"].(map[string]interface{}); ok {
		payURL, _ = data["pay_url"].(string)
	}

	c.JSON(http.StatusOK, gin.H{
		"pay_url":      payURL,
		"out_trade_no": outTradeNo,
		"order_id":     orderID,
		"amount":       req.Amount,
		"credits":      credits,
	})
}

type payApplyNotifyReq struct {
	ProKey   string `json:"pro_key"`
	TradeNo  string `json:"trade_no"`
	AlipayNo string `json:"alipay_no"` // 三方平台流水号
	PayMoney int64  `json:"pay_money"` // 支付金额（分）
	PayFlat  int    `json:"pay_flat"`  // 1=微信 2=支付宝
	UserID   int64  `json:"user_id"`
}

// PayApplyNotify 中台支付回调接口（中台异步通知，无需用户认证）。
// POST /pay/apply/notify
func PayApplyNotify(c *gin.Context) {
	var req payApplyNotifyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "msg": "参数解析失败"})
		return
	}

	if req.ProKey == "" {
		c.JSON(http.StatusOK, gin.H{"status": false, "msg": "请填写商品key"})
		return
	}
	if req.TradeNo == "" {
		c.JSON(http.StatusOK, gin.H{"status": false, "msg": "请填写订单号"})
		return
	}
	if req.AlipayNo == "" {
		c.JSON(http.StatusOK, gin.H{"status": false, "msg": "请填写三方订单号"})
		return
	}
	if req.PayMoney <= 0 {
		c.JSON(http.StatusOK, gin.H{"status": false, "msg": "请填写支付金额"})
		return
	}
	if req.PayFlat <= 0 {
		c.JSON(http.StatusOK, gin.H{"status": false, "msg": "请填写支付平台"})
		return
	}

	// 校验 pro_key
	applyKey := getSettingValue("pay_apply_key")
	if req.ProKey != applyKey {
		c.JSON(http.StatusOK, gin.H{"status": false, "msg": "商品key无效"})
		return
	}

	// 查找订单
	order := &model.PaymentOrder{}
	found, err := db.Engine.Where("out_trade_no = ?", req.TradeNo).Get(order)
	if err != nil || !found {
		c.JSON(http.StatusOK, gin.H{"status": false, "msg": "订单不存在"})
		return
	}

	// 幂等：已处理则直接返回成功
	if order.Status == "paid" {
		c.JSON(http.StatusOK, gin.H{"status": true, "msg": "already processed"})
		return
	}

	// 校验金额（允许±1分误差应对浮点）
	expectedFen := int64(order.Amount * 100)
	if req.PayMoney != expectedFen {
		c.JSON(http.StatusOK, gin.H{"status": false, "msg": fmt.Sprintf("金额不匹配: expected %d, got %d", expectedFen, req.PayMoney)})
		return
	}

	// 更新订单状态
	paidAt := time.Now()
	_, err = db.Engine.ID(order.ID).Cols("status", "trade_no", "pay_flat", "paid_at").Update(&model.PaymentOrder{
		Status:  "paid",
		TradeNo: req.AlipayNo,
		PayFlat: req.PayFlat,
		PaidAt:  &paidAt,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "msg": "更新订单失败: " + err.Error()})
		return
	}

	// 给用户充值
	if err := service.Recharge(context.Background(), order.UserID, 0, order.Credits); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "msg": "充值失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "msg": "处理成功"})
}
