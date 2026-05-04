package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// AffiliateAccount is the per-user denormalized commission balance summary.
//
// Why denormalize: the source of truth is the CommissionLog table (every
// commission event). But /api/affiliate/overview is hot and computing
// SUM(commission) for every page load is wasteful. AffiliateAccount stores
// the running totals; CommissionLog provides audit trail.
//
// Updated atomically inside the same DB transaction that writes to
// CommissionLog (see service.PayCommission).
//
// Note: we deliberately did NOT extend the existing User.AffQuota field —
// that's measured in the same "quota" unit as user balance (1 USD =
// QuotaPerUnit quota), used for the existing simple registration bonus.
// Commission is measured in RMB (matches what the user actually paid),
// which is more meaningful for cash-out. Keeping them separate avoids
// quota/RMB conversion ambiguity in the ledger.
type AffiliateAccount struct {
	UserId         int     `json:"user_id" gorm:"primaryKey;column:user_id"`
	BalanceRmb     float64 `json:"balance_rmb" gorm:"default:0;column:balance_rmb"`
	TotalEarnedRmb float64 `json:"total_earned_rmb" gorm:"default:0;column:total_earned_rmb"`
	TotalWithdrawnRmb float64 `json:"total_withdrawn_rmb" gorm:"default:0;column:total_withdrawn_rmb"`
	LastEarnedAt   int64   `json:"last_earned_at" gorm:"column:last_earned_at"`
	CreatedAt      int64   `json:"created_at" gorm:"autoCreateTime;column:created_at"`
	UpdatedAt      int64   `json:"updated_at" gorm:"autoUpdateTime;column:updated_at"`
}

func (AffiliateAccount) TableName() string {
	return "affiliate_accounts"
}

// GetOrCreateAffiliateAccount fetches the row for a user, creating one with
// zero balance if missing. Safe to call repeatedly. Used inside transactions
// from PayCommission.
//
// `db` lets the caller pass either DB or a *gorm.DB transaction.
func GetOrCreateAffiliateAccount(db *gorm.DB, userId int) (*AffiliateAccount, error) {
	var acc AffiliateAccount
	err := db.Where("user_id = ?", userId).First(&acc).Error
	if err == nil {
		return &acc, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	acc = AffiliateAccount{
		UserId:       userId,
		LastEarnedAt: time.Now().Unix(),
	}
	if err := db.Create(&acc).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

// CreditAffiliateAccount adds the given commission amount to the user's
// balance + total_earned + bumps last_earned_at. Caller is responsible for
// running this inside the same tx as the matching CommissionLog insert.
func CreditAffiliateAccount(db *gorm.DB, userId int, amountRmb float64) error {
	if amountRmb <= 0 {
		return errors.New("commission amount must be positive")
	}
	now := time.Now().Unix()
	// gorm.Expr keeps the increment atomic; avoids the read-modify-write race
	return db.Model(&AffiliateAccount{}).
		Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"balance_rmb":      gorm.Expr("balance_rmb + ?", amountRmb),
			"total_earned_rmb": gorm.Expr("total_earned_rmb + ?", amountRmb),
			"last_earned_at":   now,
			"updated_at":       now,
		}).Error
}

// DebitAffiliateAccountForWithdrawal locks `amount` for a withdrawal request.
// Use inside a tx alongside the Withdrawal insert. Returns gorm.ErrRecordNotFound
// (or similar) when the balance is insufficient — caller should surface as
// a 4xx-style error to the user.
func DebitAffiliateAccountForWithdrawal(db *gorm.DB, userId int, amountRmb float64) error {
	if amountRmb <= 0 {
		return errors.New("withdrawal amount must be positive")
	}
	res := db.Model(&AffiliateAccount{}).
		Where("user_id = ? AND balance_rmb >= ?", userId, amountRmb).
		Updates(map[string]interface{}{
			"balance_rmb":         gorm.Expr("balance_rmb - ?", amountRmb),
			"total_withdrawn_rmb": gorm.Expr("total_withdrawn_rmb + ?", amountRmb),
			"updated_at":          time.Now().Unix(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("insufficient affiliate balance")
	}
	return nil
}

// RefundAffiliateAccount reverts a withdrawal (e.g. admin rejected). Adds
// amount back to balance, decrements total_withdrawn.
func RefundAffiliateAccount(db *gorm.DB, userId int, amountRmb float64) error {
	if amountRmb <= 0 {
		return errors.New("refund amount must be positive")
	}
	return db.Model(&AffiliateAccount{}).
		Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"balance_rmb":         gorm.Expr("balance_rmb + ?", amountRmb),
			"total_withdrawn_rmb": gorm.Expr("total_withdrawn_rmb - ?", amountRmb),
			"updated_at":          time.Now().Unix(),
		}).Error
}
