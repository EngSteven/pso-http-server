/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: sleep.go
Descripcion: Implementa pausa controlada para simular esperas o inactividad
con soporte para cancelacion durante el periodo de pausa.
*/

package algorithms

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// Sleep ejecuta pausa controlada de N segundos simulando inactividad.
// Entrada: seconds (int) - duracion de la pausa en segundos (debe ser > 0)
//
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con confirmacion de pausa o error
// Descripcion: Implementa pausa controlada con sleep interruptible mediante
//
//	chequeo de cancelacion cada segundo. Reporta tiempo transcurrido
//	si se cancela antes de completar la pausa total solicitada.
func Sleep(seconds int, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETROS ===
	// Verificar que la duracion de pausa sea positiva
	if seconds <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: seconds must be > 0"}`))
	}

	// === VERIFICACION DE CANCELACION INICIAL ===
	// Comprobar cancelacion antes de iniciar pausa
	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"sleep cancelled before start"}`))
	default:
	}

	// === PAUSA INTERRUPTIBLE ===
	// Ejecutar pausa con verificacion de cancelacion cada segundo
	for i := 0; i < seconds; i++ {
		// === VERIFICACION DE CANCELACION POR SEGUNDO ===
		// Permitir cancelacion durante la pausa
		select {
		case <-cancelCh:
			// === REPORTE DE CANCELACION PARCIAL ===
			// Informar tiempo transcurrido al momento de cancelacion
			msg := fmt.Sprintf(`{"error":"sleep cancelled after %d seconds"}`, i)
			return server.NewResponse(499, "Client Closed Request", "application/json", []byte(msg))
		default:
			// === PAUSA DE UN SEGUNDO ===
			// Dormir por 1 segundo antes de siguiente verificacion
			time.Sleep(1 * time.Second)
		}
	}

	// === CONSTRUCCION DE RESPUESTA EXITOSA ===
	// Preparar respuesta JSON con confirmacion de pausa completa
	data, _ := json.MarshalIndent(map[string]interface{}{
		"seconds":    seconds,                                      // Duracion total de la pausa
		"message":    fmt.Sprintf("slept for %d seconds", seconds), // Mensaje de confirmacion
		"elapsed_ms": time.Since(start).Milliseconds(),             // Tiempo real transcurrido
	}, "", "  ")

	// === RESPUESTA EXITOSA ===
	// Retornar resultado con codigo 200 y datos JSON
	return server.NewResponse(200, "OK", "application/json", data)
}
