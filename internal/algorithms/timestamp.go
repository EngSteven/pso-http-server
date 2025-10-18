package algorithms

import (
	"encoding/json"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// GetTimestamp devuelve la fecha y hora actual en múltiples formatos.
func GetTimestamp(cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	now := time.Now()
	data, _ := json.MarshalIndent(map[string]interface{}{
		"unix":        now.Unix(),
		"unix_ms":     now.UnixMilli(),
		"iso":         now.Format(time.RFC3339),
		"local_time":  now.Format("2006-01-02 15:04:05"),
		"timezone":    now.Location().String(),
		"elapsed_ms":  time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
