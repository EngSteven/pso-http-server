package algorithms

import (
	"encoding/json"
	"math"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// Factorize descompone un número entero n en sus factores primos.
func Factorize(n int64, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if n <= 1 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: n must be > 1"}`))
	}

	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	factors := make([]int64, 0)
	num := n

	// División por 2
	for num%2 == 0 {
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"factorization cancelled"}`))
		default:
			factors = append(factors, 2)
			num /= 2
		}
	}

	// División por impares
	for i := int64(3); i <= int64(math.Sqrt(float64(num))); i += 2 {
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"factorization cancelled"}`))
		default:
			for num%i == 0 {
				factors = append(factors, i)
				num /= i
			}
		}
	}

	// Si queda un número mayor a 2, también es factor
	if num > 2 {
		factors = append(factors, num)
	}

	data, _ := json.MarshalIndent(map[string]interface{}{
		"n":          n,
		"factors":    factors,
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
