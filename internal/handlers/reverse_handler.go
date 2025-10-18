package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// ReverseHandler maneja /reverse?text=...
func ReverseHandler(req *types.Request) *types.Response {
	text := req.Query.Get("text")
	if text == "" {
		return algorithms.ReverseText("", nil) // delega validación al algoritmo
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.ReverseText(text, cancelCh)
	}

	// Usa el pool "reverse" (puedes definirlo en main.go o compartir con CPU livianos)
	return workers.HandlePoolSubmit("reverse", jobFn, workers.PriorityNormal)
}
