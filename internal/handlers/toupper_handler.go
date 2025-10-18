package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// ToUpperHandler maneja /toupper?text=...
func ToUpperHandler(req *types.Request) *types.Response {
	text := req.Query.Get("text")
	if text == "" {
		return algorithms.ToUpper("", nil) // validación delegada
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.ToUpper(text, cancelCh)
	}

	// Pool "reverse" puede compartirse entre transformaciones ligeras
	return workers.HandlePoolSubmit("reverse", jobFn, workers.PriorityNormal)
}
