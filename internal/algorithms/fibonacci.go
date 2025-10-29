/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: fibonacci.go
Descripcion: Implementa el calculo de la secuencia de Fibonacci con soporte
para cancelacion y retorna los resultados en formato JSON estructurado.
*/

package algorithms

import (
	"encoding/json"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// CalculateFibonacci genera la secuencia de Fibonacci hasta n terminos.
// Entrada: n (int) - numero de terminos a calcular (debe ser > 0)
//
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con secuencia completa o error
// Descripcion: Calcula secuencia de Fibonacci iterativamente desde F(0)=0, F(1)=1.
//
//	Cada termino es suma de los dos anteriores. Valida parametros de
//	entrada, maneja cancelacion durante calculo y retorna serie completa.
func CalculateFibonacci(n int, cancelCh <-chan struct{}) *types.Response {
	// === VALIDACION DE PARAMETRO N ===
	// Verificar que n sea positivo antes de iniciar calculo
	if n <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: n must be > 0"}`))
	}
	start := time.Now()

	// === INICIALIZACION DE SERIE ===
	// Crear array para almacenar secuencia completa
	series := make([]int, n)
	if n > 0 {
		series[0] = 0 // F(0) = 0
	}
	if n > 1 {
		series[1] = 1 // F(1) = 1

		// === CALCULO ITERATIVO DE FIBONACCI ===
		// Calcular cada termino como suma de los dos anteriores
		for i := 2; i < n; i++ {
			// === VERIFICACION DE CANCELACION EN CADA ITERACION ===
			// Permitir cancelacion durante calculos largos
			select {
			case <-cancelCh:
				return server.NewResponse(499, "Client Closed Request", "application/json",
					[]byte(`{"error":"calculation cancelled"}`))
			default:
			}
			// Formula: F(n) = F(n-1) + F(n-2)
			series[i] = series[i-1] + series[i-2]
		}
	}

	// === CONSTRUCCION DE RESPUESTA ===
	// Generar respuesta JSON con serie completa y metricas
	data, _ := json.MarshalIndent(map[string]interface{}{
		"n": n, "series": series, "elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")
	return server.NewResponse(200, "OK", "application/json", data)
}
