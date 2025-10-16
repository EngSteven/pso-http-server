package handlers

import (
	"encoding/json"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

type ReverseResponse struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

func ReverseHandler(req *types.Request) *types.Response {
	text := req.Query.Get("text")
	if text == "" {
		body := []byte(`{"error": "missing parameter: text"}`)
		return server.NewResponse(400, "Bad Request", "application/json", body)
	}

	// Reversar string
	runes := []rune(text)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	resp := ReverseResponse{Input: text, Output: string(runes)}
	body, _ := json.MarshalIndent(resp, "", "  ")
	return server.NewResponse(200, "OK", "application/json", body)
}
