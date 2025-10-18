package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// PiHandler maneja /pi?digits=D
func PiHandler(req *types.Request) *types.Response {
	digitsStr := req.Query.Get("digits")

	if digitsStr == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: digits"}`))
	}

	digits, err := strconv.Atoi(digitsStr)
	if err != nil || digits <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: digits must be integer > 0"}`))
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.CalculatePi(digits, cancelCh)
	}

	return workers.HandlePoolSubmit("pi", jobFn, workers.PriorityNormal)
}
