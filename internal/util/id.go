/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: id.go
Descripcion: Generador de identificadores unicos usando UUID v4
para tracking de requests y jobs en el sistema.
*/

package util

import "github.com/google/uuid"

// NewRequestID genera identificador unico UUID v4 para requests.
// Entrada: ninguna
// Salida: string - UUID v4 como string
// Descripcion: Genera identificador unico usando UUID v4 para tracking
//              de requests HTTP y jobs asincronos. Proporciona trazabilidad
//              unica across el sistema distribuido y logs.
func NewRequestID() string {
	return uuid.NewString()
}
