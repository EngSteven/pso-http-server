/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: http_server.go
Descripcion: Servidor HTTP principal que acepta conexiones TCP, parsea requests
y delega procesamiento a handlers via router con logging detallado.
*/

/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: http_server.go
Descripcion: Servidor HTTP principal que maneja conexiones TCP,
parsea requests, ejecuta handlers y registra metricas con logging detallado.
*/

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

// Server encapsula direccion de escucha y router para manejo de rutas.
type Server struct {
	Address string
	Router  *router.Router
}

// NewServer crea nueva instancia de servidor con router inicializado.
// Entrada: address (string) - direccion de red para escuchar (ej: ":8080")
// Salida: *Server - servidor HTTP inicializado
// Descripcion: Constructor que crea servidor HTTP con router vacio listo
//
//	para registro de rutas. Configura direccion de escucha
//	pero no inicia el servidor hasta llamar Start().
func NewServer(address string) *Server {
	// === INICIALIZACION DE SERVIDOR ===
	// Crear servidor con router vacío listo para registro de rutas
	return &Server{
		Address: address,
		Router:  router.NewRouter(),
	}
}

// Start inicia servidor TCP escuchando en direccion configurada.
// Entrada: ninguna (metodo de Server)
// Salida: error - error de inicializacion del listener
// Descripcion: Inicia servidor TCP, crea listener en direccion especificada
//
//	y acepta conexiones entrantes en loop infinito. Cada conexion
//	se procesa en goroutine separada para concurrencia.
func (s *Server) Start() error {
	// === CREACION DEL LISTENER TCP ===
	// Establecer socket TCP en la dirección configurada
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		return fmt.Errorf("error al iniciar servidor: %v", err)
	}
	log.Printf("Servidor escuchando en %s", s.Address)

	// === LOOP PRINCIPAL DE ACEPTACION ===
	// Aceptar conexiones entrantes de forma infinita
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error al aceptar conexión: %v", err)
			continue
		}

		// === INCREMENTO DE METRICAS GLOBALES ===
		// Registrar nueva conexión para estadísticas
		metrics.IncrementConnections()

		// === PROCESAMIENTO CONCURRENTE ===
		// Delegar cada conexión a goroutine separada
		go s.handleConnection(conn)
	}
}

// handleConnection procesa conexion individual parseando request y ejecutando handler.
// Entrada: conn (net.Conn) - conexion TCP establecida con cliente
// Salida: ninguna
// Descripcion: Procesa request HTTP completo desde conexion TCP, parsea headers,
//
//	busca handler apropiado, ejecuta procesamiento y envia respuesta.
//	Incluye logging detallado con timing y headers de tracking.
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// === INICIO DE MEDICION DE TIEMPO ===
	// Capturar timestamp para cálculo de latencia
	start := time.Now()
	reader := bufio.NewReader(conn)

	// === PARSING DEL REQUEST HTTP ===
	// Extraer y validar componentes del request
	request, err := ParseRequest(reader)
	if err != nil {
		// === ERROR DE PARSING - RESPUESTA 400 ===
		response := NewResponse(400, "Bad Request", "text/plain", []byte("400 Bad Request"))
		conn.Write(response.Bytes())
		log.Printf("[ERROR] parse request: %v", err)
		return
	}

	// === GENERACION DE ID DE REQUEST ===
	// Asignar ID único para tracking y logging
	request.ID = util.NewRequestID()

	// === BUSQUEDA DE HANDLER EN ROUTER ===
	// Encontrar handler correspondiente al path
	handler := s.Router.Match(request.Path)
	if handler == nil {
		// === RUTA NO ENCONTRADA - RESPUESTA 404 ===
		response := NewResponse(404, "Not Found", "text/plain", []byte("404 Not Found"))
		response.Headers["X-Request-Id"] = request.ID
		response.Headers["X-Worker-Pid"] = fmt.Sprint(os.Getpid())
		conn.Write(response.Bytes())
		log.Printf("[%s] %s %s -> 404 (%.2f ms)", request.ID, request.Method, request.Path, time.Since(start).Seconds()*1000)
		return
	}

	// === EJECUCION DEL HANDLER ===
	// Procesar request y obtener respuesta
	response := handler(request)

	// === INYECCION DE HEADERS DE TRACKING ===
	// Agregar headers para debugging y monitoreo
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers["X-Request-Id"] = request.ID
	response.Headers["X-Worker-Pid"] = fmt.Sprint(os.Getpid())

	// === ENVIO DE RESPUESTA AL CLIENTE ===
	conn.Write(response.Bytes())

	// === LOGGING CON METRICAS DE RENDIMIENTO ===
	// Registrar request completo con timing y metadata
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
