/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: timestamp_handler.go
Descripcion: Handler HTTP para obtener timestamps actuales en multiples
formatos sin requerir parametros de entrada.
*/

package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// TimestampHandler procesa requests para obtener timestamp actual.
// Entrada: req (*types.Request) - request HTTP (no requiere parametros)
// Salida: *types.Response - respuesta HTTP con timestamps en multiples formatos
// Descripcion: Handler HTTP que no requiere parametros de entrada y delega
//
//	la obtencion de timestamp al algoritmo correspondiente mediante
//	pool "timestamp" con prioridad normal.
func TimestampHandler(req *types.Request) *types.Response {
	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo sin parametros
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.GetTimestamp(cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "timestamp" (tareas rapidas) con prioridad normal
	return workers.HandlePoolSubmit("timestamp", jobFn, workers.PriorityNormal)
}
