package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

type FibonacciResponse struct {
	N      int   `json:"n"`
	Series []int `json:"series"`
}

// FibonacciHandler genera la serie de Fibonacci hasta N elementos.
func FibonacciHandler(req *types.Request) *types.Response {
	numStr := req.Query.Get("num")
	if numStr == "" {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"missing parameter: num"}`))
	}

	n, err := strconv.Atoi(numStr)
	if err != nil || n <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"invalid num parameter"}`))
	}

	series := make([]int, n)
	if n > 0 {
		series[0] = 0
	}
	if n > 1 {
		series[1] = 1
		for i := 2; i < n; i++ {
			series[i] = series[i-1] + series[i-2]
		}
	}

	resp := FibonacciResponse{N: n, Series: series}
	body, _ := json.Marshal(resp)
	return server.NewResponse(200, "OK", "application/json", body)
}
