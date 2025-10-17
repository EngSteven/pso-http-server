package algorithms

import 	(
	"encoding/json"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)


// Función independiente que calcula la serie Fibonacci y devuelve el JSON
func CalculateFibonacci(n int, cancelCh <-chan struct{}) *types.Response {
	series := make([]int, n)
	if n > 0 {
		series[0] = 0
	}
	if n > 1 {
		series[1] = 1
		for i := 2; i < n; i++ {
			select {
			case <-cancelCh:
				return server.NewResponse(500, "Canceled", "application/json", []byte(`{"error":"cancelled"}`))
			default:
			}
			series[i] = series[i-1] + series[i-2]
		}
	}
	data, _ := json.MarshalIndent(map[string]interface{}{"n": n, "series": series}, "", "  ")
	return server.NewResponse(200, "OK", "application/json", data)
}
