package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// HashHandler maneja /hash?text=...
func HashHandler(req *types.Request) *types.Response {
	text := req.Query.Get("text")
	if text == "" {
		return algorithms.HashText("", nil)
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.HashText(text, cancelCh)
	}

	// Usa un pool genérico o "hash"
	return workers.HandlePoolSubmit("hash", jobFn, workers.PriorityNormal)
}
