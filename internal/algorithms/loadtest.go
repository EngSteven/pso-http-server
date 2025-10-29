/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: loadtest.go
Descripcion: Ejecuta pruebas de carga con multiples tareas concurrentes
para medir capacidad de manejo de concurrencia del sistema.
*/

package algorithms

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// LoadTest crea multiples goroutines que ejecutan tareas simuladas en paralelo.
// Entrada: taskCount (int) - numero de tareas concurrentes a ejecutar (debe ser > 0)
//
//	sleepSeconds (int) - duracion de pausa por tarea en segundos (debe ser >= 0)
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con resultados de tareas o error
// Descripcion: Ejecuta pruebas de carga creando taskCount goroutines concurrentes.
//
//	Cada tarea duerme por sleepSeconds para simular trabajo. Maneja
//	cancelacion y recopila resultados de todas las tareas ejecutadas.
func LoadTest(taskCount, sleepSeconds int, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETROS ===
	// Verificar que parametros de carga sean validos
	if taskCount <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: tasks must be > 0"}`))
	}
	if sleepSeconds < 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: sleep must be >= 0"}`))
	}

	// === VERIFICACION DE CANCELACION TEMPRANA ===
	// Chequear cancelacion antes de iniciar prueba de carga
	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"loadtest cancelled before start"}`))
	default:
	}

	// === INICIALIZACION DE SINCRONIZACION ===
	// Configurar primitivas de concurrencia para manejo de goroutines
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]string, 0, taskCount)

	// === CREACION DE TAREAS CONCURRENTES ===
	// Lanzar taskCount goroutines para simular carga de trabajo
	for i := 1; i <= taskCount; i++ {
		// === VERIFICACION DE CANCELACION POR TAREA ===
		// Permitir cancelacion durante creacion de tareas
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(fmt.Sprintf(`{"error":"loadtest cancelled after %d/%d tasks"}`, i-1, taskCount)))
		default:
			// === LANZAMIENTO DE GOROUTINE ===
			// Crear worker concurrente con ID unico
			wg.Add(1)
			go func(taskID int) {
				defer wg.Done()

				// === EJECUCION DE TAREA CON CANCELACION ===
				// Simular trabajo o manejar cancelacion
				select {
				case <-cancelCh:
					// === REGISTRO DE TAREA CANCELADA ===
					// Marcar tarea como cancelada de forma thread-safe
					mu.Lock()
					results = append(results, fmt.Sprintf("task-%d: cancelled", taskID))
					mu.Unlock()
					return
				default:
					// === SIMULACION DE TRABAJO ===
					// Ejecutar sleep para simular carga de trabajo
					time.Sleep(time.Duration(sleepSeconds) * time.Second)
					// === REGISTRO DE TAREA COMPLETADA ===
					// Marcar tarea como completada de forma thread-safe
					mu.Lock()
					results = append(results, fmt.Sprintf("task-%d: done", taskID))
					mu.Unlock()
				}
			}(i)
		}
	}

	// === ESPERA DE FINALIZACION ===
	// Aguardar que todas las goroutines terminen
	wg.Wait()

	// === CONSTRUCCION DE RESPUESTA ===
	// Generar respuesta JSON con resultados de todas las tareas
	data, _ := json.MarshalIndent(map[string]interface{}{
		"tasks":      taskCount,
		"sleep_s":    sleepSeconds,
		"results":    results,
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
