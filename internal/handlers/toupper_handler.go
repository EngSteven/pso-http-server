/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: toupper_handler.go
Descripcion: Handler HTTP para conversion de texto a mayusculas
delegando validacion al algoritmo para manejo consistente de errores.
*/

package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// ToUpperHandler procesa requests para convertir texto a mayusculas.
// Entrada: req (*types.Request) - request HTTP con parametro text
// Salida: *types.Response - respuesta HTTP con texto convertido o error
// Descripcion: Handler HTTP que extrae parametro text desde query string.
//
//	Si text esta vacio, delega validacion al algoritmo. Caso contrario,
//	comparte pool "reverse" para transformaciones de texto ligeras.
func ToUpperHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener texto a convertir desde query string
	text := req.Query.Get("text")

	// === MANEJO DE TEXTO VACIO ===
	// Delegar validacion al algoritmo para respuesta consistente
	if text == "" {
		return algorithms.ToUpper("", nil) // validación delegada
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con texto capturado
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.ToUpper(text, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "reverse" (compartido entre transformaciones) con prioridad normal
	return workers.HandlePoolSubmit("reverse", jobFn, workers.PriorityNormal)
}
