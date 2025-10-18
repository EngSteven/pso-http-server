package algorithms

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// DeleteFile elimina un archivo existente y devuelve un resultado JSON.
func DeleteFile(name string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}

	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	if err := os.Remove(name); err != nil {
		msg := fmt.Sprintf(`{"error":"failed to delete file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}

	data, _ := json.MarshalIndent(map[string]interface{}{
		"file":       name,
		"message":    "file deleted successfully",
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
