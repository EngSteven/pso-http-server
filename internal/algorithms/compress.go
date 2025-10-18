package algorithms

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// CompressFile comprime un archivo usando gzip o xz y devuelve métricas de tiempo y tamaño.
func CompressFile(name, codec string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}
	if codec == "" {
		codec = "gzip"
	}

	// Verificar que el archivo exista
	inFile, err := os.Open(name)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"failed to open file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}
	defer inFile.Close()

	info, _ := inFile.Stat()
	inputSize := info.Size()

	var outName string
	var outFile *os.File

	switch strings.ToLower(codec) {
	case "gzip":
		outName = name + ".gz"
		outFile, err = os.Create(outName)
		if err != nil {
			msg := fmt.Sprintf(`{"error":"failed to create output file: %v"}`, err)
			return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
		}
		defer outFile.Close()

		writer := gzip.NewWriter(outFile)
		defer writer.Close()

		buf := make([]byte, 64*1024)
		reader := bufio.NewReader(inFile)
		for {
			select {
			case <-cancelCh:
				writer.Close()
				os.Remove(outName)
				return server.NewResponse(499, "Client Closed Request", "application/json",
					[]byte(`{"error":"compression cancelled"}`))
			default:
			}

			n, err := reader.Read(buf)
			if n > 0 {
				writer.Write(buf[:n])
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				msg := fmt.Sprintf(`{"error":"read error: %v"}`, err)
				return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
			}
		}

	case "xz":
		// Requiere comando `xz` instalado en el sistema
		outName = name + ".xz"
		cmd := exec.Command("xz", "-c", "-z", "-9", name)
		outFile, err = os.Create(outName)
		if err != nil {
			msg := fmt.Sprintf(`{"error":"failed to create output file: %v"}`, err)
			return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
		}
		defer outFile.Close()

		cmd.Stdout = outFile
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			msg := fmt.Sprintf(`{"error":"xz compression failed: %v"}`, err)
			return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
		}

	default:
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid codec: must be gzip or xz"}`))
	}

	outInfo, _ := os.Stat(outName)
	outputSize := outInfo.Size()

	data, _ := json.MarshalIndent(map[string]interface{}{
		"file":         name,
		"codec":        codec,
		"output_file":  filepath.Base(outName),
		"input_bytes":  inputSize,
		"output_bytes": outputSize,
		"ratio":        fmt.Sprintf("%.2f", float64(outputSize)/float64(inputSize)),
		"elapsed_ms":   time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
