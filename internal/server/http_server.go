package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/EngSteven/pso-http-server/internal/metrics"
	"github.com/EngSteven/pso-http-server/internal/router"
	"github.com/EngSteven/pso-http-server/internal/util"
)

type Server struct {
	Address string
	Router  *router.Router
}

func NewServer(address string) *Server {
	return &Server{
		Address: address,
		Router:  router.NewRouter(),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		return fmt.Errorf("error al iniciar servidor: %v", err)
	}
	log.Printf("Servidor escuchando en %s", s.Address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error al aceptar conexión: %v", err)
			continue
		}

		// 🔹 Incrementa contador global sin crear ciclo
		metrics.IncrementConnections()

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	start := time.Now()
	reader := bufio.NewReader(conn)

	request, err := ParseRequest(reader)
	if err != nil {
		response := NewResponse(400, "Bad Request", "text/plain", []byte("400 Bad Request"))
		conn.Write(response.Bytes())
		log.Printf("[ERROR] parse request: %v", err)
		return
	}

	request.ID = util.NewRequestID()

	handler := s.Router.Match(request.Path)
	if handler == nil {
		response := NewResponse(404, "Not Found", "text/plain", []byte("404 Not Found"))
		response.Headers["X-Request-Id"] = request.ID
		response.Headers["X-Worker-Pid"] = fmt.Sprint(os.Getpid())
		conn.Write(response.Bytes())
		log.Printf("[%s] %s %s -> 404 (%.2f ms)", request.ID, request.Method, request.Path, time.Since(start).Seconds()*1000)
		return
	}

	response := handler(request)

	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers["X-Request-Id"] = request.ID
	response.Headers["X-Worker-Pid"] = fmt.Sprint(os.Getpid())

	conn.Write(response.Bytes())

	duration := time.Since(start)
	log.Printf("[%s] %s %s -> %d (%s) [PID=%d] [%.2f ms]",
		request.ID,
		request.Method,
		request.Path,
		response.StatusCode,
		response.StatusText,
		os.Getpid(),
		duration.Seconds()*1000,
	)
}
