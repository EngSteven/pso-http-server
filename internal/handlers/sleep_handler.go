package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// SleepHandler maneja /sleep?seconds=s
func SleepHandler(req *types.Request) *types.Response {
	secStr := req.Query.Get("seconds")
	seconds, err := strconv.Atoi(secStr)
	if err != nil || seconds <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid or missing parameter: seconds"}`))
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.Sleep(seconds, cancelCh)
	}

	return workers.HandlePoolSubmit("sleep", jobFn, workers.PriorityNormal)
}
