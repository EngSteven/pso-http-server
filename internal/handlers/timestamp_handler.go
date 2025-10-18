package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// TimestampHandler maneja /timestamp
func TimestampHandler(req *types.Request) *types.Response {
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.GetTimestamp(cancelCh)
	}

	// Usa el pool "timestamp" o uno genérico de tareas rápidas
	return workers.HandlePoolSubmit("timestamp", jobFn, workers.PriorityNormal)
}
