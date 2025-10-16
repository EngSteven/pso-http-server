package main

import (
	"github.com/EngSteven/pso-http-server/internal/handlers"
	"github.com/EngSteven/pso-http-server/internal/server"
	"log"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := server.NewServer(":" + port)

	// Registro de rutas
	srv.Router.Handle("/help", handlers.HelpHandler)
	srv.Router.Handle("/status", handlers.StatusHandler)
	srv.Router.Handle("/reverse", handlers.ReverseHandler)
	srv.Router.Handle("/toupper", handlers.ToUpperHandler)
	srv.Router.Handle("/fibonacci", handlers.FibonacciHandler)
	srv.Router.Handle("/createfile", handlers.CreateFileHandler)
	srv.Router.Handle("/deletefile", handlers.DeleteFileHandler)

	if err := srv.Start(); err != nil {
		log.Fatalf("Error al iniciar servidor: %v", err)
	}
}
