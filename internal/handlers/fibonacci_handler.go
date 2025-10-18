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
	numStr := req.Query.Get("num")
	if numStr == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: num"}`))
	}
	n, err := strconv.Atoi(numStr)
	if err != nil || n <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid num parameter"}`))
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.CalculateFibonacci(n, cancelCh)
	}

	// Maneja el pool y los posibles errores 
	return workers.HandlePoolSubmit("fibonacci", jobFn, workers.PriorityNormal)
}