/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: matrixmul.go
Descripcion: Implementa multiplicacion de matrices NxN con datos pseudoaleatorios
calculando hash SHA256 del resultado para verificacion e integridad.
*/

package algorithms

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// MatrixMultiply crea dos matrices NxN aleatorias y calcula su producto.
// Entrada: size (int) - tamaño NxN de las matrices (debe estar entre 1 y 1000)
//
//	seed (int64) - semilla para generacion pseudoaleatoria reproducible
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con hash del resultado o error
// Descripcion: Genera matrices A y B con valores aleatorios usando semilla,
//
//	calcula producto C = A * B y genera hash SHA256 del resultado
//	para verificacion. Maneja cancelacion durante multiplicacion.
func MatrixMultiply(size int, seed int64, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETROS ===
	// Verificar que el tamaño de matriz este en rango valido
	if size <= 0 || size > 1000 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: size must be between 1 and 1000"}`))
	}

	// === INICIALIZACION DE GENERADOR ALEATORIO ===
	// Configurar semilla para reproducibilidad de resultados
	rand.Seed(seed)

	// === CREACION DE MATRICES DE ENTRADA ===
	// Crear matrices A y B con valores pseudoaleatorios
	A := make([][]float64, size)
	B := make([][]float64, size)
	for i := 0; i < size; i++ {
		A[i] = make([]float64, size)
		B[i] = make([]float64, size)
		// === POBLACION DE MATRICES ===
		// Llenar cada posicion con valores aleatorios 0-10
		for j := 0; j < size; j++ {
			A[i][j] = rand.Float64() * 10
			B[i][j] = rand.Float64() * 10
		}
	}

	// === INICIALIZACION DE MATRIZ RESULTADO ===
	// Crear matriz C para almacenar producto A * B
	C := make([][]float64, size)
	for i := 0; i < size; i++ {
		C[i] = make([]float64, size)
	}

	// === MULTIPLICACION DE MATRICES ===
	// Implementar algoritmo estandar C[i][j] = Σ(A[i][k] * B[k][j])
	for i := 0; i < size; i++ {
		// === VERIFICACION DE CANCELACION POR FILA ===
		// Permitir cancelacion entre filas para operaciones grandes
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"matrix multiplication cancelled"}`))
		default:
			// === CALCULO DE FILA DE MATRIZ RESULTADO ===
			// Calcular cada elemento C[i][j] de la fila i
			for j := 0; j < size; j++ {
				sum := 0.0
				// === PRODUCTO PUNTO ===
				// Calcular suma de productos A[i][k] * B[k][j]
				for k := 0; k < size; k++ {
					C[i][j] += A[i][k] * B[k][j]
				}
				_ = sum // Variable no usada, mantener para compatibilidad
			}
		}
	}

	// === CALCULO DE HASH DE VERIFICACION ===
	// Generar hash SHA-256 del resultado para integridad
	h := sha256.New()
	for i := range C {
		for j := range C[i] {
			// === CONVERSION A BYTES CONSISTENTE ===
			// Formatear cada valor con precision fija para hash reproducible
			valBytes := []byte(fmt.Sprintf("%.6f", C[i][j]))
			h.Write(valBytes)
		}
	}
	// === CODIFICACION HEXADECIMAL ===
	// Convertir hash binario a representacion hexadecimal
	hashHex := hex.EncodeToString(h.Sum(nil))

	// === CONSTRUCCION DE RESPUESTA ===
	// Preparar respuesta JSON con metadatos y hash de resultado
	data, _ := json.MarshalIndent(map[string]interface{}{
		"size":        size,                             // Dimension NxN de las matrices
		"seed":        seed,                             // Semilla usada para generacion
		"hash_sha256": hashHex,                          // Hash SHA256 del resultado
		"elapsed_ms":  time.Since(start).Milliseconds(), // Tiempo de procesamiento
	}, "", "  ")

	// === RESPUESTA EXITOSA ===
	// Retornar resultado con codigo 200 y datos JSON
	return server.NewResponse(200, "OK", "application/json", data)
}
