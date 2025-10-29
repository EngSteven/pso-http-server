/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: http_response.go
Descripcion: Constructor de respuestas HTTP que crea estructura Response
con headers automaticos como Content-Length y Connection close.
*/

package server

import (
	"fmt"

	"github.com/EngSteven/pso-http-server/internal/types"
)

// NewResponse crea respuesta HTTP completa con headers automaticos.
// Entrada: code (int) - codigo de estado HTTP
//
//	text (string) - texto del estado HTTP
//	contentType (string) - tipo de contenido MIME
//	body ([]byte) - cuerpo de la respuesta
//
// Salida: *types.Response - respuesta HTTP construida
// Descripcion: Factory function que crea respuesta HTTP con headers automaticos
//
//	incluyendo Content-Length calculado y Connection close por defecto.
//	Simplifica creacion de respuestas HTTP completas.
func NewResponse(code int, text, contentType string, body []byte) *types.Response {
	// === CONSTRUCCION DE HEADERS AUTOMATICOS ===
	// Crear headers esenciales para respuesta HTTP válida
	headers := map[string]string{
		"Content-Type":   contentType,                  // Tipo MIME del contenido
		"Content-Length": fmt.Sprintf("%d", len(body)), // Longitud exacta del body
		"Connection":     "close",                      // Cerrar conexión tras respuesta
	}
	// === CONSTRUCCION DE RESPUESTA COMPLETA ===
	// Ensamblar todos los componentes en estructura Response
	return &types.Response{
		StatusCode: code,
		StatusText: text,
		Headers:    headers,
		Body:       body,
	}
}
