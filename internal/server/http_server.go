package server

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/EngSteven/pso-http-server/internal/util"
	"github.com/EngSteven/pso-http-server/internal/router"
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
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	request, err := ParseRequest(reader)
	if err != nil {
		response := NewResponse(400, "Bad Request", "text/plain", []byte("400 Bad Request"))
		conn.Write(response.Bytes())
		return
	}

	request.ID = util.NewRequestID()

	handler := s.Router.Match(request.Path)
	if handler == nil {
		response := NewResponse(404, "Not Found", "text/plain", []byte("404 Not Found"))
		conn.Write(response.Bytes())
		return
	}

	response := handler(request)
	response.Headers["X-Request-Id"] = request.ID
	conn.Write(response.Bytes())
	log.Printf("[%s] %s %s %d", request.ID, request.Method, request.Path, response.StatusCode)
}
