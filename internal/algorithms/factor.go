/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: factor.go
Descripcion: Implementa factorizacion de numeros enteros en factores primos
usando division iterativa con soporte para cancelacion durante el calculo.
*/

package algorithms

import (
	"encoding/json"
	"math"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// Factorize descompone un numero entero en sus factores primos.
// Entrada: n (int64) - numero entero a factorizar (debe ser > 1)
//
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con lista de factores primos o error
// Descripcion: Descompone numero en factores primos usando division iterativa.
//
//	Inicia con division por 2, luego numeros impares hasta raiz cuadrada.
//	Maneja cancelacion durante proceso y valida parametros de entrada.
func Factorize(n int64, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETRO N ===
	// Solo numeros mayores a 1 pueden ser factorizados
	if n <= 1 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: n must be > 1"}`))
	}

	// === VERIFICACION DE CANCELACION TEMPRANA ===
	// Chequear cancelacion antes de iniciar factorizacion
	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	// === INICIALIZACION DE FACTORIZACION ===
	// Array para almacenar factores primos encontrados
	factors := make([]int64, 0)
	num := n

	// === DIVISION POR 2 (FACTOR PAR) ===
	// Extraer todos los factores de 2 primero
	for num%2 == 0 {
		// Verificar cancelacion en cada division
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"factorization cancelled"}`))
		default:
			factors = append(factors, 2)
			num /= 2
		}
	}

	// === DIVISION POR NUMEROS IMPARES ===
	// Probar divisores impares hasta la raiz cuadrada
	for i := int64(3); i <= int64(math.Sqrt(float64(num))); i += 2 {
		// Verificar cancelacion entre cada divisor candidato
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"factorization cancelled"}`))
		default:
			// Extraer todas las instancias del divisor actual
			for num%i == 0 {
				factors = append(factors, i)
				num /= i
			}
		}
	}

	// === MANEJO DE FACTOR PRIMO RESTANTE ===
	// Si queda un número mayor a 2, también es factor primo
	if num > 2 {
		factors = append(factors, num)
	}

	// === CONSTRUCCION DE RESPUESTA ===
	// Generar respuesta JSON con lista completa de factores primos
	data, _ := json.MarshalIndent(map[string]interface{}{
		"n":          n,
		"factors":    factors,
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
