/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: hash_handler.go
Descripcion: Handler HTTP para calculo de hashes de texto que maneja
casos donde el parametro text puede estar vacio delegando validacion.
*/

package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// HashHandler procesa requests para calcular hashes de texto.
// Entrada: req (*types.Request) - request HTTP con parametro text en query string
// Salida: *types.Response - respuesta HTTP con hashes calculados o error
// Descripcion: Handler HTTP que extrae parametro text desde query string.
//
//	Si text esta vacio, llama directamente al algoritmo para manejo
//	de error. Caso contrario, delega al pool "hash" con prioridad normal.
func HashHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener texto a hashear desde query string
	text := req.Query.Get("text")

	// === MANEJO DE TEXTO VACIO ===
	// Delegar validacion al algoritmo para respuesta consistente
	if text == "" {
		return algorithms.HashText("", nil)
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con texto capturado
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.HashText(text, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "hash" con prioridad normal
	return workers.HandlePoolSubmit("hash", jobFn, workers.PriorityNormal)
}
