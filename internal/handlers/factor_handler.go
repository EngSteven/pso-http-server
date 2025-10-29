/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: factor_handler.go
Descripcion: Handler HTTP para factorizacion de numeros enteros que valida
entrada numerica y ejecuta algoritmos de factorizacion mediante workers.
*/

package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// FactorHandler procesa requests para factorizar numeros enteros.
// Entrada: req (*types.Request) - request HTTP con parametro n en query string
// Salida: *types.Response - respuesta HTTP con factores primos del numero o error
// Descripcion: Handler HTTP que extrae parametro n (obligatorio), lo convierte
//
//	a int64, valida que sea mayor a 1 y delega la factorizacion
//	al algoritmo correspondiente mediante el pool "factor".
func FactorHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener numero a factorizar desde query string
	numStr := req.Query.Get("n")

	// === VALIDACION DE PRESENCIA ===
	// Verificar que el parametro n este presente
	if numStr == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: n"}`))
	}

	// === CONVERSION Y VALIDACION NUMERICA ===
	// Convertir string a int64 y validar rango
	n, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil || n <= 1 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: n must be integer > 1"}`))
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con numero validado
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.Factorize(n, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "factor" con prioridad normal
	return workers.HandlePoolSubmit("factor", jobFn, workers.PriorityNormal)
}
