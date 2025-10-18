package algorithms

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// GenerateRandom genera una lista de enteros aleatorios entre [min, max].
func GenerateRandom(count, min, max int, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// Validaciones
	if count <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: count must be > 0"}`))
	}
	if min > max {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid range: min must be <= max"}`))
	}

	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	rand.Seed(time.Now().UnixNano())
	numbers := make([]int, count)
	for i := 0; i < count; i++ {
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"generation cancelled"}`))
		default:
			numbers[i] = rand.Intn(max-min+1) + min
		}
	}

	data, _ := json.MarshalIndent(map[string]interface{}{
		"count":      count,
		"min":        min,
		"max":        max,
		"numbers":    numbers,
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
