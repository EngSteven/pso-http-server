package handlers

import (
	"encoding/json"
	"math"
	"strconv"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

func factorize(n int64) []int64 {
	factors := []int64{}
	for n%2 == 0 {
		factors = append(factors, 2)
		n /= 2
	}
	for i := int64(3); i <= int64(math.Sqrt(float64(n))); i += 2 {
		for n%i == 0 {
			factors = append(factors, i)
			n /= i
		}
	}
	if n > 2 {
		factors = append(factors, n)
	}
	return factors
}

func FactorHandler(req *types.Request) *types.Response {
	numStr := req.Query.Get("n")
	if numStr == "" {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"missing parameter n"}`))
	}
	n, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil || n <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"invalid n"}`))
	}

	pool := workers.GetPool("factor")
	job := func(cancelCh <-chan struct{}) *types.Response {
		start := time.Now()
		factors := factorize(n)
		elapsed := time.Since(start).Milliseconds()
		body, _ := json.MarshalIndent(map[string]interface{}{
			"n":          n,
			"factors":    factors,
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
