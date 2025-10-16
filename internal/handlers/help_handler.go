package handlers

import (
	"encoding/json"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

type HelpInfo struct {
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	Endpoints []string `json:"endpoints"`
}

func HelpHandler(req *types.Request) *types.Response {
	info := HelpInfo{
		Name:      "PSO HTTP Server",
		Version:   "1.0",
		Endpoints: []string{"/help", "/status", "/reverse?text=...", "/toupper?text=...", "/fibonacci?num=..."},
	}
	body, _ := json.MarshalIndent(info, "", "  ")
	return server.NewResponse(200, "OK", "application/json", body)
}
