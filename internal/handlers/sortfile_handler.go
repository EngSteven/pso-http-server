/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: sortfile_handler.go
Descripcion: Handler HTTP para ordenamiento de archivos numericos
validando parametro name requerido y algoritmo opcional.
*/

package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// SortFileHandler procesa requests para ordenar archivos numericos.
// Entrada: req (*types.Request) - request HTTP con parametros name y algo
// Salida: *types.Response - respuesta HTTP con resultado de ordenamiento o error
// Descripcion: Handler HTTP que extrae parametros name (obligatorio) y algo
//
//	(opcional) desde query string, valida que name este presente
//	y delega el ordenamiento al algoritmo mediante pool "sortfile".
func SortFileHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener archivo y algoritmo de ordenamiento desde query string
	name := req.Query.Get("name")
	algo := req.Query.Get("algo")

	// === VALIDACION DE PARAMETROS OBLIGATORIOS ===
	// Verificar que el nombre del archivo este presente
	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con parametros capturados
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.SortFile(name, algo, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "sortfile" con prioridad normal
	return workers.HandlePoolSubmit("sortfile", jobFn, workers.PriorityNormal)
}
