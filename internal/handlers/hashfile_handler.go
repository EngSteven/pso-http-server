package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// HashFileHandler maneja /hashfile?name=FILE[&algo=sha256|sha1|sha512|md5]
func HashFileHandler(req *types.Request) *types.Response {
	name := req.Query.Get("name")
	algo := req.Query.Get("algo")

	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.HashFile(name, algo, cancelCh)
	}

	return workers.HandlePoolSubmit("hashfile", jobFn, workers.PriorityNormal)
}
