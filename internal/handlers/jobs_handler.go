/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: jobs_handler.go
Descripcion: Handlers HTTP para gestion completa de jobs asincronos incluyendo
envio, consulta de estado, obtencion de resultados y cancelacion.
*/

package handlers

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/EngSteven/pso-http-server/internal/jobs"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

var globalJobMgr *jobs.JobManager

// InitializeJobManager configura el gestor global de jobs.
// Entrada: jm (*jobs.JobManager) - instancia del gestor de jobs a configurar
// Salida: ninguna
// Descripcion: Funcion de inicializacion que establece el gestor global de jobs.
//
//	Debe ser llamada desde main antes de usar cualquier handler de jobs.
//	Permite que todos los handlers accedan al mismo gestor.
func InitializeJobManager(jm *jobs.JobManager) {
	// === CONFIGURACION DE GESTOR GLOBAL ===
	// Establecer instancia global del gestor de jobs para uso en handlers
	globalJobMgr = jm
}

// queryToMap convierte url.Values a map string para parametros de jobs.
// Entrada: values (url.Values) - valores del query string de la URL
// Salida: map[string]string - mapa con parametros convertidos a strings
// Descripcion: Funcion utilitaria que convierte url.Values (que permite multiples
//
//	valores por clave) a un mapa simple string-string tomando solo
//	el primer valor de cada clave para uso en parametros de jobs.
func queryToMap(values url.Values) map[string]string {
	// === INICIALIZACION DE MAPA RESULTADO ===
	// Crear mapa para almacenar parametros convertidos
	out := make(map[string]string)

	// === CONVERSION DE PARAMETROS ===
	// Iterar sobre todos los parametros del query string
	for k, vv := range values {
		// === SELECCION DEL PRIMER VALOR ===
		// Tomar solo el primer valor si hay multiples para la misma clave
		if len(vv) > 0 {
			out[k] = vv[0]
		}
	}
	return out
}

// ------------------------------------------------------------
// /jobs/submit?task=TASK&priority=high|normal|low
// ------------------------------------------------------------
// JobsSubmitHandler procesa envio de nuevos jobs al sistema.
// Entrada: req (*types.Request) - request HTTP con parametros task, priority y otros
// Salida: *types.Response - respuesta HTTP con ID del job creado o error
// Descripcion: Handler HTTP que extrae parametro task (obligatorio) y priority
//
//	(opcional, por defecto "normal"), convierte parametros restantes
//	a mapa y delega el envio al JobManager global para ejecucion asincrona.
func JobsSubmitHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETRO OBLIGATORIO ===
	// Obtener nombre de tarea desde query string
	task := req.Query.Get("task")
	if task == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing task parameter"}`))
	}

	// === PROCESAMIENTO DE PRIORIDAD ===
	// Obtener prioridad opcional, usar normal por defecto
	priorityStr := req.Query.Get("priority")
	if priorityStr == "" {
		priorityStr = "normal"
	}

	// === MAPEO DE PRIORIDAD ===
	// Convertir string de prioridad a tipo enumerado
	var pr jobs.Priority
	switch priorityStr {
	case "high":
		pr = jobs.PriorityHigh
	case "low":
		pr = jobs.PriorityLow
	default:
		pr = jobs.PriorityNormal
	}

	// === EXTRACCION DE PARAMETROS ADICIONALES ===
	// Convertir query string a mapa y remover parametros de control
	params := queryToMap(req.Query)
	delete(params, "task")     // Remover parametro de tarea
	delete(params, "priority") // Remover parametro de prioridad

	// === ENVIO DEL JOB AL GESTOR ===
	// Solicitar creacion y encolado del job
	jobID, err := globalJobMgr.Submit(task, params, pr)

	// === MANEJO DE ERRORES ESPECIFICOS ===
	// Verificar diferentes tipos de errores posibles
	if err == jobs.ErrJobQueueFull {
		return server.NewResponse(503, "Service Unavailable", "application/json",
			[]byte(`{"error":"queue full","retry_after_ms":1000}`))
	}
	if err != nil {
		msg := fmt.Sprintf(`{"error":"%s"}`, err.Error())
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}

	// === CONSTRUCCION DE RESPUESTA EXITOSA ===
	// Crear respuesta con ID del job y estado inicial
	resp := map[string]interface{}{
		"job_id": jobID,
		"status": "queued",
	}

	// === SERIALIZACION Y RESPUESTA ===
	// Convertir a JSON y retornar respuesta exitosa
	b, _ := json.MarshalIndent(resp, "", "  ")
	return server.NewResponse(200, "OK", "application/json", b)
}

// ------------------------------------------------------------
// /jobs/status?id=JOBID
// ------------------------------------------------------------
// JobsStatusHandler consulta el estado actual de un job por ID.
// Entrada: req (*types.Request) - request HTTP con parametro id del job
// Salida: *types.Response - respuesta HTTP con estado del job o error
// Descripcion: Handler HTTP que extrae parametro id (obligatorio), consulta
//
//	el estado del job al JobManager y calcula progreso basico
//	segun el estado (pending=0%, running=50%, completed=100%).
func JobsStatusHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETRO OBLIGATORIO ===
	// Obtener ID del job desde query string
	id := req.Query.Get("id")
	if id == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing id parameter"}`))
	}

	// === CONSULTA DE METADATOS DEL JOB ===
	// Obtener informacion del job desde el gestor
	meta, err := globalJobMgr.GetMeta(id)
	if err == jobs.ErrJobNotFound {
		return server.NewResponse(404, "Not Found", "application/json",
			[]byte(`{"error":"job not found"}`))
	}

	// === CALCULO DE PROGRESO BASICO ===
	// Determinar progreso segun el estado actual del job
	progress := 0
	switch meta.Status {
	case jobs.StatusQueued:
		progress = 0 // En cola, sin iniciar
	case jobs.StatusRunning:
		progress = 50 // Ejecutandose, progreso medio
	case jobs.StatusDone:
		progress = 100 // Completado
	default:
		progress = 0 // Estado desconocido
	}

	// === CONSTRUCCION DE RESPUESTA DE ESTADO ===
	// Crear estructura con informacion del estado del job
	statusResp := map[string]interface{}{
		"id":       meta.ID,     // ID del job
		"status":   meta.Status, // Estado actual
		"progress": progress,    // Progreso calculado
		"eta_ms":   0,           // Tiempo estimado (no implementado)
	}

	// === SERIALIZACION Y RESPUESTA ===
	// Convertir a JSON y retornar estado del job
	b, _ := json.MarshalIndent(statusResp, "", "  ")
	return server.NewResponse(200, "OK", "application/json", b)
}

// ------------------------------------------------------------
// /jobs/result?id=JOBID
// ------------------------------------------------------------
// JobsResultHandler obtiene el resultado completo de un job terminado.
// Entrada: req (*types.Request) - request HTTP con parametro id del job
// Salida: *types.Response - respuesta HTTP con resultado del job o error
// Descripcion: Handler HTTP que extrae parametro id (obligatorio), consulta
//
//	el resultado del job al JobManager, decodifica la respuesta
//	almacenada y extrae el JSON del algoritmo para retornarlo al cliente.
func JobsResultHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETRO OBLIGATORIO ===
	// Obtener ID del job desde query string
	id := req.Query.Get("id")
	if id == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing id parameter"}`))
	}

	// === CONSULTA DE METADATOS DEL JOB ===
	// Obtener informacion del job desde el gestor
	meta, err := globalJobMgr.GetMeta(id)
	if err == jobs.ErrJobNotFound {
		return server.NewResponse(404, "Not Found", "application/json",
			[]byte(`{"error":"job not found"}`))
	}

	// === VERIFICACION DE ESTADO COMPLETADO ===
	// Confirmar que el job haya terminado antes de retornar resultado
	if meta.Status != jobs.StatusDone {
		msg := fmt.Sprintf(`{"error":"result not ready","status":"%s"}`, meta.Status)
		return server.NewResponse(409, "Conflict", "application/json", []byte(msg))
	}

	// === DECODIFICACION DE RESPUESTA ALMACENADA ===
	// Deserializar la respuesta completa guardada en el job
	var res types.Response
	if err := json.Unmarshal([]byte(meta.Result), &res); err != nil {
		return server.NewResponse(500, "Internal Server Error", "application/json",
			[]byte(`{"error":"invalid result format"}`))
	}

	// === EXTRACCION DEL CONTENIDO JSON ===
	// Intentar decodificar el body como JSON del algoritmo
	var body map[string]interface{}
	if err := json.Unmarshal(res.Body, &body); err == nil {
		// === RESPUESTA JSON FORMATEADA ===
		// Retornar JSON del algoritmo formateado
		b, _ := json.MarshalIndent(body, "", "  ")
		return server.NewResponse(200, "OK", "application/json", b)
	}

	// === RESPUESTA DE CONTENIDO LITERAL ===
	// Si no era JSON, devolver contenido original con headers apropiados
	return server.NewResponse(200, "OK", res.Headers["Content-Type"], res.Body)
}

// ------------------------------------------------------------
// /jobs/cancel?id=JOBID
// ------------------------------------------------------------
// JobsCancelHandler cancela la ejecucion de un job por ID.
// Entrada: req (*types.Request) - request HTTP con parametro id del job
// Salida: *types.Response - respuesta HTTP con confirmacion de cancelacion o error
// Descripcion: Handler HTTP que extrae parametro id (obligatorio) y solicita
//
//	la cancelacion del job al JobManager. Retorna confirmacion
//	si el job fue cancelado exitosamente o mensaje de error.
func JobsCancelHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETRO OBLIGATORIO ===
	// Obtener ID del job desde query string
	id := req.Query.Get("id")
	if id == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing id parameter"}`))
	}

	// === SOLICITUD DE CANCELACION ===
	// Enviar peticion de cancelacion al gestor de jobs
	err := globalJobMgr.Cancel(id)

	// === MANEJO DE RESULTADO DE CANCELACION ===
	// Procesar diferentes tipos de respuesta del gestor
	switch err {
	case nil:
		// === CANCELACION EXITOSA ===
		// Job cancelado correctamente
		resp := map[string]string{"status": "canceled"}
		b, _ := json.MarshalIndent(resp, "", "  ")
		return server.NewResponse(200, "OK", "application/json", b)

	case jobs.ErrJobNotFound:
		// === JOB NO ENCONTRADO ===
		// El ID proporcionado no existe
		return server.NewResponse(404, "Not Found", "application/json",
			[]byte(`{"error":"job not found"}`))

	case jobs.ErrJobCancelled:
		// === JOB NO CANCELABLE ===
		// El job ya termino o no puede cancelarse
		return server.NewResponse(409, "Conflict", "application/json",
			[]byte(`{"error":"not cancelable"}`))

	default:
		// === ERROR INTERNO ===
		// Otro tipo de error del gestor
		msg := fmt.Sprintf(`{"error":"%s"}`, err.Error())
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}
}
