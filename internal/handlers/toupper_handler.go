package handlers

import (
	"encoding/json"
	"strings"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

type UpperResponse struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

func ToUpperHandler(req *types.Request) *types.Response {
	text := req.Query.Get("text")
	if text == "" {
		body := []byte(`{"error": "missing parameter: text"}`)
		return server.NewResponse(400, "Bad Request", "application/json", body)
	}

	resp := UpperResponse{Input: text, Output: strings.ToUpper(text)}
	body, _ := json.MarshalIndent(resp, "", "  ")
	return server.NewResponse(200, "OK", "application/json", body)
}
