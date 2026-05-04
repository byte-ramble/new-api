package service

import (
	"errors"
	"fmt"
	"math"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"gorm.io/gorm"
)

// affiliate.go — multi-level commission service.
//
// Layers ON TOP OF the existing User.AffCode / InviterId / AffQuota fields.
// We deliberately do NOT touch the legacy AffQuota path (which credits a
// fixed quota bonus on registration) — both can co-exist:
//   AffQuota:    one-time bonus for inviting a user that signs up
//   Commission:  recurring percentage of every topup that user makes
//
// The two solve different problems and target different psychologies:
//   AffQuota   = "thanks for bringing a friend, here's $1 of free credit"
//   Commission = "this person tops up ¥100/mo? You get ¥15/mo as long as
//                 they keep using us." — recurring revenue, much stickier.

// PayCommission credits L1 (and optionally L2) inviters when `userId` tops up
// `topupRmb` via `channel`. Safe to call from any topup-success site;
// no-ops cleanly when:
//   - the affiliate program is disabled
//   - the topping-up user has no inviter
//   - L1 has no inviter (for L2)
//   - the daily cap would be exceeded
//   - self-invite (same user), unless AllowSelfInvite is true
//
// Errors are returned but should be logged + swallowed by callers — the
// commission system must never block a successful topup. The caller
// should wrap the call in a recover() for safety:
//
//   defer func() { if r := recover(); r != nil { common.SysError(...) } }()
//   if err := service.PayCommission(userId, amountRmb, "stripe"); err != nil {
//       common.SysError(fmt.Sprintf("commission: %v", err))
//   }
func PayCommission(userId int, topupRmb float64, channel string) error {
	cfg := operation_setting.GetAffiliateSetting()
	if !cfg.Enabled {
		return nil
	}
	if topupRmb <= 0 {
		return nil
	}

	// Resolve the inviter chain BEFORE entering the transaction.
	//
	// Why: GetUserById grabs its own connection from model.DB, and gorm's
	// Transaction holds an exclusive connection for its callback. With the
	// SQLite test harness pinning the pool to a single connection, a nested
	// GetUserById from inside the tx callback deadlocks (and at scale on
	// MySQL/Postgres it would still serialize unnecessarily).
	user, err := model.GetUserById(userId, true)
	if err != nil {
		return fmt.Errorf("commission: load topping-up user: %w", err)
	}
	if user.InviterId == 0 {
		return nil // no inviter → no commission
	}
	if !cfg.AllowSelfInvite && user.InviterId == userId {
		return nil // pathological data — skip silently
	}

	// Pre-resolve L2 (inviter's inviter) outside the tx, same reason.
	var l2InviterId int
	if cfg.Level2RatePct > 0 {
		l1, err := model.GetUserById(user.InviterId, true)
		if err != nil {
			common.SysError(fmt.Sprintf("commission: load L1 user #%d for L2: %v", user.InviterId, err))
		} else if l1.InviterId != 0 && l1.InviterId != userId && l1.InviterId != user.InviterId {
			l2InviterId = l1.InviterId
		}
	}

	// Run all DB writes inside a single transaction so commission_log +
	// affiliate_account stay consistent.
	return model.DB.Transaction(func(tx *gorm.DB) error {
		// ── Level 1 ──────────────────────────────────────────────
		if err := payOneLevel(tx, cfg, user.InviterId, userId, topupRmb, 1, cfg.Level1RatePct, channel); err != nil {
			return err
		}
		// ── Level 2 ──────────────────────────────────────────────
		if l2InviterId == 0 {
			return nil
		}
		if err := payOneLevel(tx, cfg, l2InviterId, userId, topupRmb, 2, cfg.Level2RatePct, channel); err != nil {
			// L2 failure isn't fatal for L1; log and keep the L1 credit by
			// returning nil (otherwise the transaction rolls back both levels).
			common.SysError(fmt.Sprintf("commission: L2 payout failed (L1 still credited): %v", err))
			return nil
		}
		return nil
	})
}

// payOneLevel does the math + DB writes for a single inviter level. Caller
// must wrap in a transaction. Skips when amount rounds to zero, when daily
// cap is exceeded, or when GetOrCreateAffiliateAccount fails.
func payOneLevel(
	tx *gorm.DB,
	cfg *operation_setting.AffiliateSetting,
	earnerId, sourceUserId int,
	topupRmb float64,
	level int,
	ratePct float64,
	channel string,
) error {
	if ratePct <= 0 {
		return nil
	}
	commission := round2(topupRmb * ratePct / 100.0)
	if commission <= 0 {
		return nil
	}

	// Daily cap check — skip if this earner is already over their 24h limit.
	if cfg.MaxDailyCommissionRmb > 0 {
		paidToday, err := model.SumCommissionPaidToday(tx, earnerId)
		if err != nil {
			return fmt.Errorf("daily cap check: %w", err)
		}
		if paidToday+commission > cfg.MaxDailyCommissionRmb {
			common.SysLog(fmt.Sprintf(
				"affiliate L%d earner=%d source=%d skipped: would exceed daily cap (%.2f + %.2f > %.2f)",
				level, earnerId, sourceUserId, paidToday, commission, cfg.MaxDailyCommissionRmb,
			))
			return nil
		}
	}

	// Ensure the account row exists, then credit + log.
	if _, err := model.GetOrCreateAffiliateAccount(tx, earnerId); err != nil {
		return fmt.Errorf("get/create account: %w", err)
	}
	if err := model.CreditAffiliateAccount(tx, earnerId, commission); err != nil {
		return fmt.Errorf("credit account: %w", err)
	}
	logRow := &model.CommissionLog{
		EarnerId:       earnerId,
		SourceUserId:   sourceUserId,
		Level:          level,
		TopupAmountRmb: topupRmb,
		CommissionRmb:  commission,
		RatePct:        ratePct,
		Status:         "paid",
		TopupChannel:   channel,
	}
	if err := model.InsertCommissionLog(tx, logRow); err != nil {
		return fmt.Errorf("insert log: %w", err)
	}
	common.SysLog(fmt.Sprintf(
		"affiliate L%d paid: earner=%d source=%d topup=¥%.2f rate=%.2f%% commission=¥%.2f channel=%s",
		level, earnerId, sourceUserId, topupRmb, ratePct, commission, channel,
	))
	return nil
}

// RequestWithdrawal validates + locks balance + inserts a pending withdrawal.
// Returns the inserted Withdrawal row (or an error suitable for surfacing
// directly to the user).
func RequestWithdrawal(userId int, amountRmb float64, method, account, userNote string) (*model.Withdrawal, error) {
	cfg := operation_setting.GetAffiliateSetting()
	if !cfg.Enabled {
		return nil, errors.New("affiliate program is currently disabled")
	}
	if amountRmb <= 0 {
		return nil, errors.New("withdrawal amount must be positive")
	}
	if cfg.MinWithdrawalRmb > 0 && amountRmb < cfg.MinWithdrawalRmb {
		return nil, fmt.Errorf("minimum withdrawal is ¥%.2f", cfg.MinWithdrawalRmb)
	}
	switch method {
	case model.WithdrawalMethodAlipay,
		model.WithdrawalMethodWechat,
		model.WithdrawalMethodBank,
		model.WithdrawalMethodBalance:
		// ok
	default:
		return nil, fmt.Errorf("unsupported withdrawal method: %s", method)
	}
	if account == "" {
		return nil, errors.New("withdrawal account is required")
	}

	w := &model.Withdrawal{
		UserId:    userId,
		AmountRmb: amountRmb,
		FeeRmb:    cfg.WithdrawalFeeRmb,
		Method:    method,
		Account:   account,
		UserNote:  userNote,
	}
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := model.DebitAffiliateAccountForWithdrawal(tx, userId, amountRmb); err != nil {
			return err
		}
		return model.InsertWithdrawal(tx, w)
	})
	if err != nil {
		return nil, err
	}
	return w, nil
}

// ApproveWithdrawal flips a pending request to approved (admin marked paid
// out off-platform). Idempotent.
func ApproveWithdrawal(withdrawalId, adminId int, note string) error {
	return model.MarkWithdrawalApproved(model.DB, withdrawalId, adminId, note)
}

// RejectWithdrawal flips a pending request to rejected AND refunds the
// locked balance back to the user's affiliate account, in one transaction.
func RejectWithdrawal(withdrawalId, adminId int, note string) error {
	return model.DB.Transaction(func(tx *gorm.DB) error {
		w, err := model.GetWithdrawalById(tx, withdrawalId)
		if err != nil {
			return err
		}
		if w.Status != model.WithdrawalStatusPending {
			return errors.New("withdrawal not in pending state")
		}
		if err := model.MarkWithdrawalRejected(tx, withdrawalId, adminId, note); err != nil {
			return err
		}
		return model.RefundAffiliateAccount(tx, w.UserId, w.AmountRmb)
	})
}

// round2 rounds to 2 decimal places (cents) using math.Round (round half
// away from zero). Note that float64 cannot exactly represent every decimal
// value (1.005 is actually stored as ~1.00499999...), so callers must
// accept ¥0.01 worst-case drift for commission amounts. For larger sums
// or settlement that demands exact cents, switch to shopspring/decimal
// (already a project dep).
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
