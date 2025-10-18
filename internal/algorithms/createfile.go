package algorithms

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// CreateFile genera un archivo con contenido repetido 'repeat' veces.
func CreateFile(name, content string, repeat int, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if name == "" || content == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameters: name or content"}`))
	}
	if repeat <= 0 {
		repeat = 1
	}

	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	full := strings.Repeat(content+"\n", repeat)
	err := os.WriteFile(name, []byte(full), 0644)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"failed to create file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}

	data, _ := json.MarshalIndent(map[string]interface{}{
		"file":       name,
		"message":    "file created successfully",
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
