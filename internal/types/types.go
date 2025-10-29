/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: types.go
Descripcion: Tipos fundamentales del servidor HTTP incluyendo estructuras
Request, Response y HandlerFunc con metodo de serializacion HTTP.
*/

package types

import (
	"bytes"
	"fmt"
	"net/url"
)

// Request representa request HTTP entrante con metodo, ruta, query y headers.
type Request struct {
	Method  string
	Path    string
	Query   url.Values
	Headers map[string]string
	ID      string
}

// Response representa respuesta HTTP con codigo, headers y body.
type Response struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       []byte
}

// HandlerFunc define firma de funciones handler que procesan requests.
type HandlerFunc func(req *Request) *Response

// Bytes serializa respuesta HTTP en formato de texto listo para transmision.
// Entrada: ninguna (metodo de Response)
// Salida: []byte - respuesta HTTP serializada como bytes
// Descripcion: Convierte estructura Response a formato HTTP/1.0 valido incluyendo
//
//	linea de estado, headers y body. Formatea segun especificacion HTTP
//	con CRLF apropiados para transmision TCP directa.
func (r *Response) Bytes() []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "HTTP/1.0 %d %s\r\n", r.StatusCode, r.StatusText)
	for k, v := range r.Headers {
		fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
	}
	buf.WriteString("\r\n")
	buf.Write(r.Body)
	return buf.Bytes()
}
