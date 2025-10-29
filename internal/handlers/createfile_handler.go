/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: createfile_handler.go
Descripcion: Handler HTTP para creacion de archivos con contenido repetido
que procesa parametros name, content y repeat del query string.
*/

package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// CreateFileHandler procesa requests para crear archivos con contenido repetido.
// Entrada: req (*types.Request) - request HTTP con parametros name, content y repeat
// Salida: *types.Response - respuesta HTTP con confirmacion de creacion o error
// Descripcion: Handler HTTP que extrae parametros name, content (obligatorios) y repeat
//
//	(opcional, por defecto 1) desde query string. Convierte repeat a entero
//	y delega la creacion al algoritmo mediante pool "createfile".
func CreateFileHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener parametros de creacion de archivo desde query string
	name := req.Query.Get("name")
	content := req.Query.Get("content")
	repeat := 1

	// === PROCESAMIENTO DE PARAMETRO OPCIONAL ===
	// Configurar numero de repeticiones si se proporciona
	if r := req.Query.Get("repeat"); r != "" {
		// === VALIDACION Y CONVERSION ===
		// Convertir a entero y validar que sea positivo
		if val, err := strconv.Atoi(r); err == nil && val > 0 {
			repeat = val
		}
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con parametros procesados
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.CreateFile(name, content, repeat, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "createfile" con prioridad normal
	return workers.HandlePoolSubmit("createfile", jobFn, workers.PriorityNormal)
}
