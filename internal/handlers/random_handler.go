/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: random_handler.go
Descripcion: Handler HTTP para generar numeros aleatorios en rangos
especificados con parametros count, min y max configurables.
*/

package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// RandomHandler procesa requests para generar numeros aleatorios.
// Entrada: req (*types.Request) - request HTTP con parametros count, min, max
// Salida: *types.Response - respuesta HTTP con numeros generados o error
// Descripcion: Handler HTTP que extrae parametros count, min, max desde query
//
//	string, los convierte a enteros y delega la generacion al
//	algoritmo correspondiente mediante pool "random".
func RandomHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener cantidad y rango de numeros aleatorios desde query string
	countStr := req.Query.Get("count")
	minStr := req.Query.Get("min")
	maxStr := req.Query.Get("max")

	// === CONVERSION DE PARAMETROS ===
	// Convertir strings a enteros, usar 0 si conversion falla
	count, _ := strconv.Atoi(countStr)
	min, _ := strconv.Atoi(minStr)
	max, _ := strconv.Atoi(maxStr)

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con parametros procesados
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.GenerateRandom(count, min, max, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "random" con prioridad normal
	return workers.HandlePoolSubmit("random", jobFn, workers.PriorityNormal)
}
