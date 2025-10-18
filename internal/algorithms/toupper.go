package algorithms

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// ToUpper convierte texto a mayúsculas y devuelve un JSON estándar.
func ToUpper(text string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if text == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: text"}`))
	}

	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	output := strings.ToUpper(text)
	data, _ := json.MarshalIndent(map[string]interface{}{
		"input":       text,
		"output":      output,
		"elapsed_ms":  time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
