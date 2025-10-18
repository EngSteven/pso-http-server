package algorithms

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// SimulateWork ejecuta una simulación que "ocupa" el CPU/worker por cierto tiempo.
func SimulateWork(seconds int, taskName string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if seconds <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: seconds must be > 0"}`))
	}

	if taskName == "" {
		taskName = "generic"
	}

	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"simulation cancelled before start"}`))
	default:
	}

	// Simular trabajo con chequeo de cancelación
	for i := 0; i < seconds; i++ {
		select {
		case <-cancelCh:
			msg := fmt.Sprintf(`{"task":"%s","error":"simulation cancelled after %d seconds"}`, taskName, i)
			return server.NewResponse(499, "Client Closed Request", "application/json", []byte(msg))
		default:
			time.Sleep(1 * time.Second)
		}
	}

	data, _ := json.MarshalIndent(map[string]interface{}{
		"task":       taskName,
		"seconds":    seconds,
		"message":    fmt.Sprintf("simulation for task '%s' completed successfully", taskName),
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
