package handlers

import (
	"encoding/json"
	"os"
	"runtime"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

var startTime = time.Now()

type Status struct {
	UptimeSeconds float64 `json:"uptime_seconds"`
	PID           int     `json:"pid"`
	GoRoutines    int     `json:"goroutines"`
	GoVersion     string  `json:"go_version"`
}

// StatusHandler devuelve información básica del proceso.
func StatusHandler(req *types.Request) *types.Response {
	status := Status{
		UptimeSeconds: time.Since(startTime).Seconds(),
		PID:           os.Getpid(),
		GoRoutines:    runtime.NumGoroutine(),
		GoVersion:     runtime.Version(),
	}
	body, _ := json.MarshalIndent(status, "", "  ")
	return server.NewResponse(200, "OK", "application/json", body)
}
