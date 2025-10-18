package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// CompressHandler maneja /compress?name=FILE&codec=gzip|xz
func CompressHandler(req *types.Request) *types.Response {
	name := req.Query.Get("name")
	codec := req.Query.Get("codec")

	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.CompressFile(name, codec, cancelCh)
	}

	return workers.HandlePoolSubmit("compress", jobFn, workers.PriorityNormal)
}
