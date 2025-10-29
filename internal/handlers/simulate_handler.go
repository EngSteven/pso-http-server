/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: simulate_handler.go
Descripcion: Handler HTTP para simulacion de trabajo computacional
validando parametro seconds y permitiendo nombre de tarea opcional.
*/

package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// SimulateHandler procesa requests para simular trabajo computacional.
// Entrada: req (*types.Request) - request HTTP con parametros seconds y task
// Salida: *types.Response - respuesta HTTP con resultado de simulacion o error
// Descripcion: Handler HTTP que extrae parametro seconds (obligatorio) y task
//
//	(opcional), convierte seconds a entero, valida que sea positivo
//	y delega la simulacion al algoritmo mediante pool "simulate".
func SimulateHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener duracion y nombre de tarea desde query string
	secStr := req.Query.Get("seconds")
	taskName := req.Query.Get("task")

	// === CONVERSION Y VALIDACION ===
	// Convertir seconds a entero y validar que sea positivo
	seconds, err := strconv.Atoi(secStr)
	if err != nil || seconds <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid or missing parameter: seconds"}`))
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con parametros procesados
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.SimulateWork(seconds, taskName, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "simulate" con prioridad normal
	return workers.HandlePoolSubmit("simulate", jobFn, workers.PriorityNormal)
}
