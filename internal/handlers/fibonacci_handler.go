package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
	"github.com/EngSteven/pso-http-server/internal/algorithms"
)

type FibonacciResponse struct {
	N      int   `json:"n"`
	Series []int `json:"series"`
}

func FibonacciHandler(req *types.Request) *types.Response {
	// 1️⃣ Leer y validar parámetros
	numStr := req.Query.Get("num")
	if numStr == "" {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"missing parameter: num"}`))
	}
	n, err := strconv.Atoi(numStr)
	if err != nil || n <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"invalid num parameter"}`))
	}

	// 2️⃣ Preparar función del job
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.CalculateFibonacci(n, cancelCh)
	}

	// 3️⃣ Intentar enviar al pool "fibonacci"
	pool := workers.GetPool("fibonacci")
	if pool == nil {
		// fallback inline si no hay pool
		return jobFn(nil)
	}

	resp, err := pool.SubmitAndWait(jobFn, workers.PriorityNormal)
	if err != nil {
		switch err {
		case workers.ErrQueueFull:
			return server.NewResponse(503, "Service Unavailable", "application/json", []byte(`{"error":"queue full"}`))
		case workers.ErrTimeout:
			return server.NewResponse(500, "Internal Server Error", "application/json", []byte(`{"error":"job timeout"}`))
		default:
			return server.NewResponse(500, "Internal Server Error", "application/json", []byte(`{"error":"unknown error"}`))
		}
	}

	if resp == nil {
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(`{"error":"empty job result"}`))
	}

	return resp
}