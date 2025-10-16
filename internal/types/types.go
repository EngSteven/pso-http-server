package types

import (
	"bytes"
	"fmt"
	"net/url"
)

type Request struct {
	Method  string
	Path    string
	Query   url.Values
	Headers map[string]string
	ID      string
}

type Response struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       []byte
}

type HandlerFunc func(req *Request) *Response

// Bytes serializa la respuesta HTTP en formato texto listo para enviar por red.
func (r *Response) Bytes() []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "HTTP/1.0 %d %s\r\n", r.StatusCode, r.StatusText)
	for k, v := range r.Headers {
		fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
	}
	buf.WriteString("\r\n")
	buf.Write(r.Body)
	return buf.Bytes()
}
