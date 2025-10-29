/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: loadtest_handler.go
Descripcion: Handler HTTP para ejecutar pruebas de carga concurrente
validando parametros de numero de tareas y tiempo de sleep.
*/

package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// LoadTestHandler procesa requests para ejecutar pruebas de carga.
// Entrada: req (*types.Request) - request HTTP con parametros tasks y sleep
// Salida: *types.Response - respuesta HTTP con resultados de las tareas o error
// Descripcion: Handler HTTP que extrae parametros tasks y sleep, convierte
//
//	a enteros, valida que tasks sea positivo y delega la ejecucion
//	al algoritmo correspondiente mediante pool "loadtest".
func LoadTestHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener numero de tareas y tiempo de sleep desde query string
	taskStr := req.Query.Get("tasks")
	sleepStr := req.Query.Get("sleep")

	// === CONVERSION DE PARAMETROS ===
	// Convertir strings a enteros, usar 0 si conversion falla
	taskCount, _ := strconv.Atoi(taskStr)
	sleepSeconds, _ := strconv.Atoi(sleepStr)

	// === VALIDACION DE PARAMETROS ===
	// Verificar que el numero de tareas sea positivo
	if taskCount <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: tasks must be > 0"}`))
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con parametros procesados
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.LoadTest(taskCount, sleepSeconds, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "loadtest" con prioridad normal
	return workers.HandlePoolSubmit("loadtest", jobFn, workers.PriorityNormal)
}
