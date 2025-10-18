package algorithms

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// WordCount analiza un archivo de texto y devuelve el conteo de líneas, palabras y bytes.
func WordCount(name string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}

	file, err := os.Open(name)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"failed to open file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var (
		lines  int
		words  int
		bytesN int64
	)

	buf := make([]byte, 32*1024) // 32KB buffer

	for {
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"operation cancelled while reading"}`))
		default:
		}

		n, err := reader.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			bytesN += int64(n)
			lines += strings.Count(string(chunk), "\n")
			words += len(strings.Fields(string(chunk)))
		}
		if err != nil {
			break
		}
	}

	data, _ := json.MarshalIndent(map[string]interface{}{
		"file":       name,
		"lines":      lines,
		"words":      words,
		"bytes":      bytesN,
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
