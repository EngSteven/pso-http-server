/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: main.go
Descripcion: Archivo principal del servidor HTTP PSO que inicializa y configura
el servidor con pools de workers dinamicos para ejecutar algoritmos computacionales
intensivos de forma asincrona y eficiente.
*/

package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/EngSteven/pso-http-server/internal/handlers"
	"github.com/EngSteven/pso-http-server/internal/jobs"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// getenvInt obtiene una variable de entorno como entero con valor por defecto.
// Entrada: key (string) - nombre de la variable de entorno
//
//	def (int) - valor por defecto si la variable no existe
//
// Salida: int - valor de la variable de entorno convertido a entero o valor por defecto
// Descripcion: Lee una variable de entorno y la convierte a entero. Si la variable
//
//	no existe o no es un numero valido, retorna el valor por defecto.
func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// getConfigForCommand permite configuracion dinamica por comando individual.
// Entrada: cmd (string) - nombre del comando para buscar configuracion especifica
//
//	defaultWorkers (int) - numero por defecto de workers
//	defaultQueue (int) - tamaño por defecto de la cola
//	defaultTimeout (int) - timeout por defecto en milisegundos
//
// Salida: workersN (int) - numero efectivo de workers
//
//	queueN (int) - tamaño efectivo de la cola
//	timeoutN (int) - timeout efectivo en milisegundos
//
// Descripcion: Busca variables de entorno especificas para un comando (WORKERS_<CMD>,
//
//	QUEUE_<CMD>, TIMEOUT_<CMD>). Si no las encuentra, usa valores por defecto.
//	Permite personalizar la configuracion de pools por tipo de operacion.
func getConfigForCommand(cmd string, defaultWorkers, defaultQueue, defaultTimeout int) (workersN, queueN, timeoutN int) {
	upper := strings.ToUpper(cmd)
	workersN = getenvInt("WORKERS_"+upper, defaultWorkers)
	queueN = getenvInt("QUEUE_"+upper, defaultQueue)
	timeoutN = getenvInt("TIMEOUT_"+upper, defaultTimeout)
	return
}

// main es la funcion principal del servidor HTTP PSO.
// Entrada: ninguna
// Salida: ninguna
// Descripcion: Inicializa y configura el servidor HTTP completo incluyendo:
//   - Configuracion de parametros del servidor y directorios
//   - Inicializacion de pools de workers con configuracion dinamica
//   - Configuracion del job manager con persistencia
//   - Registro de todas las rutas y handlers HTTP
//   - Inicio del servidor en el puerto especificado
func main() {
	// CONFIGURACION GENERAL DEL SERVIDOR
	// Establece puerto, directorio de datos y parametros globales del JobManager

	// Puerto del servidor (por defecto 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Directorio para almacenar archivos de datos (por defecto "data")
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}

	// Crear directorio de datos si no existe
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	// Parametros globales del JobManager para control de carga
	qDepth := getenvInt("QUEUE_DEPTH", 50)  // Profundidad maxima de cola
	maxTotal := getenvInt("MAX_TOTAL", 150) // Maximo total de jobs concurrentes

	// CONFIGURACION DE COMANDOS
	// Define workers, queue y timeout por tipo de operacion
	// CPU-bound: operaciones computacionalmente intensivas
	// IO-bound: operaciones de entrada/salida intensivas
	// Utilitarios: operaciones ligeras de texto y utilidades

	commands := map[string]struct {
		workers int
		queue   int
		timeout int
	}{
		// CPU-bound
		"fibonacci":  {2, 5, 3000},
		"isprime":    {2, 3, 3000},
		"factor":     {2, 3, 3000},
		"pi":         {1, 2, 5000},
		"matrixmul":  {1, 2, 5000},
		"mandelbrot": {1, 2, 5000},

		// IO-bound
		"sortfile":  {1, 2, 4000},
		"wordcount": {1, 2, 4000},
		"grep":      {1, 2, 4000},
		"hashfile":  {1, 2, 4000},
		"compress":  {1, 2, 4000},

		// Utilitarios / Texto
		"createfile": {2, 5, 3000},
		"deletefile": {2, 5, 3000},
		"reverse":    {2, 5, 3000},
		"toupper":    {2, 5, 3000},
		"random":     {2, 5, 3000},
		"timestamp":  {2, 5, 3000},
		"hash":       {2, 5, 3000},
		"simulate":   {2, 5, 3000},
		"sleep":      {2, 5, 3000},
		"loadtest":   {2, 3, 3000},
	}

	// Inicializar pools de workers con configuracion dinamica
	// Permite override por variables de entorno especificas por comando
	for cmd, cfg := range commands {
		w, q, t := getConfigForCommand(cmd, cfg.workers, cfg.queue, cfg.timeout)
		workers.InitPool(cmd, w, q)
		workers.SetTimeout(cmd, t)
	}

	// Mostrar configuracion efectiva en logs
	log.Println("============================================================")
	log.Printf("Servidor escuchando en http://localhost:%s", port)
	log.Printf("Ruta de datos: %s", dataDir)
	log.Println("Configuración dinámica de pools y timeouts:")
	for cmd, cfg := range commands {
		w, q, t := getConfigForCommand(cmd, cfg.workers, cfg.queue, cfg.timeout)
		log.Printf("  %-12s → workers=%d | queue=%d | timeout=%dms", cmd, w, q, t)
	}
	log.Println("============================================================")

	// INICIALIZACION DE JOB MANAGER
	// Configura el gestor de trabajos con persistencia en archivo JSONL

	journalPath := filepath.Join(dataDir, "jobs_journal.jsonl")
	jobMgr, err := jobs.NewJobManager(journalPath, qDepth, maxTotal)
	if err != nil {
		log.Fatalf("failed to init job manager: %v", err)
	}
	handlers.InitializeJobManager(jobMgr)

	// REGISTRO DE RUTAS
	// Configura todos los endpoints del servidor HTTP

	srv := server.NewServer(":" + port)

	// Rutas de informacion y monitoreo
	srv.Router.Handle("/help", handlers.HelpHandler)       // Ayuda y documentacion
	srv.Router.Handle("/status", handlers.StatusHandler)   // Estado del servidor
	srv.Router.Handle("/metrics", handlers.MetricsHandler) // Metricas de rendimiento

	// Comandos utilitarios y de proposito general
	srv.Router.Handle("/fibonacci", handlers.FibonacciHandler)   // Secuencia de Fibonacci
	srv.Router.Handle("/createfile", handlers.CreateFileHandler) // Crear archivos
	srv.Router.Handle("/deletefile", handlers.DeleteFileHandler) // Eliminar archivos
	srv.Router.Handle("/reverse", handlers.ReverseHandler)       // Invertir texto
	srv.Router.Handle("/toupper", handlers.ToUpperHandler)       // Convertir a mayusculas
	srv.Router.Handle("/random", handlers.RandomHandler)         // Generar numeros aleatorios
	srv.Router.Handle("/timestamp", handlers.TimestampHandler)   // Obtener timestamp actual
	srv.Router.Handle("/hash", handlers.HashHandler)             // Calcular hash de texto
	srv.Router.Handle("/simulate", handlers.SimulateHandler)     // Simular operaciones
	srv.Router.Handle("/sleep", handlers.SleepHandler)           // Pausar ejecucion
	srv.Router.Handle("/loadtest", handlers.LoadTestHandler)     // Pruebas de carga

	// Algoritmos CPU-intensivos para calculo matematico
	srv.Router.Handle("/isprime", handlers.IsPrimeHandler)       // Verificar si es primo
	srv.Router.Handle("/factor", handlers.FactorHandler)         // Factorizacion de numeros
	srv.Router.Handle("/pi", handlers.PiHandler)                 // Calculo de Pi
	srv.Router.Handle("/matrixmul", handlers.MatrixHandler)      // Multiplicacion de matrices
	srv.Router.Handle("/mandelbrot", handlers.MandelbrotHandler) // Conjunto de Mandelbrot

	// Operaciones IO-intensivas para manejo de archivos
	srv.Router.Handle("/sortfile", handlers.SortFileHandler)   // Ordenar contenido de archivos
	srv.Router.Handle("/wordcount", handlers.WordCountHandler) // Contar palabras en archivos
	srv.Router.Handle("/grep", handlers.GrepHandler)           // Buscar patrones en archivos
	srv.Router.Handle("/hashfile", handlers.HashFileHandler)   // Calcular hash de archivos
	srv.Router.Handle("/compress", handlers.CompressHandler)   // Comprimir archivos

	// Endpoints para gestion de trabajos asincronos
	srv.Router.Handle("/jobs/submit", handlers.JobsSubmitHandler) // Enviar nuevo trabajo
	srv.Router.Handle("/jobs/status", handlers.JobsStatusHandler) // Consultar estado de trabajo
	srv.Router.Handle("/jobs/result", handlers.JobsResultHandler) // Obtener resultado de trabajo
	srv.Router.Handle("/jobs/cancel", handlers.JobsCancelHandler) // Cancelar trabajo en ejecucion

	// EJECUCION DEL SERVIDOR
	// Inicia el servidor HTTP y escucha peticiones

	if err := srv.Start(); err != nil {
		log.Fatalf("Error al iniciar servidor: %v", err)
	}
}
