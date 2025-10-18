package handlers

import (
	"strconv"
	"time"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// MatrixHandler maneja /matrixmul?size=N&seed=S
func MatrixHandler(req *types.Request) *types.Response {
	sizeStr := req.Query.Get("size")
	seedStr := req.Query.Get("seed")

	if sizeStr == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: size"}`))
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: size must be integer > 0"}`))
	}

	var seed int64 = time.Now().UnixNano()
	if seedStr != "" {
		if s, err := strconv.ParseInt(seedStr, 10, 64); err == nil {
			seed = s
		}
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.MatrixMultiply(size, seed, cancelCh)
	}

	return workers.HandlePoolSubmit("matrixmul", jobFn, workers.PriorityNormal)
}
