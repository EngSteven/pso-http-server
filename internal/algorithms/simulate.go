/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: simulate.go
Descripcion: Simula trabajo computacional durante tiempo especificado
para pruebas de carga y ocupacion de workers con cancelacion.
*/

package algorithms

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// SimulateWork ejecuta simulacion de trabajo ocupando CPU/worker por tiempo definido.
// Entrada: seconds (int) - duracion de la simulacion en segundos (debe ser > 0)
//
//	taskName (string) - nombre identificador de la tarea
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con resultado de simulacion o error
// Descripcion: Simula trabajo computacional durante tiempo especificado usando
//
//	sleep de 1 segundo por iteracion. Permite nombrar tarea y reporta
//	progreso si se cancela parcialmente durante ejecucion.
func SimulateWork(seconds int, taskName string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETROS ===
	// Verificar que la duracion sea positiva
	if seconds <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: seconds must be > 0"}`))
	}

	// === NOMBRE DE TAREA POR DEFECTO ===
	// Asignar nombre generico si no se proporciona
	if taskName == "" {
		taskName = "generic"
	}

	// === VERIFICACION DE CANCELACION INICIAL ===
	// Comprobar cancelacion antes de iniciar simulacion
	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"simulation cancelled before start"}`))
	default:
	}

	// === SIMULACION DE TRABAJO ===
	// Ejecutar trabajo simulado con verificacion de cancelacion
	for i := 0; i < seconds; i++ {
		// === VERIFICACION DE CANCELACION POR SEGUNDO ===
		// Permitir cancelacion cada segundo de simulacion
		select {
		case <-cancelCh:
			// === REPORTE DE CANCELACION PARCIAL ===
			// Informar progreso parcial al momento de cancelacion
			msg := fmt.Sprintf(`{"task":"%s","error":"simulation cancelled after %d seconds"}`, taskName, i)
			return server.NewResponse(499, "Client Closed Request", "application/json", []byte(msg))
		default:
			// === OCUPACION DE RECURSO ===
			// Simular trabajo usando sleep de 1 segundo
			time.Sleep(1 * time.Second)
		}
	}

	// === CONSTRUCCION DE RESPUESTA EXITOSA ===
	// Preparar respuesta JSON con resultado de simulacion completa
	data, _ := json.MarshalIndent(map[string]interface{}{
		"task":       taskName,                                                                 // Nombre de la tarea simulada
		"seconds":    seconds,                                                                  // Duracion total configurada
		"message":    fmt.Sprintf("simulation for task '%s' completed successfully", taskName), // Mensaje de exito
		"elapsed_ms": time.Since(start).Milliseconds(),                                         // Tiempo real transcurrido
	}, "", "  ")

	// === RESPUESTA EXITOSA ===
	// Retornar resultado con codigo 200 y datos JSON
	return server.NewResponse(200, "OK", "application/json", data)
}
