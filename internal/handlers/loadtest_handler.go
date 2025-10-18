package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// LoadTestHandler maneja /loadtest?tasks=n&sleep=x
func LoadTestHandler(req *types.Request) *types.Response {
	taskStr := req.Query.Get("tasks")
	sleepStr := req.Query.Get("sleep")

	taskCount, _ := strconv.Atoi(taskStr)
	sleepSeconds, _ := strconv.Atoi(sleepStr)

	if taskCount <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: tasks must be > 0"}`))
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.LoadTest(taskCount, sleepSeconds, cancelCh)
	}

	return workers.HandlePoolSubmit("loadtest", jobFn, workers.PriorityNormal)
}
