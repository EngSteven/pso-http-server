package util

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type LogLevel string

const (
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

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

func Info(msg string, fields any)  { Log(LevelInfo, msg, fields) }
func Warn(msg string, fields any)  { Log(LevelWarn, msg, fields) }
func Error(msg string, fields any) { Log(LevelError, msg, fields) }
