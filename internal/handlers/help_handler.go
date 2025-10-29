/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: help_handler.go
Descripcion: Handler HTTP que proporciona informacion de ayuda del servidor
incluyendo endpoints disponibles, comandos de jobs y notas de uso.
*/

package handlers

import (
	"encoding/json"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// HelpInfo estructura JSON con informacion completa del servidor.
type HelpInfo struct {
	Name          string   `json:"name"`
	Version       string   `json:"version"`
	Description   string   `json:"description"`
	HTTPEndpoints []string `json:"http_endpoints"`
	JobCommands   []string `json:"job_commands"`
	Notes         []string `json:"notes"`
}

// HelpHandler proporciona documentacion completa de la API del servidor.
// Entrada: req (*types.Request) - request HTTP (no requiere parametros)
// Salida: *types.Response - respuesta HTTP con documentacion de la API en JSON
// Descripcion: Handler HTTP que retorna informacion estatica sobre endpoints
//
//	disponibles, comandos de jobs, version del servidor y notas de uso.
//	No procesa parametros de entrada, siempre retorna la misma informacion.
func HelpHandler(req *types.Request) *types.Response {
	// === CONSTRUCCION DE INFORMACION ESTATICA ===
	// Crear estructura completa con documentacion del servidor
	info := HelpInfo{
		Name:        "PSO HTTP Server",
		Version:     "1.0",
		Description: "Servidor HTTP concurrente con soporte para ejecución directa o asincrónica (Jobs) de algoritmos CPU e IO bound.",

		// === LISTADO DE ENDPOINTS HTTP ===
		// Documentar todos los endpoints disponibles organizados por categoria
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

		// === COMANDOS DISPONIBLES PARA JOBS ===
		// Listar todos los comandos que pueden ejecutarse via jobs
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

		// === NOTAS Y OBSERVACIONES ===
		// Proporcionar informacion adicional sobre uso y caracteristicas
		Notes: []string{
			"Todos los endpoints soportan HTTP/1.0 y devuelven JSON.",
			"Los comandos listados en 'job_commands' pueden ejecutarse vía /jobs/submit.",
			"El servidor soporta concurrencia configurable por variables de entorno.",
			"Cada job soporta cancelación y timeout mediante JobManager.",
			"Los endpoints /metrics y /status exponen información en tiempo real de los pools.",
		},
	}

	// === SERIALIZACION JSON ===
	// Convertir estructura de ayuda a JSON formateado
	body, _ := json.MarshalIndent(info, "", "  ")

	// === RESPUESTA EXITOSA ===
	// Retornar documentacion completa con codigo 200
	return server.NewResponse(200, "OK", "application/json", body)
}
