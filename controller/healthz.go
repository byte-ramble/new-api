package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

// Healthz is a Kubernetes-style liveness probe.
// It only confirms the process is alive and able to serve HTTP — no external
// dependency check. Use /ready for readiness (DB / Redis health).
//
// Registered at both /health and /healthz to match common conventions
// (k8s default probes use /healthz; many LBs default to /health).
func Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": common.Version,
	})
}

// Readyz is a Kubernetes-style readiness probe.
// It checks that critical dependencies (database, and Redis when enabled) are
// reachable. Returns 503 with per-dependency detail when any check fails so
// load balancers stop routing traffic to this instance until it recovers.
//
// Notes:
//   - Database ping reuses model.PingDB() which has a built-in 10s cache, so
//     hammering /ready won't pressure the DB.
//   - Redis ping uses a 2s timeout to avoid hanging the probe.
//   - The DB check is wrapped in a recover so an early-startup probe (before
//     model.InitDB has run) reports "down" instead of crashing the process.
func Readyz(c *gin.Context) {
	checks := gin.H{}
	healthy := true

	// Database check (panic-safe — model.PingDB will deref nil model.DB if
	// the probe lands before InitDB).
	func() {
		defer func() {
			if r := recover(); r != nil {
				checks["database"] = gin.H{"status": "down", "error": "database not initialized"}
				healthy = false
			}
		}()
		if err := model.PingDB(); err != nil {
			checks["database"] = gin.H{"status": "down", "error": err.Error()}
			healthy = false
		} else {
			checks["database"] = gin.H{"status": "up"}
		}
	}()

	// Redis check (only when enabled)
	if common.RedisEnabled && common.RDB != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if _, err := common.RDB.Ping(ctx).Result(); err != nil {
			checks["redis"] = gin.H{"status": "down", "error": err.Error()}
			healthy = false
		} else {
			checks["redis"] = gin.H{"status": "up"}
		}
	}

	statusCode := http.StatusOK
	statusText := "ok"
	if !healthy {
		statusCode = http.StatusServiceUnavailable
		statusText = "unavailable"
	}
	c.JSON(statusCode, gin.H{
		"status":  statusText,
		"checks":  checks,
		"version": common.Version,
	})
}
