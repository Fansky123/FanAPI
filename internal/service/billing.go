package service

import (
	"context"
	"fmt"
	"strconv"

	"fanapi/internal/db"
	"fanapi/internal/model"
)

// WriteTx writes a billing transaction and syncs the user's DB balance.
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
		// UPDATE … RETURNING balance in one round-trip; captures the resulting
		// balance atomically for the audit trail.
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

// GetBalance returns the user's current balance from DB.
func GetBalance(ctx context.Context, userID int64) (int64, error) {
	user := &model.User{}
	found, err := db.Engine.Where("id = ?", userID).Cols("balance").Get(user)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("user not found")
	}
	return user.Balance, nil
}

// Recharge adds credits to a user's balance (admin operation).
// Balance update is handled inside WriteTx; do NOT update here separately.
func Recharge(ctx context.Context, userID, adminID, credits int64) error {
	return WriteTx(ctx, userID, 0, 0, "", "recharge", credits, 0, nil)
}

// ListTransactions returns paginated billing history for a user.
func ListTransactions(ctx context.Context, userID int64, page, pageSize int) ([]model.BillingTransaction, error) {
	var txs []model.BillingTransaction
	err := db.Engine.Where("user_id = ?", userID).
		Desc("created_at").
		Limit(pageSize, (page-1)*pageSize).
		Find(&txs)
	return txs, err
}

// CountTransactions returns the total number of billing records for a user.
func CountTransactions(ctx context.Context, userID int64) (int64, error) {
	count, err := db.Engine.Where("user_id = ?", userID).Count(&model.BillingTransaction{})
	return count, err
}
