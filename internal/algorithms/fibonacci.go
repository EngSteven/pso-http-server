package algorithms

import 	(
	"time"
	"encoding/json"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)


// Función independiente que calcula la serie Fibonacci y devuelve el JSON
func CalculateFibonacci(n int, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()
	series := make([]int, n)
	if n > 0 {
		series[0] = 0
	}
	if n > 1 {
		series[1] = 1
		for i := 2; i < n; i++ {
			select {
			case <-cancelCh:
				return server.NewResponse(499, "Client Closed Request", "application/json", 
					[]byte(`{"error":"calculation cancelled"}`))
			default:
			}
			series[i] = series[i-1] + series[i-2]
		}
	}
	data, _ := json.MarshalIndent(map[string]interface{}{
		"n": n, "series": series, "elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")
	return server.NewResponse(200, "OK", "application/json", data)
}
