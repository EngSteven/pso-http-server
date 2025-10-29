/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: isprime.go
Descripcion: Implementa pruebas de primalidad usando division trial y Miller-Rabin
con soporte para cancelacion durante calculos computacionalmente intensivos.
*/

package algorithms

import (
	"crypto/rand"
	"encoding/json"
	"math"
	"math/big"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// IsPrime determina si un numero es primo usando metodos seleccionables.
// Entrada: n (int64) - numero a verificar si es primo (debe ser > 1)
//
//	method (string) - metodo de verificacion ("trial" o "miller")
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con resultado de primalidad o error
// Descripcion: Verifica primalidad usando division trial (deterministico) o
//
//	Miller-Rabin (probabilistico). Division trial verifica hasta raiz
//	cuadrada, Miller-Rabin usa 5 iteraciones para alta confiabilidad.
func IsPrime(n int64, method string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETRO N ===
	// Solo numeros mayores a 1 pueden ser evaluados para primalidad
	if n <= 1 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: n must be > 1"}`))
	}

	// === ESTABLECIMIENTO DE METODO POR DEFECTO ===
	// Usar division trial como metodo predeterminado si no se especifica
	if method == "" {
		method = "trial"
	}

	// === VERIFICACION DE CANCELACION TEMPRANA ===
	// Chequear cancelacion antes de iniciar calculo computacionalmente intensivo
	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled"}`))
	default:
	}

	// === SELECCION Y EJECUCION DE ALGORITMO ===
	// Ejecutar test de primalidad segun metodo especificado
	isPrime := false
	switch method {
	case "trial":
		// Metodo deterministico: division trial hasta raiz cuadrada
		isPrime = trialDivision(n, cancelCh)
	case "miller":
		// Metodo probabilistico: Miller-Rabin con 5 iteraciones
		isPrime = millerRabin(n, 5, cancelCh)
	default:
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid method: must be 'trial' or 'miller'"}`))
	}

	// === CONSTRUCCION DE RESPUESTA ===
	// Generar respuesta JSON con resultado de primalidad y metadatos
	data, _ := json.MarshalIndent(map[string]interface{}{
		"n":          n,
		"method":     method,
		"is_prime":   isPrime,
		"elapsed_ms": time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}

// --- Métodos auxiliares ---

// trialDivision implementa prueba de primalidad deterministico.
// Entrada: n (int64) - numero a verificar
//
//	cancelCh (<-chan struct{}) - canal para cancelacion
//
// Salida: bool - true si es primo, false si no es primo o se cancelo
// Descripcion: Verifica divisibilidad desde 3 hasta raiz cuadrada del numero
//
//	usando solo numeros impares. Maneja cancelacion entre divisores.
func trialDivision(n int64, cancelCh <-chan struct{}) bool {
	// === CASOS ESPECIALES BASICOS ===
	// Manejar casos triviales antes del algoritmo principal
	if n < 2 {
		return false // 0 y 1 no son primos
	}
	if n == 2 {
		return true // 2 es el unico primo par
	}
	if n%2 == 0 {
		return false // Todos los pares > 2 no son primos
	}

	// === CALCULO DE LIMITE DE DIVISION ===
	// Solo necesitamos probar hasta la raiz cuadrada del numero
	limit := int64(math.Sqrt(float64(n)))

	// === DIVISION TRIAL CON NUMEROS IMPARES ===
	// Probar divisibilidad con todos los impares desde 3 hasta limit
	for i := int64(3); i <= limit; i += 2 {
		// === VERIFICACION DE CANCELACION POR DIVISOR ===
		// Permitir cancelacion entre cada divisor probado
		select {
		case <-cancelCh:
			return false // Retornar false si se cancela
		default:
			// === TEST DE DIVISIBILIDAD ===
			// Si n es divisible por i, entonces no es primo
			if n%i == 0 {
				return false
			}
		}
	}
	// === CONFIRMACION DE PRIMALIDAD ===
	// Si no se encontro ningun divisor, el numero es primo
	return true
}

// millerRabin implementa test probabilistico de primalidad.
// Entrada: n (int64) - numero a verificar
//
//	k (int) - numero de iteraciones para confiabilidad
//	cancelCh (<-chan struct{}) - canal para cancelacion
//
// Salida: bool - true si probablemente es primo, false si es compuesto o se cancelo
// Descripcion: Test de Miller-Rabin con k iteraciones usando bases aleatorias.
//
//	Descompone n-1 como 2^r * d y verifica condiciones de primalidad.
//	Mayor k aumenta confiabilidad pero tambien tiempo de ejecucion.
func millerRabin(n int64, k int, cancelCh <-chan struct{}) bool {
	// === CASOS ESPECIALES BASICOS ===
	// Manejar casos triviales antes del algoritmo probabilistico
	if n < 2 {
		return false
	}
	if n == 2 || n == 3 {
		return true
	}
	if n%2 == 0 {
		return false
	}

	// === DESCOMPOSICION DE N-1 COMO 2^r * d ===
	// Escribir n-1 en la forma 2^r * d donde d es impar
	d := n - 1
	r := 0
	for d%2 == 0 {
		d /= 2
		r++
	}

	// === ITERACIONES PROBABILISTICAS ===
	// Ejecutar k rondas del test Miller-Rabin para mayor confiabilidad
	for i := 0; i < k; i++ {
		// === VERIFICACION DE CANCELACION POR ITERACION ===
		// Permitir cancelacion entre cada ronda del test
		select {
		case <-cancelCh:
			return false
		default:
			// === GENERACION DE BASE ALEATORIA ===
			// Generar numero aleatorio a en rango [2, n-2]
			a, _ := rand.Int(rand.Reader, big.NewInt(n-4))
			a.Add(a, big.NewInt(2))

			// === CALCULO DE POTENCIA INICIAL ===
			// Calcular x = a^d mod n
			x := new(big.Int).Exp(a, big.NewInt(d), big.NewInt(n))

			// === PRIMERA CONDICION DE MILLER-RABIN ===
			// Si x ≡ 1 (mod n) o x ≡ n-1 (mod n), continuar con siguiente iteracion
			if x.Cmp(big.NewInt(1)) == 0 || x.Cmp(big.NewInt(n-1)) == 0 {
				continue
			}

			// === ELEVACIONES SUCESIVAS AL CUADRADO ===
			// Elevar x al cuadrado r-1 veces
			cont := false
			for j := 0; j < r-1; j++ {
				x.Exp(x, big.NewInt(2), big.NewInt(n))
				// Si x ≡ n-1 (mod n), el numero pasa esta ronda
				if x.Cmp(big.NewInt(n-1)) == 0 {
					cont = true
					break
				}
			}
			// === VERIFICACION DE FALLO DEL TEST ===
			// Si no se encontro x ≡ n-1 (mod n), el numero es compuesto
			if !cont {
				return false
			}
		}
	}
	// === CONCLUSION PROBABILISTICA ===
	// Si paso todas las k rondas, probablemente es primo
	return true
}
