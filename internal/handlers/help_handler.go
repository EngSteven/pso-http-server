package handlers

import (
	"encoding/json"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

type HelpInfo struct {
	Name           string   `json:"name"`
	Version        string   `json:"version"`
	Description    string   `json:"description"`
	HTTPEndpoints  []string `json:"http_endpoints"`
	JobCommands    []string `json:"job_commands"`
	Notes          []string `json:"notes"`
}

// HelpHandler devuelve información general y endpoints disponibles.
func HelpHandler(req *types.Request) *types.Response {
	info := HelpInfo{
		Name:        "PSO HTTP Server",
		Version:     "1.0",
		Description: "Servidor HTTP concurrente con soporte para ejecución de algoritmos vía endpoints directos o jobs asincrónicos.",
		HTTPEndpoints: []string{
			"/help",
			"/status",
			"/metrics",
			"/reverse?text=...",
			"/toupper?text=...",
			"/fibonacci?num=...",
			"/createfile?name=...&content=...&repeat=x",
			"/deletefile?name=...",
			"/jobs/submit?task=TASK&<params>",
			"/jobs/status?id=JOBID",
			"/jobs/result?id=JOBID",
			"/jobs/cancel?id=JOBID",
		},
		JobCommands: []string{
			"fibonacci",
			"createfile",
			"deletefile",
			"reverse",
			// futuros algoritmos:
			// "isprime",
			// "factor",
			// "pi",
			// "matrixmul",
			// "mandelbrot",
		},
		Notes: []string{
			"Todos los endpoints soportan HTTP/1.0 y devuelven JSON.",
			"Los comandos listados en 'job_commands' pueden ejecutarse vía /jobs/submit.",
			"Los tiempos y concurrencia son configurables mediante variables de entorno.",
		},
	}

	body, _ := json.MarshalIndent(info, "", "  ")
	return server.NewResponse(200, "OK", "application/json", body)
}
