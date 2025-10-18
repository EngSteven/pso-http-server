package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// IsPrimeHandler maneja /isprime?n=NUM[&method=trial|miller]
func IsPrimeHandler(req *types.Request) *types.Response {
	numStr := req.Query.Get("n")
	method := req.Query.Get("method")

	if numStr == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: n"}`))
	}

	n, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil || n <= 1 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: n must be integer > 1"}`))
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.IsPrime(n, method, cancelCh)
	}

	return workers.HandlePoolSubmit("isprime", jobFn, workers.PriorityNormal)
}
