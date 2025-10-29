/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: mandelbrot_handler.go
Descripcion: Handler HTTP para generar conjunto de Mandelbrot con parametros
configurables de dimensiones, iteraciones y opcion de guardar archivo.
*/

package handlers

import (
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// MandelbrotHandler procesa requests para generar fractales Mandelbrot.
// Entrada: req (*types.Request) - request HTTP con parametros width, height, max_iter, save
// Salida: *types.Response - respuesta HTTP con fractal generado o error
// Descripcion: Handler HTTP que extrae parametros width, height, max_iter (obligatorios)
//
//	y save (opcional), convierte a enteros y booleano respectivamente,
//	valida dimensiones e iteraciones y delega al algoritmo mediante pool "mandelbrot".
func MandelbrotHandler(req *types.Request) *types.Response {
	// === EXTRACCION DE PARAMETROS ===
	// Obtener dimensiones, iteraciones y opcion de guardado desde query string
	widthStr := req.Query.Get("width")
	heightStr := req.Query.Get("height")
	iterStr := req.Query.Get("max_iter")
	saveStr := req.Query.Get("save")

	// === CONVERSION DE PARAMETROS NUMERICOS ===
	// Convertir strings a enteros, usar 0 si conversion falla
	width, _ := strconv.Atoi(widthStr)
	height, _ := strconv.Atoi(heightStr)
	maxIter, _ := strconv.Atoi(iterStr)

	// === CONVERSION DE PARAMETRO BOOLEANO ===
	// Interpretar valores "true" o "1" como verdadero
	saveFile := (saveStr == "true" || saveStr == "1")

	// === VALIDACION DE PARAMETROS OBLIGATORIOS ===
	// Verificar que dimensiones e iteraciones sean positivas
	if width <= 0 || height <= 0 || maxIter <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing or invalid parameters: width, height, max_iter"}`))
	}

	// === CREACION DE FUNCION DE TRABAJO ===
	// Encapsular llamada al algoritmo con parametros validados
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		return algorithms.Mandelbrot(width, height, maxIter, saveFile, cancelCh)
	}

	// === DELEGACION AL POOL DE WORKERS ===
	// Enviar trabajo al pool "mandelbrot" con prioridad normal
	return workers.HandlePoolSubmit("mandelbrot", jobFn, workers.PriorityNormal)
}
