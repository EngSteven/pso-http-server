/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: hashfile.go
Descripcion: Calcula hash criptografico de archivos usando algoritmos seleccionables
(SHA256, SHA1, SHA512, MD5) con lectura eficiente por bloques y cancelacion.
*/

package algorithms

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"os"
	"strings"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// HashFile calcula el hash criptografico de un archivo completo.
// Entrada: name (string) - ruta del archivo a procesar
//
//	algo (string) - algoritmo de hash ("sha256", "sha1", "sha512", "md5")
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con hash calculado o error
// Descripcion: Calcula hash criptografico de archivo usando algoritmo seleccionable.
//
//	Lee archivo por bloques de 64KB para eficiencia con archivos grandes.
//	Maneja cancelacion durante lectura y retorna hash en hexadecimal.
func HashFile(name, algo string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETROS ===
	// Verificar que se proporcione nombre de archivo
	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}
	// Establecer algoritmo por defecto si no se especifica
	if algo == "" {
		algo = "sha256"
	}

	// === SELECCION DE ALGORITMO DE HASH ===
	// Configurar hasher segun algoritmo especificado
	var h hash.Hash
	switch strings.ToLower(algo) {
	case "sha256":
		h = sha256.New() // Hash SHA256 (mas comun y seguro)
	case "sha1":
		h = sha1.New() // Hash SHA1 (legacy)
	case "sha512":
		h = sha512.New() // Hash SHA512 (mayor seguridad)
	case "md5":
		h = md5.New() // Hash MD5 (solo para compatibilidad)
	default:
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid algorithm: must be sha256, sha1, sha512, or md5"}`))
	}

	// === APERTURA Y VALIDACION DE ARCHIVO ===
	// Abrir archivo y manejar errores de acceso
	file, err := os.Open(name)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"failed to open file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}
	defer file.Close()

	// === LECTURA Y HASH POR BLOQUES ===
	// Usar buffer de 64KB para archivos grandes y eficiencia de memoria
	reader := bufio.NewReader(file)
	buf := make([]byte, 64*1024) // 64 KB buffer
	var totalBytes int64

	// === PROCESAMIENTO ITERATIVO CON CANCELACION ===
	// Leer archivo por bloques y actualizar hash incrementalmente
	for {
		// Verificar cancelacion en cada bloque leido
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"operation cancelled while reading"}`))
		default:
		}

		// === LECTURA DE BLOQUE Y ACTUALIZACION DE HASH ===
		// Leer siguiente bloque y alimentar al hasher
		n, err := reader.Read(buf)
		if n > 0 {
			h.Write(buf[:n])       // Actualizar hash con datos leidos
			totalBytes += int64(n) // Contar bytes procesados
		}
		if err != nil {
			break // Fin de archivo o error de lectura
		}
	}

	// === FINALIZACION DE HASH ===
	// Obtener hash final en formato hexadecimal
	hashHex := hex.EncodeToString(h.Sum(nil))

	// === CONSTRUCCION DE RESPUESTA ===
	// Generar respuesta JSON con hash y metadatos del procesamiento
	data, _ := json.MarshalIndent(map[string]interface{}{
		"file":       name,
		"algorithm":  algo,
		"hash_hex":   hashHex,
		"bytes_read": totalBytes,
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
