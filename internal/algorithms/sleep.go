package algorithms

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// Sleep ejecuta una pausa controlada de N segundos (simula inactividad o espera).
func Sleep(seconds int, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if seconds <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: seconds must be > 0"}`))
	}

	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"sleep cancelled before start"}`))
	default:
	}

	for i := 0; i < seconds; i++ {
		select {
		case <-cancelCh:
			msg := fmt.Sprintf(`{"error":"sleep cancelled after %d seconds"}`, i)
			return server.NewResponse(499, "Client Closed Request", "application/json", []byte(msg))
		default:
			time.Sleep(1 * time.Second)
		}
	}

	data, _ := json.MarshalIndent(map[string]interface{}{
		"seconds":    seconds,
		"message":    fmt.Sprintf("slept for %d seconds", seconds),
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
