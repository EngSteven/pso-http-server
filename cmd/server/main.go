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

// getenvInt obtiene una variable de entorno como entero, con valor por defecto.
func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// getConfigForCommand permite sobreescribir dinámicamente la configuración de cada comando.
// Busca WORKERS_<CMD>, QUEUE_<CMD> y TIMEOUT_<CMD> en el entorno.
func getConfigForCommand(cmd string, defaultWorkers, defaultQueue, defaultTimeout int) (workersN, queueN, timeoutN int) {
	upper := strings.ToUpper(cmd)
	workersN = getenvInt("WORKERS_"+upper, defaultWorkers)
	queueN = getenvInt("QUEUE_"+upper, defaultQueue)
	timeoutN = getenvInt("TIMEOUT_"+upper, defaultTimeout)
	return
}

func main() {
	// ============================================================
	// CONFIGURACIÓN GENERAL DEL SERVIDOR
	// ============================================================

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}

	// Crear directorio de datos si no existe
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	// Parámetros globales del JobManager
	qDepth := getenvInt("QUEUE_DEPTH", 50)
	maxTotal := getenvInt("MAX_TOTAL", 150)

	// ============================================================
	// CONFIGURACIÓN DE COMANDOS (CPU, IO y utilitarios)
	// ============================================================

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

	// Inicializar pools con overrides dinámicos
	for cmd, cfg := range commands {
		w, q, t := getConfigForCommand(cmd, cfg.workers, cfg.queue, cfg.timeout)
		workers.InitPool(cmd, w, q)
		workers.SetTimeout(cmd, t)
	}

	// Mostrar configuración efectiva
	log.Println("============================================================")
	log.Printf("Servidor escuchando en http://localhost:%s", port)
	log.Printf("Ruta de datos: %s", dataDir)
	log.Println("Configuración dinámica de pools y timeouts:")
	for cmd, cfg := range commands {
		w, q, t := getConfigForCommand(cmd, cfg.workers, cfg.queue, cfg.timeout)
		log.Printf("  %-12s → workers=%d | queue=%d | timeout=%dms", cmd, w, q, t)
	}
	log.Println("============================================================")

	// ============================================================
	// INICIALIZACIÓN DE JOB MANAGER
	// ============================================================

	journalPath := filepath.Join(dataDir, "jobs_journal.jsonl")
	jobMgr, err := jobs.NewJobManager(journalPath, qDepth, maxTotal)
	if err != nil {
		log.Fatalf("failed to init job manager: %v", err)
	}
	handlers.InitializeJobManager(jobMgr)

	// ============================================================
	// REGISTRO DE RUTAS
	// ============================================================

	srv := server.NewServer(":" + port)

	// Rutas base
	srv.Router.Handle("/help", handlers.HelpHandler)
	srv.Router.Handle("/status", handlers.StatusHandler)
	srv.Router.Handle("/metrics", handlers.MetricsHandler)

	// Comandos utilitarios
	srv.Router.Handle("/fibonacci", handlers.FibonacciHandler)
	srv.Router.Handle("/createfile", handlers.CreateFileHandler)
	srv.Router.Handle("/deletefile", handlers.DeleteFileHandler)
	srv.Router.Handle("/reverse", handlers.ReverseHandler)
	srv.Router.Handle("/toupper", handlers.ToUpperHandler)
	srv.Router.Handle("/random", handlers.RandomHandler)
	srv.Router.Handle("/timestamp", handlers.TimestampHandler)
	srv.Router.Handle("/hash", handlers.HashHandler)
	srv.Router.Handle("/simulate", handlers.SimulateHandler)
	srv.Router.Handle("/sleep", handlers.SleepHandler)
	srv.Router.Handle("/loadtest", handlers.LoadTestHandler)

	// CPU-bound
	srv.Router.Handle("/isprime", handlers.IsPrimeHandler)
	srv.Router.Handle("/factor", handlers.FactorHandler)
	srv.Router.Handle("/pi", handlers.PiHandler)
	srv.Router.Handle("/matrixmul", handlers.MatrixHandler)
	srv.Router.Handle("/mandelbrot", handlers.MandelbrotHandler)

	// IO-bound
	srv.Router.Handle("/sortfile", handlers.SortFileHandler)
	srv.Router.Handle("/wordcount", handlers.WordCountHandler)
	srv.Router.Handle("/grep", handlers.GrepHandler)
	srv.Router.Handle("/hashfile", handlers.HashFileHandler)
	srv.Router.Handle("/compress", handlers.CompressHandler)

	// Jobs endpoints
	srv.Router.Handle("/jobs/submit", handlers.JobsSubmitHandler)
	srv.Router.Handle("/jobs/status", handlers.JobsStatusHandler)
	srv.Router.Handle("/jobs/result", handlers.JobsResultHandler)
	srv.Router.Handle("/jobs/cancel", handlers.JobsCancelHandler)

	// ============================================================
	// EJECUCIÓN DEL SERVIDOR
	// ============================================================

	if err := srv.Start(); err != nil {
		log.Fatalf("Error al iniciar servidor: %v", err)
	}
}
