package handlers

import (
	"encoding/json"
	"os"
	"runtime"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
	"github.com/EngSteven/pso-http-server/internal/metrics"
)

var startTime = time.Now()

// Status estructura JSON del estado general del sistema
type Status struct {
	UptimeSeconds   float64                     `json:"uptime_seconds"`
	PID             int                         `json:"pid"`
	Hostname        string                      `json:"hostname"`
	GoRoutines      int                         `json:"goroutines"`
	GoVersion       string                      `json:"go_version"`
	ConnectionsSeen int64                       `json:"connections_seen"`
	Pools           map[string]workers.PoolInfo `json:"pools"`
	Timestamp       string                      `json:"timestamp"`
}

// StatusHandler devuelve información detallada del proceso y pools activos.
func StatusHandler(req *types.Request) *types.Response {
	pools := make(map[string]workers.PoolInfo)

	// Recorre dinámicamente todos los pools registrados
	for name, pool := range workers.GetAllPools() {
		if pool != nil {
			info := pool.Info()
			pools[name] = info
		}
	}

	hostname, _ := os.Hostname()

	status := Status{
		UptimeSeconds:   time.Since(startTime).Seconds(),
		PID:             os.Getpid(),
		Hostname:        hostname,
		GoRoutines:      runtime.NumGoroutine(),
		GoVersion:       runtime.Version(),
		ConnectionsSeen: metrics.GetTotalConnections(),
		Pools:           pools,
		Timestamp:       time.Now().Format(time.RFC3339Nano),
	}

	body, _ := json.MarshalIndent(status, "", "  ")
	return server.NewResponse(200, "OK", "application/json", body)
}
