/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: reverse.go
Descripcion: Invierte cadenas de texto caracter por caracter usando runes
para soporte completo de caracteres UTF-8 y Unicode.
*/

package algorithms

import (
	"encoding/json"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// ReverseText invierte texto caracter por caracter respetando UTF-8.
// Entrada: text (string) - texto a invertir
//
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con texto invertido o error
// Descripcion: Invierte cadena de texto caracter por caracter usando runes
//
//	para soporte completo de caracteres UTF-8 y Unicode. Valida
//	que el texto no este vacio y maneja cancelacion antes del proceso.
func ReverseText(text string, cancelCh <-chan struct{}) *types.Response {
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

	// === CONVERSION A RUNES UTF-8 ===
	// Convertir string a slice de runes para soporte Unicode completo
	runes := []rune(text)

	// === INVERSION POR INTERCAMBIO ===
	// Intercambiar caracteres desde extremos hacia el centro
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		// === INTERCAMBIO DE POSICIONES ===
		// Swap de runes en posiciones simetricas
		runes[i], runes[j] = runes[j], runes[i]
	}

	// === CONSTRUCCION DE RESPUESTA ===
	// Preparar respuesta JSON con texto original e invertido
	data, _ := json.MarshalIndent(map[string]interface{}{
		"input":      text,                             // Texto original
		"output":     string(runes),                    // Texto invertido
		"elapsed_ms": time.Since(start).Milliseconds(), // Tiempo de procesamiento
	}, "", "  ")

	// === RESPUESTA EXITOSA ===
	// Retornar resultado con codigo 200 y datos JSON
	return server.NewResponse(200, "OK", "application/json", data)
}
