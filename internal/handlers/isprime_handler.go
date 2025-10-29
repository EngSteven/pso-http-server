/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: isprime_handler.go
Descripcion: Handler HTTP para pruebas de primalidad que soporta diferentes
metodos de verificacion como trial division y Miller-Rabin.
*/

package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// IsPrimeHandler procesa requests para verificar si un numero es primo.
// Entrada: req (*types.Request) - request HTTP con parametros n y method
// Salida: *types.Response - respuesta HTTP con resultado de primalidad o error
// Descripcion: Handler HTTP que extrae parametro n (obligatorio) y method
//
//	(opcional). Convierte n a int64, valida que sea mayor a 1
//	y delega la verificacion al algoritmo mediante pool "isprime".
func IsPrimeHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener numero y metodo de verificacion desde query string
	numStr := req.Query.Get("n")
	method := req.Query.Get("method")

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
	// Encapsular llamada al algoritmo con parametros validados
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.IsPrime(n, method, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "isprime" con prioridad normal
	return workers.HandlePoolSubmit("isprime", jobFn, workers.PriorityNormal)
}
