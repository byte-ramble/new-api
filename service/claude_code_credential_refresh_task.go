package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"

	"github.com/bytedance/gopkg/util/gopool"
)

const (
	claudeCodeCredentialRefreshTickInterval = 10 * time.Minute
	claudeCodeCredentialRefreshThreshold    = 24 * time.Hour
	claudeCodeCredentialRefreshBatchSize    = 200
	claudeCodeCredentialRefreshTimeout      = 15 * time.Second
)

var (
	claudeCodeCredentialRefreshOnce    sync.Once
	claudeCodeCredentialRefreshRunning atomic.Bool
)

func shouldAutoRefreshClaudeCodeChannelStatus(status int) bool {
	return status == common.ChannelStatusEnabled || status == common.ChannelStatusAutoDisabled
}

func StartClaudeCodeCredentialAutoRefreshTask() {
	claudeCodeCredentialRefreshOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}

		gopool.Go(func() {
			logger.LogInfo(context.Background(), fmt.Sprintf("claude_code credential auto-refresh task started: tick=%s threshold=%s", claudeCodeCredentialRefreshTickInterval, claudeCodeCredentialRefreshThreshold))

			ticker := time.NewTicker(claudeCodeCredentialRefreshTickInterval)
			defer ticker.Stop()

			runClaudeCodeCredentialAutoRefreshOnce()
			for range ticker.C {
				runClaudeCodeCredentialAutoRefreshOnce()
			}
		})
	})
}

func runClaudeCodeCredentialAutoRefreshOnce() {
	if !claudeCodeCredentialRefreshRunning.CompareAndSwap(false, true) {
		return
	}
	defer claudeCodeCredentialRefreshRunning.Store(false)

	ctx := context.Background()
	now := time.Now()

	var refreshed int
	var scanned int

	offset := 0
	for {
		var channels []*model.Channel
		err := model.DB.
			Select("id", "name", "key", "status", "channel_info").
			Where("type = ? AND (status = ? OR status = ?)",
				constant.ChannelTypeClaudeCode,
				common.ChannelStatusEnabled,
				common.ChannelStatusAutoDisabled,
			).
			Order("id asc").
			Limit(claudeCodeCredentialRefreshBatchSize).
			Offset(offset).
			Find(&channels).Error
		if err != nil {
			logger.LogError(ctx, fmt.Sprintf("claude_code credential auto-refresh: query channels failed: %v", err))
			return
		}
		if len(channels) == 0 {
			break
		}
		offset += claudeCodeCredentialRefreshBatchSize

		for _, ch := range channels {
			if ch == nil {
				continue
			}
			scanned++
			if ch.ChannelInfo.IsMultiKey {
				continue
			}

			rawKey := strings.TrimSpace(ch.Key)
			if rawKey == "" {
				continue
			}

			oauthKey, err := parseClaudeCodeOAuthKey(rawKey)
			if err != nil {
				continue
			}

			refreshToken := strings.TrimSpace(oauthKey.RefreshToken)
			if refreshToken == "" {
				continue
			}

			expiredAtRaw := strings.TrimSpace(oauthKey.Expired)
			expiredAt, err := time.Parse(time.RFC3339, expiredAtRaw)
			if err == nil && !expiredAt.IsZero() && expiredAt.Sub(now) > claudeCodeCredentialRefreshThreshold {
				continue
			}

			refreshCtx, cancel := context.WithTimeout(ctx, claudeCodeCredentialRefreshTimeout)
			newKey, _, err := RefreshClaudeCodeChannelCredential(refreshCtx, ch.Id, ClaudeCodeCredentialRefreshOptions{ResetCaches: false})
			cancel()
			if err != nil {
				logger.LogWarn(ctx, fmt.Sprintf("claude_code credential auto-refresh: channel_id=%d name=%s refresh failed: %v", ch.Id, ch.Name, err))
				continue
			}

			refreshed++
			logger.LogInfo(ctx, fmt.Sprintf("claude_code credential auto-refresh: channel_id=%d name=%s refreshed, expires_at=%s", ch.Id, ch.Name, newKey.Expired))
		}
	}

	if refreshed > 0 {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.LogWarn(ctx, fmt.Sprintf("claude_code credential auto-refresh: InitChannelCache panic: %v", r))
				}
			}()
			model.InitChannelCache()
		}()
		ResetProxyClientCache()
	}

	if common.DebugEnabled {
		logger.LogDebug(ctx, "claude_code credential auto-refresh: scanned=%d refreshed=%d", scanned, refreshed)
	}
}
