/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: pi.go
Descripcion: Calcula aproximaciones de Pi usando el algoritmo de Chudnovsky
con precision configurable y soporte para cancelacion durante iteraciones.
*/

package algorithms

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// CalculatePi calcula Pi con precision especificada usando algoritmo de Chudnovsky.
// Entrada: digits (int) - numero de digitos decimales de precision (1-100000)
//
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con aproximacion de Pi o error
// Descripcion: Calcula Pi usando algoritmo de Chudnovsky con aritmetica de
//
//	precision arbitraria. Converge rapidamente (14 digitos por termino).
//	Limita precision maxima y maneja cancelacion durante iteraciones.
func CalculatePi(digits int, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETROS ===
	// Verificar que el numero de digitos este en rango valido
	if digits <= 0 || digits > 100000 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: digits must be between 1 and 10000"}`))
	}

	// === VERIFICACION DE CANCELACION INICIAL ===
	// Comprobar cancelacion antes de iniciar calculos costosos
	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled before start"}`))
	default:
	}

	// === CONFIGURACION DE PRECISION ===
	// Establecer precision en bits mayor que digitos solicitados
	prec := uint(digits * 4)

	// === CALCULO DE PI CON CHUDNOVSKY ===
	// Llamar algoritmo de Chudnovsky para aproximacion de alta precision
	bigPi := chudnovskyPi(prec, cancelCh)

	// === FORMATEO DE RESULTADO ===
	// Convertir resultado a string con precision especificada
	piStr := bigPi.Text('f', digits)

	// === CONSTRUCCION DE RESPUESTA ===
	// Preparar respuesta JSON con aproximacion y metadatos
	data, _ := json.MarshalIndent(map[string]interface{}{
		"digits":     digits,                           // Numero de digitos calculados
		"approx_pi":  piStr,                            // Aproximacion de Pi como string
		"elapsed_ms": time.Since(start).Milliseconds(), // Tiempo de calculo
	}, "", "  ")

	// === RESPUESTA EXITOSA ===
	// Retornar resultado con codigo 200 y datos JSON
	return server.NewResponse(200, "OK", "application/json", data)
}

// --- Implementación del algoritmo de Chudnovsky (iterativa truncada) ---

// chudnovskyPi implementa el algoritmo de Chudnovsky para calculo de Pi.
// Entrada: prec (uint) - precision en bits para aritmetica de punto flotante
//
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *big.Float - aproximacion de Pi con precision especificada
// Descripcion: Implementa serie de Chudnovsky que converge rapidamente a Pi
//
//	(aproximadamente 14 digitos correctos por termino). Usa aritmetica
//	de precision arbitraria y limita iteraciones para evitar excesos.
func chudnovskyPi(prec uint, cancelCh <-chan struct{}) *big.Float {
	// === CONFIGURACION DE PRECISION ===
	// Establecer precision alta para aritmetica de punto flotante
	bigPi := new(big.Float).SetPrec(prec)
	sum := new(big.Float).SetPrec(prec).SetFloat64(0)

	// === CALCULO DE NUMERO DE TERMINOS ===
	// Determinar iteraciones necesarias basado en precision
	terms := int(prec / 14) // ~14 digitos de precision por termino
	if terms > 200 {
		terms = 200 // === LIMITACION DE ITERACIONES ===
		// Evitar excesos computacionales para precision extrema
	}

	// === SERIE DE CHUDNOVSKY ===
	// Iterar sobre terminos de la serie para aproximar Pi
	for n := 0; n < terms; n++ {
		// === VERIFICACION DE CANCELACION POR TERMINO ===
		// Permitir cancelacion entre iteraciones costosas
		select {
		case <-cancelCh:
			return big.NewFloat(0)
		default:
		}

		// === CALCULO DE SIGNO ALTERNANTE ===
		// (-1)^n para alternar signos en la serie
		sign := int64(1)
		if n%2 != 0 {
			sign = -1
		}

		// === CALCULO DE FACTORIALES ===
		// factorial(6n) / (factorial(3n) * factorial(n)^3)
		f6n := new(big.Int).MulRange(1, int64(6*n)) // 6n!
		f3n := new(big.Int).MulRange(1, int64(3*n)) // 3n!
		fn := new(big.Int).MulRange(1, int64(n))    // n!

		// === CONSTRUCCION DEL DENOMINADOR ===
		// 3n! * (n!)^3 para denominador del termino
		den := new(big.Int).Mul(f3n, new(big.Int).Mul(fn, new(big.Int).Mul(fn, fn)))

		// === CONSTRUCCION DEL NUMERADOR BASE ===
		// 6n! * (-1)^n para numerador inicial
		num := new(big.Int).Mul(f6n, big.NewInt(sign))

		// === CONSTANTES DE CHUDNOVSKY ===
		// Termino lineal: 545140134*n + 13591409
		a := new(big.Int).Mul(big.NewInt(545140134), big.NewInt(int64(n)))
		a.Add(a, big.NewInt(13591409))

		// === POTENCIA DE RAMANUJAN ===
		// 640320^(3n) denominador adicional
		kpow := new(big.Int).Exp(big.NewInt(640320), big.NewInt(int64(3*n)), nil)

		// === CONSTRUCCION DEL TERMINO FINAL ===
		// Combinar todos los componentes del termino n
		termNum := new(big.Float).SetPrec(prec).SetInt(new(big.Int).Mul(num, a))
		termDen := new(big.Float).SetPrec(prec).SetInt(new(big.Int).Mul(den, kpow))
		term := new(big.Float).SetPrec(prec).Quo(termNum, termDen)

		// === ACUMULACION DE SUMA ===
		// Agregar termino actual a la suma total
		sum.Add(sum, term)
	}

	// === FORMULA FINAL DE CHUDNOVSKY ===
	// π ≈ 426880 * sqrt(10005) / sum
	c1 := new(big.Float).SetPrec(prec).SetFloat64(426880) // Constante multiplicativa
	c2 := new(big.Float).SetPrec(prec).SetFloat64(10005)  // Base para raiz cuadrada

	// === CALCULO DE RAIZ CUADRADA ===
	// sqrt(10005) con precision configurada
	sqrtC2 := new(big.Float).Sqrt(c2)

	// === NUMERADOR FINAL ===
	// 426880 * sqrt(10005)
	numer := new(big.Float).Mul(c1, sqrtC2)

	// === DIVISION FINAL ===
	// Calcular Pi = numerador / suma de serie
	bigPi.Quo(numer, sum)

	return bigPi
}
