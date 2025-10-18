package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// CreateFileHandler controla /createfile?name=...&content=...&repeat=...
func CreateFileHandler(req *types.Request) *types.Response {
	name := req.Query.Get("name")
	content := req.Query.Get("content")
	repeat := 1

	if r := req.Query.Get("repeat"); r != "" {
		if val, err := strconv.Atoi(r); err == nil && val > 0 {
			repeat = val
		}
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.CreateFile(name, content, repeat, cancelCh)
	}

	return workers.HandlePoolSubmit("createfile", jobFn, workers.PriorityNormal)
}
