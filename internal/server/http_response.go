package server

import (
	"fmt"

	"github.com/EngSteven/pso-http-server/internal/types"
)

// NewResponse crea una nueva respuesta HTTP lista para serializar
func NewResponse(code int, text, contentType string, body []byte) *types.Response {
	headers := map[string]string{
		"Content-Type":   contentType,
		"Content-Length": fmt.Sprintf("%d", len(body)),
		"Connection":     "close",
	}
	return &types.Response{
		StatusCode: code,
		StatusText: text,
		Headers:    headers,
		Body:       body,
	}
}
