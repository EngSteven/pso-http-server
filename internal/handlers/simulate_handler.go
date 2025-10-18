package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// SimulateHandler maneja /simulate?seconds=s&task=name
func SimulateHandler(req *types.Request) *types.Response {
	secStr := req.Query.Get("seconds")
	taskName := req.Query.Get("task")

	seconds, err := strconv.Atoi(secStr)
	if err != nil || seconds <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid or missing parameter: seconds"}`))
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.SimulateWork(seconds, taskName, cancelCh)
	}

	return workers.HandlePoolSubmit("simulate", jobFn, workers.PriorityNormal)
}
