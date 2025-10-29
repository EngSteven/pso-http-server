/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: logger.go
Descripcion: Sistema de logging estructurado en JSON con niveles configurables
y funciones de conveniencia para Info, Warn y Error.
*/

package util

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// LogLevel define niveles de severidad para el sistema de logging.
type LogLevel string

const (
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// logEntry estructura JSON para entradas de log con timestamp RFC3339.
type logEntry struct {
	Time    string   `json:"time"`
	Level   LogLevel `json:"level"`
	Message string   `json:"message"`
	Fields  any      `json:"fields,omitempty"`
}

var (
	mu       sync.Mutex
	logLevel = LevelInfo
)

// SetLogLevel configura nivel minimo de logging global.
// Entrada: level (string) - nivel de logging ("info", "warn", "error")
// Salida: ninguna
// Descripcion: Establece nivel minimo global de logging. Solo se emitiran
//
//	logs de nivel igual o superior al configurado. Niveles
//	reconocidos: "info" (por defecto), "warn", "error".
func SetLogLevel(level string) {
	switch level {
	case "warn":
		logLevel = LevelWarn
	case "error":
		logLevel = LevelError
	default:
		logLevel = LevelInfo
	}
}

// Log emite entrada de log estructurada en JSON con timestamp.
// Entrada: level (LogLevel) - nivel de severidad del log
//
//	msg (string) - mensaje principal del log
//	fields (any) - campos adicionales para contexto
//
// Salida: ninguna
// Descripcion: Emite entrada de log en formato JSON estructurado con timestamp
//
//	RFC3339Nano, nivel, mensaje y campos opcionales. Thread-safe
//	usando mutex para escritura atomica a stdout.
func Log(level LogLevel, msg string, fields any) {
	mu.Lock()
	defer mu.Unlock()

	entry := logEntry{
		Time:    time.Now().Format(time.RFC3339Nano),
		Level:   level,
		Message: msg,
		Fields:  fields,
	}

	data, _ := json.Marshal(entry)
	fmt.Fprintln(os.Stdout, string(data))
}

// Funciones de conveniencia para niveles de log comunes.
// Info - Entrada: msg (string), fields (any) - Salida: ninguna
//
//	Descripcion: Emite log de nivel INFO para eventos informativos
//
// Warn - Entrada: msg (string), fields (any) - Salida: ninguna
//
//	Descripcion: Emite log de nivel WARN para situaciones de atencion
//
// Error - Entrada: msg (string), fields (any) - Salida: ninguna
//
//	Descripcion: Emite log de nivel ERROR para errores del sistema
func Info(msg string, fields any)  { Log(LevelInfo, msg, fields) }
func Warn(msg string, fields any)  { Log(LevelWarn, msg, fields) }
func Error(msg string, fields any) { Log(LevelError, msg, fields) }
