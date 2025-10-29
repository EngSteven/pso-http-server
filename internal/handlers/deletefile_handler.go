/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: deletefile_handler.go
Descripcion: Handler HTTP para eliminacion de archivos que valida
parametros requeridos y ejecuta la operacion mediante workers.
*/

package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// DeleteFileHandler procesa requests para eliminar archivos del sistema.
// Entrada: req (*types.Request) - request HTTP con parametro name en query string
// Salida: *types.Response - respuesta HTTP con confirmacion de eliminacion o error
// Descripcion: Handler HTTP que extrae parametro name (obligatorio) desde query string,
//
//	valida que este presente y delega la eliminacion al algoritmo
//	correspondiente mediante el pool de workers "createfile".
func DeleteFileHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener nombre del archivo a eliminar desde query string
	name := req.Query.Get("name")

	// === VALIDACION DE PARAMETROS OBLIGATORIOS ===
	// Verificar que el nombre del archivo este presente
	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con parametro capturado
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.DeleteFile(name, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "createfile" (compartido) con prioridad normal
	return workers.HandlePoolSubmit("createfile", jobFn, workers.PriorityNormal)
}
