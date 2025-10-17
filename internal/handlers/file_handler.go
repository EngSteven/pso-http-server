package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

type FileResponse struct {
	FileName string `json:"file"`
	Message  string `json:"message"`
}

// CreateFileHandler crea un archivo con contenido repetido x veces.
func CreateFileHandler(req *types.Request) *types.Response {
	name := req.Query.Get("name")
	content := req.Query.Get("content")
	repeatStr := req.Query.Get("repeat")

	if name == "" || content == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error": "missing parameters: name or content"}`))
	}

	repeat := 1
	if repeatStr != "" {
		r, err := strconv.Atoi(repeatStr)
		if err == nil && r > 0 {
			repeat = r
		}
	}

	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		// Verificar cancelación antes de empezar
		select {
		case <-cancelCh:
			return server.NewResponse(500, "Canceled", "application/json", []byte(`{"error":"cancelled"}`))
		default:
		}

		fullContent := strings.Repeat(content+"\n", repeat)
		err := os.WriteFile(name, []byte(fullContent), 0644)
		if err != nil {
			msg := fmt.Sprintf(`{"error":"failed to create file: %v"}`, err)
			return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
		}
		resp := FileResponse{FileName: name, Message: "file created successfully"}
		body, _ := json.MarshalIndent(resp, "", "  ")
		return server.NewResponse(200, "OK", "application/json", body)
	}

	p := workers.GetPool("createfile")
	if p == nil {
		// Pasar nil como cancelCh cuando no hay pool
		return jobFn(nil)
	}
	resp, err := p.SubmitAndWait(jobFn, 30000)
	if err == workers.ErrQueueFull {
		return server.NewResponse(503, "Service Unavailable", "application/json", []byte(`{"error":"queue full", "retry_after_ms":1000}`))
	}
	if err == workers.ErrTimeout {
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(`{"error":"job timeout"}`))
	}
	if resp == nil {
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(`{"error":"empty job result"}`))
	}
	return resp
}

// DeleteFileHandler elimina un archivo existente.
func DeleteFileHandler(req *types.Request) *types.Response {
	name := req.Query.Get("name")
	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error": "missing parameter: name"}`))
	}

	err := os.Remove(name)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"failed to delete file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}

	resp := FileResponse{FileName: name, Message: "file deleted successfully"}
	body, _ := json.MarshalIndent(resp, "", "  ")
	return server.NewResponse(200, "OK", "application/json", body)
}
