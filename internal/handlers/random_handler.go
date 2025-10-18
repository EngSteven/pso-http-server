package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// RandomHandler maneja /random?count=n&min=a&max=b
func RandomHandler(req *types.Request) *types.Response {
	countStr := req.Query.Get("count")
	minStr := req.Query.Get("min")
	maxStr := req.Query.Get("max")

	count, _ := strconv.Atoi(countStr)
	min, _ := strconv.Atoi(minStr)
	max, _ := strconv.Atoi(maxStr)

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.GenerateRandom(count, min, max, cancelCh)
	}

	return workers.HandlePoolSubmit("random", jobFn, workers.PriorityNormal)
}
