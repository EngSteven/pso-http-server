/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: status_handler.go
Descripcion: Handler HTTP que proporciona estado detallado del servidor
incluyendo uptime, recursos del sistema y estado de todos los pools.
*/

package handlers

import (
	"encoding/json"
	"os"
	"runtime"
	"time"

	"github.com/EngSteven/pso-http-server/internal/metrics"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

var startTime = time.Now()

// Status estructura JSON con informacion completa del estado del sistema.
type Status struct {
	UptimeSeconds   float64                     `json:"uptime_seconds"`
	PID             int                         `json:"pid"`
	Hostname        string                      `json:"hostname"`
	GoRoutines      int                         `json:"goroutines"`
	GoVersion       string                      `json:"go_version"`
	ConnectionsSeen int64                       `json:"connections_seen"`
	Pools           map[string]workers.PoolInfo `json:"pools"`
	Timestamp       string                      `json:"timestamp"`
}

// StatusHandler recopila y retorna estado completo del servidor.
// Entrada: req (*types.Request) - request HTTP (no requiere parametros)
// Salida: *types.Response - respuesta HTTP con estado del sistema en JSON
// Descripcion: Handler HTTP que recopila metricas completas del sistema incluyendo
//
//	uptime, PID, hostname, goroutines, version Go, conexiones vistas,
//	estado de todos los pools de workers y timestamp actual.
func StatusHandler(req *types.Request) *types.Response {
	// === INICIALIZACION DE ESTRUCTURA DE POOLS ===
	// Crear mapa para almacenar informacion de cada pool
	pools := make(map[string]workers.PoolInfo)

	// === RECOLECCION DE ESTADO DE POOLS ===
	// Recorrer dinámicamente todos los pools registrados
	for name, pool := range workers.GetAllPools() {
		if pool != nil {
			// === EXTRACCION DE INFORMACION DEL POOL ===
			// Obtener estadisticas actuales del pool
			info := pool.Info()
			pools[name] = info
		}
	}

	// === OBTENCION DE INFORMACION DEL SISTEMA ===
	// Recopilar hostname y otras metricas del sistema
	hostname, _ := os.Hostname()

	// === CONSTRUCCION DE ESTADO COMPLETO ===
	// Crear estructura con todas las metricas del servidor
	status := Status{
		UptimeSeconds:   time.Since(startTime).Seconds(),     // Tiempo de funcionamiento
		PID:             os.Getpid(),                         // Process ID
		Hostname:        hostname,                            // Nombre del host
		GoRoutines:      runtime.NumGoroutine(),              // Goroutines activas
		GoVersion:       runtime.Version(),                   // Version de Go
		ConnectionsSeen: metrics.GetTotalConnections(),       // Conexiones totales
		Pools:           pools,                               // Estado de pools
		Timestamp:       time.Now().Format(time.RFC3339Nano), // Timestamp actual
	}

	// === SERIALIZACION JSON ===
	// Convertir estructura de estado a JSON formateado
	body, _ := json.MarshalIndent(status, "", "  ")

	// === RESPUESTA EXITOSA ===
	// Retornar estado completo con codigo 200
	return server.NewResponse(200, "OK", "application/json", body)
}
