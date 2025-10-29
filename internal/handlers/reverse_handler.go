/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: reverse_handler.go
Descripcion: Handler HTTP para inversion de texto que delega validacion
de parametros al algoritmo correspondiente para manejo consistente.
*/

package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// ReverseHandler procesa requests para invertir texto.
// Entrada: req (*types.Request) - request HTTP con parametro text
// Salida: *types.Response - respuesta HTTP con texto invertido o error
// Descripcion: Handler HTTP que extrae parametro text desde query string.
//
//	Si text esta vacio, delega validacion directamente al algoritmo.
//	Caso contrario, usa pool "reverse" para procesamiento con prioridad normal.
func ReverseHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener texto a invertir desde query string
	text := req.Query.Get("text")

	// === MANEJO DE TEXTO VACIO ===
	// Delegar validacion al algoritmo para respuesta consistente
	if text == "" {
		return algorithms.ReverseText("", nil) // delega validación al algoritmo
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con texto capturado
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.ReverseText(text, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "reverse" (compartido con transformaciones) con prioridad normal
	return workers.HandlePoolSubmit("reverse", jobFn, workers.PriorityNormal)
}
