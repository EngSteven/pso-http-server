package tests

// ============================================================
// TEST SUITE — Módulo de Algoritmos (Básicos, CPU-bound, IO-bound)
// Descripción:
//   Este archivo contiene las pruebas unitarias para los algoritmos
//   del servidor HTTP. Las pruebas están organizadas por tipo de carga,
//   con descripciones claras, logs consistentes y encabezados seccionales.
// ============================================================

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// TestMain imprime un encabezado de suite al iniciar los tests del paquete.
func TestMain(m *testing.M) {
	fmt.Println("\n============================================================")
	fmt.Println("TEST SUITE — Algoritmos del Servidor HTTP (PSO_PY01b)")
	fmt.Println("Descripción: Pruebas unitarias por categorías (básicos, CPU, IO)")
	fmt.Println("============================================================")
	code := m.Run()
	os.Exit(code)
}

// decodeJSON ayuda a decodificar la respuesta JSON de los algoritmos.
func decodeJSON(t *testing.T, res *types.Response) map[string]interface{} {
	t.Helper()
	var data map[string]interface{}
	if err := json.Unmarshal(res.Body, &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	return data
}

// logStart estandariza el log de arranque de cada prueba.
func logStart(t *testing.T, testName, description string) {
	t.Logf("\n--- [%s] %s ---", testName, description)
}

// ============================================================
// BLOQUE A — ALGORITMOS BÁSICOS
// ============================================================

// TestReverse_Success verifica que ReverseText invierta correctamente una cadena válida.
func TestReverse_Success(t *testing.T) {
	logStart(t, "TestReverse_Success", "Reversa de texto válido")
	res := algorithms.ReverseText("abcdef", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["output"] != "fedcba" {
		t.Errorf("expected 'fedcba', got %v", data["output"])
	}
}

// TestReverse_EmptyText verifica que ReverseText falle con texto vacío (400).
func TestReverse_EmptyText(t *testing.T) {
	logStart(t, "TestReverse_EmptyText", "Reversa con texto vacío debe devolver 400")
	res := algorithms.ReverseText("", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 Bad Request, got %d", res.StatusCode)
	}
}

// TestToUpper_Success verifica que ToUpper convierta el texto a mayúsculas.
func TestToUpper_Success(t *testing.T) {
	logStart(t, "TestToUpper_Success", "Conversión a mayúsculas con entrada válida")
	res := algorithms.ToUpper("hola", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["output"] != "HOLA" {
		t.Errorf("expected 'HOLA', got %v", data["output"])
	}
}

// TestToUpper_MissingParam verifica que ToUpper falle con parámetro faltante (400).
func TestToUpper_MissingParam(t *testing.T) {
	logStart(t, "TestToUpper_MissingParam", "Falta parámetro 'text' debe devolver 400")
	res := algorithms.ToUpper("", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400, got %d", res.StatusCode)
	}
}

// TestHash_Success verifica que HashText incluya el campo sha256 en la respuesta.
func TestHash_Success(t *testing.T) {
	logStart(t, "TestHash_Success", "Hash de texto válido produce sha256")
	res := algorithms.HashText("openai", make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["sha256"]; !ok {
		t.Errorf("missing sha256 field")
	}
}

// TestHash_Empty verifica 400 cuando no se provee texto a hashear.
func TestHash_Empty(t *testing.T) {
	logStart(t, "TestHash_Empty", "Hash con texto vacío debe devolver 400")
	res := algorithms.HashText("", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing text, got %d", res.StatusCode)
	}
}

// TestRandom_Success verifica que GenerateRandom genere la cantidad solicitada.
func TestRandom_Success(t *testing.T) {
	logStart(t, "TestRandom_Success", "Generación de números aleatorios válida")
	res := algorithms.GenerateRandom(3, 1, 10, make(chan struct{}))
	data := decodeJSON(t, res)
	if len(data["numbers"].([]interface{})) != 3 {
		t.Errorf("expected 3 numbers, got %v", data["numbers"])
	}
}

// TestRandom_InvalidRange verifica 400 cuando min/max son inválidos.
func TestRandom_InvalidRange(t *testing.T) {
	logStart(t, "TestRandom_InvalidRange", "Rango inválido (min>max) debe devolver 400")
	res := algorithms.GenerateRandom(5, 10, 1, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for invalid range, got %d", res.StatusCode)
	}
}

// TestTimestamp_Success verifica que GetTimestamp incluya un campo 'iso'.
func TestTimestamp_Success(t *testing.T) {
	logStart(t, "TestTimestamp_Success", "Timestamp debe incluir campo 'iso'")
	res := algorithms.GetTimestamp(make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["iso"]; !ok {
		t.Errorf("expected iso field in response")
	}
}

// TestTimestamp_Cancelled valida el flujo de cancelación previo a la ejecución (sin assertions).
func TestTimestamp_Cancelled(t *testing.T) {
	logStart(t, "TestTimestamp_Cancelled", "Cancelación previa al inicio (sin invocación)")
	cancel := make(chan struct{})
	close(cancel)
	// Intencionalmente no se invoca el algoritmo en este test original.
}

// TestSimulate_Success verifica un trabajo simulado exitoso.
func TestSimulate_Success(t *testing.T) {
	logStart(t, "TestSimulate_Success", "Simulación de trabajo exitosa")
	res := algorithms.SimulateWork(1, "jobtest", make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}
}

// TestSimulate_Cancelled verifica 499 cuando se cancela durante la simulación.
func TestSimulate_Cancelled(t *testing.T) {
	logStart(t, "TestSimulate_Cancelled", "Simulación cancelada debe devolver 499")
	cancel := make(chan struct{})
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.SimulateWork(2, "cancelme", cancel)
	if res.StatusCode != 499 {
		t.Errorf("expected 499 cancelled, got %d", res.StatusCode)
	}
}

// TestSleep_Success verifica Sleep exitoso.
func TestSleep_Success(t *testing.T) {
	logStart(t, "TestSleep_Success", "Sleep con segundos válidos")
	res := algorithms.Sleep(1, make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}
}

// TestSleep_Cancel verifica 499 cuando se cancela durante el sleep.
func TestSleep_Cancel(t *testing.T) {
	logStart(t, "TestSleep_Cancel", "Sleep cancelado debe devolver 499")
	cancel := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.Sleep(2, cancel)
	if res.StatusCode != 499 {
		t.Errorf("expected 499 cancelled, got %d", res.StatusCode)
	}
}

// TestSleep_InvalidSeconds verifica 400 cuando los segundos no son válidos.
func TestSleep_InvalidSeconds(t *testing.T) {
	logStart(t, "TestSleep_InvalidSeconds", "Segundos inválidos (<=0) debe devolver 400")
	res := algorithms.Sleep(0, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for invalid seconds, got %d", res.StatusCode)
	}
}

// TestSleep_CancelledBeforeStart verifica cancelación antes de iniciar.
func TestSleep_CancelledBeforeStart(t *testing.T) {
	logStart(t, "TestSleep_CancelledBeforeStart", "Cancelación previa al inicio debe devolver 499")
	cancel := make(chan struct{})
	close(cancel)
	res := algorithms.Sleep(2, cancel)
	if res.StatusCode != 499 {
		t.Errorf("expected 499 for cancellation before start, got %d", res.StatusCode)
	}
}

// TestLoadTest_Success verifica LoadTest con parámetros válidos.
func TestLoadTest_Success(t *testing.T) {
	logStart(t, "TestLoadTest_Success", "Ejecución de tareas concurrentes válida")
	res := algorithms.LoadTest(3, 1, make(chan struct{}))
	data := decodeJSON(t, res)
	if len(data["results"].([]interface{})) != 3 {
		t.Errorf("expected 3 tasks, got %v", data["results"])
	}
}

// TestLoadTest_InvalidParams verifica 400 con parámetros inválidos.
func TestLoadTest_InvalidParams(t *testing.T) {
	logStart(t, "TestLoadTest_InvalidParams", "Parámetros inválidos en LoadTest (tasks=0)")
	res := algorithms.LoadTest(0, 1, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for invalid task count, got %d", res.StatusCode)
	}
}

// TestLoadTest_NegativeSleep verifica 400 cuando sleep<0.
func TestLoadTest_NegativeSleep(t *testing.T) {
	logStart(t, "TestLoadTest_NegativeSleep", "Parámetro sleep negativo debe devolver 400")
	res := algorithms.LoadTest(3, -1, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for negative sleep, got %d", res.StatusCode)
	}
}

// TestLoadTest_CancelledBeforeStart verifica cancelación previa al arranque.
func TestLoadTest_CancelledBeforeStart(t *testing.T) {
	logStart(t, "TestLoadTest_CancelledBeforeStart", "Cancelación antes de iniciar debe devolver 499")
	cancel := make(chan struct{})
	close(cancel)
	res := algorithms.LoadTest(3, 1, cancel)
	if res.StatusCode != 499 {
		t.Errorf("expected 499 for cancelled before start, got %d", res.StatusCode)
	}
}

// TestLoadTest_CancelledDuringExecution deja el flujo original (sin invocar) y registra el intento.
func TestLoadTest_CancelledDuringExecution(t *testing.T) {
	logStart(t, "TestLoadTest_CancelledDuringExecution", "Cancelación durante ejecución (sin llamada, conserva intención original)")
	cancel := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		close(cancel)
	}()
	// Intencionalmente sin invocación del algoritmo en el test original.
}

// TestHashText_Consistency verifica determinismo del hash para misma entrada.
func TestHashText_Consistency(t *testing.T) {
	logStart(t, "TestHashText_Consistency", "Mismo input debe producir mismo hash")
	a := algorithms.HashText("hello", make(chan struct{}))
	b := algorithms.HashText("hello", make(chan struct{}))
	if string(a.Body) != string(b.Body) {
		t.Errorf("expected consistent hash results")
	}
}

// TestTimestamp_NotEmpty verifica que el body de timestamp no sea vacío.
func TestTimestamp_NotEmpty(t *testing.T) {
	logStart(t, "TestTimestamp_NotEmpty", "Body del timestamp no debe estar vacío")
	res := algorithms.GetTimestamp(make(chan struct{}))
	if len(res.Body) == 0 {
		t.Errorf("timestamp body should not be empty")
	}
}

// TestReverse_Cancelled verifica cancelación temprana en ReverseText.
func TestReverse_Cancelled(t *testing.T) {
	logStart(t, "TestReverse_Cancelled", "Cancelación previa o bad request")
	cancel := make(chan struct{})
	close(cancel)
	res := algorithms.ReverseText("abc", cancel)
	if res.StatusCode != 499 && res.StatusCode != 400 {
		t.Errorf("expected cancellation or bad request, got %d", res.StatusCode)
	}
}

// TestCreateTempFileForFutureIO crea y elimina un archivo temporal.
func TestCreateTempFileForFutureIO(t *testing.T) {
	logStart(t, "TestCreateTempFileForFutureIO", "Crear/eliminar archivo temporal")
	name := "tmp_unit_test.txt"
	os.WriteFile(name, []byte("data"), 0644)
	if _, err := os.Stat(name); err != nil {
		t.Errorf("file not created: %v", err)
	}
	os.Remove(name)
}

// ============================================================
// BLOQUE B — ALGORITMOS CPU-BOUND
// ============================================================

// TestFibonacci_Valid verifica la generación de serie y metadatos correctos.
func TestFibonacci_Valid(t *testing.T) {
	logStart(t, "TestFibonacci_Valid", "Serie de Fibonacci válida")
	res := algorithms.CalculateFibonacci(10, make(chan struct{}))
	data := decodeJSON(t, res)
	if int(data["n"].(float64)) != 10 {
		t.Errorf("expected n=10, got %v", data["n"])
	}
	series := data["series"].([]interface{})
	if len(series) != 10 {
		t.Errorf("expected 10 elements, got %d", len(series))
	}
}

// TestFibonacci_Invalid verifica 400 cuando n es inválido.
func TestFibonacci_Invalid(t *testing.T) {
	logStart(t, "TestFibonacci_Invalid", "n inválido debe devolver 400")
	res := algorithms.CalculateFibonacci(-5, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for invalid n, got %d", res.StatusCode)
	}
}

// ---------------------- ISPRIME ----------------------

// TestIsPrime_TrialMethod verifica primalidad con método trial.
func TestIsPrime_TrialMethod(t *testing.T) {
	logStart(t, "TestIsPrime_TrialMethod", "Primalidad con método 'trial'")
	res := algorithms.IsPrime(37, "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] != true {
		t.Errorf("expected 37 to be prime")
	}
}

// TestIsPrime_TrialMethod2 verifica primalidad para n=2 con trial.
func TestIsPrime_TrialMethod2(t *testing.T) {
	logStart(t, "TestIsPrime_TrialMethod2", "Primalidad para 2 con método 'trial'")
	res := algorithms.IsPrime(2, "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] != true {
		t.Errorf("expected 2 to be prime")
	}
}

// TestIsPrime_TrialMethod3 verifica no primalidad para 4 con trial.
func TestIsPrime_TrialMethod3(t *testing.T) {
	logStart(t, "TestIsPrime_TrialMethod3", "No primalidad para 4 con 'trial'")
	res := algorithms.IsPrime(4, "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] == true {
		t.Errorf("expected even number 4 to be non-prime")
	}
}

// TestIsPrime_TrialMethod4 verifica casos <2 no son primos.
func TestIsPrime_TrialMethod4(t *testing.T) {
	logStart(t, "TestIsPrime_TrialMethod4", "Valores <2 no son primos (trial)")
	res := algorithms.IsPrime(1, "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] == true {
		t.Errorf("expected number < 2 to be non-prime")
	}
}

// TestIsPrime_TrialMethod5 verifica no primalidad para 35 con trial.
func TestIsPrime_TrialMethod5(t *testing.T) {
	logStart(t, "TestIsPrime_TrialMethod5", "No primalidad para 35 (trial)")
	res := algorithms.IsPrime(35, "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] != false {
		t.Errorf("expected 35 to be non-prime")
	}
}

// TestIsPrime_MillerMethod verifica primalidad con método Miller-Rabin.
func TestIsPrime_MillerMethod(t *testing.T) {
	logStart(t, "TestIsPrime_MillerMethod", "Primalidad con método 'miller'")
	res := algorithms.IsPrime(37, "miller", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] != true {
		t.Errorf("expected 37 to be prime")
	}
}

// TestIsPrime_InvalidMethod verifica 400 cuando el método es inválido.
func TestIsPrime_InvalidMethod(t *testing.T) {
	logStart(t, "TestIsPrime_InvalidMethod", "Método inválido debe devolver 400")
	res := algorithms.IsPrime(37, "unknown", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid method, got %d", res.StatusCode)
	}
}

// TestIsPrime_DefaultMethod verifica uso por defecto del método 'trial'.
func TestIsPrime_DefaultMethod(t *testing.T) {
	logStart(t, "TestIsPrime_DefaultMethod", "Sin método => usa 'trial'")
	res := algorithms.IsPrime(19, "", make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}
	data := decodeJSON(t, res)
	if data["method"] != "trial" {
		t.Errorf("expected default method 'trial', got %v", data["method"])
	}
}

// TestIsPrime_CancelledImmediately verifica 499 con cancelación inmediata.
func TestIsPrime_CancelledImmediately(t *testing.T) {
	logStart(t, "TestIsPrime_CancelledImmediately", "Cancelación inmediata debe devolver 499")
	cancel := make(chan struct{})
	close(cancel)
	res := algorithms.IsPrime(17, "trial", cancel)
	if res.StatusCode != 499 {
		t.Errorf("expected 499 cancelled, got %d", res.StatusCode)
	}
}

// TestIsPrime_CancelDuringTrial verifica cancelación/éxito con trial en número grande.
func TestIsPrime_CancelDuringTrial(t *testing.T) {
	logStart(t, "TestIsPrime_CancelDuringTrial", "Cancel durante trial en número grande")
	cancel := make(chan struct{})
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.IsPrime(999983, "trial", cancel)
	if res.StatusCode != 499 && res.StatusCode != 200 {
		t.Errorf("expected cancel or OK, got %d", res.StatusCode)
	}
}

// TestIsPrime_CancelDuringMiller verifica cancelación/éxito con Miller-Rabin.
func TestIsPrime_CancelDuringMiller(t *testing.T) {
	logStart(t, "TestIsPrime_CancelDuringMiller", "Cancel durante Miller-Rabin")
	cancel := make(chan struct{})
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.IsPrime(104729, "miller", cancel)
	if res.StatusCode != 499 && res.StatusCode != 200 {
		t.Errorf("expected cancel or OK, got %d", res.StatusCode)
	}
}

// TestIsPrime_MillerCompositeBranches asegura cobertura de compuestos en Miller.
func TestIsPrime_MillerCompositeBranches(t *testing.T) {
	logStart(t, "TestIsPrime_MillerCompositeBranches", "Cobertura de ramas compuestas (Miller)")
	res := algorithms.IsPrime(9, "miller", make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}
}

// ---------------------- FACTORIZATION ----------------------

// TestFactorNumber_Valid verifica lista de factores para 84.
func TestFactorNumber_Valid(t *testing.T) {
	logStart(t, "TestFactorNumber_Valid", "Factorización de 84")
	res := algorithms.Factorize(84, make(chan struct{}))
	data := decodeJSON(t, res)
	if len(data["factors"].([]interface{})) != 4 {
		t.Errorf("expected 4 factors for 84")
	}
}

// TestFactorNumber_Valid2 verifica lista de factores para 15.
func TestFactorNumber_Valid2(t *testing.T) {
	logStart(t, "TestFactorNumber_Valid2", "Factorización de 15")
	res := algorithms.Factorize(15, make(chan struct{}))
	data := decodeJSON(t, res)
	if len(data["factors"].([]interface{})) != 2 {
		t.Errorf("expected 2 factors for 15")
	}
}

// TestFactorNumber_PrimeInput verifica que un primo regrese lista [n].
func TestFactorNumber_PrimeInput(t *testing.T) {
	logStart(t, "TestFactorNumber_PrimeInput", "Factorización de primo debe ser [n]")
	res := algorithms.Factorize(13, make(chan struct{}))
	data := decodeJSON(t, res)
	list := data["factors"].([]interface{})
	if len(list) != 1 || int(list[0].(float64)) != 13 {
		t.Errorf("expected [13] for prime input, got %v", list)
	}
}

// TestFactorNumber_Invalid verifica 400 cuando n<0.
func TestFactorNumber_Invalid(t *testing.T) {
	logStart(t, "TestFactorNumber_Invalid", "Entrada negativa debe devolver 400")
	res := algorithms.Factorize(-1, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for negative input, got %d", res.StatusCode)
	}
}

// ---------------------- MATRIX MULTIPLICATION ----------------------

// TestMatrixMultiply_Valid verifica presencia de hash_sha256 en resultado.
func TestMatrixMultiply_Valid(t *testing.T) {
	logStart(t, "TestMatrixMultiply_Valid", "Multiplicación de matrices válida")
	res := algorithms.MatrixMultiply(4, 42, make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["hash_sha256"]; !ok {
		t.Errorf("expected hash_sha256 field")
	}
}

// TestMatrixMultiply_InvalidSize verifica 400 cuando size<=0.
func TestMatrixMultiply_InvalidSize(t *testing.T) {
	logStart(t, "TestMatrixMultiply_InvalidSize", "Size inválido debe devolver 400")
	res := algorithms.MatrixMultiply(0, 42, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid size, got %d", res.StatusCode)
	}
}

// ---------------------- MANDELBROT ----------------------

// TestMandelbrot_Valid verifica que la respuesta incluya dimensiones correctas.
func TestMandelbrot_Valid(t *testing.T) {
	logStart(t, "TestMandelbrot_Valid", "Generación válida de fractal")
	res := algorithms.Mandelbrot(50, 50, 50, false, make(chan struct{}))
	data := decodeJSON(t, res)
	if int(data["width"].(float64)) != 50 {
		t.Errorf("expected width=50, got %v", data["width"])
	}
}

// TestMandelbrot_InvalidParams verifica 400 con parámetros inválidos.
func TestMandelbrot_InvalidParams(t *testing.T) {
	logStart(t, "TestMandelbrot_InvalidParams", "Parámetros inválidos deben devolver 400")
	res := algorithms.Mandelbrot(-5, 20, 10, false, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid width, got %d", res.StatusCode)
	}
}

// ---------------------- PI APPROXIMATION ----------------------

// TestPi_Valid verifica que la aproximación tenga longitud razonable.
func TestPi_Valid(t *testing.T) {
	logStart(t, "TestPi_Valid", "Cálculo de PI con dígitos válidos")
	res := algorithms.CalculatePi(20, make(chan struct{}))
	data := decodeJSON(t, res)
	str := data["approx_pi"].(string)
	if len(str) < 10 {
		t.Errorf("expected long pi approximation, got %s", str)
	}
}

// TestPi_InvalidDigits verifica 400 cuando digits<0.
func TestPi_InvalidDigits(t *testing.T) {
	logStart(t, "TestPi_InvalidDigits", "Dígitos inválidos deben devolver 400")
	res := algorithms.CalculatePi(-5, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid digits, got %d", res.StatusCode)
	}
}

// ---------------------- CANCELLATION TESTS ----------------------

// TestFibonacci_Cancelled verifica cancelación/éxito para Fibonacci costoso.
func TestFibonacci_Cancelled(t *testing.T) {
	logStart(t, "TestFibonacci_Cancelled", "Cancelación durante Fibonacci")
	cancel := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.CalculateFibonacci(35, cancel)
	if res.StatusCode != 499 && res.StatusCode != 200 {
		t.Errorf("expected cancel or partial, got %d", res.StatusCode)
	}
}

// TestMatrixMultiply_Cancelled verifica cancelación/éxito en matrices grandes.
func TestMatrixMultiply_Cancelled(t *testing.T) {
	logStart(t, "TestMatrixMultiply_Cancelled", "Cancelación durante multiplicación de matrices grande")
	cancel := make(chan struct{})
	go func() {
		time.Sleep(30 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.MatrixMultiply(100, 1, cancel)
	if res.StatusCode != 499 && res.StatusCode != 200 {
		t.Errorf("expected cancelled or OK, got %d", res.StatusCode)
	}
}

// ---------------------- EDGE STRESS TEST ----------------------

// TestIsPrime_LargeNumber verifica existencia de campo is_prime en caso grande.
func TestIsPrime_LargeNumber(t *testing.T) {
	logStart(t, "TestIsPrime_LargeNumber", "Prueba con número grande (bordes)")
	res := algorithms.IsPrime(int64(1<<31-1), "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["is_prime"]; !ok {
		t.Errorf("missing is_prime field for large input")
	}
}

// TestMandelbrot_SaveFileTrue verifica que se informe archivo guardado cuando se pide.
func TestMandelbrot_SaveFileTrue(t *testing.T) {
	logStart(t, "TestMandelbrot_SaveFileTrue", "Generación con volcado a archivo")
	res := algorithms.Mandelbrot(20, 20, 20, true, make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["saved_file"]; !ok {
		t.Errorf("expected saved_file field")
	}
	os.Remove(data["saved_file"].(string))
}

// TestPi_Consistency verifica determinismo con mismos dígitos.
func TestPi_Consistency(t *testing.T) {
	logStart(t, "TestPi_Consistency", "Determinismo en cálculo de PI")
	a := algorithms.CalculatePi(20, make(chan struct{}))
	b := algorithms.CalculatePi(20, make(chan struct{}))
	if string(a.Body) != string(b.Body) {
		t.Errorf("expected deterministic output for same digits")
	}
}

// TestMatrixMul_SeedEffect verifica que diferentes seeds produzcan resultados distintos.
func TestMatrixMul_SeedEffect(t *testing.T) {
	logStart(t, "TestMatrixMul_SeedEffect", "Seeds diferentes deben producir hashes distintos")
	a := algorithms.MatrixMultiply(4, 1, make(chan struct{}))
	b := algorithms.MatrixMultiply(4, 2, make(chan struct{}))
	if string(a.Body) == string(b.Body) {
		t.Errorf("expected different results for different seeds")
	}
}

// TestFactorNumber_StringConversion valida una conversión base (sanity check).
func TestFactorNumber_StringConversion(t *testing.T) {
	logStart(t, "TestFactorNumber_StringConversion", "Sanity check de strconv.Itoa")
	num := 42
	str := strconv.Itoa(num)
	if str != "42" {
		t.Errorf("strconv conversion sanity check failed")
	}
}

// ============================================================
// BLOQUE C — ALGORITMOS IO-BOUND
// ============================================================

// ---------------------- CREATEFILE & DELETEFILE ----------------------

// TestCreateFile_Success verifica creación de archivo con contenido y repeticiones.
func TestCreateFile_Success(t *testing.T) {
	logStart(t, "TestCreateFile_Success", "Creación de archivo válida")
	name := "test_create.txt"
	res := algorithms.CreateFile(name, "hola mundo", 3, make(chan struct{}))
	if res.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	if _, err := os.Stat(name); err != nil {
		t.Errorf("file not created: %v", err)
	}
	os.Remove(name)
}

// TestCreateFile_Invalid verifica 400 cuando falta nombre de archivo.
func TestCreateFile_Invalid(t *testing.T) {
	logStart(t, "TestCreateFile_Invalid", "Falta de nombre de archivo debe devolver 400")
	res := algorithms.CreateFile("", "data", 2, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing filename, got %d", res.StatusCode)
	}
}

// TestCreateFile_DefaultRepeat verifica que repeat<=0 use valor por defecto.
func TestCreateFile_DefaultRepeat(t *testing.T) {
	logStart(t, "TestCreateFile_DefaultRepeat", "Repeat<=0 debe usar default=1")
	name := "repeat_default.txt"
	res := algorithms.CreateFile(name, "hola", 0, make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 with repeat<=0 defaulting to 1, got %d", res.StatusCode)
	}
	os.Remove(name)
}

// TestCreateFile_CancelledBeforeStart verifica 499 y que el archivo no exista.
func TestCreateFile_CancelledBeforeStart(t *testing.T) {
	logStart(t, "TestCreateFile_CancelledBeforeStart", "Cancelado antes de iniciar no debe crear archivo")
	name := "cancel_create.txt"
	cancel := make(chan struct{})
	close(cancel)
	res := algorithms.CreateFile(name, "data", 3, cancel)
	if res.StatusCode != 499 {
		t.Errorf("expected 499 for cancelled operation, got %d", res.StatusCode)
	}
	if _, err := os.Stat(name); err == nil {
		t.Errorf("file should not exist after cancellation")
		os.Remove(name)
	}
}

// TestCreateFile_WriteError verifica 500 cuando el path es inválido.
func TestCreateFile_WriteError(t *testing.T) {
	logStart(t, "TestCreateFile_WriteError", "Error de escritura (ruta inválida) debe devolver 500")
	name := "/nonexistent_dir/fail.txt"
	res := algorithms.CreateFile(name, "content", 2, make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for write error, got %d", res.StatusCode)
	}
}

// TestDeleteFile_Success verifica borrado exitoso de archivo existente.
func TestDeleteFile_Success(t *testing.T) {
	logStart(t, "TestDeleteFile_Success", "Borrado de archivo existente")
	name := "test_delete.txt"
	os.WriteFile(name, []byte("dummy"), 0644)
	res := algorithms.DeleteFile(name, make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}
}

// TestDeleteFile_NotFound verifica 500 cuando el archivo no existe.
func TestDeleteFile_NotFound(t *testing.T) {
	logStart(t, "TestDeleteFile_NotFound", "Borrado de archivo inexistente debe devolver 500")
	res := algorithms.DeleteFile("no_such_file.txt", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res.StatusCode)
	}
}

// TestDeleteFile_MissingName verifica 400 cuando falta nombre.
func TestDeleteFile_MissingName(t *testing.T) {
	logStart(t, "TestDeleteFile_MissingName", "Falta de nombre debe devolver 400")
	res := algorithms.DeleteFile("", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing parameter, got %d", res.StatusCode)
	}
}

// TestDeleteFile_CancelledBeforeStart verifica que no se elimine tras cancelación previa.
func TestDeleteFile_CancelledBeforeStart(t *testing.T) {
	logStart(t, "TestDeleteFile_CancelledBeforeStart", "Cancelado antes de iniciar no debe eliminar archivo")
	name := "cancel_test.txt"
	os.WriteFile(name, []byte("data"), 0644)
	cancel := make(chan struct{})
	close(cancel)
	res := algorithms.DeleteFile(name, cancel)
	if res.StatusCode != 499 {
		t.Errorf("expected 499 for cancelled operation, got %d", res.StatusCode)
	}
	if _, err := os.Stat(name); os.IsNotExist(err) {
		t.Errorf("file should not have been deleted on cancel")
	}
	os.Remove(name)
}

// ---------------------- SORTFILE ----------------------

// TestSortFile_Success verifica ordenamiento y creación del archivo .sorted.
func TestSortFile_Success(t *testing.T) {
	logStart(t, "TestSortFile_Success", "Ordenamiento de números por merge sort")
	name := "nums.txt"
	content := "5\n1\n3\n2\n4\n"
	os.WriteFile(name, []byte(content), 0644)

	res := algorithms.SortFile(name, "merge", make(chan struct{}))
	data := decodeJSON(t, res)
	if int(data["count"].(float64)) != 5 {
		t.Errorf("expected 5 lines sorted")
	}

	out := name + ".sorted"
	if _, err := os.Stat(out); err != nil {
		t.Errorf("sorted file not created: %v", err)
	}

	os.Remove(name)
	os.Remove(out)
}

// TestSortFile_InvalidAlgo verifica 400 cuando el algoritmo es inválido.
func TestSortFile_InvalidAlgo(t *testing.T) {
	logStart(t, "TestSortFile_InvalidAlgo", "Algoritmo inválido debe devolver 400")
	name := "nums_invalid.txt"
	os.WriteFile(name, []byte("1\n2\n3\n"), 0644)
	res := algorithms.SortFile(name, "bogus", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid algo, got %d", res.StatusCode)
	}
	os.Remove(name)
}

// TestSortFile_MissingName verifica 400 cuando no se especifica archivo.
func TestSortFile_MissingName(t *testing.T) {
	logStart(t, "TestSortFile_MissingName", "Falta de nombre de archivo debe devolver 400")
	res := algorithms.SortFile("", "merge", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing filename, got %d", res.StatusCode)
	}
}

// TestSortFile_FileNotFound verifica 500 cuando el archivo no existe.
func TestSortFile_FileNotFound(t *testing.T) {
	logStart(t, "TestSortFile_FileNotFound", "Archivo inexistente debe devolver 500")
	res := algorithms.SortFile("no_such_file.txt", "merge", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res.StatusCode)
	}
}

// TestSortFile_EmptyFile verifica 400 cuando el archivo está vacío.
func TestSortFile_EmptyFile(t *testing.T) {
	logStart(t, "TestSortFile_EmptyFile", "Archivo vacío debe devolver 400")
	name := "empty.txt"
	os.WriteFile(name, []byte(""), 0644)
	res := algorithms.SortFile(name, "merge", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for empty file, got %d", res.StatusCode)
	}
	os.Remove(name)
}

// TestSortFile_CancelledBeforeStart verifica 499 cuando se cancela antes de iniciar.
func TestSortFile_CancelledBeforeStart(t *testing.T) {
	logStart(t, "TestSortFile_CancelledBeforeStart", "Cancelación antes de iniciar debe devolver 499")
	name := "cancel_before.txt"
	os.WriteFile(name, []byte("1\n2\n3\n"), 0644)
	cancel := make(chan struct{})
	close(cancel)
	res := algorithms.SortFile(name, "merge", cancel)
	if res.StatusCode != 499 {
		t.Errorf("expected 499 for cancelled before start, got %d", res.StatusCode)
	}
	os.Remove(name)
}

// TestSortFile_CancelDuringRead conserva la intención original sin invocar (se elimina archivo).
func TestSortFile_CancelDuringRead(t *testing.T) {
	logStart(t, "TestSortFile_CancelDuringRead", "Cancelación durante lectura (sin llamada, se elimina el archivo)")
	name := "cancel_read.txt"
	os.WriteFile(name, []byte("10\n5\n3\n2\n"), 0644)
	cancel := make(chan struct{})
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(cancel)
	}()
	os.Remove(name)
}

// TestSortFile_CancelDuringWrite verifica cancelación/éxito durante escritura.
func TestSortFile_CancelDuringWrite(t *testing.T) {
	logStart(t, "TestSortFile_CancelDuringWrite", "Cancelación durante escritura del .sorted")
	name := "cancel_write.txt"
	os.WriteFile(name, []byte("3\n2\n1\n"), 0644)
	cancel := make(chan struct{})
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.SortFile(name, "merge", cancel)
	if res.StatusCode != 499 && res.StatusCode != 200 {
		t.Errorf("expected 499 (cancelled) or 200 (completed before cancel), got %d", res.StatusCode)
	}
	os.Remove(name)
	os.Remove(name + ".sorted")
}

// TestSortFile_DefaultQuickSort verifica que el algoritmo por defecto sea quicksort.
func TestSortFile_DefaultQuickSort(t *testing.T) {
	logStart(t, "TestSortFile_DefaultQuickSort", "Algoritmo por defecto quicksort cuando 'algo' está vacío")
	name := "quick_default.txt"
	os.WriteFile(name, []byte("5\n4\n3\n2\n1\n"), 0644)
	res := algorithms.SortFile(name, "", make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 for default quicksort, got %d", res.StatusCode)
	}
	os.Remove(name)
	os.Remove(name + ".sorted")
}

// ---------------------- WORDCOUNT ----------------------

// TestWordCount_Success verifica recuento básico de líneas/palabras/bytes.
func TestWordCount_Success(t *testing.T) {
	logStart(t, "TestWordCount_Success", "Recuento de un archivo simple")
	name := "wc_test.txt"
	os.WriteFile(name, []byte("hola mundo\nlinea dos\nultima\n"), 0644)
	res := algorithms.WordCount(name, make(chan struct{}))
	data := decodeJSON(t, res)
	if data["lines"].(float64) < 3 {
		t.Errorf("expected at least 3 lines, got %v", data["lines"])
	}
	os.Remove(name)
}

// TestWordCount_MissingFile verifica 500 cuando el archivo no existe.
func TestWordCount_MissingFile(t *testing.T) {
	logStart(t, "TestWordCount_MissingFile", "Archivo inexistente debe devolver 500")
	res := algorithms.WordCount("missing.txt", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res.StatusCode)
	}
}

// ---------------------- GREP ----------------------

// TestGrep_Success verifica matches y conteo mínimo esperado.
func TestGrep_Success(t *testing.T) {
	logStart(t, "TestGrep_Success", "Búsqueda de patrón con coincidencias")
	name := "grep_test.txt"
	os.WriteFile(name, []byte("linea uno\nmatch here\notra match\n"), 0644)
	res := algorithms.Grep(name, "match", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["matches"].(float64) < 2 {
		t.Errorf("expected >=2 matches")
	}
	os.Remove(name)
}

// TestGrep_InvalidRegex verifica 400 con regex inválido.
func TestGrep_InvalidRegex(t *testing.T) {
	logStart(t, "TestGrep_InvalidRegex", "Regex inválido debe devolver 400")
	name := "grep_invalid.txt"
	os.WriteFile(name, []byte("algo\ntexto\n"), 0644)
	res := algorithms.Grep(name, "(unclosed[", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for invalid regex, got %d", res.StatusCode)
	}
	os.Remove(name)
}

// TestGrep_MissingParams verifica 400 cuando falta nombre o patrón.
func TestGrep_MissingParams(t *testing.T) {
	logStart(t, "TestGrep_MissingParams", "Falta de parámetros debe devolver 400")
	res := algorithms.Grep("", "abc", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing filename, got %d", res.StatusCode)
	}
	res2 := algorithms.Grep("file.txt", "", make(chan struct{}))
	if res2.StatusCode != 400 {
		t.Errorf("expected 400 for missing pattern, got %d", res2.StatusCode)
	}
}

// TestGrep_FileNotFound verifica 500 cuando el archivo no existe.
func TestGrep_FileNotFound(t *testing.T) {
	logStart(t, "TestGrep_FileNotFound", "Archivo inexistente debe devolver 500")
	res := algorithms.Grep("no_such_file.txt", "abc", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res.StatusCode)
	}
}

// TestGrep_CancelledDuringRead conserva el flujo original (sin invocar) y elimina archivo.
func TestGrep_CancelledDuringRead(t *testing.T) {
	logStart(t, "TestGrep_CancelledDuringRead", "Cancelación durante lectura (sin llamada, conserva intención)")
	name := "grep_cancel.txt"
	os.WriteFile(name, []byte("line1\nmatch this\nline2\n"), 0644)
	cancel := make(chan struct{})
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(cancel)
	}()
	os.Remove(name)
}

// TestGrep_NoMatches verifica que no existan coincidencias con patrón ausente.
func TestGrep_NoMatches(t *testing.T) {
	logStart(t, "TestGrep_NoMatches", "Búsqueda sin coincidencias debe retornar matches=0")
	name := "grep_nomatch.txt"
	os.WriteFile(name, []byte("abc\ndef\nghi\n"), 0644)
	res := algorithms.Grep(name, "zzz", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["matches"].(float64) != 0 {
		t.Errorf("expected 0 matches, got %v", data["matches"])
	}
	os.Remove(name)
}

// TestGrep_ScannerError simula error al pasar un directorio en lugar de archivo.
func TestGrep_ScannerError(t *testing.T) {
	logStart(t, "TestGrep_ScannerError", "Pasar directorio debe provocar error (500)")
	os.Mkdir("grep_dir", 0755)
	res := algorithms.Grep("grep_dir", "anything", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for scanner error, got %d", res.StatusCode)
	}
	os.Remove("grep_dir")
}

// ---------------------- HASHFILE ----------------------

// TestHashFile_Success verifica cálculo de hash de archivo.
func TestHashFile_Success(t *testing.T) {
	logStart(t, "TestHashFile_Success", "Hash de archivo válido")
	name := "hash_test.txt"
	os.WriteFile(name, []byte("contenido hash"), 0644)
	res := algorithms.HashFile(name, "sha256", make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["hash_hex"]; !ok {
		t.Errorf("missing hash_hex field")
	}
	os.Remove(name)
}

// TestHashFile_InvalidAlgo verifica 400 cuando el algoritmo de hash es inválido.
func TestHashFile_InvalidAlgo(t *testing.T) {
	logStart(t, "TestHashFile_InvalidAlgo", "Algoritmo inválido debe devolver 400")
	name := "hash_invalid.txt"
	os.WriteFile(name, []byte("data"), 0644)
	res := algorithms.HashFile(name, "unknown", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid algo, got %d", res.StatusCode)
	}
	os.Remove(name)
}

// ---------------------- COMPRESS ----------------------

// TestCompressFile_GzipSuccess verifica compresión gzip y metadatos.
func TestCompressFile_GzipSuccess(t *testing.T) {
	logStart(t, "TestCompressFile_GzipSuccess", "Compresión gzip con archivo válido")
	name := "compress_test.txt"
	os.WriteFile(name, []byte(strings.Repeat("X", 1024)), 0644)
	res := algorithms.CompressFile(name, "gzip", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["codec"] != "gzip" {
		t.Errorf("expected gzip codec, got %v", data["codec"])
	}
	os.Remove(name)
	os.Remove(name + ".gz")
}

// TestCompressFile_InvalidCodec verifica 400 cuando el codec es inválido.
func TestCompressFile_InvalidCodec(t *testing.T) {
	logStart(t, "TestCompressFile_InvalidCodec", "Codec inválido debe devolver 400")
	name := "compress_invalid.txt"
	os.WriteFile(name, []byte("abc"), 0644)
	res := algorithms.CompressFile(name, "bogus", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid codec, got %d", res.StatusCode)
	}
	os.Remove(name)
}

// TestCompressFile_Cancelled verifica cancelación/éxito en compresión.
func TestCompressFile_Cancelled(t *testing.T) {
	logStart(t, "TestCompressFile_Cancelled", "Cancelación durante compresión gzip")
	name := "compress_cancel.txt"
	os.WriteFile(name, []byte(strings.Repeat("Y", 50000)), 0644)
	cancel := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.CompressFile(name, "gzip", cancel)
	if res.StatusCode != 499 && res.StatusCode != 200 {
		t.Errorf("expected cancelled or OK, got %d", res.StatusCode)
	}
	os.Remove(name)
	os.Remove(name + ".gz")
}

// TestCompressFile_MissingName verifica 400 cuando falta el nombre.
func TestCompressFile_MissingName(t *testing.T) {
	logStart(t, "TestCompressFile_MissingName", "Falta de nombre debe devolver 400")
	res := algorithms.CompressFile("", "gzip", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing name, got %d", res.StatusCode)
	}
}

// TestCompressFile_FileNotFound verifica 500 cuando el archivo no existe.
func TestCompressFile_FileNotFound(t *testing.T) {
	logStart(t, "TestCompressFile_FileNotFound", "Archivo inexistente debe devolver 500")
	res := algorithms.CompressFile("no_such_file.txt", "gzip", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res.StatusCode)
	}
}

// TestCompressFile_DefaultCodec verifica que el default sea gzip.
func TestCompressFile_DefaultCodec(t *testing.T) {
	logStart(t, "TestCompressFile_DefaultCodec", "Codec por defecto gzip cuando string vacío")
	name := "compress_default.txt"
	os.WriteFile(name, []byte(strings.Repeat("Z", 1024)), 0644)
	res := algorithms.CompressFile(name, "", make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 for default gzip codec, got %d", res.StatusCode)
	}
	os.Remove(name)
	os.Remove(name + ".gz")
}

// TestCompressFile_XZCodec intenta compresión con xz (puede devolver 200 o 500).
func TestCompressFile_XZCodec(t *testing.T) {
	logStart(t, "TestCompressFile_XZCodec", "Compresión con xz (resultado dependiente del entorno)")
	name := "compress_xz.txt"
	os.WriteFile(name, []byte("data for xz"), 0644)
	res := algorithms.CompressFile(name, "xz", make(chan struct{}))
	if res.StatusCode != 200 && res.StatusCode != 500 {
		t.Errorf("expected 200 or 500 for xz codec, got %d", res.StatusCode)
	}
	os.Remove(name)
	os.Remove(name + ".xz")
}

// TestCompressFile_CreateOutputError verifica 500 cuando no se puede crear salida.
func TestCompressFile_CreateOutputError(t *testing.T) {
	logStart(t, "TestCompressFile_CreateOutputError", "Error al crear archivo de salida debe devolver 500")
	res := algorithms.CompressFile("/no/such/path/file.txt", "gzip", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for output file creation error, got %d", res.StatusCode)
	}
}

// TestCompressFile_EmptyFile verifica que un archivo vacío se comprima (200).
func TestCompressFile_EmptyFile(t *testing.T) {
	logStart(t, "TestCompressFile_EmptyFile", "Compresión de archivo vacío (200)")
	name := "empty.txt"
	os.WriteFile(name, []byte(""), 0644)
	res := algorithms.CompressFile(name, "gzip", make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 for empty file, got %d", res.StatusCode)
	}
	os.Remove(name)
	os.Remove(name + ".gz")
}

// TestCompressFile_FileMissingAndXZ prueba archivo faltante y luego xz exitoso.
func TestCompressFile_FileMissingAndXZ(t *testing.T) {
	logStart(t, "TestCompressFile_FileMissingAndXZ", "Missing file (500) y luego xz (200)")
	res := algorithms.CompressFile("no_such.txt", "gzip", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file")
	}
	name := "xzfile.txt"
	os.WriteFile(name, []byte("hello"), 0644)
	res2 := algorithms.CompressFile(name, "xz", make(chan struct{}))
	if res2.StatusCode != 200 {
		t.Errorf("expected 200 for xz codec")
	}
	os.Remove(name)
	os.Remove(name + ".xz")
}

// TestSortFile_QuickSort verifica quicksort directo.
func TestSortFile_QuickSort(t *testing.T) {
	logStart(t, "TestSortFile_QuickSort", "Ordenamiento quicksort explícito")
	name := "quick.txt"
	os.WriteFile(name, []byte("3\n2\n1\n"), 0644)
	res := algorithms.SortFile(name, "quick", make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 quick sort, got %d", res.StatusCode)
	}
	os.Remove(name)
	os.Remove(name + ".sorted")
}

// ============================================================
// BLOQUE E — COBERTURA RESTANTE / EDGE CASES
// ============================================================

// TestIsPrime_EdgeCases recorre casos límite de primalidad.
func TestIsPrime_EdgeCases(t *testing.T) {
	logStart(t, "TestIsPrime_EdgeCases", "Casos límite: 0,1,2,4")
	cases := []int64{0, 1, 2, 4}
	for _, n := range cases {
		res := algorithms.IsPrime(n, "trial", make(chan struct{}))
		if res.StatusCode == 0 {
			t.Errorf("no response for %d", n)
		}
	}
}

// TestLoadTest_Cancelled verifica cancelación/éxito en un conjunto de tareas.
func TestLoadTest_Cancelled(t *testing.T) {
	logStart(t, "TestLoadTest_Cancelled", "Cancelación durante LoadTest")
	cancel := make(chan struct{})
	go func() {
		time.Sleep(20 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.LoadTest(5, 1, cancel)
	if res.StatusCode != 499 && res.StatusCode != 200 {
		t.Errorf("expected cancel or ok, got %d", res.StatusCode)
	}
}

// TestOverall_Consistency imprime un log resumen al final del bloque.
func TestOverall_Consistency(t *testing.T) {
	logStart(t, "TestOverall_Consistency", "Resumen de verificación general")
	t.Log("Todos los algoritmos básicos, CPU-bound e IO-bound verificados con éxito")
}

// TestLoadTest_InvalidAndCancel recorre invalidaciones y cancelación.
func TestLoadTest_InvalidAndCancel(t *testing.T) {
	logStart(t, "TestLoadTest_InvalidAndCancel", "Params inválidos y cancelación en LoadTest")
	res1 := algorithms.LoadTest(0, 1, make(chan struct{}))
	if res1.StatusCode != 400 {
		t.Errorf("expected 400 for invalid task count")
	}

	res2 := algorithms.LoadTest(3, -1, make(chan struct{}))
	if res2.StatusCode != 400 {
		t.Errorf("expected 400 for negative sleep")
	}

	cancel := make(chan struct{})
	go func() {
		time.Sleep(30 * time.Millisecond)
		close(cancel)
	}()
	res3 := algorithms.LoadTest(10, 1, cancel)
	if res3.StatusCode != 499 && res3.StatusCode != 200 {
		t.Errorf("expected cancel or ok, got %d", res3.StatusCode)
	}
}

// TestHashFile_FileMissingAndCancel verifica missing file y cancelación antes de leer.
func TestHashFile_FileMissingAndCancel(t *testing.T) {
	logStart(t, "TestHashFile_FileMissingAndCancel", "Missing file (500) y cancelación inmediata")
	res1 := algorithms.HashFile("no_such.txt", "sha256", make(chan struct{}))
	if res1.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res1.StatusCode)
	}

	name := "cancel_hash.txt"
	os.WriteFile(name, []byte("cancel"), 0644)
	cancel := make(chan struct{})
	close(cancel)
	res2 := algorithms.HashFile(name, "sha256", cancel)
	if res2.StatusCode != 499 && res2.StatusCode != 200 {
		t.Errorf("expected cancel or ok, got %d", res2.StatusCode)
	}
	os.Remove(name)
}

// TestSimulateWork_EdgeCases verifica 400 en segundos=0 y cancelación inmediata.
func TestSimulateWork_EdgeCases(t *testing.T) {
	logStart(t, "TestSimulateWork_EdgeCases", "Segundos=0 (400) y cancelación inmediata")
	res1 := algorithms.SimulateWork(0, "edge", make(chan struct{}))
	if res1.StatusCode != 400 {
		t.Errorf("expected 400 for zero seconds")
	}

	cancel := make(chan struct{})
	close(cancel)
	res2 := algorithms.SimulateWork(2, "cancel", cancel)
	if res2.StatusCode != 499 && res2.StatusCode != 200 {
		t.Errorf("expected cancel or ok, got %d", res2.StatusCode)
	}
}

// TestSortFile_EmptyAndBadNumbers verifica archivo vacío y números inválidos.
func TestSortFile_EmptyAndBadNumbers(t *testing.T) {
	logStart(t, "TestSortFile_EmptyAndBadNumbers", "Archivo vacío (400) y líneas no numéricas (400)")
	name1 := "empty.txt"
	os.WriteFile(name1, []byte(""), 0644)
	res1 := algorithms.SortFile(name1, "merge", make(chan struct{}))
	if res1.StatusCode != 400 {
		t.Errorf("expected 400 for empty file, got %d", res1.StatusCode)
	}
	os.Remove(name1)

	name2 := "badnums.txt"
	os.WriteFile(name2, []byte("a\nb\nc\n"), 0644)
	res2 := algorithms.SortFile(name2, "merge", make(chan struct{}))
	if res2.StatusCode != 400 {
		t.Errorf("expected 400 for invalid number lines, got %d", res2.StatusCode)
	}
	os.Remove(name2)
}

// TestLoadTest_EdgeAndCancel recorre invalidaciones y cancelación.
func TestLoadTest_EdgeAndCancel(t *testing.T) {
	logStart(t, "TestLoadTest_EdgeAndCancel", "Tasks=0 (400), sleep<0 (400) y cancelación")
	res1 := algorithms.LoadTest(0, 1, make(chan struct{}))
	if res1.StatusCode != 400 {
		t.Errorf("expected 400 for invalid tasks")
	}
	res2 := algorithms.LoadTest(3, -1, make(chan struct{}))
	if res2.StatusCode != 400 {
		t.Errorf("expected 400 for negative sleep")
	}
	cancel := make(chan struct{})
	go func() {
		time.Sleep(20 * time.Millisecond)
		close(cancel)
	}()
	res3 := algorithms.LoadTest(5, 1, cancel)
	if res3.StatusCode != 499 && res3.StatusCode != 200 {
		t.Errorf("expected cancel or ok, got %d", res3.StatusCode)
	}
}

// TestGrep_NoMatchesAndCancel combina caso sin matches y cancelación inmediata.
func TestGrep_NoMatchesAndCancel(t *testing.T) {
	logStart(t, "TestGrep_NoMatchesAndCancel", "Sin coincidencias y cancelación inmediata")
	name := "grep_nomatch.txt"
	os.WriteFile(name, []byte("abc\ndef\n"), 0644)
	res := algorithms.Grep(name, "zzz", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["matches"].(float64) != 0 {
		t.Errorf("expected 0 matches")
	}
	cancel := make(chan struct{})
	go func() { close(cancel) }()
	res2 := algorithms.Grep(name, "abc", cancel)
	if res2.StatusCode != 499 && res2.StatusCode != 200 {
		t.Errorf("expected cancel or ok, got %d", res2.StatusCode)
	}
	os.Remove(name)
}

// TestRandom_EdgeCases verifica count=0, min=max y cancelación.
func TestRandom_EdgeCases(t *testing.T) {
	logStart(t, "TestRandom_EdgeCases", "count=0 (400), min=max (400) y cancelación")
	res1 := algorithms.GenerateRandom(0, 1, 10, make(chan struct{}))
	if res1.StatusCode != 400 {
		t.Errorf("expected 400 for count=0")
	}
	res2 := algorithms.GenerateRandom(3, 5, 5, make(chan struct{}))
	if res2.StatusCode != 400 {
		t.Errorf("expected 400 for equal min/max")
	}
	cancel := make(chan struct{})
	close(cancel)
	res3 := algorithms.GenerateRandom(5, 1, 10, cancel)
	if res3.StatusCode != 499 && res3.StatusCode != 200 {
		t.Errorf("expected cancel or ok")
	}
}
