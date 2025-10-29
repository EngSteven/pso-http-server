/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: random.go
Descripcion: Genera listas de numeros enteros aleatorios en rangos especificados
con validacion de parametros y soporte para cancelacion durante generacion.
*/

package algorithms

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// GenerateRandom crea una lista de enteros aleatorios en rango [min, max].
// Entrada: count (int) - cantidad de numeros a generar (debe ser > 0)
//
//	min (int) - valor minimo del rango (debe ser < max)
//	max (int) - valor maximo del rango (debe ser > min)
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con array de numeros o error
// Descripcion: Genera lista de enteros pseudoaleatorios en rango especificado.
//
//	Usa semilla basada en tiempo actual, valida parametros de entrada
//	y maneja cancelacion durante generacion de cada numero.
func GenerateRandom(count, min, max int, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETROS ===
	// Verificar que el contador sea positivo
	if count <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: count must be > 0"}`))
	}

	// === VALIDACION DE RANGO ===
	// Verificar que el rango minimo-maximo sea valido
	if min >= max {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid range: min must be < max"}`))
	}

	// === VERIFICACION DE CANCELACION INICIAL ===
	// Comprobar cancelacion antes de generar numeros
	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	// === INICIALIZACION DEL GENERADOR ===
	// Configurar semilla basada en tiempo actual para aleatoriedad
	rand.Seed(time.Now().UnixNano())

	// === CREACION DE ARRAY RESULTADO ===
	// Reservar espacio para los numeros generados
	numbers := make([]int, count)

	// === GENERACION DE NUMEROS ALEATORIOS ===
	// Producir cada numero en el rango especificado
	for i := 0; i < count; i++ {
		// === VERIFICACION DE CANCELACION POR NUMERO ===
		// Permitir cancelacion durante generacion intensiva
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"generation cancelled"}`))
		default:
			// === CALCULO DE NUMERO ALEATORIO ===
			// Generar entero en rango [min, max] inclusive
			numbers[i] = rand.Intn(max-min+1) + min
		}
	}

	// === CONSTRUCCION DE RESPUESTA ===
	// Preparar respuesta JSON con numeros generados y metadatos
	data, _ := json.MarshalIndent(map[string]interface{}{
		"count":      count,                            // Cantidad de numeros generados
		"min":        min,                              // Valor minimo del rango
		"max":        max,                              // Valor maximo del rango
		"numbers":    numbers,                          // Array de numeros aleatorios
		"elapsed_ms": time.Since(start).Milliseconds(), // Tiempo de generacion
	}, "", "  ")

	// === RESPUESTA EXITOSA ===
	// Retornar resultado con codigo 200 y datos JSON
	return server.NewResponse(200, "OK", "application/json", data)
}
