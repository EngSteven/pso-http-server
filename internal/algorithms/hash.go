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

// HashText calcula hashes comunes (MD5, SHA1, SHA256, SHA512) del texto.
func HashText(text string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if text == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: text"}`))
	}

	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	md5Sum := md5.Sum([]byte(text))
	sha1Sum := sha1.Sum([]byte(text))
	sha256Sum := sha256.Sum256([]byte(text))
	sha512Sum := sha512.Sum512([]byte(text))

	data, _ := json.MarshalIndent(map[string]interface{}{
		"input":       text,
		"md5":         hex.EncodeToString(md5Sum[:]),
		"sha1":        hex.EncodeToString(sha1Sum[:]),
		"sha256":      hex.EncodeToString(sha256Sum[:]),
		"sha512":      hex.EncodeToString(sha512Sum[:]),
		"elapsed_ms":  time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
