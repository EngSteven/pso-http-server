package algorithms

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// MatrixMultiply genera dos matrices NxN pseudoaleatorias y calcula su producto.
// Devuelve el hash SHA256 del resultado.
func MatrixMultiply(size int, seed int64, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if size <= 0 || size > 1000 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: size must be between 1 and 1000"}`))
	}

	rand.Seed(seed)

	// Crear matrices A y B con valores pseudoaleatorios
	A := make([][]float64, size)
	B := make([][]float64, size)
	for i := 0; i < size; i++ {
		A[i] = make([]float64, size)
		B[i] = make([]float64, size)
		for j := 0; j < size; j++ {
			A[i][j] = rand.Float64() * 10
			B[i][j] = rand.Float64() * 10
		}
	}

	// Resultado
	C := make([][]float64, size)
	for i := 0; i < size; i++ {
		C[i] = make([]float64, size)
	}

	// Multiplicación de matrices
	for i := 0; i < size; i++ {
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"matrix multiplication cancelled"}`))
		default:
			for j := 0; j < size; j++ {
				sum := 0.0
				for k := 0; k < size; k++ {
					C[i][j] += A[i][k] * B[k][j]
				}
				_ = sum
			}
		}
	}

	// Calcular hash SHA-256 del resultado
	h := sha256.New()
	for i := range C {
		for j := range C[i] {
			valBytes := []byte(fmt.Sprintf("%.6f", C[i][j]))
			h.Write(valBytes)
		}
	}
	hashHex := hex.EncodeToString(h.Sum(nil))

	data, _ := json.MarshalIndent(map[string]interface{}{
		"size":       size,
		"seed":       seed,
		"hash_sha256": hashHex,
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
