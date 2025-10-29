/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: helpers.go
Descripcion: Funciones auxiliares para pools de workers que manejan
submission de jobs y configuracion de timeouts con fallbacks.
*/

package workers

import (
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// HandlePoolSubmit ejecuta job en pool especificado con manejo de errores.
// Entrada: poolName (string) - nombre del pool de workers
//
//	job (JobFunc) - funcion ejecutable del job
//	priority (int) - prioridad de ejecucion
//
// Salida: *types.Response - respuesta HTTP del job o error mapeado
// Descripcion: Ejecuta job en pool especificado con fallback inline si pool
//
//	no existe. Mapea errores internos a codigos HTTP apropiados
//	(503 para cola llena, 408 para timeout, 500 para otros).
func HandlePoolSubmit(poolName string, job JobFunc, priority int) *types.Response {
	// === VALIDACION DE JOB FUNCTION ===
	// Verificar que job no sea nulo antes de procesamiento
	if job == nil {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"nil job function"}`))
	}

	// === BUSQUEDA DEL POOL ESPECIALIZADO ===
	// Intentar obtener pool específico para el tipo de job
	pool := GetPool(poolName)
	if pool == nil {
		// === FALLBACK INLINE - SIN POOL DISPONIBLE ===
		// Ejecutar directamente si no hay pool especializado
		return job(nil)
	}

	// === EJECUCION EN POOL CON ESPERA ===
	// Delegar job al pool y esperar resultado
	resp, err := pool.SubmitAndWait(job, priority)
	if err != nil {
		// === MAPEO DE ERRORES A CODIGOS HTTP ===
		// Convertir errores internos a respuestas HTTP apropiadas
		switch err {
		case ErrQueueFull:
			// === COLA LLENA - BACKPRESSURE ===
			return server.NewResponse(503, "Service Unavailable", "application/json",
				[]byte(`{"error":"queue full"}`))
		case ErrTimeout:
			// === TIMEOUT DE JOB ===
			return server.NewResponse(408, "Internal Server Error", "application/json",
				[]byte(`{"error":"job timeout"}`))
		default:
			// === ERROR DESCONOCIDO ===
			return server.NewResponse(500, "Internal Server Error", "application/json",
				[]byte(`{"error":"unknown error"}`))
		}
	}

	// === VALIDACION DE RESPUESTA ===
	// Verificar que job retornó resultado válido
	if resp == nil {
		return server.NewResponse(500, "Internal Server Error", "application/json",
			[]byte(`{"error":"empty job result"}`))
	}

	return resp
}

// SetTimeout configura timeout personalizado para comando especifico.
func SetTimeout(name string, ms int) {
	// === ACTUALIZACION DE TIMEOUT GLOBAL ===
	// Modificar configuración de timeout para algoritmo específico
	defaultTimeouts[name] = ms
}
