/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: pi_handler.go
Descripcion: Handler HTTP para calculo de Pi con precision configurable
validando parametro digits y delegando al algoritmo correspondiente.
*/

package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// PiHandler procesa requests para calculo de Pi con precision especifica.
// Entrada: req (*types.Request) - request HTTP con parametro digits
// Salida: *types.Response - respuesta HTTP con aproximacion de Pi o error
// Descripcion: Handler HTTP que extrae parametro digits (obligatorio),
//
//	lo convierte a entero, valida que sea positivo y delega
//	el calculo al algoritmo correspondiente mediante pool "pi".
func PiHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener numero de digitos de precision desde query string
	digitsStr := req.Query.Get("digits")

	// === VALIDACION DE PRESENCIA ===
	// Verificar que el parametro digits este presente
	if digitsStr == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: digits"}`))
	}

	// === CONVERSION Y VALIDACION NUMERICA ===
	// Convertir a entero y validar que sea positivo
	digits, err := strconv.Atoi(digitsStr)
	if err != nil || digits <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: digits must be integer > 0"}`))
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con precision validada
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.CalculatePi(digits, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "pi" con prioridad normal
	return workers.HandlePoolSubmit("pi", jobFn, workers.PriorityNormal)
}
