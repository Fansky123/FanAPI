package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fanapi/internal/db"
	"fanapi/internal/model"
)

// CreateOrUpdateOcpcRecord 记录用户广告来源。若该用户已有记录则忽略（先到先得）。
// platformID 为 ocpc_platforms.id（0 表示未绑定具体账户）。
func CreateOrUpdateOcpcRecord(ctx context.Context, userID, platformID int64, bdVid, qhClickID, sourceID, ip, ua string) {
	if bdVid == "" && qhClickID == "" && sourceID == "" {
		return // 无追踪参数，不记录
	}
	exists, _ := db.Engine.Where("user_id = ?", userID).Exist(&model.OcpcRecord{})
	if exists {
		return // 用户已有追踪记录，不覆盖（保留最早的广告来源）
	}
	rec := &model.OcpcRecord{
		UserID:     userID,
		PlatformID: platformID,
		BdVid:      bdVid,
		QhClickID:  qhClickID,
		SourceID:   sourceID,
		IP:         ip,
		UA:         ua,
		AddTime:    time.Now().Unix(),
	}
	if _, err := db.Engine.Insert(rec); err != nil {
		log.Printf("[ocpc] insert record error user_id=%d: %v", userID, err)
	}
}

// MarkOcpcOrder 在用户付款成功后记录订单金额，供后续 OCPC 上报使用。
func MarkOcpcOrder(ctx context.Context, userID int64, amountYuan float64) {
	rec := &model.OcpcRecord{}
	found, _ := db.Engine.Where("user_id = ?", userID).Get(rec)
	if !found {
		return
	}
	if rec.OrderAmount > 0 && !rec.OrderIsUploaded {
		db.Engine.Where("user_id = ?", userID).Cols("order_amount").
			Update(&model.OcpcRecord{OrderAmount: rec.OrderAmount + amountYuan})
		return
	}
	if rec.OrderIsUploaded {
		db.Engine.Where("user_id = ?", userID).
			Cols("order_amount", "order_is_uploaded", "order_uploaded_at", "order_ret_json").
			Update(&model.OcpcRecord{OrderAmount: amountYuan})
		return
	}
	db.Engine.Where("user_id = ?", userID).Cols("order_amount").
		Update(&model.OcpcRecord{OrderAmount: amountYuan})
}

// UploadOcpcConversions 将所有待上报的转化数据推送到对应的广告平台。
// 按各记录绑定的 platform_id 路由到具体账户配置。
func UploadOcpcConversions(ctx context.Context) (regOK, regFail, orderOK, orderFail int) {
	var records []model.OcpcRecord
	if err := db.Engine.
		Where("(bd_vid != '' OR qh_click_id != '' OR source_id != '')").
		Find(&records); err != nil {
		log.Printf("[ocpc] query records error: %v", err)
		return
	}

	// 预加载所有启用的平台配置
	platformCache := map[int64]*model.OcpcPlatform{}
	loadPlatform := func(id int64) *model.OcpcPlatform {
		if id == 0 {
			return nil
		}
		if p, ok := platformCache[id]; ok {
			return p
		}
		p := &model.OcpcPlatform{}
		found, _ := db.Engine.Where("id = ? AND enabled = true", id).Get(p)
		if !found {
			platformCache[id] = nil
			return nil
		}
		platformCache[id] = p
		return p
	}

	for i := range records {
		rec := &records[i]
		plat := loadPlatform(rec.PlatformID)

		// ── 注册转化 ────────────────────────────────────────────────
		if !rec.RegIsUploaded {
			uploaded, retJSON := uploadRecord(rec, plat, "register", 0)
			if uploaded {
				now := time.Now()
				db.Engine.Where("id = ?", rec.ID).
					Cols("reg_is_uploaded", "reg_uploaded_at", "reg_ret_json").
					Update(&model.OcpcRecord{RegIsUploaded: true, RegUploadedAt: &now, RegRetJSON: retJSON})
				regOK++
			} else if plat != nil {
				regFail++
			}
		}

		// ── 订单转化 ────────────────────────────────────────────────
		if rec.OrderAmount > 0 && !rec.OrderIsUploaded {
			uploaded, retJSON := uploadRecord(rec, plat, "order", rec.OrderAmount)
			if uploaded {
				now := time.Now()
				db.Engine.Where("id = ?", rec.ID).
					Cols("order_is_uploaded", "order_uploaded_at", "order_ret_json").
					Update(&model.OcpcRecord{OrderIsUploaded: true, OrderUploadedAt: &now, OrderRetJSON: retJSON})
				orderOK++
			} else if plat != nil {
				orderFail++
			}
		}
	}
	return
}

// uploadRecord 根据平台配置上报单条记录，返回是否成功及 JSON 摘要。
func uploadRecord(rec *model.OcpcRecord, plat *model.OcpcPlatform, convertType string, amountYuan float64) (bool, string) {
	if plat == nil {
		return false, `{"error":"no platform config"}`
	}
	switch plat.Platform {
	case "baidu":
		if rec.BdVid == "" {
			return false, `{"error":"no bd_vid"}`
		}
		ok, rj := uploadBaiduOcpc(rec, plat, convertType, amountYuan)
		return ok, mergeRetJSON("", "baidu", rj)
	case "360":
		if rec.QhClickID == "" && rec.SourceID == "" {
			return false, `{"error":"no 360 click id"}`
		}
		ok, rj := upload360Ocpc(rec, plat, convertType, amountYuan)
		return ok, mergeRetJSON("", "360", rj)
	default:
		return false, fmt.Sprintf(`{"error":"unknown platform %s"}`, plat.Platform)
	}
}

// ── 百度 OCPC ────────────────────────────────────────────────────────────────

func uploadBaiduOcpc(rec *model.OcpcRecord, plat *model.OcpcPlatform, convertType string, amountYuan float64) (bool, string) {
	if plat.BaiduToken == "" || plat.BaiduPageURL == "" {
		return false, `{"error":"baidu config missing"}`
	}

	regType := plat.BaiduRegType
	if regType == 0 {
		regType = 68 // 默认：注册/关注
	}
	orderType := plat.BaiduOrderType
	if orderType == 0 {
		orderType = 10 // 默认：购买
	}
	newType := regType
	if convertType == "order" {
		newType = orderType
	}

	logidURL := fmt.Sprintf("%s?bd_vid=%s", strings.TrimRight(plat.BaiduPageURL, "/"), rec.BdVid)
	item := map[string]interface{}{
		"logidUrl": logidURL,
		"newType":  newType,
	}
	if convertType == "order" && amountYuan > 0 {
		item["convertValue"] = int64(amountYuan * 100) // 百度单位：分
	}

	payload := map[string]interface{}{
		"token":           plat.BaiduToken,
		"conversionTypes": []interface{}{item},
	}
	body, _ := json.Marshal(payload)

	const apiURL = "https://ocpc.baidu.com/ocpcapi/api/uploadConvertData"
	for i := 0; i < 3; i++ {
		resp, err := http.Post(apiURL, "application/json; charset=UTF-8", bytes.NewReader(body)) //nolint:noctx
		if err != nil {
			log.Printf("[ocpc/baidu] post error (attempt %d): %v", i+1, err)
			continue
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		var ret map[string]interface{}
		if err := json.Unmarshal(respBody, &ret); err != nil {
			continue
		}
		header, _ := ret["header"].(map[string]interface{})
		if header == nil {
			continue
		}
		status, _ := header["status"].(float64)
		if int(status) == 4 {
			continue // 服务端异常，重试
		}
		retStr, _ := json.Marshal(header)
		return int(status) == 0, string(retStr)
	}
	return false, `{"error":"baidu upload failed after retries"}`
}

// ── 360 OCPC ─────────────────────────────────────────────────────────────────

func upload360Ocpc(rec *model.OcpcRecord, plat *model.OcpcPlatform, convertType string, amountYuan float64) (bool, string) {
	if plat.E360Key == "" || plat.E360Secret == "" {
		return false, `{"error":"360 config missing"}`
	}

	soType, _ := strconv.Atoi(plat.E360SoType)
	if soType == 0 {
		soType = 1
	}

	clickID := rec.QhClickID
	if clickID == "" {
		clickID = rec.SourceID
	}
	if clickID == "" {
		return false, `{"error":"no 360 click id"}`
	}

	nowUnix := time.Now().Unix()
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%d", nowUnix)))
	transID := fmt.Sprintf("%x", h.Sum(nil))

	eventStr := plat.E360RegEvent
	if eventStr == "" {
		eventStr = "REGISTERED"
	}
	if convertType == "order" {
		eventStr = plat.E360OrderEvent
		if eventStr == "" {
			eventStr = "ORDER"
		}
	}

	detail := map[string]interface{}{
		"qhclickid":  clickID,
		"trans_id":   transID,
		"event":      eventStr,
		"event_time": nowUnix,
	}
	dataIndustry := "ocpc_ps_convert"
	if soType == 2 {
		dataIndustry = "ocpc_zs_convert"
		detail["jzqs"] = plat.E360Jzqs
	}
	if soType == 1 && convertType == "order" && amountYuan > 0 {
		detail["event_param"] = map[string]interface{}{
			"value": int64(amountYuan * 100),
		}
	}

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"request_time":  nowUnix,
			"data_industry": dataIndustry,
			"data_detail":   detail,
		},
	}

	postData, _ := json.Marshal(payload)
	postStr := strings.ReplaceAll(string(postData), " ", "")

	sig := md5.New()
	sig.Write([]byte(plat.E360Secret + postStr))
	sign := fmt.Sprintf("%x", sig.Sum(nil))

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost,
		"https://convert.dop.360.cn/uploadWebConvert", strings.NewReader(postStr))
	req.Header.Set("App-Key", plat.E360Key)
	req.Header.Set("App-Sign", sign)
	req.Header.Set("Content-Type", "application/json;charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var ret map[string]interface{}
	json.Unmarshal(respBody, &ret) //nolint:errcheck

	retStr, _ := json.Marshal(ret)
	errno, _ := ret["errno"].(float64)
	return int(errno) == 0, string(retStr)
}

func mergeRetJSON(existing, key, value string) string {
	merged := map[string]json.RawMessage{}
	if existing != "" {
		json.Unmarshal([]byte(existing), &merged) //nolint:errcheck
	}
	merged[key] = json.RawMessage(value)
	b, _ := json.Marshal(merged)
	return string(b)
}

// ── 定时上报调度器 ─────────────────────────────────────────────────────────────

// StartOcpcScheduler 启动 OCPC 定时上报后台协程。
// 每分钟检查一次系统设置，到达配置间隔时自动触发 UploadOcpcConversions。
func StartOcpcScheduler(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ocpcSchedulerTick(ctx)
			}
		}
	}()
}

func ocpcSchedulerTick(ctx context.Context) {
	if getOcpcSetting("ocpc_schedule_enabled") != "true" {
		return
	}
	interval, err := strconv.Atoi(getOcpcSetting("ocpc_schedule_interval"))
	if err != nil || interval < 1 {
		interval = 30
	}
	var lastRun time.Time
	if ts, err := strconv.ParseInt(getOcpcSetting("ocpc_last_run_at"), 10, 64); err == nil {
		lastRun = time.Unix(ts, 0)
	}
	if time.Since(lastRun) < time.Duration(interval)*time.Minute {
		return
	}
	// 先更新时间戳，防止并发重复触发
	upsertOcpcSetting("ocpc_last_run_at", strconv.FormatInt(time.Now().Unix(), 10))
	regOK, regFail, orderOK, orderFail := UploadOcpcConversions(ctx)
	log.Printf("[ocpc/scheduler] upload done: reg_ok=%d reg_fail=%d order_ok=%d order_fail=%d",
		regOK, regFail, orderOK, orderFail)
}

func getOcpcSetting(key string) string {
	s := &model.SystemSetting{}
	if found, _ := db.Engine.Where("key = ?", key).Get(s); found {
		return s.Value
	}
	return ""
}

func upsertOcpcSetting(key, value string) {
	s := &model.SystemSetting{}
	found, _ := db.Engine.Where("key = ?", key).Get(s)
	if found {
		db.Engine.Where("key = ?", key).Cols("value").Update(&model.SystemSetting{Value: value}) //nolint:errcheck
	} else {
		db.Engine.Insert(&model.SystemSetting{Key: key, Value: value}) //nolint:errcheck
	}
}
