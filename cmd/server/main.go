package main

import (
	"log"
	"os"

	"github.com/EngSteven/pso-http-server/internal/handlers"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := server.NewServer(":" + port)

	// 🧩 Inicializar pools antes de registrar las rutas
	workers.InitPool("fibonacci", 2, 5)
	workers.InitPool("createfile", 2, 5)

	// 🗺️ Registro de rutas
	srv.Router.Handle("/help", handlers.HelpHandler)
	srv.Router.Handle("/status", handlers.StatusHandler)
	srv.Router.Handle("/reverse", handlers.ReverseHandler)
	srv.Router.Handle("/toupper", handlers.ToUpperHandler)
	srv.Router.Handle("/fibonacci", handlers.FibonacciHandler)
	srv.Router.Handle("/createfile", handlers.CreateFileHandler)
	srv.Router.Handle("/deletefile", handlers.DeleteFileHandler)

	log.Printf("Servidor escuchando en http://localhost:%s\n", port)

	if err := srv.Start(); err != nil {
		log.Fatalf("Error al iniciar servidor: %v", err)
	}
}
