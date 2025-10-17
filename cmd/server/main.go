package main

import (
	"log"
	"os"

	"github.com/EngSteven/pso-http-server/internal/handlers"
	"github.com/EngSteven/pso-http-server/internal/jobs"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// create server
	srv := server.NewServer(":" + port)

	// init pools
	workers.InitPool("fibonacci", 2, 5)
	workers.InitPool("createfile", 2, 5)

	workers.InitPool("isprime", 2, 5)
	workers.InitPool("factor", 2, 5)
	workers.InitPool("pi", 1, 2)
	workers.InitPool("matrixmul", 2, 3)
	workers.InitPool("mandelbrot", 2, 2) 


	// init job manager: journal path and queue depth per priority (e.g., 50 each, max total 150)
	jobMgr, err := jobs.NewJobManager("data/jobs_journal.jsonl", 50, 150)
	if err != nil {
		log.Fatalf("failed to init job manager: %v", err)
	}
	handlers.InitializeJobManager(jobMgr)

	// register routes
	srv.Router.Handle("/help", handlers.HelpHandler)
	srv.Router.Handle("/status", handlers.StatusHandler)
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
