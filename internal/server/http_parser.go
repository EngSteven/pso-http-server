package server

import (
	"bufio"
	"fmt"
	"net/url"
	"strings"

	"github.com/EngSteven/pso-http-server/internal/types"
)

func ParseRequest(reader *bufio.Reader) (*types.Request, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error leyendo request line: %v", err)
	}
	line = strings.TrimSpace(line)

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("línea inválida: %s", line)
	}

	method, target, version := parts[0], parts[1], parts[2]
	if version != "HTTP/1.0" && version != "HTTP/1.1" {
		return nil, fmt.Errorf("versión no soportada: %s", version)
	}
	if method != "GET" {
		return nil, fmt.Errorf("solo se soporta GET")
	}

	u, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("URL inválida: %v", err)
	}

	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		colon := strings.Index(line, ":")
		if colon == -1 {
			continue
		}
		key := strings.TrimSpace(line[:colon])
		value := strings.TrimSpace(line[colon+1:])
		headers[strings.ToLower(key)] = value
	}

	req := &types.Request{
		Method:  method,
		Path:    u.Path,
		Query:   u.Query(),
		Headers: headers,
	}
	return req, nil
}
