/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: fibonacci_handler.go
Descripcion: Handler HTTP para calculo de secuencia Fibonacci que incluye
soporte para prioridades de ejecucion y validacion de parametros numericos.
*/

package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// FibonacciResponse estructura para respuesta JSON con serie Fibonacci.
type FibonacciResponse struct {
	N      int   `json:"n"`
	Series []int `json:"series"`
}

// FibonacciHandler procesa requests para calcular secuencia Fibonacci.
// Entrada: req (*types.Request) - request HTTP con parametros num y priority
// Salida: *types.Response - respuesta HTTP con secuencia Fibonacci o error
// Descripcion: Handler HTTP que extrae parametro num (obligatorio) y priority
//
//	(opcional). Convierte num a entero, valida que sea positivo,
//	determina prioridad (high/normal/low) y delega al algoritmo
//	mediante pool "fibonacci" con la prioridad especificada.
func FibonacciHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener numero de terminos Fibonacci desde query string
	numStr := req.Query.Get("num")

	// === VALIDACION DE PRESENCIA ===
	// Verificar que el parametro num este presente
	if numStr == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: num"}`))
	}

	// === CONVERSION Y VALIDACION NUMERICA ===
	// Convertir a entero y validar que sea positivo
	n, err := strconv.Atoi(numStr)
	if err != nil || n <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid num parameter"}`))
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con numero validado
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.CalculateFibonacci(n, cancelCh)
	}

	// === PROCESAMIENTO DE PRIORIDAD ===
	// Obtener prioridad opcional, usar normal por defecto
	prioStr := req.Query.Get("priority")
	prio := workers.PriorityNormal
	switch prioStr {
	case "high":
		prio = workers.PriorityHigh
	case "low":
		prio = workers.PriorityLow
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "fibonacci" con prioridad especificada
	return workers.HandlePoolSubmit("fibonacci", jobFn, prio)
}
