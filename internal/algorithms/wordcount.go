/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: wordcount.go
Descripcion: Analiza archivos de texto contando lineas, palabras y bytes
usando lectura eficiente por bloques con soporte para cancelacion.
*/

package algorithms

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// WordCount analiza archivo de texto contando lineas, palabras y bytes totales.
// Entrada: name (string) - ruta del archivo a analizar
//
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con estadisticas del archivo o error
// Descripcion: Analiza archivo contando lineas (por saltos de linea), palabras
//
//	(separadas por espacios) y bytes totales. Lee por bloques de 32KB
//	para eficiencia con archivos grandes y maneja cancelacion.
func WordCount(name string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETROS ===
	// Verificar que se proporcione nombre de archivo
	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}

	// === APERTURA DE ARCHIVO ===
	// Abrir archivo para lectura y analisis
	file, err := os.Open(name)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"failed to open file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}
	defer file.Close()

	// === CONFIGURACION DE BUFFER DE LECTURA ===
	// Crear reader con buffer para lectura eficiente
	reader := bufio.NewReader(file)
	var (
		lines  int   // Contador de lineas
		words  int   // Contador de palabras
		bytesN int64 // Contador de bytes
	)

	// === BUFFER DE PROCESAMIENTO ===
	// Buffer de 32KB para lectura por bloques eficiente
	buf := make([]byte, 32*1024)

	// === PROCESAMIENTO POR BLOQUES ===
	// Leer archivo en chunks para manejar archivos grandes
	for {
		// === VERIFICACION DE CANCELACION POR BLOQUE ===
		// Permitir cancelacion durante lectura intensiva
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"operation cancelled while reading"}`))
		default:
		}

		// === LECTURA DE BLOQUE ===
		// Leer siguiente chunk del archivo
		n, err := reader.Read(buf)
		if n > 0 {
			// === PROCESAMIENTO DE CHUNK ===
			// Analizar contenido del bloque leido
			chunk := buf[:n]
			bytesN += int64(n) // Acumular bytes leidos

			// === CONTEO DE LINEAS ===
			// Contar saltos de linea en el chunk
			lines += strings.Count(string(chunk), "\n")

			// === CONTEO DE PALABRAS ===
			// Separar por espacios y contar campos
			words += len(strings.Fields(string(chunk)))
		}

		// === VERIFICACION DE FIN DE ARCHIVO ===
		// Terminar si se alcanza EOF o error
		if err != nil {
			break
		}
	}

	// === CONSTRUCCION DE RESPUESTA ===
	// Preparar respuesta JSON con estadisticas del archivo
	data, _ := json.MarshalIndent(map[string]interface{}{
		"file":       name,                             // Archivo analizado
		"lines":      lines,                            // Total de lineas
		"words":      words,                            // Total de palabras
		"bytes":      bytesN,                           // Total de bytes
		"elapsed_ms": time.Since(start).Milliseconds(), // Tiempo de procesamiento
	}, "", "  ")

	// === RESPUESTA EXITOSA ===
	// Retornar resultado con codigo 200 y estadisticas JSON
	return server.NewResponse(200, "OK", "application/json", data)
}
