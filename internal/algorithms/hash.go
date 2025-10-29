/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: hash.go
Descripcion: Calcula hashes criptograficos de texto usando algoritmos MD5, SHA1,
SHA256 y SHA512 retornando todos los valores en formato hexadecimal.
*/

package algorithms

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// HashText calcula multiples hashes criptograficos de un texto.
// Entrada: text (string) - texto del cual calcular hashes
//
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con hashes calculados o error
// Descripcion: Genera simultaneamente hashes MD5, SHA1, SHA256 y SHA512 del texto
//
//	de entrada. Valida que el texto no este vacio, maneja cancelacion
//	y retorna todos los hashes en formato hexadecimal para comparacion.
func HashText(text string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETRO TEXTO ===
	// Verificar que se proporcione texto para procesar
	if text == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: text"}`))
	}

	// === VERIFICACION DE CANCELACION TEMPRANA ===
	// Chequear cancelacion antes de calcular hashes
	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	// === CALCULO SIMULTANEO DE MULTIPLES HASHES ===
	// Generar hashes MD5, SHA1, SHA256 y SHA512 del mismo texto
	md5Sum := md5.Sum([]byte(text))          // Hash MD5 (128 bits)
	sha1Sum := sha1.Sum([]byte(text))        // Hash SHA1 (160 bits)
	sha256Sum := sha256.Sum256([]byte(text)) // Hash SHA256 (256 bits)
	sha512Sum := sha512.Sum512([]byte(text)) // Hash SHA512 (512 bits)

	// === CONSTRUCCION DE RESPUESTA ===
	// Generar respuesta JSON con todos los hashes en formato hexadecimal
	data, _ := json.MarshalIndent(map[string]interface{}{
		"input":      text,
		"md5":        hex.EncodeToString(md5Sum[:]),
		"sha1":       hex.EncodeToString(sha1Sum[:]),
		"sha256":     hex.EncodeToString(sha256Sum[:]),
		"sha512":     hex.EncodeToString(sha512Sum[:]),
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
