/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: createfile.go
Descripcion: Crea archivos con contenido repetido especificado por el usuario
con validacion de parametros y soporte para cancelacion de operaciones.
*/

package algorithms

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// CreateFile genera un archivo con contenido de texto repetido n veces.
// Entrada: name (string) - nombre del archivo a crear
//
//	content (string) - texto a escribir en el archivo
//	repeat (int) - numero de veces a repetir el contenido (minimo 1)
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con confirmacion de creacion o error
// Descripcion: Crea archivo con contenido repetido especificado, valida parametros
//
//	obligatorios, maneja errores de escritura y soporta cancelacion.
//	Agrega salto de linea despues de cada repeticion del contenido.
func CreateFile(name, content string, repeat int, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETROS OBLIGATORIOS ===
	// Verificar que nombre y contenido sean proporcionados
	if name == "" || content == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameters: name or content"}`))
	}
	// === NORMALIZACION DE PARAMETRO REPEAT ===
	// Asegurar minimo 1 repeticion si valor invalido
	if repeat <= 0 {
		repeat = 1
	}

	// === VERIFICACION DE CANCELACION TEMPRANA ===
	// Chequear cancelacion antes de iniciar operacion de archivo
	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	// === CONSTRUCCION Y ESCRITURA DE CONTENIDO ===
	// Generar contenido final con repeticiones y saltos de linea
	full := strings.Repeat(content+"\n", repeat)
	err := os.WriteFile(name, []byte(full), 0644)
	// === MANEJO DE ERRORES DE ESCRITURA ===
	// Retornar error 500 si falla creacion del archivo
	if err != nil {
		msg := fmt.Sprintf(`{"error":"failed to create file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}

	// === RESPUESTA DE CONFIRMACION ===
	// Generar respuesta exitosa con confirmacion y metricas
	data, _ := json.MarshalIndent(map[string]interface{}{
		"file":       name,
		"message":    "file created successfully",
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
