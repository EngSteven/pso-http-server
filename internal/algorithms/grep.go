package algorithms

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// Grep busca líneas que coincidan con un patrón regex en un archivo.
// Devuelve el número total de coincidencias y las primeras 10 líneas encontradas.
func Grep(name, pattern string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if name == "" || pattern == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameters: name or pattern"}`))
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"invalid regex pattern: %v"}`, err)
		return server.NewResponse(400, "Bad Request", "application/json", []byte(msg))
	}

	file, err := os.Open(name)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"failed to open file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // hasta 10 MB por línea

	matches := 0
	samples := make([]string, 0, 10)

	for scanner.Scan() {
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"operation cancelled while reading"}`))
		default:
		}

		line := scanner.Text()
		if re.MatchString(line) {
			matches++
			if len(samples) < 10 {
				samples = append(samples, line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		msg := fmt.Sprintf(`{"error":"failed while reading: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}

	data, _ := json.MarshalIndent(map[string]interface{}{
		"file":          name,
		"pattern":       pattern,
		"matches":       matches,
		"sample_lines":  samples,
		"elapsed_ms":    time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
