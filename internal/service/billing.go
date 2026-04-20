package service

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"fanapi/internal/billing"
	"fanapi/internal/db"
	"fanapi/internal/model"
)

// WriteTx 写入一条计费流水并同步更新用户的 DB 余额。
// poolKeyID 为本次请求使用的号池 Key ID（0 表示未使用号池）。
// cost 为支付给上游的进价成本（若暂不记录可传 0）。
//
// DB 余额权威策略：
//   - "hold"    ：仅插入流水记录，不动 DB（Redis 已原子扣款，不要重复扣 DB）
//   - "settle"  ：将实际费用写入 DB（Redis 已由 Charge+Refund 组合处理好）
//   - "charge"  ：直接一次性扣费（图片/视频/音频），DB 同步扣款
//   - "refund"  ：退款加回 DB
//   - "recharge"：充值加到 DB
func WriteTx(ctx context.Context, userID, channelID, apiKeyID, poolKeyID int64, corrID, txType string, credits, cost int64, metrics model.JSON) error {
	tx := &model.BillingTransaction{
		UserID:    userID,
		ChannelID: channelID,
		APIKeyID:  apiKeyID,
		PoolKeyID: poolKeyID,
		CorrID:    corrID,
		Type:      txType,
		Credits:   credits,
		Cost:      cost,
		Metrics:   metrics,
	}

	// 仅以下类型同步 DB 余额：
	// - hold    预扣时同步扣除 DB
	// - settle  结算时扣除输出部分
	// - charge  直接扣除（图片/视频/音频）
	// - refund  恢复不应扣除的金额
	// - recharge 充值
	var delta int64
	switch txType {
	case "charge", "settle", "hold":
		delta = -credits
	case "refund", "recharge":
		delta = credits
	}

	if delta != 0 {
		// 单条 SQL 内原子地更新并返回新余额，用于审计日志。
		rows, err := db.Engine.QueryString(
			"UPDATE users SET balance = balance + $1 WHERE id = $2 RETURNING balance",
			delta, userID,
		)
		if err != nil {
			return err
		}
		if len(rows) > 0 {
			if balStr, ok := rows[0]["balance"]; ok {
				tx.BalanceAfter, _ = strconv.ParseInt(balStr, 10, 64)
			}
		}
	}

	if _, err := db.Engine.Insert(tx); err != nil {
		return err
	}

	// 消费类交易触发邀请返佣和号商收益：
	//   hold    — 输入费预扣（input_from_response=false 时精确，=true 时为估算）
	//   settle  — 输出费或差额补扣
	//   charge  — 图片/视频/音频一次性扣费
	//   refund  — 退款时反向扣回已发放的返佣/收益（传负值）
	switch txType {
	case "charge", "settle", "hold":
		go applyPostBillingHooks(userID, poolKeyID, credits, cost)
	case "refund":
		go applyPostBillingHooks(userID, poolKeyID, -credits, -cost)
	}
	return nil
}

// applyPostBillingHooks 在消费发生后异步处理：
//  1. 邀请返佣：若用户有邀请人，按比例将 credits 加入邀请人的冻结余额
//  2. 号商收益：若本次请求使用了号商的 Key，按比例计入号商可提现余额
//
// credits/cost 可为负值（refund 场景），表示回扣已发放的返佣/收益，不会使余额低于 0。
func applyPostBillingHooks(userID, poolKeyID, credits, cost int64) {
	ctx := context.Background()

	// ── 邀请返佣 ─────────────────────────────────────────────────────────
	if credits != 0 {
		var inviterID int64
		var rebateRatio *float64
		rows, err := db.Engine.QueryString(
			"SELECT inviter_id, rebate_ratio FROM users WHERE id = $1", userID,
		)
		if err == nil && len(rows) > 0 {
			if s := rows[0]["inviter_id"]; s != "" {
				inviterID, _ = strconv.ParseInt(s, 10, 64)
			}
			if s := rows[0]["rebate_ratio"]; s != "" {
				var r float64
				if _, err2 := fmt.Sscanf(s, "%f", &r); err2 == nil {
					rebateRatio = &r
				}
			}
		}

		if inviterID > 0 {
			ratio := getRebateRatio(ctx, rebateRatio)
			if ratio > 0 {
				rebateCredits := int64(float64(credits) * ratio)
				if rebateCredits != 0 {
					var sql string
					if rebateCredits > 0 {
						sql = "UPDATE users SET frozen_balance = frozen_balance + $1 WHERE id = $2"
					} else {
						// 回扣：floor 0，不允许透支冻结余额
						sql = "UPDATE users SET frozen_balance = GREATEST(0, frozen_balance + $1) WHERE id = $2"
					}
					if _, err := db.Engine.Exec(sql, rebateCredits, inviterID); err != nil {
						log.Printf("[billing] apply inviter rebate failed user=%d inviter=%d err=%v", userID, inviterID, err)
					}
				}
			}
		}
	}

	// ── 号商收益 ──────────────────────────────────────────────────────────
	if poolKeyID > 0 && cost != 0 {
		rows, err := db.Engine.QueryString(
			"SELECT vendor_id FROM pool_keys WHERE id = $1", poolKeyID,
		)
		if err == nil && len(rows) > 0 {
			vendorIDStr := rows[0]["vendor_id"]
			if vendorIDStr != "" {
				vendorID, _ := strconv.ParseInt(vendorIDStr, 10, 64)
				if vendorID > 0 {
					commission := getVendorCommission(ctx, vendorID)
					// 号商到手 = cost * (1 - commission)；负值时回扣
					vendorEarns := int64(float64(cost) * (1 - commission))
					if vendorEarns != 0 {
						var sql string
						if vendorEarns > 0 {
							sql = "UPDATE vendors SET balance = balance + $1 WHERE id = $2"
						} else {
							// 回扣：floor 0
							sql = "UPDATE vendors SET balance = GREATEST(0, balance + $1) WHERE id = $2"
						}
						if _, err2 := db.Engine.Exec(sql, vendorEarns, vendorID); err2 != nil {
							log.Printf("[billing] apply vendor earning failed vendor=%d err=%v", vendorID, err2)
						}
					}
				}
			}
		}
	}
}

// getRebateRatio 返回有效的返佣比例：优先使用用户个人设置，否则读取系统默认值。
func getRebateRatio(ctx context.Context, userRatio *float64) float64 {
	if userRatio != nil {
		return *userRatio
	}
	s := &model.SystemSetting{}
	if found, _ := db.Engine.Where("key = ?", "default_rebate_ratio").Get(s); found && s.Value != "" {
		var r float64
		if _, err := fmt.Sscanf(s.Value, "%f", &r); err == nil {
			return r
		}
	}
	return 0
}

// getVendorCommission 返回有效的平台手续费比例：优先使用号商个人设置，否则读取系统默认值。
func getVendorCommission(ctx context.Context, vendorID int64) float64 {
	rows, _ := db.Engine.QueryString("SELECT commission_ratio FROM vendors WHERE id = $1", vendorID)
	if len(rows) > 0 && rows[0]["commission_ratio"] != "" {
		var r float64
		if _, err := fmt.Sscanf(rows[0]["commission_ratio"], "%f", &r); err == nil {
			return r
		}
	}
	s := &model.SystemSetting{}
	if found, _ := db.Engine.Where("key = ?", "default_vendor_commission").Get(s); found && s.Value != "" {
		var r float64
		if _, err := fmt.Sscanf(s.Value, "%f", &r); err == nil {
			return r
		}
	}
	return 0
}

// GetBalance 从 DB 返回用户的当前余额。
func GetBalance(ctx context.Context, userID int64) (int64, error) {
	user := &model.User{}
	found, err := db.Engine.Where("id = ?", userID).Cols("balance").Get(user)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("用户不存在")
	}
	return user.Balance, nil
}

// Recharge 为用户增加 credits（管理员操作）。
// 余额更新已在 WriteTx 内完成，但运行时扣费走 Redis，需要立刻同步缓存。
func Recharge(ctx context.Context, userID, adminID, credits int64) error {
	if err := WriteTx(ctx, userID, 0, 0, 0, "", "recharge", credits, 0, nil); err != nil {
		return err
	}
	_, err := billing.SyncBalanceToRedis(ctx, userID)
	return err
}

// ListTransactions 返回用户的分页计费历史。corrID/taskID 非空时分别按对应字段过滤。
func ListTransactions(ctx context.Context, userID int64, page, pageSize int, corrID, taskID string) ([]model.BillingTransaction, error) {
	var txs []model.BillingTransaction
	sess := db.Engine.Where("user_id = ?", userID)
	if corrID != "" {
		sess.And("corr_id = ?", corrID)
	}
	if taskID != "" {
		sess.And("metrics->>'task_id' = ?", taskID)
	}
	err := sess.Desc("created_at").
		Limit(pageSize, (page-1)*pageSize).
		Find(&txs)
	return txs, err
}

// CountTransactions 返回用户的计费记录总数。corrID/taskID 非空时分别按对应字段过滤。
func CountTransactions(ctx context.Context, userID int64, corrID, taskID string) (int64, error) {
	sess := db.Engine.Where("user_id = ?", userID)
	if corrID != "" {
		sess.And("corr_id = ?", corrID)
	}
	if taskID != "" {
		sess.And("metrics->>'task_id' = ?", taskID)
	}
	count, err := sess.Count(&model.BillingTransaction{})
	return count, err
}
