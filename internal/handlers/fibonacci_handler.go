package handlers

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

type FibonacciResponse struct {
	N      int   `json:"n"`
	Series []int `json:"series"`
}

func FibonacciHandler(req *types.Request) *types.Response {
	// parámetros y validación como antes
	numStr := req.Query.Get("num")
	if numStr == "" {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"missing parameter: num"}`))
	}
	n, err := strconv.Atoi(numStr)
	if err != nil || n <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"invalid num parameter"}`))
	}

	// preparar job function
	jobFn := func(cancelCh <-chan struct{}) *types.Response {
		time.Sleep(2 * time.Second)
		
		// Verificar cancelación después del sleep
		select {
		case <-cancelCh:
			return server.NewResponse(500, "Canceled", "application/json", []byte(`{"error":"cancelled"}`))
		default:
		}
		
		series := make([]int, n)
		if n > 0 {
			series[0] = 0
		}
		if n > 1 {
			series[1] = 1
			for i := 2; i < n; i++ {
				// Verificar cancelación en cada iteración
				select {
				case <-cancelCh:
					return server.NewResponse(500, "Canceled", "application/json", []byte(`{"error":"cancelled"}`))
				default:
				}
				series[i] = series[i-1] + series[i-2]
			}
		}
		resp := FibonacciResponse{N: n, Series: series}
		body, _ := json.MarshalIndent(resp, "", "  ")
		return server.NewResponse(200, "OK", "application/json", body)
	}

	// intentamos enviar al pool "fibonacci"
	p := workers.GetPool("fibonacci")
	if p == nil {
		// si no existe pool, ejecutar inline (fallback) - pasar nil como cancelCh
		return jobFn(nil)
	}
	// espera hasta 30s como ejemplo (30000 ms)
	resp, err := p.SubmitAndWait(jobFn, 30000)
	if err == workers.ErrQueueFull {
		return server.NewResponse(503, "Service Unavailable", "application/json", []byte(`{"error":"queue full", "retry_after_ms":1000}`))
	}
	if err == workers.ErrTimeout {
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(`{"error":"job timeout"}`))
	}
	// resp no debe ser nil, pero por seguridad:
	if resp == nil {
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(`{"error":"empty job result"}`))
	}
	return resp
}