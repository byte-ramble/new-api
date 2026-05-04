package service

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withAffiliateSettings overrides the live AffiliateSetting for the duration
// of one test, restoring it via t.Cleanup. The package's existing TestMain
// (in task_billing_test.go) sets up an in-memory SQLite, so DB calls work.
func withAffiliateSettings(t *testing.T, mod func(*operation_setting.AffiliateSetting)) {
	t.Helper()
	cfg := operation_setting.GetAffiliateSetting()
	prev := *cfg
	mod(cfg)
	t.Cleanup(func() {
		cfg.Enabled = prev.Enabled
		cfg.Level1RatePct = prev.Level1RatePct
		cfg.Level2RatePct = prev.Level2RatePct
		cfg.MaxDailyCommissionRmb = prev.MaxDailyCommissionRmb
		cfg.MinWithdrawalRmb = prev.MinWithdrawalRmb
		cfg.AllowSelfInvite = prev.AllowSelfInvite
	})
}

// migrateAffiliateTablesForTest is a no-op placeholder if the package's
// TestMain doesn't already migrate our 3 new tables. Calling AutoMigrate
// repeatedly is safe.
func migrateAffiliateTablesForTest(t *testing.T) {
	t.Helper()
	require.NoError(t, model.DB.AutoMigrate(
		&model.User{},
		&model.AffiliateAccount{},
		&model.CommissionLog{},
		&model.Withdrawal{},
	))
}

func TestRound2(t *testing.T) {
	// We deliberately exclude classic float-imprecise cases (1.005, 0.1+0.2,
	// etc.) — those are float64 storage quirks, not rounding bugs. The
	// guarantee we care about: amounts that ARE exactly representable round
	// to 2 decimals.
	cases := []struct{ in, want float64 }{
		{0, 0},
		{1, 1},
		{1.004, 1.00}, // truly < 1.005
		{1.006, 1.01},
		{15.0 / 100 * 100, 15.0},
		{0.001, 0.00},
		{99.999, 100.00},
	}
	for _, c := range cases {
		got := round2(c.in)
		assert.InDelta(t, c.want, got, 0.0001, "round2(%v)", c.in)
	}
}

func TestPayCommission_Disabled(t *testing.T) {
	migrateAffiliateTablesForTest(t)
	withAffiliateSettings(t, func(s *operation_setting.AffiliateSetting) { s.Enabled = false })
	err := PayCommission(1, 100, "test")
	assert.NoError(t, err, "disabled program should silently no-op")
}

func TestPayCommission_NoInviter(t *testing.T) {
	migrateAffiliateTablesForTest(t)
	withAffiliateSettings(t, func(s *operation_setting.AffiliateSetting) {
		s.Enabled = true
		s.Level1RatePct = 15
	})
	// Create a user with NO inviter
	u := &model.User{Username: "noinviter_user", Password: "x", Email: "n@x.com", Status: 1, AffCode: "AFFN1"}
	require.NoError(t, model.DB.Create(u).Error)
	t.Cleanup(func() { model.DB.Unscoped().Delete(u) })

	err := PayCommission(u.Id, 100, "test")
	assert.NoError(t, err)

	var n int64
	model.DB.Model(&model.CommissionLog{}).Where("source_user_id = ?", u.Id).Count(&n)
	assert.Equal(t, int64(0), n, "no commission row should be created")
}

func TestPayCommission_L1AndL2HappyPath(t *testing.T) {
	migrateAffiliateTablesForTest(t)
	withAffiliateSettings(t, func(s *operation_setting.AffiliateSetting) {
		s.Enabled = true
		s.Level1RatePct = 15
		s.Level2RatePct = 5
		s.MaxDailyCommissionRmb = 0 // unlimited
	})
	// Build chain: grandInviter → inviter → invitee
	grand := &model.User{Username: "grandparent_aff", Password: "x", Email: "g@x.com", Status: 1, AffCode: "AFFG1"}
	require.NoError(t, model.DB.Create(grand).Error)
	parent := &model.User{Username: "parent_aff", Password: "x", Email: "p@x.com", Status: 1, InviterId: grand.Id, AffCode: "AFFP1"}
	require.NoError(t, model.DB.Create(parent).Error)
	child := &model.User{Username: "child_aff", Password: "x", Email: "c@x.com", Status: 1, InviterId: parent.Id, AffCode: "AFFC1"}
	require.NoError(t, model.DB.Create(child).Error)
	t.Cleanup(func() {
		model.DB.Unscoped().Delete(child)
		model.DB.Unscoped().Delete(parent)
		model.DB.Unscoped().Delete(grand)
		model.DB.Where("source_user_id = ?", child.Id).Delete(&model.CommissionLog{})
		model.DB.Where("user_id IN ?", []int{grand.Id, parent.Id}).Delete(&model.AffiliateAccount{})
	})

	// child tops up ¥100 → parent gets ¥15 (L1), grand gets ¥5 (L2)
	require.NoError(t, PayCommission(child.Id, 100, "test"))

	var parentAcc model.AffiliateAccount
	require.NoError(t, model.DB.Where("user_id = ?", parent.Id).First(&parentAcc).Error)
	assert.InDelta(t, 15.00, parentAcc.BalanceRmb, 0.001, "L1 balance")
	assert.InDelta(t, 15.00, parentAcc.TotalEarnedRmb, 0.001, "L1 total earned")

	var grandAcc model.AffiliateAccount
	require.NoError(t, model.DB.Where("user_id = ?", grand.Id).First(&grandAcc).Error)
	assert.InDelta(t, 5.00, grandAcc.BalanceRmb, 0.001, "L2 balance")

	// Two commission_log rows
	var logs []model.CommissionLog
	require.NoError(t, model.DB.Where("source_user_id = ?", child.Id).Order("level ASC").Find(&logs).Error)
	require.Len(t, logs, 2)
	assert.Equal(t, 1, logs[0].Level)
	assert.Equal(t, parent.Id, logs[0].EarnerId)
	assert.InDelta(t, 15.0, logs[0].CommissionRmb, 0.001)
	assert.Equal(t, 2, logs[1].Level)
	assert.Equal(t, grand.Id, logs[1].EarnerId)
	assert.InDelta(t, 5.0, logs[1].CommissionRmb, 0.001)
}

func TestPayCommission_DailyCap(t *testing.T) {
	migrateAffiliateTablesForTest(t)
	withAffiliateSettings(t, func(s *operation_setting.AffiliateSetting) {
		s.Enabled = true
		s.Level1RatePct = 15
		s.Level2RatePct = 0 // single level — easier to reason about cap
		s.MaxDailyCommissionRmb = 20.0
	})
	parent := &model.User{Username: "cap_parent", Password: "x", Email: "cap_p@x.com", Status: 1, AffCode: "AFFCP1"}
	require.NoError(t, model.DB.Create(parent).Error)
	child := &model.User{Username: "cap_child", Password: "x", Email: "cap_c@x.com", Status: 1, InviterId: parent.Id, AffCode: "AFFCC1"}
	require.NoError(t, model.DB.Create(child).Error)
	t.Cleanup(func() {
		model.DB.Unscoped().Delete(child)
		model.DB.Unscoped().Delete(parent)
		model.DB.Where("source_user_id = ?", child.Id).Delete(&model.CommissionLog{})
		model.DB.Where("user_id = ?", parent.Id).Delete(&model.AffiliateAccount{})
	})

	// First topup: ¥100 × 15% = ¥15 → ok (under cap)
	require.NoError(t, PayCommission(child.Id, 100, "test"))
	// Second topup: would add ¥15 → total ¥30 > cap ¥20 → silently skipped
	require.NoError(t, PayCommission(child.Id, 100, "test"))

	var acc model.AffiliateAccount
	require.NoError(t, model.DB.Where("user_id = ?", parent.Id).First(&acc).Error)
	assert.InDelta(t, 15.0, acc.BalanceRmb, 0.001, "cap should have skipped second event")

	var n int64
	model.DB.Model(&model.CommissionLog{}).Where("earner_id = ?", parent.Id).Count(&n)
	assert.Equal(t, int64(1), n, "only first event should be logged")
}

func TestRequestWithdrawal_Validation(t *testing.T) {
	migrateAffiliateTablesForTest(t)
	withAffiliateSettings(t, func(s *operation_setting.AffiliateSetting) {
		s.Enabled = true
		s.MinWithdrawalRmb = 100
	})

	_, err := RequestWithdrawal(1, 0, "alipay", "x", "")
	assert.ErrorContains(t, err, "positive")

	_, err = RequestWithdrawal(1, 50, "alipay", "x", "") // < min
	assert.ErrorContains(t, err, "minimum")

	_, err = RequestWithdrawal(1, 200, "wonky", "x", "")
	assert.ErrorContains(t, err, "unsupported")

	_, err = RequestWithdrawal(1, 200, "alipay", "", "") // empty account
	assert.ErrorContains(t, err, "account is required")
}

func TestRequestWithdrawal_HappyPathLocksBalance(t *testing.T) {
	migrateAffiliateTablesForTest(t)
	withAffiliateSettings(t, func(s *operation_setting.AffiliateSetting) {
		s.Enabled = true
		s.MinWithdrawalRmb = 50
		s.WithdrawalFeeRmb = 2.0
	})

	// Seed: an account with ¥500 balance
	u := &model.User{Username: "wd_user", Password: "x", Email: "wd@x.com", Status: 1, AffCode: "AFFWD1"}
	require.NoError(t, model.DB.Create(u).Error)
	require.NoError(t, model.DB.Create(&model.AffiliateAccount{UserId: u.Id, BalanceRmb: 500, TotalEarnedRmb: 500}).Error)
	t.Cleanup(func() {
		model.DB.Unscoped().Delete(u)
		model.DB.Where("user_id = ?", u.Id).Delete(&model.AffiliateAccount{})
		model.DB.Where("user_id = ?", u.Id).Delete(&model.Withdrawal{})
	})

	w, err := RequestWithdrawal(u.Id, 200, "alipay", "test@alipay", "first cashout")
	require.NoError(t, err)
	assert.Equal(t, model.WithdrawalStatusPending, w.Status)
	assert.InDelta(t, 200.0, w.AmountRmb, 0.001)
	assert.InDelta(t, 2.0, w.FeeRmb, 0.001)
	assert.InDelta(t, 198.0, w.NetRmb, 0.001)

	var acc model.AffiliateAccount
	require.NoError(t, model.DB.Where("user_id = ?", u.Id).First(&acc).Error)
	assert.InDelta(t, 300.0, acc.BalanceRmb, 0.001, "balance should be locked")
	assert.InDelta(t, 200.0, acc.TotalWithdrawnRmb, 0.001, "total_withdrawn bumped")

	// Reject → balance refunded
	require.NoError(t, RejectWithdrawal(w.Id, 999, "test reject"))
	require.NoError(t, model.DB.Where("user_id = ?", u.Id).First(&acc).Error)
	assert.InDelta(t, 500.0, acc.BalanceRmb, 0.001, "balance should be back to 500")
	assert.InDelta(t, 0.0, acc.TotalWithdrawnRmb, 0.001, "total_withdrawn rolled back")
}
