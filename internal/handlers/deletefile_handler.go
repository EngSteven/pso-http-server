package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// DeleteFileHandler controla /deletefile?name=...
func DeleteFileHandler(req *types.Request) *types.Response {
	name := req.Query.Get("name")
	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.DeleteFile(name, cancelCh)
	}

	return workers.HandlePoolSubmit("createfile", jobFn, workers.PriorityNormal)
}
