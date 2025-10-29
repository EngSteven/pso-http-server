/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: sleep_handler.go
Descripcion: Handler HTTP para pausas controladas validando parametro
seconds y ejecutando sleep mediante el pool correspondiente.
*/

package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// SleepHandler procesa requests para ejecutar pausas controladas.
// Entrada: req (*types.Request) - request HTTP con parametro seconds
// Salida: *types.Response - respuesta HTTP con confirmacion de pausa o error
// Descripcion: Handler HTTP que extrae parametro seconds (obligatorio),
//
//	lo convierte a entero, valida que sea positivo y delega
//	la pausa al algoritmo correspondiente mediante pool "sleep".
func SleepHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener duracion de pausa desde query string
	secStr := req.Query.Get("seconds")

	// === CONVERSION Y VALIDACION ===
	// Convertir seconds a entero y validar que sea positivo
	seconds, err := strconv.Atoi(secStr)
	if err != nil || seconds <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid or missing parameter: seconds"}`))
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con duracion validada
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.Sleep(seconds, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "sleep" con prioridad normal
	return workers.HandlePoolSubmit("sleep", jobFn, workers.PriorityNormal)
}
