package handlers

import (
	"encoding/json"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

type HelpInfo struct {
	Name          string   `json:"name"`
	Version       string   `json:"version"`
	Description   string   `json:"description"`
	HTTPEndpoints []string `json:"http_endpoints"`
	JobCommands   []string `json:"job_commands"`
	Notes         []string `json:"notes"`
}

// HelpHandler devuelve información general y endpoints disponibles.
func HelpHandler(req *types.Request) *types.Response {
	info := HelpInfo{
		Name:        "PSO HTTP Server",
		Version:     "1.0",
		Description: "Servidor HTTP concurrente con soporte para ejecución directa o asincrónica (Jobs) de algoritmos CPU e IO bound.",
		HTTPEndpoints: []string{
			// --- Sistema ---
			"/help",
			"/status",
			"/metrics",

			// --- Básicos ---
			"/reverse?text=...",
			"/toupper?text=...",
			"/lowercase?text=...",
			"/hash?text=...",
			"/random?count=...&min=...&max=...",
			"/timestamp",
			"/simulate?seconds=...&task=...",
			"/sleep?seconds=...",
			"/loadtest?tasks=...&sleep=...",

			// --- CPU-bound ---
			"/fibonacci?num=...",
			"/isprime?n=...&method=trial|miller",
			"/factor?n=...",
			"/matrixmul?size=...&seed=...",
			"/mandelbrot?width=...&height=...&max_iter=...&save=true",
			"/pi?digits=...",

			// --- IO-bound ---
			"/createfile?name=...&content=...&repeat=...",
			"/deletefile?name=...",
			"/sortfile?name=...&algo=merge|quick",
			"/wordcount?name=...",
			"/grep?name=...&pattern=...",
			"/hashfile?name=...&algo=sha256|sha512|md5",
			"/compress?name=...&codec=gzip|xz",

			// --- Jobs API ---
			"/jobs/submit?task=TASK&<params>",
			"/jobs/status?id=JOBID",
			"/jobs/result?id=JOBID",
			"/jobs/cancel?id=JOBID",
		},
		JobCommands: []string{
			// Básicos
			"reverse", "toupper", "lowercase", "hash",
			"random", "timestamp", "simulate", "sleep", "loadtest",

			// CPU-bound
			"fibonacci", "isprime", "factor", "matrixmul", "mandelbrot", "pi",

			// IO-bound
			"createfile", "deletefile", "sortfile", "wordcount",
			"grep", "hashfile", "compress",
		},
		Notes: []string{
			"Todos los endpoints soportan HTTP/1.0 y devuelven JSON.",
			"Los comandos listados en 'job_commands' pueden ejecutarse vía /jobs/submit.",
			"El servidor soporta concurrencia configurable por variables de entorno.",
			"Cada job soporta cancelación y timeout mediante JobManager.",
			"Los endpoints /metrics y /status exponen información en tiempo real de los pools.",
		},
	}

	body, _ := json.MarshalIndent(info, "", "  ")
	return server.NewResponse(200, "OK", "application/json", body)
}
