/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: types.go
Descripcion: Define tipos y constantes para el sistema de jobs incluyendo
estados, prioridades y estructura de metadata con timestamps.
*/

package jobs

import "time"

// Job status constants
const (
	// === ESTADOS DEL CICLO DE VIDA DEL JOB ===
	// Estados que representan las fases de ejecución
	StatusQueued   = "queued"   // Job encolado esperando procesamiento
	StatusRunning  = "running"  // Job en ejecución activa
	StatusDone     = "done"     // Job completado exitosamente
	StatusError    = "error"    // Job falló con error
	StatusCanceled = "canceled" // Job cancelado por usuario
	StatusTimeout  = "timeout"  // Job cancelado por timeout
)

// Priority define niveles de prioridad para ejecucion de jobs.
type Priority string

const (
	// === NIVELES DE PRIORIDAD ===
	// Sistema de tres niveles para gestión de carga
	PriorityHigh   Priority = "high"   // Alta prioridad - procesamiento preferente
	PriorityNormal Priority = "normal" // Prioridad normal - flujo estándar
	PriorityLow    Priority = "low"    // Baja prioridad - procesamiento diferido
)

// JobMeta contiene metadata completa y estado actual de un job.
// Se persiste en journal para recuperacion tras reinicio del servidor.
type JobMeta struct {
	// === IDENTIFICACION Y COMANDO ===
	ID      string            `json:"id"`      // ID único del job
	Command string            `json:"command"` // Comando/algoritmo a ejecutar
	Params  map[string]string `json:"params"`  // Parámetros del comando

	// === CONTROL DE EJECUCION ===
	Priority  Priority `json:"priority"`             // Prioridad de procesamiento
	Status    string   `json:"status"`               // Estado actual del job
	TimeoutMs int      `json:"timeout_ms,omitempty"` // Timeout específico del job

	// === RESULTADOS Y ERRORES ===
	Error  string `json:"error,omitempty"`  // Mensaje de error si falló
	Result string `json:"result,omitempty"` // Resultado serializado en JSON

	// === TIMESTAMPS DE SEGUIMIENTO ===
	CreatedAt   time.Time `json:"created_at"`             // Momento de creación
	UpdatedAt   time.Time `json:"updated_at"`             // Última actualización
	SubmittedAt time.Time `json:"submitted_at,omitempty"` // Momento de envío al worker
}
