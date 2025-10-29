/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: toupper.go
Descripcion: Convierte texto a mayusculas respetando caracteres Unicode
y retorna resultado con formato JSON estructurado.
*/

package algorithms

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// ToUpper convierte texto a mayusculas usando transformacion Unicode.
// Entrada: text (string) - texto a convertir a mayusculas
//
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con texto convertido o error
// Descripcion: Convierte texto a mayusculas respetando caracteres Unicode y
//
//	acentos internacionales. Valida que el texto no este vacio,
//	maneja cancelacion y retorna texto original junto con convertido.
func ToUpper(text string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETROS ===
	// Verificar que el texto no este vacio
	if text == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: text"}`))
	}

	// === VERIFICACION DE CANCELACION ===
	// Comprobar cancelacion antes de procesar texto
	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	// === CONVERSION A MAYUSCULAS ===
	// Usar funcion nativa Go que respeta Unicode y acentos
	output := strings.ToUpper(text)

	// === CONSTRUCCION DE RESPUESTA ===
	// Preparar respuesta JSON con texto original y convertido
	data, _ := json.MarshalIndent(map[string]interface{}{
		"input":      text,                             // Texto original
		"output":     output,                           // Texto convertido a mayusculas
		"elapsed_ms": time.Since(start).Milliseconds(), // Tiempo de procesamiento
	}, "", "  ")

	// === RESPUESTA EXITOSA ===
	// Retornar resultado con codigo 200 y datos JSON
	return server.NewResponse(200, "OK", "application/json", data)
}
