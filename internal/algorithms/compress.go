/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: compress.go
Descripcion: Implementa compresion de archivos usando algoritmos gzip y xz
con metricas de rendimiento y soporte para cancelacion de operaciones.
*/

package algorithms

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// CompressFile comprime un archivo usando gzip (nativo) o xz (comando externo).
// Entrada: name (string) - ruta del archivo a comprimir
//
//	codec (string) - algoritmo de compresion ("gzip" o "xz", por defecto "gzip")
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con metricas de compresion o error
// Descripcion: Comprime archivo usando algoritmo especificado, calcula metricas de
//
//	tamaño original vs comprimido, ratio de compresion y tiempo transcurrido.
//	Soporta cancelacion durante el proceso y limpia archivos parciales.
func CompressFile(name, codec string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETROS ===
	// Verificar que nombre de archivo sea proporcionado
	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}
	// Establecer codec por defecto si no se especifica
	if codec == "" {
		codec = "gzip"
	}

	// === VERIFICACION DE ARCHIVO DE ENTRADA ===
	// Verificar que el archivo exista
	inFile, err := os.Open(name)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"failed to open file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}
	defer inFile.Close()

	// === OBTENCION DE METADATOS DEL ARCHIVO ===
	// Obtener tamaño del archivo original para calcular ratio de compresion
	info, _ := inFile.Stat()
	inputSize := info.Size()

	var outName string
	var outFile *os.File

	// === SELECCION DE ALGORITMO DE COMPRESION ===
	// Ejecutar compresion segun el codec especificado
	switch strings.ToLower(codec) {
	case "gzip":
		// === COMPRESION GZIP NATIVA ===
		// Usar libreria compress/gzip de Go para compresion eficiente
		outName = name + ".gz"
		outFile, err = os.Create(outName)
		if err != nil {
			msg := fmt.Sprintf(`{"error":"failed to create output file: %v"}`, err)
			return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
		}
		defer outFile.Close()

		// Crear escritor gzip con compresion automatica
		writer := gzip.NewWriter(outFile)
		defer writer.Close()

		// === LECTURA Y COMPRESION POR BLOQUES ===
		// Usar buffer de 64KB para procesamiento eficiente de archivos grandes
		buf := make([]byte, 64*1024)
		reader := bufio.NewReader(inFile)
		for {
			// === VERIFICACION DE CANCELACION ===
			// Chequear cancelacion en cada iteracion para permitir interrupcion
			select {
			case <-cancelCh:
				writer.Close()
				os.Remove(outName) // Limpiar archivo parcial
				return server.NewResponse(499, "Client Closed Request", "application/json",
					[]byte(`{"error":"compression cancelled"}`))
			default:
			}

			// === LECTURA Y ESCRITURA DE DATOS ===
			// Leer bloque de datos y escribir al archivo comprimido
			n, err := reader.Read(buf)
			if n > 0 {
				writer.Write(buf[:n])
			}
			if err == io.EOF {
				break // Fin del archivo alcanzado
			}
			if err != nil {
				msg := fmt.Sprintf(`{"error":"read error: %v"}`, err)
				return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
			}
		}

	case "xz":
		// === COMPRESION XZ EXTERNA ===
		// Requiere comando `xz` instalado en el sistema
		outName = name + ".xz"
		cmd := exec.Command("xz", "-c", "-z", "-9", name) // Maxima compresion (-9)
		outFile, err = os.Create(outName)
		if err != nil {
			msg := fmt.Sprintf(`{"error":"failed to create output file: %v"}`, err)
			return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
		}
		defer outFile.Close()

		// === CONFIGURACION DE COMANDO EXTERNO ===
		// Redirigir salida del comando xz al archivo de destino
		cmd.Stdout = outFile
		cmd.Stderr = os.Stderr

		// Ejecutar comando xz y manejar errores
		if err := cmd.Run(); err != nil {
			msg := fmt.Sprintf(`{"error":"xz compression failed: %v"}`, err)
			return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
		}

	default:
		// === VALIDACION DE CODEC ===
		// Rechazar codecs no soportados
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid codec: must be gzip or xz"}`))
	}

	// === CALCULO DE METRICAS DE COMPRESION ===
	// Obtener tamaño del archivo comprimido y calcular ratio
	outInfo, _ := os.Stat(outName)
	outputSize := outInfo.Size()

	// === CONSTRUCCION DE RESPUESTA ===
	// Generar respuesta JSON con metricas detalladas de compresion
	data, _ := json.MarshalIndent(map[string]interface{}{
		"file":         name,
		"codec":        codec,
		"output_file":  filepath.Base(outName),
		"input_bytes":  inputSize,
		"output_bytes": outputSize,
		"ratio":        fmt.Sprintf("%.2f", float64(outputSize)/float64(inputSize)),
		"elapsed_ms":   time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
