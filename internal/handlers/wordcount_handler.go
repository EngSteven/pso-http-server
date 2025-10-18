package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// WordCountHandler maneja /wordcount?name=FILE
func WordCountHandler(req *types.Request) *types.Response {
	name := req.Query.Get("name")
	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.WordCount(name, cancelCh)
	}

	return workers.HandlePoolSubmit("wordcount", jobFn, workers.PriorityNormal)
}
