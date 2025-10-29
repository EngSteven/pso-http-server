/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: compress_handler.go
Descripcion: Handler HTTP para compresion de archivos que procesa parametros
de entrada, valida datos y delega trabajo al pool de workers correspondiente.
*/

package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// CompressHandler procesa requests HTTP para compresion de archivos.
// Entrada: req (*types.Request) - request HTTP con parametros name y codec en query string
// Salida: *types.Response - respuesta HTTP con resultado de compresion o error
// Descripcion: Handler HTTP que extrae parametros name (obligatorio) y codec (opcional)
//
//	desde query string, valida que name este presente y delega la compresion
//	al algoritmo correspondiente mediante el pool de workers "compress".
func CompressHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener parametros de compresion desde query string
	name := req.Query.Get("name")
	codec := req.Query.Get("codec")

	// === VALIDACION DE PARAMETROS OBLIGATORIOS ===
	// Verificar que el nombre del archivo este presente
	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con parametros capturados
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.CompressFile(name, codec, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "compress" con prioridad normal
	return workers.HandlePoolSubmit("compress", jobFn, workers.PriorityNormal)
}
