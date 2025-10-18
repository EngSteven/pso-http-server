package algorithms

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// CalculatePi calcula π usando el algoritmo de Chudnovsky truncado.
func CalculatePi(digits int, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if digits <= 0 || digits > 10000 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: digits must be between 1 and 10000"}`))
	}

	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled before start"}`))
	default:
	}

	// Configurar precisión (un poco mayor que los dígitos solicitados)
	prec := uint(digits * 4)
	bigPi := chudnovskyPi(prec, cancelCh)

	// Formatear el resultado a string truncado
	piStr := bigPi.Text('f', digits)

	data, _ := json.MarshalIndent(map[string]interface{}{
		"digits":     digits,
		"approx_pi":  piStr,
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}

// --- Implementación del algoritmo de Chudnovsky (iterativa truncada) ---

func chudnovskyPi(prec uint, cancelCh <-chan struct{}) *big.Float {
	// Configurar precisión alta
	bigPi := new(big.Float).SetPrec(prec)


	sum := new(big.Float).SetPrec(prec).SetFloat64(0)

	// Número de términos (truncado según precisión)
	terms := int(prec / 14)
	if terms > 200 {
		terms = 200 // limitar iteraciones para evitar excesos
	}

	for n := 0; n < terms; n++ {
		select {
		case <-cancelCh:
			return big.NewFloat(0)
		default:
		}

		// (-1)^n
		sign := int64(1)
		if n%2 != 0 {
			sign = -1
		}

		// factorial(6n) / ((factorial(3n) * factorial(n)^3)
		f6n := new(big.Int).MulRange(1, int64(6*n))
		f3n := new(big.Int).MulRange(1, int64(3*n))
		fn := new(big.Int).MulRange(1, int64(n))

		den := new(big.Int).Mul(f3n, new(big.Int).Mul(fn, new(big.Int).Mul(fn, fn)))
		num := new(big.Int).Mul(f6n, big.NewInt(sign))

		// Constantes grandes
		a := new(big.Int).Mul(big.NewInt(545140134), big.NewInt(int64(n)))
		a.Add(a, big.NewInt(13591409))

		// k4^(3n)
		kpow := new(big.Int).Exp(big.NewInt(640320), big.NewInt(int64(3*n)), nil)

		termNum := new(big.Float).SetPrec(prec).SetInt(new(big.Int).Mul(num, a))
		termDen := new(big.Float).SetPrec(prec).SetInt(new(big.Int).Mul(den, kpow))
		term := new(big.Float).SetPrec(prec).Quo(termNum, termDen)

		sum.Add(sum, term)
	}

	// Calcular π ≈ 426880 * sqrt(10005) / sum
	c1 := new(big.Float).SetPrec(prec).SetFloat64(426880)
	c2 := new(big.Float).SetPrec(prec).SetFloat64(10005)
	sqrtC2 := new(big.Float).Sqrt(c2)
	numer := new(big.Float).Mul(c1, sqrtC2)

	bigPi.Quo(numer, sum)
	return bigPi
}
