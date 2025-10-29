/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: grep_handler.go
Descripcion: Handler HTTP para busqueda de patrones regex en archivos
validando que ambos parametros name y pattern esten presentes.
*/

package handlers

import (
	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// GrepHandler procesa requests para buscar patrones en archivos.
// Entrada: req (*types.Request) - request HTTP con parametros name y pattern
// Salida: *types.Response - respuesta HTTP con coincidencias encontradas o error
// Descripcion: Handler HTTP que extrae parametros name y pattern (ambos obligatorios)
//
//	desde query string, valida que ambos esten presentes y delega
//	la busqueda al algoritmo correspondiente mediante pool "grep".
func GrepHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener archivo y patron de busqueda desde query string
	name := req.Query.Get("name")
	pattern := req.Query.Get("pattern")

	// === VALIDACION DE PARAMETROS OBLIGATORIOS ===
	// Verificar que ambos parametros esten presentes
	if name == "" || pattern == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameters: name or pattern"}`))
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con parametros validados
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.Grep(name, pattern, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "grep" con prioridad normal
	return workers.HandlePoolSubmit("grep", jobFn, workers.PriorityNormal)
}
