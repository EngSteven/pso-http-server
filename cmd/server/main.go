package main

import (
	"log"
	"os"
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/handlers"
	"github.com/EngSteven/pso-http-server/internal/jobs"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// configuraciones dinámicas
	workersFib := getenvInt("WORKERS_FIBONACCI", 2)
	queueFib := getenvInt("QUEUE_FIBONACCI", 5)
	qDepth := getenvInt("QUEUE_DEPTH", 50)
	maxTotal := getenvInt("MAX_TOTAL", 150)

	timeoutFib := getenvInt("TIMEOUT_FIBONACCI", 3000)
	workers.SetTimeout("fibonacci", timeoutFib)

	// crear servidor
	srv := server.NewServer(":" + port)

	// init pools (puedes añadir más)
	workers.InitPool("fibonacci", workersFib, queueFib)
	workers.InitPool("createfile", 2, 5)
	workers.InitPool("isprime", 2, 5)
	workers.InitPool("factor", 2, 5)
	workers.InitPool("pi", 1, 2)
	workers.InitPool("matrixmul", 2, 3)
	workers.InitPool("mandelbrot", 2, 2)

	// job manager con configuraciones dinámicas
	jobMgr, err := jobs.NewJobManager("data/jobs_journal.jsonl", qDepth, maxTotal)
	if err != nil {
		log.Fatalf("failed to init job manager: %v", err)
	}
	handlers.InitializeJobManager(jobMgr)

	// register routes
	srv.Router.Handle("/help", handlers.HelpHandler)
	srv.Router.Handle("/status", handlers.StatusHandler)
	srv.Router.Handle("/metrics", handlers.MetricsHandler)
	
	srv.Router.Handle("/reverse", handlers.ReverseHandler)
	srv.Router.Handle("/toupper", handlers.ToUpperHandler)
	srv.Router.Handle("/fibonacci", handlers.FibonacciHandler)
	srv.Router.Handle("/createfile", handlers.CreateFileHandler)
	srv.Router.Handle("/deletefile", handlers.DeleteFileHandler)

	//  CPU-bound
	srv.Router.Handle("/isprime", handlers.IsPrimeHandler)
	srv.Router.Handle("/factor", handlers.FactorHandler)
	srv.Router.Handle("/pi", handlers.PiHandler)
	srv.Router.Handle("/matrixmul", handlers.MatrixHandler)
	srv.Router.Handle("/mandelbrot", handlers.MandelbrotHandler)


	// jobs endpoints
	srv.Router.Handle("/jobs/submit", handlers.JobsSubmitHandler)
	srv.Router.Handle("/jobs/status", handlers.JobsStatusHandler)
	srv.Router.Handle("/jobs/result", handlers.JobsResultHandler)
	srv.Router.Handle("/jobs/cancel", handlers.JobsCancelHandler)

	log.Printf("🚀 Servidor escuchando en http://localhost:%s\n", port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Error al iniciar servidor: %v", err)
	}
}
