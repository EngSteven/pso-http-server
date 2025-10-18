package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// MandelbrotHandler maneja /mandelbrot?width=W&height=H&max_iter=I[&save=true]
func MandelbrotHandler(req *types.Request) *types.Response {
	widthStr := req.Query.Get("width")
	heightStr := req.Query.Get("height")
	iterStr := req.Query.Get("max_iter")
	saveStr := req.Query.Get("save")

	width, _ := strconv.Atoi(widthStr)
	height, _ := strconv.Atoi(heightStr)
	maxIter, _ := strconv.Atoi(iterStr)
	saveFile := (saveStr == "true" || saveStr == "1")

	if width <= 0 || height <= 0 || maxIter <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing or invalid parameters: width, height, max_iter"}`))
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.Mandelbrot(width, height, maxIter, saveFile, cancelCh)
	}

	return workers.HandlePoolSubmit("mandelbrot", jobFn, workers.PriorityNormal)
}
