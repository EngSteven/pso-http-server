package handlers

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

func multiplyMatrix(n int) uint64 {
	A := make([][]float64, n)
	B := make([][]float64, n)
	C := make([][]float64, n)
	for i := 0; i < n; i++ {
		A[i] = make([]float64, n)
		B[i] = make([]float64, n)
		C[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			A[i][j] = rand.Float64()
			B[i][j] = rand.Float64()
		}
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			sum := 0.0
			for k := 0; k < n; k++ {
				sum += A[i][k] * B[k][j]
			}
			C[i][j] = sum
		}
	}
	var checksum uint64
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			checksum += uint64(C[i][j] * 1e6)
		}
	}
	return checksum
}

func MatrixHandler(req *types.Request) *types.Response {
	sizeStr := req.Query.Get("size")
	if sizeStr == "" {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"missing parameter size"}`))
	}
	n, err := strconv.Atoi(sizeStr)
	if err != nil || n <= 0 || n > 256 {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"invalid size"}`))
	}

	pool := workers.GetPool("matrixmul")
	job := func(cancelCh <-chan struct{}) *types.Response {
		start := time.Now()
		checksum := multiplyMatrix(n)
		elapsed := time.Since(start).Milliseconds()
		body, _ := json.MarshalIndent(map[string]interface{}{
			"size":        n,
			"checksum":    checksum,
			"elapsed_ms":  elapsed,
			"description": "Matrix multiplication complete",
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
