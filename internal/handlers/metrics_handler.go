/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: metrics_handler.go
Descripcion: Handler HTTP que proporciona metricas detalladas de rendimiento
de todos los pools de workers incluyendo latencias y estadisticas de uso.
*/

package handlers

import (
	"encoding/json"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// CommandMetrics estructura con metricas detalladas por pool de workers.
type CommandMetrics struct {
	WorkersTotal   int     `json:"workers_total"`
	BusyWorkers    int32   `json:"busy_workers"`
	QueueLength    int     `json:"queue_length"`
	TotalProcessed int64   `json:"total_processed"`
	AvgLatencyMs   float64 `json:"avg_latency_ms"`
	P50Ms          int64   `json:"p50_ms"`
	P95Ms          int64   `json:"p95_ms"`
}

// Metrics estructura JSON principal para endpoint de metricas.
type Metrics struct {
	Timestamp string                    `json:"timestamp"`
	Commands  map[string]CommandMetrics `json:"commands"`
}

// MetricsHandler recopila y retorna metricas de todos los pools activos.
// Entrada: req (*types.Request) - request HTTP (no requiere parametros)
// Salida: *types.Response - respuesta HTTP con metricas detalladas en JSON
// Descripcion: Handler HTTP que itera sobre todos los pools de workers registrados,
//
//	extrae estadisticas de rendimiento (workers totales, ocupados, cola,
//	latencias, percentiles) y las retorna en formato JSON estructurado.
func MetricsHandler(req *types.Request) *types.Response {
	// === OBTENCION DE POOLS ACTIVOS ===
	// Recuperar todos los pools de workers registrados
	pools := workers.GetAllPools()
	metricsData := make(map[string]CommandMetrics)

	// === ITERACION SOBRE POOLS ===
	// Extraer metricas de cada pool registrado
	for name, pool := range pools {
		if pool != nil {
			// === EXTRACCION DE INFORMACION DEL POOL ===
			// Obtener estadisticas actuales del pool
			info := pool.Info()

			// === CONSTRUCCION DE METRICAS ===
			// Mapear informacion del pool a estructura de metricas
			metricsData[name] = CommandMetrics{
				WorkersTotal:   info.Workers,
				BusyWorkers:    info.BusyWorkers,
				QueueLength:    info.QueueLength,
				TotalProcessed: info.TotalProcessed,
				AvgLatencyMs:   info.AvgLatencyMs,
				P50Ms:          info.P50Ms,
				P95Ms:          info.P95Ms,
			}
		}
	}

	// === CONSTRUCCION DE RESPUESTA ===
	// Crear estructura principal con timestamp y metricas
	data := Metrics{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Commands:  metricsData,
	}

	// === SERIALIZACION JSON ===
	// Convertir estructura a JSON formateado
	body, _ := json.MarshalIndent(data, "", "  ")

	// === RESPUESTA EXITOSA ===
	// Retornar metricas con codigo 200
	return server.NewResponse(200, "OK", "application/json", body)
}
