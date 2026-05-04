package model

import (
	"time"

	"gorm.io/gorm"
)

// CommissionLog records every individual commission event (earnings + reversals).
//
// One topup typically produces 0-2 rows: zero if the topping-up user has no
// inviter (and no L2), one for the direct inviter (L1), one for the inviter's
// inviter (L2). Refund reversals add a negative-amount row referencing the
// original log row by RefId.
//
// This is the source of truth — AffiliateAccount summaries can always be
// rebuilt from a SUM over this table.
type CommissionLog struct {
	Id              int     `json:"id" gorm:"primaryKey"`
	EarnerId        int     `json:"earner_id" gorm:"index;not null;column:earner_id"`     // user receiving the commission
	SourceUserId    int     `json:"source_user_id" gorm:"index;not null;column:source_user_id"` // user whose topup triggered this
	Level           int     `json:"level" gorm:"not null;default:1;column:level"`         // 1 = direct inviter, 2 = inviter's inviter
	TopupAmountRmb  float64 `json:"topup_amount_rmb" gorm:"not null;column:topup_amount_rmb"`
	CommissionRmb   float64 `json:"commission_rmb" gorm:"not null;column:commission_rmb"`
	RatePct         float64 `json:"rate_pct" gorm:"not null;column:rate_pct"` // rate applied at time of event
	Status          string  `json:"status" gorm:"type:varchar(16);default:'paid';column:status;index"` // 'paid' | 'frozen' | 'reversed'
	RefId           *int    `json:"ref_id,omitempty" gorm:"column:ref_id"`                // reversal pointer back to original log
	TopupChannel    string  `json:"topup_channel" gorm:"type:varchar(32);column:topup_channel"` // 'epay'|'stripe'|'creem'|'redeem'|...
	Note            string  `json:"note,omitempty" gorm:"type:varchar(255);column:note"`
	CreatedAt       int64   `json:"created_at" gorm:"autoCreateTime;index;column:created_at"`
}

func (CommissionLog) TableName() string {
	return "commission_logs"
}

// InsertCommissionLog persists a new commission event. Use inside the tx that
// also credits AffiliateAccount.
func InsertCommissionLog(db *gorm.DB, log *CommissionLog) error {
	return db.Create(log).Error
}

// SumCommissionPaidToday returns the total commission paid to `earnerId` in
// the rolling 24-hour window. Used by the daily cap anti-fraud rule.
func SumCommissionPaidToday(db *gorm.DB, earnerId int) (float64, error) {
	cutoff := time.Now().Add(-24 * time.Hour).Unix()
	var total float64
	err := db.Model(&CommissionLog{}).
		Select("COALESCE(SUM(commission_rmb), 0)").
		Where("earner_id = ? AND status = 'paid' AND created_at >= ?", earnerId, cutoff).
		Row().Scan(&total)
	return total, err
}

// CountCommissionLogsForUser returns total rows + paged slice for a user's
// "earnings history" page.
func CountCommissionLogsForUser(db *gorm.DB, earnerId int) (int64, error) {
	var n int64
	err := db.Model(&CommissionLog{}).
		Where("earner_id = ?", earnerId).
		Count(&n).Error
	return n, err
}

// ListCommissionLogsForUser returns paginated rows newest-first.
// limit + offset are caller-validated.
func ListCommissionLogsForUser(db *gorm.DB, earnerId int, limit, offset int) ([]CommissionLog, error) {
	var rows []CommissionLog
	err := db.
		Where("earner_id = ?", earnerId).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error
	return rows, err
}
