/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: hashfile_handler.go
Descripcion: Handler HTTP para calculo de hashes de archivos completos
con soporte para diferentes algoritmos como SHA256, SHA1, MD5.
*/

package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// HashFileHandler procesa requests para calcular hash de archivos.
// Entrada: req (*types.Request) - request HTTP con parametros name y algo
// Salida: *types.Response - respuesta HTTP con hash del archivo o error
// Descripcion: Handler HTTP que extrae parametros name (obligatorio) y algo
//
//	(opcional) desde query string, valida que name este presente
//	y delega el calculo al algoritmo mediante pool "hashfile".
func HashFileHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener archivo y algoritmo de hash desde query string
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
		return algorithms.HashFile(name, algo, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "hashfile" con prioridad normal
	return workers.HandlePoolSubmit("hashfile", jobFn, workers.PriorityNormal)
}
