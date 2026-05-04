package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// affiliate.go — HTTP endpoints for the OmniRouter multi-level commission
// program. See omnirouter-docs/operations/affiliate-program.md for the
// design + data model.
//
// Routes:
//
//   /api/affiliate/             (UserAuth)
//     GET   /overview            balance + recent activity
//     GET   /log                 paginated commission ledger
//     GET   /withdrawals         user's own withdrawal history
//     POST  /withdrawal          submit a new withdrawal request
//
//   /api/affiliate/admin/       (AdminAuth)
//     GET   /withdrawals         pending queue (oldest first)
//     POST  /withdrawals/:id/approve
//     POST  /withdrawals/:id/reject

// ============================================================================
// User-facing endpoints
// ============================================================================

// GetAffiliateOverview returns the dashboard summary the user-center page
// shows on first paint:
//   - account balances (current / total earned / total withdrawn)
//   - 5 most recent commission events
//   - count of pending withdrawals (so the UI can badge "1 awaiting review")
func GetAffiliateOverview(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		common.ApiErrorMsg(c, "unauthenticated")
		return
	}

	cfg := operation_setting.GetAffiliateSetting()
	if !cfg.Enabled {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"enabled": false,
			},
		})
		return
	}

	// Account row may not exist yet (user hasn't earned anything). Treat as
	// zero-balance instead of erroring.
	var acc model.AffiliateAccount
	err := model.DB.Where("user_id = ?", userId).First(&acc).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		common.ApiError(c, err)
		return
	}

	recent, err := model.ListCommissionLogsForUser(model.DB, userId, 5, 0)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	var pendingWithdrawals int64
	if err := model.DB.Model(&model.Withdrawal{}).
		Where("user_id = ? AND status = ?", userId, model.WithdrawalStatusPending).
		Count(&pendingWithdrawals).Error; err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enabled":               true,
			"balance_rmb":           acc.BalanceRmb,
			"total_earned_rmb":      acc.TotalEarnedRmb,
			"total_withdrawn_rmb":   acc.TotalWithdrawnRmb,
			"last_earned_at":        acc.LastEarnedAt,
			"recent_commissions":    recent,
			"pending_withdrawals":   pendingWithdrawals,
			"settings": gin.H{
				"level1_rate_pct":          cfg.Level1RatePct,
				"level2_rate_pct":          cfg.Level2RatePct,
				"min_withdrawal_rmb":       cfg.MinWithdrawalRmb,
				"withdrawal_fee_rmb":       cfg.WithdrawalFeeRmb,
				"max_daily_commission_rmb": cfg.MaxDailyCommissionRmb,
			},
		},
	})
}

// GetAffiliateCommissionLog returns paginated commission events for the
// authenticated user. Caller-controlled page / page_size with sane caps.
func GetAffiliateCommissionLog(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		common.ApiErrorMsg(c, "unauthenticated")
		return
	}

	page, pageSize := parsePaging(c, 20, 100)

	total, err := model.CountCommissionLogsForUser(model.DB, userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	rows, err := model.ListCommissionLogsForUser(model.DB, userId, pageSize, (page-1)*pageSize)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items":     rows,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetAffiliateWithdrawals returns the user's own withdrawal history.
func GetAffiliateWithdrawals(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		common.ApiErrorMsg(c, "unauthenticated")
		return
	}
	page, pageSize := parsePaging(c, 20, 100)

	var total int64
	if err := model.DB.Model(&model.Withdrawal{}).Where("user_id = ?", userId).Count(&total).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	rows, err := model.ListWithdrawalsForUser(model.DB, userId, pageSize, (page-1)*pageSize)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items":     rows,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

type withdrawalRequest struct {
	AmountRmb float64 `json:"amount_rmb" binding:"required"`
	Method    string  `json:"method" binding:"required"`
	Account   string  `json:"account" binding:"required"`
	UserNote  string  `json:"user_note"`
}

// PostAffiliateWithdrawal submits a new withdrawal request. Validation +
// balance lock happen inside service.RequestWithdrawal.
func PostAffiliateWithdrawal(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		common.ApiErrorMsg(c, "unauthenticated")
		return
	}

	var req withdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}

	w, err := service.RequestWithdrawal(userId, req.AmountRmb, req.Method, req.Account, req.UserNote)
	if err != nil {
		// Validation / insufficient-balance errors are user-readable; return 400.
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    w,
	})
}

// ============================================================================
// Admin-facing endpoints
// ============================================================================

// AdminListWithdrawals returns the pending queue (or filterable by status).
// Default status=pending, sorted oldest first for fair processing.
func AdminListWithdrawals(c *gin.Context) {
	status := c.DefaultQuery("status", model.WithdrawalStatusPending)
	page, pageSize := parsePaging(c, 50, 200)

	var total int64
	tx := model.DB.Model(&model.Withdrawal{})
	if status != "" && status != "all" {
		tx = tx.Where("status = ?", status)
	}
	if err := tx.Count(&total).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	var rows []model.Withdrawal
	q := model.DB.Model(&model.Withdrawal{})
	if status != "" && status != "all" {
		q = q.Where("status = ?", status)
	}
	if err := q.
		Order("created_at ASC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&rows).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items":     rows,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"status":    status,
		},
	})
}

type adminReviewRequest struct {
	Note string `json:"note"`
}

// AdminApproveWithdrawal flips a pending row to approved (admin marked
// the actual cash payout as done off-platform).
func AdminApproveWithdrawal(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		common.ApiErrorMsg(c, "invalid id")
		return
	}
	adminId := c.GetInt("id")
	var req adminReviewRequest
	_ = c.ShouldBindJSON(&req)
	if err := service.ApproveWithdrawal(id, adminId, req.Note); err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AdminRejectWithdrawal rejects a pending row AND refunds the locked
// balance (in one transaction inside service.RejectWithdrawal).
func AdminRejectWithdrawal(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		common.ApiErrorMsg(c, "invalid id")
		return
	}
	adminId := c.GetInt("id")
	var req adminReviewRequest
	_ = c.ShouldBindJSON(&req)
	if err := service.RejectWithdrawal(id, adminId, req.Note); err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ============================================================================
// helpers
// ============================================================================

// parsePaging extracts ?page= and ?page_size= with defaults + caps.
func parsePaging(c *gin.Context, defaultSize, maxSize int) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(defaultSize)))
	if pageSize < 1 {
		pageSize = defaultSize
	}
	if pageSize > maxSize {
		pageSize = maxSize
	}
	return
}
