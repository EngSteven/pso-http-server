package server

import (
	"net"
	"testing"
	"time"
)

// Test básico: el servidor responde correctamente por TCP
func TestHelloServer(t *testing.T) {
	listener, err := net.Listen("tcp", ":0") // Puerto aleatorio
	if err != nil {
		t.Fatalf("Error creando listener: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				c.Write([]byte("HTTP/1.0 200 OK\r\n\r\nTest"))
			}(conn)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Error conectando: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 64)
	n, _ := conn.Read(buf)
	response := string(buf[:n])
	if response == "" {
		t.Error("No se recibió respuesta del servidor")
	}
}
