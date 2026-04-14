package service

import (
	"context"
	"fmt"
	"strconv"

	"fanapi/internal/db"
	"fanapi/internal/model"
)

// WriteTx 写入一条计费流水并同步更新用户的 DB 余额。
// cost 为支付给上游的进价成本（若暂不记录可传 0）。
//
// DB 余额权威笪略：
//   - "hold"    ：仅插入流水记录，不动 DB（Redis 已原子扣款，不要重复扣 DB）
//   - "settle"  ：将实际费用写入 DB（Redis 已由 Charge+Refund 组合处理好）
//   - "charge"  ：直接一次性扣费（图片/视频/音频），DB 同步扣款
//   - "refund"  ：退款加回 DB
//   - "recharge"：充値加到 DB
func WriteTx(ctx context.Context, userID, channelID, apiKeyID int64, corrID, txType string, credits, cost int64, metrics model.JSON) error {
	tx := &model.BillingTransaction{
		UserID:    userID,
		ChannelID: channelID,
		APIKeyID:  apiKeyID,
		CorrID:    corrID,
		Type:      txType,
		Credits:   credits,
		Cost:      cost,
		Metrics:   metrics,
	}

	// 仅以下类型同步 DB 余额：
	// - hold    预扣时同步扣除 DB（输入 token 在请求时即可精确计算）
	// - settle  结算时扣除输出部分（或 input_from_response=true 时扣除差额）
	// - charge  直接扣除（图片/视频/音频）
	// - refund  恢复不应扣除的金额
	// - recharge 充値
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
	// "hold" 不修改 DB 余额，balance_after 保持 0（前端显示 —），
	// 避免与 settle 后的 DB 余额混淆。

	_, err := db.Engine.Insert(tx)
	return err
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
// 余额更新已在 WriteTx 内完成，请勿在此处重复更新 DB。
func Recharge(ctx context.Context, userID, adminID, credits int64) error {
	return WriteTx(ctx, userID, 0, 0, "", "recharge", credits, 0, nil)
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
