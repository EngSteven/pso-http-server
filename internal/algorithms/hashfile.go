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

// HashFile calcula el hash de un archivo usando el algoritmo indicado (sha256 por defecto).
func HashFile(name, algo string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}
	if algo == "" {
		algo = "sha256"
	}

	var h hash.Hash
	switch strings.ToLower(algo) {
	case "sha256":
		h = sha256.New()
	case "sha1":
		h = sha1.New()
	case "sha512":
		h = sha512.New()
	case "md5":
		h = md5.New()
	default:
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid algorithm: must be sha256, sha1, sha512, or md5"}`))
	}

	file, err := os.Open(name)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"failed to open file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	buf := make([]byte, 64*1024) // 64 KB buffer
	var totalBytes int64

	for {
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"operation cancelled while reading"}`))
		default:
		}

		n, err := reader.Read(buf)
		if n > 0 {
			h.Write(buf[:n])
			totalBytes += int64(n)
		}
		if err != nil {
			break
		}
	}

	hashHex := hex.EncodeToString(h.Sum(nil))

	data, _ := json.MarshalIndent(map[string]interface{}{
		"file":        name,
		"algorithm":   algo,
		"hash_hex":    hashHex,
		"bytes_read":  totalBytes,
		"elapsed_ms":  time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
