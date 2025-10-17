package handlers

import (
	"encoding/json"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

func isPrimeTrial(n int64) bool {
	if n < 2 {
		return false
	}
	if n%2 == 0 {
		return n == 2
	}
	limit := int64(math.Sqrt(float64(n)))
	for i := int64(3); i <= limit; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func millerRabin(n int64, k int) bool {
	if n < 4 {
		return n == 2 || n == 3
	}
	bigN := big.NewInt(n)
	return bigN.ProbablyPrime(k)
}

func IsPrimeHandler(req *types.Request) *types.Response {
	numStr := req.Query.Get("n")
	if numStr == "" {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"missing parameter n"}`))
	}
	n, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil || n <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"invalid n"}`))
	}

	pool := workers.GetPool("isprime")
	job := func(cancelCh <-chan struct{}) *types.Response {
		start := time.Now()
		isPrime := isPrimeTrial(n)
		if n > 1_000_000 {
			isPrime = millerRabin(n, 10)
		}
		elapsed := time.Since(start).Milliseconds()
		body, _ := json.MarshalIndent(map[string]interface{}{
			"n":          n,
			"is_prime":   isPrime,
			"elapsed_ms": elapsed,
		}, "", "  ")
		return server.NewResponse(200, "OK", "application/json", body)
	}

	if pool == nil {
		return job(nil)
	}

	resp, err := pool.SubmitAndWait(job, workers.PriorityNormal)
	if err != nil {
		return server.NewResponse(503, "Service Unavailable", "application/json", []byte(`{"error":"queue full"}`))
	}
	return resp
}
