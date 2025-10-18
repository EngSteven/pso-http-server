package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// SortFileHandler maneja /sortfile?name=FILE&algo=merge|quick
func SortFileHandler(req *types.Request) *types.Response {
	name := req.Query.Get("name")
	algo := req.Query.Get("algo")

	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.SortFile(name, algo, cancelCh)
	}

	return workers.HandlePoolSubmit("sortfile", jobFn, workers.PriorityNormal)
}
