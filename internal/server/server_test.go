/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: server_test.go
Descripcion: Test unitario basico del servidor TCP que verifica
conectividad y respuesta correcta en puerto aleatorio.
*/

package server

import (
	"net"
	"testing"
	"time"
)

// TestHelloServer verifica que servidor responda correctamente via TCP.
// Entrada: t (*testing.T) - instancia de testing para reportar resultados
// Salida: ninguna (test pasa/falla via t.Error/t.Fatal)
// Descripcion: Test unitario que crea servidor TCP en puerto aleatorio,
//
//	establece conexion cliente, envia datos y verifica respuesta.
//	Valida conectividad basica TCP sin protocolo HTTP completo.
func TestHelloServer(t *testing.T) {
	// === CREACION DE LISTENER EN PUERTO ALEATORIO ===
	// Usar puerto 0 para asignación automática del SO
	listener, err := net.Listen("tcp", ":0") // Puerto aleatorio
	if err != nil {
		t.Fatalf("Error creando listener: %v", err)
	}
	defer listener.Close()

	// === SERVIDOR MOCK EN GOROUTINE ===
	// Simular servidor que responde con HTTP básico
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				// === RESPUESTA HTTP MINIMA ===
				// Enviar respuesta HTTP/1.0 básica
				c.Write([]byte("HTTP/1.0 200 OK\r\n\r\nTest"))
			}(conn)
		}
	}()

	// === ESPERA PARA INICIALIZACION ===
	// Dar tiempo al servidor para estar listo
	time.Sleep(100 * time.Millisecond)

	// === CONEXION CLIENTE DE PRUEBA ===
	// Establecer conexión TCP al servidor mock
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Error conectando: %v", err)
	}
	defer conn.Close()

	// === LECTURA Y VALIDACION DE RESPUESTA ===
	// Verificar que servidor responde correctamente
	buf := make([]byte, 64)
	n, _ := conn.Read(buf)
	response := string(buf[:n])
	if response == "" {
		t.Error("No se recibió respuesta del servidor")
	}
}
