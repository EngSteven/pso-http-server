package handlers

import (
	"encoding/json"
	"os"
	"runtime"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

var startTime = time.Now()

type Status struct {
	UptimeSeconds float64 `json:"uptime_seconds"`
	PID           int     `json:"pid"`
	GoRoutines    int     `json:"goroutines"`
	GoVersion     string  `json:"go_version"`
	Pools map[string]workers.PoolInfo `json:"pools"`
}

// StatusHandler devuelve información básica del proceso.
func StatusHandler(req *types.Request) *types.Response {
	pools := make(map[string]workers.PoolInfo)
	if p := workers.GetPool("fibonacci"); p != nil {
		info := p.Info()
		pools["fibonacci"] = info
	}
	if p := workers.GetPool("createfile"); p != nil {
		info := p.Info()
		pools["createfile"] = info
	}

	if p := workers.GetPool("pi"); p != nil {
		info := p.Info()
		pools["pi"] = info
	}

	status := Status{
		UptimeSeconds: time.Since(startTime).Seconds(),
		PID:           os.Getpid(),
		GoRoutines:    runtime.NumGoroutine(),
		GoVersion:     runtime.Version(),
		Pools:         pools,
	}
	body, _ := json.MarshalIndent(status, "", "  ")
	return server.NewResponse(200, "OK", "application/json", body)
}
