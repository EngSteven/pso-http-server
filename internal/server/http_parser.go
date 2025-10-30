/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: http_parser.go
Descripcion: Parser HTTP que procesa requests desde bufio.Reader extrayendo
metodo, URL, headers y query parameters segun especificacion HTTP/1.0-1.1.
*/

package server

import (
	"bufio"
	"fmt"
	"net/url"
	"strings"
	"errors"
	"io"

	"github.com/EngSteven/pso-http-server/internal/types"
)

// ParseRequest parsea request HTTP completo desde reader buffereado.
// Entrada: reader (*bufio.Reader) - reader buffereado con datos HTTP
// Salida: (*types.Request, error) - request parseado o error de parsing
// Descripcion: Parsea request HTTP segun especificacion HTTP/1.0-1.1 extrayendo
//
//	linea de request (metodo, URL, version), headers y query parameters.
//	Valida formato y construye estructura Request para procesamiento.
func ParseRequest(reader *bufio.Reader) (*types.Request, error) {
	// === LECTURA DE REQUEST LINE ===
	line, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			// Cliente cerró la conexión, no es error real
			return nil, nil
		}
		return nil, fmt.Errorf("error leyendo request line: %v", err)
	}
	line = strings.TrimSpace(line)

	// Si la línea viene vacía (por conexiones cerradas sin request)
	if line == "" {
		return nil, nil
	}

	// === PARSING DE COMPONENTES REQUEST LINE ===
	// Dividir en exactamente 3 partes: método URL versión
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("línea inválida: %s", line)
	}

	method, target, version := parts[0], parts[1], parts[2]
	// === VALIDACION DE VERSION HTTP ===
	// Solo soportar HTTP/1.0 y HTTP/1.1
	if version != "HTTP/1.0" && version != "HTTP/1.1" {
		return nil, fmt.Errorf("versión no soportada: %s", version)
	}
	// === VALIDACION DE METODO ===
	// Restricción a solo método GET
	if method != "GET" {
		return nil, fmt.Errorf("solo se soporta GET")
	}

	// === PARSING DE URL Y QUERY PARAMETERS ===
	// Extraer path y query string usando url.Parse
	u, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("URL inválida: %v", err)
	}

	// === LECTURA DE HEADERS HTTP ===
	// Procesar headers línea por línea hasta línea vacía
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		// === DETECCION DE FIN DE HEADERS ===
		// Línea vacía indica fin de headers
		if line == "" {
			break
		}
		// === PARSING DE HEADER KEY-VALUE ===
		// Buscar separador ':' y extraer clave/valor
		colon := strings.Index(line, ":")
		if colon == -1 {
			continue
		}
		key := strings.TrimSpace(line[:colon])
		value := strings.TrimSpace(line[colon+1:])
		// === NORMALIZACION DE HEADERS ===
		// Convertir keys a minúsculas para consistencia
		headers[strings.ToLower(key)] = value
	}

	// === CONSTRUCCION DE REQUEST FINAL ===
	// Crear estructura Request con todos los componentes
	req := &types.Request{
		Method:  method,
		Path:    u.Path,
		Query:   u.Query(),
		Headers: headers,
	}
	return req, nil
}
