package algorithms

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"math"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// IsPrimeHandler selecciona el método y ejecuta la prueba de primalidad.
func IsPrime(n int64, method string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if n <= 1 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: n must be > 1"}`))
	}

	if method == "" {
		method = "trial"
	}

	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	isPrime := false
	switch method {
	case "trial":
		isPrime = trialDivision(n, cancelCh)
	case "miller":
		isPrime = millerRabin(n, 5, cancelCh)
	default:
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid method: must be 'trial' or 'miller'"}`))
	}

	data, _ := json.MarshalIndent(map[string]interface{}{
		"n":          n,
		"method":     method,
		"is_prime":   isPrime,
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}

// --- Métodos auxiliares ---

// trialDivision: prueba simple hasta √n
func trialDivision(n int64, cancelCh <-chan struct{}) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}

	limit := int64(math.Sqrt(float64(n)))
	for i := int64(3); i <= limit; i += 2 {
		select {
		case <-cancelCh:
			return false
		default:
			if n%i == 0 {
				return false
			}
		}
	}
	return true
}

// millerRabin: test probabilístico de primalidad
func millerRabin(n int64, k int, cancelCh <-chan struct{}) bool {
	if n < 2 {
		return false
	}
	if n == 2 || n == 3 {
		return true
	}
	if n%2 == 0 {
		return false
	}

	// escribir n-1 como 2^r * d
	d := n - 1
	r := 0
	for d%2 == 0 {
		d /= 2
		r++
	}

	for i := 0; i < k; i++ {
		select {
		case <-cancelCh:
			return false
		default:
			a, _ := rand.Int(rand.Reader, big.NewInt(n-4))
			a.Add(a, big.NewInt(2))
			x := new(big.Int).Exp(a, big.NewInt(d), big.NewInt(n))
			if x.Cmp(big.NewInt(1)) == 0 || x.Cmp(big.NewInt(n-1)) == 0 {
				continue
			}
			cont := false
			for j := 0; j < r-1; j++ {
				x.Exp(x, big.NewInt(2), big.NewInt(n))
				if x.Cmp(big.NewInt(n-1)) == 0 {
					cont = true
					break
				}
			}
			if !cont {
				return false
			}
		}
	}
	return true
}
