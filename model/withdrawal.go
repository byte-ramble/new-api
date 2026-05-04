package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Withdrawal is a user-initiated cashout request against their affiliate
// balance. Status moves from 'pending' → 'approved' (admin marks paid out
// off-platform) or 'pending' → 'rejected' (admin refunds the locked balance).
//
// Lifecycle states:
//   pending   — user submitted, balance locked; awaiting admin
//   approved  — admin paid out; terminal success
//   rejected  — admin denied; balance refunded; terminal failure
//   reversed  — post-approval reversal (rare, e.g. payout bounced); admin tool
type Withdrawal struct {
	Id           int     `json:"id" gorm:"primaryKey"`
	UserId       int     `json:"user_id" gorm:"index;not null;column:user_id"`
	AmountRmb    float64 `json:"amount_rmb" gorm:"not null;column:amount_rmb"`
	FeeRmb       float64 `json:"fee_rmb" gorm:"default:0;column:fee_rmb"`
	NetRmb       float64 `json:"net_rmb" gorm:"default:0;column:net_rmb"` // amount - fee, what user actually receives
	Method       string  `json:"method" gorm:"type:varchar(16);not null;column:method"`     // 'alipay' | 'wechat' | 'bank' | 'balance'
	Account      string  `json:"account" gorm:"type:varchar(255);not null;column:account"`  // alipay UID / bank account / etc
	Status       string  `json:"status" gorm:"type:varchar(16);default:'pending';index;column:status"`
	UserNote     string  `json:"user_note" gorm:"type:varchar(255);column:user_note"`
	AdminNote    string  `json:"admin_note" gorm:"type:varchar(255);column:admin_note"`
	ProcessedBy  int     `json:"processed_by" gorm:"column:processed_by"` // admin user id, 0 if pending
	ProcessedAt  int64   `json:"processed_at" gorm:"column:processed_at"` // unix; 0 if pending
	CreatedAt    int64   `json:"created_at" gorm:"autoCreateTime;index;column:created_at"`
	UpdatedAt    int64   `json:"updated_at" gorm:"autoUpdateTime;column:updated_at"`
}

func (Withdrawal) TableName() string {
	return "withdrawals"
}

const (
	WithdrawalStatusPending  = "pending"
	WithdrawalStatusApproved = "approved"
	WithdrawalStatusRejected = "rejected"
	WithdrawalStatusReversed = "reversed"

	WithdrawalMethodAlipay  = "alipay"
	WithdrawalMethodWechat  = "wechat"
	WithdrawalMethodBank    = "bank"
	WithdrawalMethodBalance = "balance" // return to topup balance, no real cash payout
)

// InsertWithdrawal persists a new pending request. Caller must have already
// debited AffiliateAccount inside the same transaction.
func InsertWithdrawal(db *gorm.DB, w *Withdrawal) error {
	if w.Status == "" {
		w.Status = WithdrawalStatusPending
	}
	w.NetRmb = w.AmountRmb - w.FeeRmb
	if w.NetRmb < 0 {
		return errors.New("withdrawal fee exceeds amount")
	}
	return db.Create(w).Error
}

// GetWithdrawalById loads one row.
func GetWithdrawalById(db *gorm.DB, id int) (*Withdrawal, error) {
	var w Withdrawal
	if err := db.First(&w, id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

// MarkWithdrawalApproved transitions a pending request to approved. Idempotent
// via the status-in-where guard so concurrent admin clicks can't double-process.
func MarkWithdrawalApproved(db *gorm.DB, id int, adminId int, note string) error {
	res := db.Model(&Withdrawal{}).
		Where("id = ? AND status = ?", id, WithdrawalStatusPending).
		Updates(map[string]interface{}{
			"status":       WithdrawalStatusApproved,
			"processed_by": adminId,
			"processed_at": time.Now().Unix(),
			"admin_note":   note,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("withdrawal not in pending state")
	}
	return nil
}

// MarkWithdrawalRejected transitions to rejected. Caller must also call
// RefundAffiliateAccount inside the same transaction to return the locked
// balance.
func MarkWithdrawalRejected(db *gorm.DB, id int, adminId int, note string) error {
	res := db.Model(&Withdrawal{}).
		Where("id = ? AND status = ?", id, WithdrawalStatusPending).
		Updates(map[string]interface{}{
			"status":       WithdrawalStatusRejected,
			"processed_by": adminId,
			"processed_at": time.Now().Unix(),
			"admin_note":   note,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("withdrawal not in pending state")
	}
	return nil
}

// ListWithdrawalsForUser returns user's withdrawal history newest-first.
func ListWithdrawalsForUser(db *gorm.DB, userId, limit, offset int) ([]Withdrawal, error) {
	var rows []Withdrawal
	err := db.
		Where("user_id = ?", userId).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error
	return rows, err
}

// ListPendingWithdrawals returns all currently-pending requests for the
// admin review queue. Cap to 200 to avoid OOM on big backlogs.
func ListPendingWithdrawals(db *gorm.DB, limit int) ([]Withdrawal, error) {
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	var rows []Withdrawal
	err := db.
		Where("status = ?", WithdrawalStatusPending).
		Order("created_at ASC"). // oldest first — fairness
		Limit(limit).
		Find(&rows).Error
	return rows, err
}
