package service

import (
	"context"
	"fmt"

	"fanapi/internal/db"
	"fanapi/internal/model"
)

// WriteTx writes a billing transaction and syncs the user's DB balance.
func WriteTx(ctx context.Context, userID, channelID, apiKeyID int64, corrID, txType string, credits int64, metrics model.JSON) error {
	tx := &model.BillingTransaction{
		UserID:    userID,
		ChannelID: channelID,
		APIKeyID:  apiKeyID,
		CorrID:    corrID,
		Type:      txType,
		Credits:   credits,
		Metrics:   metrics,
	}
	_, err := db.Engine.Insert(tx)
	if err != nil {
		return err
	}

	// Sync DB balance: deduct for charge/hold/settle, add for refund/recharge.
	var delta int64
	switch txType {
	case "charge", "hold", "settle":
		delta = -credits
	case "refund", "recharge":
		delta = credits
	}
	if delta != 0 {
		_, err = db.Engine.Exec(
			"UPDATE users SET balance = balance + $1 WHERE id = $2",
			delta, userID,
		)
	}
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
func Recharge(ctx context.Context, userID, adminID, credits int64) error {
	_, err := db.Engine.Exec(
		"UPDATE users SET balance = balance + $1 WHERE id = $2",
		credits, userID,
	)
	if err != nil {
		return err
	}
	return WriteTx(ctx, userID, 0, 0, "", "recharge", credits, nil)
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
