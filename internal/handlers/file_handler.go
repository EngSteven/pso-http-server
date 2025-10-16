package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
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
