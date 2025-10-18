package handlers

import (
	"encoding/json"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// CommandMetrics representa las métricas por comando (pool)
type CommandMetrics struct {
	WorkersTotal   int     `json:"workers_total"`
	BusyWorkers    int32   `json:"busy_workers"`
	QueueLength    int     `json:"queue_length"`
	TotalProcessed int64   `json:"total_processed"`
	AvgLatencyMs   float64 `json:"avg_latency_ms"`
	P50Ms          int64   `json:"p50_ms"`
	P95Ms          int64   `json:"p95_ms"`
}

// Metrics estructura JSON del endpoint /metrics
type Metrics struct {
	Timestamp string                       `json:"timestamp"`
	Commands  map[string]CommandMetrics    `json:"commands"`
}

// MetricsHandler devuelve métricas agregadas por tipo de comando
func MetricsHandler(req *types.Request) *types.Response {
	pools := workers.GetAllPools()
	metricsData := make(map[string]CommandMetrics)

	for name, pool := range pools {
		if pool != nil {
			info := pool.Info()
			metricsData[name] = CommandMetrics{
				WorkersTotal:   info.Workers,
				BusyWorkers:    info.BusyWorkers,
				QueueLength:    info.QueueLength,
				TotalProcessed: info.TotalProcessed,
				AvgLatencyMs:   info.AvgLatencyMs,
				P50Ms:          info.P50Ms,
				P95Ms:          info.P95Ms,
			}
		}
	}

	data := Metrics{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Commands:  metricsData,
	}

	body, _ := json.MarshalIndent(data, "", "  ")
	return server.NewResponse(200, "OK", "application/json", body)
}
