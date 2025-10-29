/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: timestamp.go
Descripcion: Genera timestamps en multiples formatos incluyendo Unix, ISO 8601
y formatos locales con informacion de zona horaria.
*/

package algorithms

import (
	"encoding/json"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// GetTimestamp obtiene la fecha y hora actual en multiples formatos estandar.
// Entrada: cancelCh (<-chan struct{}) - canal para cancelacion de operacion
// Salida: *types.Response - respuesta HTTP con timestamps en varios formatos
// Descripcion: Genera timestamp actual en formatos Unix (segundos y milisegundos),
//
//	ISO 8601, formato local legible y zona horaria. Operacion rapida
//	con soporte para cancelacion y metricas de tiempo de ejecucion.
func GetTimestamp(cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VERIFICACION DE CANCELACION ===
	// Comprobar cancelacion antes de generar timestamps
	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	// === CAPTURA DE TIEMPO ACTUAL ===
	// Obtener timestamp preciso del momento actual
	now := time.Now()

	// === CONSTRUCCION DE RESPUESTA CON MULTIPLES FORMATOS ===
	// Preparar respuesta JSON con timestamps en diferentes formatos
	data, _ := json.MarshalIndent(map[string]interface{}{
		"unix":       now.Unix(),                        // Timestamp Unix en segundos
		"unix_ms":    now.UnixMilli(),                   // Timestamp Unix en milisegundos
		"iso":        now.Format(time.RFC3339),          // Formato ISO 8601 estandar
		"local_time": now.Format("2006-01-02 15:04:05"), // Formato local legible
		"timezone":   now.Location().String(),           // Informacion de zona horaria
		"elapsed_ms": time.Since(start).Milliseconds(),  // Tiempo de procesamiento
	}, "", "  ")

	// === RESPUESTA EXITOSA ===
	// Retornar resultado con codigo 200 y timestamps JSON
	return server.NewResponse(200, "OK", "application/json", data)
}
