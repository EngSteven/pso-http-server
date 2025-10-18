package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// GrepHandler maneja /grep?name=FILE&pattern=REGEX
func GrepHandler(req *types.Request) *types.Response {
	name := req.Query.Get("name")
	pattern := req.Query.Get("pattern")

	if name == "" || pattern == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameters: name or pattern"}`))
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.Grep(name, pattern, cancelCh)
	}

	return workers.HandlePoolSubmit("grep", jobFn, workers.PriorityNormal)
}
