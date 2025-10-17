package handlers

import (
	"encoding/json"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

func factorialFloat(n int) float64 {
	if n == 0 {
		return 1.0
	}
	result := 1.0
	for i := 2; i <= n; i++ {
		result *= float64(i)
	}
	return result
}

func computePi(digits int) string {
	if digits < 10 {
		digits = 10
	}
	sum := big.NewFloat(0)
	prec := uint(digits * 4)
	sum.SetPrec(prec)

	for k := 0; k < 8; k++ {
		num := new(big.Float).SetPrec(prec).SetFloat64(math.Pow(-1, float64(k)) *
			factorialFloat(6*k) * (13591409 + 545140134*float64(k)))
		den := new(big.Float).SetPrec(prec).SetFloat64(
			factorialFloat(3*k) * math.Pow(factorialFloat(k), 3) * math.Pow(640320, float64(3*k)+1.5))
		term := new(big.Float).Quo(num, den)
		sum.Add(sum, term)
	}

	sqrt := new(big.Float).SetPrec(prec).SetFloat64(math.Sqrt(10005))
	const426880 := new(big.Float).SetPrec(prec).SetFloat64(426880)
	pi := new(big.Float).Quo(const426880.Mul(const426880, sqrt), sum)
	text := pi.Text('f', digits)
	if len(text) > digits+2 {
		text = text[:digits+2]
	}
	return text
}

func PiHandler(req *types.Request) *types.Response {
	dStr := req.Query.Get("digits")
	if dStr == "" {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"missing parameter digits"}`))
	}
	d, err := strconv.Atoi(dStr)
	if err != nil || d <= 0 || d > 5000 {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"invalid digits"}`))
	}

	pool := workers.GetPool("pi")
	job := func(cancelCh <-chan struct{}) *types.Response {
		start := time.Now()
		value := computePi(d)
		elapsed := time.Since(start).Milliseconds()
		body, _ := json.MarshalIndent(map[string]interface{}{
			"digits":     d,
			"value":      value,
			"elapsed_ms": elapsed,
		}, "", "  ")
		return server.NewResponse(200, "OK", "application/json", body)
	}

	if pool == nil {
		return job(nil)
	}

	resp, err := pool.SubmitAndWait(job, workers.PriorityNormal)
	if err != nil {
		return server.NewResponse(503, "Service Unavailable", "application/json", []byte(`{"error":"queue full"}`))
	}
	return resp
}
