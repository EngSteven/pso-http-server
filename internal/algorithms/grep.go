/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: grep.go
Descripcion: Busca patrones regex en archivos de texto linea por linea
retornando numero de coincidencias y muestras de las primeras 10 lineas.
*/

package algorithms

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// Grep busca lineas que coincidan con un patron regex en un archivo.
// Entrada: name (string) - ruta del archivo donde buscar
//
//	pattern (string) - expresion regular para buscar coincidencias
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con coincidencias encontradas o error
// Descripcion: Busca patrones regex en archivo de texto linea por linea usando
//
//	buffer de 10MB para lineas extensas. Retorna total de coincidencias
//	y muestra primeras 10 lineas que coinciden con el patron.
func Grep(name, pattern string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETROS OBLIGATORIOS ===
	// Verificar que nombre de archivo y patron sean proporcionados
	if name == "" || pattern == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameters: name or pattern"}`))
	}

	// === COMPILACION Y VALIDACION DE REGEX ===
	// Compilar expresion regular y validar sintaxis
	re, err := regexp.Compile(pattern)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"invalid regex pattern: %v"}`, err)
		return server.NewResponse(400, "Bad Request", "application/json", []byte(msg))
	}

	// === APERTURA Y VALIDACION DE ARCHIVO ===
	// Abrir archivo para lectura linea por linea
	file, err := os.Open(name)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"failed to open file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}
	defer file.Close()

	// === CONFIGURACION DE SCANNER PARA LINEAS GRANDES ===
	// Scanner con buffer expandido para manejar lineas de hasta 10MB
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // hasta 10 MB por línea

	// === INICIALIZACION DE CONTADORES Y RESULTADOS ===
	// Variables para almacenar coincidencias y muestras
	matches := 0
	samples := make([]string, 0, 10) // Primeras 10 lineas que coinciden

	// === PROCESAMIENTO LINEA POR LINEA ===
	// Escanear archivo completo buscando coincidencias con regex
	for scanner.Scan() {
		// === VERIFICACION DE CANCELACION POR LINEA ===
		// Permitir cancelacion entre cada linea procesada
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"operation cancelled while reading"}`))
		default:
		}

		// === EVALUACION DE PATRON EN LINEA ACTUAL ===
		// Verificar si la linea coincide con el patron regex
		line := scanner.Text()
		if re.MatchString(line) {
			matches++
			// === RECOLECCION DE MUESTRAS ===
			// Guardar hasta 10 lineas como ejemplos de coincidencias
			if len(samples) < 10 {
				samples = append(samples, line)
			}
		}
	}

	// === MANEJO DE ERRORES DE SCANNER ===
	// Verificar errores durante la lectura del archivo
	if err := scanner.Err(); err != nil {
		msg := fmt.Sprintf(`{"error":"failed while reading: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}

	// === CONSTRUCCION DE RESPUESTA ===
	// Generar respuesta JSON con resultados de busqueda
	data, _ := json.MarshalIndent(map[string]interface{}{
		"file":         name,
		"pattern":      pattern,
		"matches":      matches,
		"sample_lines": samples,
		"elapsed_ms":   time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
