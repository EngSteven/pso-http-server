/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: matrixmul_handler.go
Descripcion: Handler HTTP para multiplicacion de matrices que genera
matrices aleatorias con semilla configurable para reproducibilidad.
*/

package handlers

import (
	"strconv"
	"time"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// MatrixHandler procesa requests para multiplicacion de matrices.
// Entrada: req (*types.Request) - request HTTP con parametros size y seed
// Salida: *types.Response - respuesta HTTP con hash del resultado o error
// Descripcion: Handler HTTP que extrae parametro size (obligatorio) y seed
//
//	(opcional, timestamp por defecto), valida que size sea positivo
//	y delega el calculo al algoritmo mediante pool "matrixmul".
func MatrixHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener tamaño de matriz y semilla desde query string
	sizeStr := req.Query.Get("size")
	seedStr := req.Query.Get("seed")

	// === VALIDACION DE PARAMETROS OBLIGATORIOS ===
	// Verificar que el parametro size este presente
	if sizeStr == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: size"}`))
	}

	// === CONVERSION Y VALIDACION NUMERICA ===
	// Convertir size a entero y validar que sea positivo
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: size must be integer > 0"}`))
	}

	// === PROCESAMIENTO DE SEMILLA ===
	// Usar timestamp actual como semilla por defecto
	var seed int64 = time.Now().UnixNano()
	if seedStr != "" {
		// === CONVERSION DE SEMILLA OPCIONAL ===
		// Usar semilla proporcionada si es valida
		if s, err := strconv.ParseInt(seedStr, 10, 64); err == nil {
			seed = s
		}
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con parametros procesados
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.MatrixMultiply(size, seed, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "matrixmul" con prioridad normal
	return workers.HandlePoolSubmit("matrixmul", jobFn, workers.PriorityNormal)
}
