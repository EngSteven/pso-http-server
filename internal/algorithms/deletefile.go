/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: deletefile.go
Descripcion: Elimina archivos del sistema con validacion de parametros
y manejo de errores, retornando confirmacion en formato JSON.
*/

package algorithms

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// DeleteFile elimina un archivo del sistema de archivos.
// Entrada: name (string) - nombre/ruta del archivo a eliminar
//
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con confirmacion de eliminacion o error
// Descripcion: Elimina archivo especificado del sistema, valida que el nombre
//
//	sea proporcionado, maneja errores de eliminacion y soporta
//	cancelacion antes de la operacion.
func DeleteFile(name string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETRO NOMBRE ===
	// Verificar que el nombre del archivo sea proporcionado
	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}

	// === VERIFICACION DE CANCELACION TEMPRANA ===
	// Chequear cancelacion antes de ejecutar operacion destructiva
	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	// === ELIMINACION DEL ARCHIVO ===
	// Ejecutar eliminacion y manejar errores del sistema de archivos
	if err := os.Remove(name); err != nil {
		msg := fmt.Sprintf(`{"error":"failed to delete file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}

	// === RESPUESTA DE CONFIRMACION ===
	// Generar respuesta exitosa con confirmacion de eliminacion
	data, _ := json.MarshalIndent(map[string]interface{}{
		"file":       name,
		"message":    "file deleted successfully",
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
