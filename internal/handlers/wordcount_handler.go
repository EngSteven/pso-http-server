/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: wordcount_handler.go
Descripcion: Handler HTTP para analisis de archivos de texto contando
lineas, palabras y bytes con validacion de parametro requerido.
*/

package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// WordCountHandler procesa requests para analizar archivos de texto.
// Entrada: req (*types.Request) - request HTTP con parametro name
// Salida: *types.Response - respuesta HTTP con estadisticas del archivo o error
// Descripcion: Handler HTTP que extrae parametro name (obligatorio) desde
//
//	query string, valida que este presente y delega el analisis
//	al algoritmo correspondiente mediante pool "wordcount".
func WordCountHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener nombre del archivo a analizar desde query string
	name := req.Query.Get("name")

	// === VALIDACION DE PARAMETROS OBLIGATORIOS ===
	// Verificar que el nombre del archivo este presente
	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con archivo capturado
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.WordCount(name, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "wordcount" con prioridad normal
	return workers.HandlePoolSubmit("wordcount", jobFn, workers.PriorityNormal)
}
