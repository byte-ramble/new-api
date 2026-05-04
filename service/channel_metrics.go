package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// channel_metrics.go — Prometheus counters for channel auto-disable / recover.
//
// Kept in the service package (not middleware) to avoid the
//   service → middleware → service
// import cycle introduced by middleware/auth.go pulling in service.
//
// Naming follows the omnirouter_* convention used by middleware/metrics.go.

var (
	channelAutoDisabledTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "omnirouter",
			Subsystem: "channel",
			Name:      "auto_disabled_total",
			Help:      "Channels auto-disabled by the gateway, by channel name and classified reason.",
		},
		[]string{"channel", "reason"},
	)

	channelRecoveredTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "omnirouter",
			Subsystem: "channel",
			Name:      "recovered_total",
			Help:      "Channels recovered (re-enabled) by the auto-test loop or admin action.",
		},
		[]string{"channel"},
	)
)

// RecordChannelAutoDisabled increments the auto-disable counter for a channel.
// Called from service.DisableChannel after the DB status update succeeds.
// `reason` should be one of the ChannelErrorReason constants (low cardinality).
func RecordChannelAutoDisabled(channel, reason string) {
	channelAutoDisabledTotal.WithLabelValues(channel, reason).Inc()
}

// RecordChannelRecovered increments the recovered counter for a channel.
// Called from service.EnableChannel after the DB status update succeeds.
func RecordChannelRecovered(channel string) {
	channelRecoveredTotal.WithLabelValues(channel).Inc()
}
