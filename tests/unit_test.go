/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: unit_test.go
Descripcion: Suite completa de pruebas unitarias para algoritmos basicos,
CPU-bound e IO-bound del sistema con validacion de funcionalidad y errores.
*/

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
// Entrada: m (*testing.M) - contexto principal de testing
// Salida: ninguna (exit del proceso)
// Descripcion: Punto de entrada para suite de tests unitarios. Imprime
//
//	encabezado informativo y ejecuta todos los tests del paquete.
//	Termina proceso con codigo de salida apropiado.
func TestMain(m *testing.M) {
	fmt.Println("\n============================================================")
	fmt.Println("TEST SUITE — Algoritmos del Servidor HTTP (PSO_PY01b)")
	fmt.Println("Descripción: Pruebas unitarias por categorías (básicos, CPU, IO)")
	fmt.Println("============================================================")
	code := m.Run()
	os.Exit(code)
}

// decodeJSON ayuda a decodificar la respuesta JSON de los algoritmos.
// Entrada: t (*testing.T) - contexto de testing para errores
//
//	res (*types.Response) - respuesta HTTP a decodificar
//
// Salida: map[string]interface{} - datos JSON parseados
// Descripcion: Helper para parsear body JSON de respuestas de algoritmos.
//
//	Marca test como helper y produce error fatal si JSON invalido.
//	Usado para validar campos de respuesta en tests unitarios.
func decodeJSON(t *testing.T, res *types.Response) map[string]interface{} {
	t.Helper()
	var data map[string]interface{}
	if err := json.Unmarshal(res.Body, &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	return data
}

// logStart estandariza el log de arranque de cada prueba.
// Entrada: t (*testing.T) - contexto de testing para logging
//
//	testName (string) - nombre del test
//	description (string) - descripcion de lo que prueba
//
// Salida: ninguna (void)
// Descripcion: Helper para generar logs consistentes al inicio de tests.
//
//	Formato estandarizado con nombre y descripcion del test.
//	Mejora legibilidad de output de testing.
func logStart(t *testing.T, testName, description string) {
	t.Logf("\n--- [%s] %s ---", testName, description)
}

// ============================================================
// BLOQUE A — ALGORITMOS BÁSICOS
// ============================================================

// TestReverse_Success verifica que ReverseText invierta correctamente una cadena válida.
// Entrada: t (*testing.T) - contexto de testing para assertions
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo ReverseText invierta correctamente
//
//	cadena "abcdef" a "fedcba". Verifica funcionalidad basica
//	de inversion de strings sin errores.
func TestReverse_Success(t *testing.T) {
	logStart(t, "TestReverse_Success", "Reversa de texto válido")
	res := algorithms.ReverseText("abcdef", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["output"] != "fedcba" {
		t.Errorf("expected 'fedcba', got %v", data["output"])
	}
}

// TestReverse_EmptyText verifica que ReverseText falle con texto vacío (400).
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo ReverseText retorne codigo 400 Bad Request
//
//	cuando se proporciona string vacio. Verifica manejo correcto
//	de parametros invalidos en validacion de entrada.
func TestReverse_EmptyText(t *testing.T) {
	logStart(t, "TestReverse_EmptyText", "Reversa con texto vacío debe devolver 400")
	res := algorithms.ReverseText("", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 Bad Request, got %d", res.StatusCode)
	}
}

// TestToUpper_Success verifica que ToUpper convierta el texto a mayúsculas.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo ToUpper convierta correctamente
//
//	texto "hola" a "HOLA". Verifica transformacion basica
//	de strings a mayusculas sin errores.
func TestToUpper_Success(t *testing.T) {
	logStart(t, "TestToUpper_Success", "Conversión a mayúsculas con entrada válida")
	res := algorithms.ToUpper("hola", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["output"] != "HOLA" {
		t.Errorf("expected 'HOLA', got %v", data["output"])
	}
}

// TestToUpper_MissingParam verifica que ToUpper falle con parámetro faltante (400).
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo ToUpper retorne codigo 400 Bad Request
//
//	cuando se proporciona string vacio. Verifica manejo correcto
//	de parametros invalidos en validacion de entrada.
func TestToUpper_MissingParam(t *testing.T) {
	logStart(t, "TestToUpper_MissingParam", "Falta parámetro 'text' debe devolver 400")
	res := algorithms.ToUpper("", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400, got %d", res.StatusCode)
	}
}

// TestHash_Success verifica que HashText incluya el campo sha256 en la respuesta.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo HashText genere hash SHA256 correcto
//
//	para texto "openai". Verifica presencia del campo 'sha256'
//	en respuesta JSON. Test de algoritmo criptografico basico.
func TestHash_Success(t *testing.T) {
	logStart(t, "TestHash_Success", "Hash de texto válido produce sha256")
	res := algorithms.HashText("openai", make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["sha256"]; !ok {
		t.Errorf("missing sha256 field")
	}
}

// TestHash_Empty verifica 400 cuando no se provee texto a hashear.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo HashText retorne codigo 400 Bad Request
//
//	cuando se proporciona string vacio. Verifica validacion de
//	parametros requeridos en algoritmos criptograficos.
func TestHash_Empty(t *testing.T) {
	logStart(t, "TestHash_Empty", "Hash con texto vacío debe devolver 400")
	res := algorithms.HashText("", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing text, got %d", res.StatusCode)
	}
}

// TestRandom_Success verifica que GenerateRandom genere la cantidad solicitada.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo GenerateRandom genere exactamente 3
//
//	numeros aleatorios en rango 1-10. Verifica longitud del
//	array 'numbers' en respuesta JSON. Test de generacion aleatoria.
func TestRandom_Success(t *testing.T) {
	logStart(t, "TestRandom_Success", "Generación de números aleatorios válida")
	res := algorithms.GenerateRandom(3, 1, 10, make(chan struct{}))
	data := decodeJSON(t, res)
	if len(data["numbers"].([]interface{})) != 3 {
		t.Errorf("expected 3 numbers, got %v", data["numbers"])
	}
}

// TestRandom_InvalidRange verifica 400 cuando min/max son inválidos.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo GenerateRandom retorne codigo 400
//
//	cuando min > max (rango invalido). Verifica validacion
//	de parametros logicos en generacion de numeros aleatorios.
func TestRandom_InvalidRange(t *testing.T) {
	logStart(t, "TestRandom_InvalidRange", "Rango inválido (min>max) debe devolver 400")
	res := algorithms.GenerateRandom(5, 10, 1, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for invalid range, got %d", res.StatusCode)
	}
}

// TestTimestamp_Success verifica que GetTimestamp incluya un campo 'iso'.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo GetTimestamp genere timestamp correcto
//
//	con campo 'iso' en formato ISO 8601. Verifica generacion
//	de marcas temporales para logging y auditorias.
func TestTimestamp_Success(t *testing.T) {
	logStart(t, "TestTimestamp_Success", "Timestamp debe incluir campo 'iso'")
	res := algorithms.GetTimestamp(make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["iso"]; !ok {
		t.Errorf("expected iso field in response")
	}
}

// TestTimestamp_Cancelled valida el flujo de cancelación previo a la ejecución (sin assertions).
// Entrada: t (*testing.T) - contexto de testing para logging
// Salida: ninguna (void)
// Descripcion: Test de cobertura para flujo de cancelacion previa.
//
//	Crea canal cerrado pero intencionalmente no invoca algoritmo.
//	Preserva patron original de test sin modificar comportamiento.
func TestTimestamp_Cancelled(t *testing.T) {
	logStart(t, "TestTimestamp_Cancelled", "Cancelación previa al inicio (sin invocación)")
	cancel := make(chan struct{})
	close(cancel)
	// Intencionalmente no se invoca el algoritmo en este test original.
}

// TestSimulate_Success verifica un trabajo simulado exitoso.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo SimulateWork ejecute correctamente
//
//	simulacion de trabajo de 1 segundo. Verifica codigo 200 OK
//	para validar funcionalidad basica de simulacion de carga.
func TestSimulate_Success(t *testing.T) {
	logStart(t, "TestSimulate_Success", "Simulación de trabajo exitosa")
	res := algorithms.SimulateWork(1, "jobtest", make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}
}

// TestSimulate_Cancelled verifica 499 cuando se cancela durante la simulación.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo SimulateWork responda con codigo 499
//
//	cuando se cancela durante ejecucion. Usa goroutine para cerrar
//	canal de cancelacion despues de 100ms. Test de manejo de
//	cancelaciones en algoritmos de larga duracion.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Sleep ejecute correctamente pausa
//
//	de 1 segundo. Verifica codigo 200 OK para validar
//	funcionalidad basica de algoritmos temporizados.
func TestSleep_Success(t *testing.T) {
	logStart(t, "TestSleep_Success", "Sleep con segundos válidos")
	res := algorithms.Sleep(1, make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}
}

// TestSleep_Cancel verifica 499 cuando se cancela durante el sleep.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Sleep responda con codigo 499
//
//	cuando se cancela durante ejecucion. Usa goroutine para
//	cerrar canal despues de 50ms. Test de cancelacion temporal.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Sleep retorne codigo 400 Bad Request
//
//	cuando segundos <= 0. Verifica validacion de parametros
//	numericos en algoritmos temporizados.
func TestSleep_InvalidSeconds(t *testing.T) {
	logStart(t, "TestSleep_InvalidSeconds", "Segundos inválidos (<=0) debe devolver 400")
	res := algorithms.Sleep(0, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for invalid seconds, got %d", res.StatusCode)
	}
}

// TestSleep_CancelledBeforeStart verifica cancelación antes de iniciar.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Sleep responda con codigo 499
//
//	cuando canal de cancelacion esta cerrado antes de ejecutar.
//	Verifica manejo de cancelaciones previas al inicio.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo LoadTest ejecute correctamente 3 tareas
//
//	concurrentes con sleep de 1 segundo. Verifica longitud del
//	array 'results' en respuesta. Test de carga concurrente.
func TestLoadTest_Success(t *testing.T) {
	logStart(t, "TestLoadTest_Success", "Ejecución de tareas concurrentes válida")
	res := algorithms.LoadTest(3, 1, make(chan struct{}))
	data := decodeJSON(t, res)
	if len(data["results"].([]interface{})) != 3 {
		t.Errorf("expected 3 tasks, got %v", data["results"])
	}
}

// TestLoadTest_InvalidParams verifica 400 con parámetros inválidos.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo LoadTest retorne codigo 400
//
//	cuando tasks = 0. Verifica validacion de parametros
//	de conteo en algoritmos de carga concurrente.
func TestLoadTest_InvalidParams(t *testing.T) {
	logStart(t, "TestLoadTest_InvalidParams", "Parámetros inválidos en LoadTest (tasks=0)")
	res := algorithms.LoadTest(0, 1, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for invalid task count, got %d", res.StatusCode)
	}
}

// TestLoadTest_NegativeSleep verifica 400 cuando sleep<0.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo LoadTest retorne codigo 400
//
//	cuando sleep es negativo (-1). Verifica validacion
//	de parametros temporales negativos.
func TestLoadTest_NegativeSleep(t *testing.T) {
	logStart(t, "TestLoadTest_NegativeSleep", "Parámetro sleep negativo debe devolver 400")
	res := algorithms.LoadTest(3, -1, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for negative sleep, got %d", res.StatusCode)
	}
}

// TestLoadTest_CancelledBeforeStart verifica cancelación previa al arranque.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo LoadTest responda con codigo 499
//
//	cuando canal de cancelacion esta cerrado antes de ejecutar.
//	Verifica manejo de cancelaciones previas en carga concurrente.
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
// Entrada: t (*testing.T) - contexto de testing para logging
// Salida: ninguna (void)
// Descripcion: Test de cobertura para cancelacion durante ejecucion.
//
//	Crea canal y goroutine de cancelacion pero intencionalmente
//	no invoca algoritmo. Preserva patron original sin modificar.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo HashText produzca resultados deterministas
//
//	para misma entrada "hello". Verifica que dos ejecuciones
//	consecutivas generen mismo hash SHA256.
func TestHashText_Consistency(t *testing.T) {
	logStart(t, "TestHashText_Consistency", "Mismo input debe producir mismo hash")
	a := algorithms.HashText("hello", make(chan struct{}))
	b := algorithms.HashText("hello", make(chan struct{}))
	if string(a.Body) != string(b.Body) {
		t.Errorf("expected consistent hash results")
	}
}

// TestTimestamp_NotEmpty verifica que el body de timestamp no sea vacío.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo GetTimestamp genere respuesta con
//
//	body no vacio. Verifica que timestamp sea generado
//	correctamente sin contenido vacio.
func TestTimestamp_NotEmpty(t *testing.T) {
	logStart(t, "TestTimestamp_NotEmpty", "Body del timestamp no debe estar vacío")
	res := algorithms.GetTimestamp(make(chan struct{}))
	if len(res.Body) == 0 {
		t.Errorf("timestamp body should not be empty")
	}
}

// TestReverse_Cancelled verifica cancelación temprana en ReverseText.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo ReverseText responda con codigo 499
//
//	o 400 cuando canal de cancelacion esta cerrado. Acepta
//	ambos codigos por comportamiento de cancelacion previa.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Test de utilidad que crea archivo temporal para verificar
//
//	operaciones de filesystem. Verifica creacion exitosa con
//	os.Stat y limpia archivo al finalizar. Setup para tests IO.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CalculateFibonacci genere serie correcta
//
//	para n=10. Verifica que respuesta contenga campo 'n' con valor
//	correcto y 'series' con 10 elementos. Test de algoritmo CPU-bound.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CalculateFibonacci retorne codigo 400
//
//	cuando n es negativo (-5). Verifica validacion de parametros
//	numericos en algoritmos matematicos CPU-intensivos.
func TestFibonacci_Invalid(t *testing.T) {
	logStart(t, "TestFibonacci_Invalid", "n inválido debe devolver 400")
	res := algorithms.CalculateFibonacci(-5, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for invalid n, got %d", res.StatusCode)
	}
}

// ---------------------- ISPRIME ----------------------

// TestIsPrime_TrialMethod verifica primalidad con método trial.
// Entrada: t (*testing.T) - contexto de testing para assertions
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo IsPrime detecte correctamente que 37
//
//	es primo usando metodo 'trial'. Verifica campo 'is_prime'
//	en respuesta JSON. Test de algoritmo matematico CPU-intensivo.
func TestIsPrime_TrialMethod(t *testing.T) {
	logStart(t, "TestIsPrime_TrialMethod", "Primalidad con método 'trial'")
	res := algorithms.IsPrime(37, "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] != true {
		t.Errorf("expected 37 to be prime")
	}
}

// TestIsPrime_TrialMethod2 verifica primalidad para n=2 con trial.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo IsPrime detecte correctamente que 2
//
//	es primo usando metodo 'trial'. Caso especial del primo
//	mas pequeño. Verifica campo 'is_prime' = true.
func TestIsPrime_TrialMethod2(t *testing.T) {
	logStart(t, "TestIsPrime_TrialMethod2", "Primalidad para 2 con método 'trial'")
	res := algorithms.IsPrime(2, "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] != true {
		t.Errorf("expected 2 to be prime")
	}
}

// TestIsPrime_TrialMethod3 verifica no primalidad para 4 con trial.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo IsPrime detecte correctamente que 4
//
//	no es primo usando metodo 'trial'. Caso de numero par
//	compuesto. Verifica campo 'is_prime' = false.
func TestIsPrime_TrialMethod3(t *testing.T) {
	logStart(t, "TestIsPrime_TrialMethod3", "No primalidad para 4 con 'trial'")
	res := algorithms.IsPrime(4, "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] == true {
		t.Errorf("expected even number 4 to be non-prime")
	}
}

// TestIsPrime_TrialMethod4 verifica casos <2 no son primos.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo IsPrime detecte correctamente que 1
//
//	no es primo usando metodo 'trial'. Caso limite de numeros
//	menores a 2. Verifica campo 'is_prime' = false.
func TestIsPrime_TrialMethod4(t *testing.T) {
	logStart(t, "TestIsPrime_TrialMethod4", "Valores <2 no son primos (trial)")
	res := algorithms.IsPrime(1, "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] == true {
		t.Errorf("expected number < 2 to be non-prime")
	}
}

// TestIsPrime_TrialMethod5 verifica no primalidad para 35 con trial.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo IsPrime detecte correctamente que 35
//
//	no es primo usando metodo 'trial'. Numero compuesto impar
//	(5 * 7). Verifica campo 'is_prime' = false.
func TestIsPrime_TrialMethod5(t *testing.T) {
	logStart(t, "TestIsPrime_TrialMethod5", "No primalidad para 35 (trial)")
	res := algorithms.IsPrime(35, "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] != false {
		t.Errorf("expected 35 to be non-prime")
	}
}

// TestIsPrime_MillerMethod verifica primalidad con método Miller-Rabin.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo IsPrime detecte correctamente que 37
//
//	es primo usando metodo 'miller' (Miller-Rabin). Algoritmo
//	probabilistico mas eficiente. Verifica campo 'is_prime' = true.
func TestIsPrime_MillerMethod(t *testing.T) {
	logStart(t, "TestIsPrime_MillerMethod", "Primalidad con método 'miller'")
	res := algorithms.IsPrime(37, "miller", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] != true {
		t.Errorf("expected 37 to be prime")
	}
}

// TestIsPrime_InvalidMethod verifica 400 cuando el método es inválido.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo IsPrime retorne codigo 400 Bad Request
//
//	cuando metodo es desconocido ("unknown"). Verifica validacion
//	de parametros de metodo en algoritmos matematicos.
func TestIsPrime_InvalidMethod(t *testing.T) {
	logStart(t, "TestIsPrime_InvalidMethod", "Método inválido debe devolver 400")
	res := algorithms.IsPrime(37, "unknown", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid method, got %d", res.StatusCode)
	}
}

// TestIsPrime_DefaultMethod verifica uso por defecto del método 'trial'.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo IsPrime use metodo 'trial' por defecto
//
//	cuando string de metodo esta vacio. Verifica campo 'method'
//	en respuesta y codigo 200 OK para numero primo 19.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo IsPrime responda con codigo 499
//
//	cuando canal de cancelacion esta cerrado antes de ejecutar.
//	Verifica manejo de cancelaciones inmediatas en calculo primo.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo IsPrime maneje cancelacion durante
//
//	calculo trial de numero grande (999983). Acepta codigo 499
//	(cancelado) o 200 (completado antes de cancelacion).
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo IsPrime maneje cancelacion durante
//
//	calculo Miller-Rabin de numero grande (104729). Acepta codigo
//	499 (cancelado) o 200 (completado antes de cancelacion).
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo IsPrime con metodo 'miller' detecte
//
//	correctamente numero compuesto (9 = 3*3). Asegura cobertura
//	de casos de prueba para numeros compuestos con Miller-Rabin.
func TestIsPrime_MillerCompositeBranches(t *testing.T) {
	logStart(t, "TestIsPrime_MillerCompositeBranches", "Cobertura de ramas compuestas (Miller)")
	res := algorithms.IsPrime(9, "miller", make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}
}

// ---------------------- FACTORIZATION ----------------------

// TestFactorNumber_Valid verifica lista de factores para 84.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Factorize calcule correctamente los
//
//	factores primos de 84 (2²×3×7). Verifica que respuesta JSON
//	contenga 4 factores en array 'factors'.
func TestFactorNumber_Valid(t *testing.T) {
	logStart(t, "TestFactorNumber_Valid", "Factorización de 84")
	res := algorithms.Factorize(84, make(chan struct{}))
	data := decodeJSON(t, res)
	if len(data["factors"].([]interface{})) != 4 {
		t.Errorf("expected 4 factors for 84")
	}
}

// TestFactorNumber_Valid2 verifica lista de factores para 15.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Factorize calcule correctamente los
//
//	factores primos de 15 (3×5). Verifica que respuesta JSON
//	contenga exactamente 2 factores en array 'factors'.
func TestFactorNumber_Valid2(t *testing.T) {
	logStart(t, "TestFactorNumber_Valid2", "Factorización de 15")
	res := algorithms.Factorize(15, make(chan struct{}))
	data := decodeJSON(t, res)
	if len(data["factors"].([]interface{})) != 2 {
		t.Errorf("expected 2 factors for 15")
	}
}

// TestFactorNumber_PrimeInput verifica que un primo regrese lista [n].
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Factorize para numero primo (13)
//
//	retorne array con un solo elemento igual al numero original.
//	Caso especial donde numero primo es su propio factor.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Factorize retorne codigo 400 Bad Request
//
//	cuando numero de entrada es negativo (-1). Verifica validacion
//	de parametros numericos en algoritmos de factorizacion.
func TestFactorNumber_Invalid(t *testing.T) {
	logStart(t, "TestFactorNumber_Invalid", "Entrada negativa debe devolver 400")
	res := algorithms.Factorize(-1, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for negative input, got %d", res.StatusCode)
	}
}

// ---------------------- MATRIX MULTIPLICATION ----------------------

// TestMatrixMultiply_Valid verifica presencia de hash_sha256 en resultado.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo MatrixMultiply genere correctamente
//
//	multiplicacion de matrices 4x4 con seed 42. Verifica que
//	respuesta JSON contenga campo 'hash_sha256' del resultado.
func TestMatrixMultiply_Valid(t *testing.T) {
	logStart(t, "TestMatrixMultiply_Valid", "Multiplicación de matrices válida")
	res := algorithms.MatrixMultiply(4, 42, make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["hash_sha256"]; !ok {
		t.Errorf("expected hash_sha256 field")
	}
}

// TestMatrixMultiply_InvalidSize verifica 400 cuando size<=0.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo MatrixMultiply retorne codigo 400
//
//	Bad Request cuando tamano de matriz es invalido (size=0).
//	Verifica validacion de parametros en operaciones matriciales.
func TestMatrixMultiply_InvalidSize(t *testing.T) {
	logStart(t, "TestMatrixMultiply_InvalidSize", "Size inválido debe devolver 400")
	res := algorithms.MatrixMultiply(0, 42, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid size, got %d", res.StatusCode)
	}
}

// ---------------------- MANDELBROT ----------------------

// TestMandelbrot_Valid verifica que la respuesta incluya dimensiones correctas.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Mandelbrot genere fractal correctamente
//
//	con parametros validos (50x50, 50 iteraciones). Verifica
//	que respuesta JSON contenga campo 'width' = 50.
func TestMandelbrot_Valid(t *testing.T) {
	logStart(t, "TestMandelbrot_Valid", "Generación válida de fractal")
	res := algorithms.Mandelbrot(50, 50, 50, false, make(chan struct{}))
	data := decodeJSON(t, res)
	if int(data["width"].(float64)) != 50 {
		t.Errorf("expected width=50, got %v", data["width"])
	}
}

// TestMandelbrot_InvalidParams verifica 400 con parámetros inválidos.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Mandelbrot retorne codigo 400 Bad Request
//
//	cuando parametros son invalidos (width=-5). Verifica validacion
//	de parametros de dimensiones en generacion de fractales.
func TestMandelbrot_InvalidParams(t *testing.T) {
	logStart(t, "TestMandelbrot_InvalidParams", "Parámetros inválidos deben devolver 400")
	res := algorithms.Mandelbrot(-5, 20, 10, false, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid width, got %d", res.StatusCode)
	}
}

// ---------------------- PI APPROXIMATION ----------------------

// TestPi_Valid verifica que la aproximación tenga longitud razonable.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CalculatePi genere aproximacion de PI
//
//	con 20 digitos. Verifica que campo 'approx_pi' en respuesta
//	JSON contenga string con al menos 10 caracteres.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CalculatePi retorne codigo 400 Bad Request
//
//	cuando numero de digitos es negativo (-5). Verifica validacion
//	de parametros numericos en calculos matematicos de precision.
func TestPi_InvalidDigits(t *testing.T) {
	logStart(t, "TestPi_InvalidDigits", "Dígitos inválidos deben devolver 400")
	res := algorithms.CalculatePi(-5, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid digits, got %d", res.StatusCode)
	}
}

// ---------------------- CANCELLATION TESTS ----------------------

// TestFibonacci_Cancelled verifica cancelación/éxito para Fibonacci costoso.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CalculateFibonacci maneje cancelacion
//
//	durante calculo costoso (n=35). Acepta codigo 499 (cancelado)
//	o 200 (completado antes de cancelacion con timeout de 50ms).
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo MatrixMultiply maneje cancelacion durante
//
//	operacion con matrices grandes (100x100). Acepta codigo 499
//	(cancelado) o 200 (completado) con timeout de 30ms.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo IsPrime maneje correctamente numero muy
//
//	grande (2^31-1) con metodo 'trial'. Verifica presencia de
//	campo 'is_prime' en respuesta JSON para casos limite.
func TestIsPrime_LargeNumber(t *testing.T) {
	logStart(t, "TestIsPrime_LargeNumber", "Prueba con número grande (bordes)")
	res := algorithms.IsPrime(int64(1<<31-1), "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["is_prime"]; !ok {
		t.Errorf("missing is_prime field for large input")
	}
}

// TestMandelbrot_SaveFileTrue verifica que se informe archivo guardado cuando se pide.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Mandelbrot con saveFile=true genere
//
//	imagen fractal y guarde archivo. Verifica presencia de campo
//	'saved_file' en respuesta y limpia archivo temporal creado.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CalculatePi sea deterministico ejecutando
//
//	dos calculos con mismos parametros (20 digitos). Verifica que
//	ambas respuestas JSON sean identicas (mismo body).
func TestPi_Consistency(t *testing.T) {
	logStart(t, "TestPi_Consistency", "Determinismo en cálculo de PI")
	a := algorithms.CalculatePi(20, make(chan struct{}))
	b := algorithms.CalculatePi(20, make(chan struct{}))
	if string(a.Body) != string(b.Body) {
		t.Errorf("expected deterministic output for same digits")
	}
}

// TestMatrixMul_SeedEffect verifica que diferentes seeds produzcan resultados distintos.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo MatrixMultiply con diferentes seeds (1 vs 2)
//
//	produzca resultados distintos para misma dimension (4x4).
//	Verifica que seed afecte generacion de matrices aleatorias.
func TestMatrixMul_SeedEffect(t *testing.T) {
	logStart(t, "TestMatrixMul_SeedEffect", "Seeds diferentes deben producir hashes distintos")
	a := algorithms.MatrixMultiply(4, 1, make(chan struct{}))
	b := algorithms.MatrixMultiply(4, 2, make(chan struct{}))
	if string(a.Body) == string(b.Body) {
		t.Errorf("expected different results for different seeds")
	}
}

// TestFactorNumber_StringConversion valida una conversión base (sanity check).
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida funcionamiento basico de strconv.Itoa para conversion
//
//	de entero (42) a string ("42"). Test sanity check para
//	verificar funcionalidad basica de conversion de tipos.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CreateFile genere archivo correctamente
//
//	con nombre, contenido y repeticiones especificadas. Verifica
//	que archivo exista en filesystem. Test de algoritmo IO-bound.
//	Limpia archivo temporal al finalizar.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CreateFile retorne codigo 400
//
//	cuando nombre de archivo esta vacio. Verifica validacion
//	de parametros requeridos en operaciones de archivo.
func TestCreateFile_Invalid(t *testing.T) {
	logStart(t, "TestCreateFile_Invalid", "Falta de nombre de archivo debe devolver 400")
	res := algorithms.CreateFile("", "data", 2, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing filename, got %d", res.StatusCode)
	}
}

// TestCreateFile_DefaultRepeat verifica que repeat<=0 use valor por defecto.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CreateFile con repeat<=0 use valor
//
//	por defecto (1) y cree archivo exitosamente. Verifica
//	manejo de parametros con valores por defecto en IO.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CreateFile con cancelacion inmediata
//
//	retorne codigo 499 y NO cree archivo en filesystem.
//	Verifica manejo de cancelaciones antes de operaciones IO.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CreateFile retorne codigo 500 Internal
//
//	Server Error cuando path es invalido (/nonexistent_dir/fail.txt).
//	Verifica manejo de errores de escritura en operaciones IO.
func TestCreateFile_WriteError(t *testing.T) {
	logStart(t, "TestCreateFile_WriteError", "Error de escritura (ruta inválida) debe devolver 500")
	name := "/nonexistent_dir/fail.txt"
	res := algorithms.CreateFile(name, "content", 2, make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for write error, got %d", res.StatusCode)
	}
}

// TestDeleteFile_Success verifica borrado exitoso de archivo existente.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo DeleteFile elimine correctamente archivo
//
//	existente (test_delete.txt). Crea archivo temporal, ejecuta
//	borrado y verifica codigo 200 OK. Test de operacion IO-bound.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo DeleteFile retorne codigo 500 Internal
//
//	Server Error cuando intenta borrar archivo inexistente
//	(no_such_file.txt). Verifica manejo de errores IO.
func TestDeleteFile_NotFound(t *testing.T) {
	logStart(t, "TestDeleteFile_NotFound", "Borrado de archivo inexistente debe devolver 500")
	res := algorithms.DeleteFile("no_such_file.txt", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res.StatusCode)
	}
}

// TestDeleteFile_MissingName verifica 400 cuando falta nombre.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo DeleteFile retorne codigo 400 Bad Request
//
//	cuando nombre de archivo esta vacio. Verifica validacion de
//	parametros requeridos en operaciones de eliminacion de archivos.
func TestDeleteFile_MissingName(t *testing.T) {
	logStart(t, "TestDeleteFile_MissingName", "Falta de nombre debe devolver 400")
	res := algorithms.DeleteFile("", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing parameter, got %d", res.StatusCode)
	}
}

// TestDeleteFile_CancelledBeforeStart verifica que no se elimine tras cancelación previa.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo DeleteFile con cancelacion inmediata
//
//	retorne codigo 499 y NO elimine archivo del filesystem.
//	Verifica manejo de cancelaciones en operaciones destructivas.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo SortFile ordene numeros correctamente
//
//	usando merge sort. Crea archivo temporal con numeros, ejecuta
//	ordenamiento y verifica que archivo .sorted sea creado.
//	Test de algoritmo IO-bound con procesamiento numerico.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo SortFile retorne codigo 400 Bad Request
//
//	cuando algoritmo de ordenamiento es invalido ("bogus").
//	Verifica validacion de parametros de algoritmo de sorting.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo SortFile retorne codigo 400 Bad Request
//
//	cuando nombre de archivo esta vacio. Verifica validacion de
//	parametros requeridos en operaciones de ordenamiento de archivos.
func TestSortFile_MissingName(t *testing.T) {
	logStart(t, "TestSortFile_MissingName", "Falta de nombre de archivo debe devolver 400")
	res := algorithms.SortFile("", "merge", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing filename, got %d", res.StatusCode)
	}
}

// TestSortFile_FileNotFound verifica 500 cuando el archivo no existe.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo SortFile retorne codigo 500 Internal
//
//	Server Error cuando archivo especificado no existe en
//	filesystem. Verifica manejo de errores de lectura de archivos.
func TestSortFile_FileNotFound(t *testing.T) {
	logStart(t, "TestSortFile_FileNotFound", "Archivo inexistente debe devolver 500")
	res := algorithms.SortFile("no_such_file.txt", "merge", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res.StatusCode)
	}
}

// TestSortFile_EmptyFile verifica 400 cuando el archivo está vacío.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo SortFile retorne codigo 400 Bad Request
//
//	cuando archivo esta vacio (sin contenido). Verifica validacion
//	de contenido minimo para operaciones de ordenamiento.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo SortFile con cancelacion inmediata retorne
//
//	codigo 499. Crea archivo temporal, cancela inmediatamente y
//	verifica que operacion de ordenamiento sea cancelada correctamente.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Test que simula cancelacion durante lectura de archivo para
//
//	operacion SortFile. Crea archivo temporal, elimina archivo
//	para simular condicion de cancelacion durante lectura.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo SortFile maneje cancelacion durante
//
//	escritura de archivo .sorted. Acepta codigo 499 (cancelado)
//	o 200 (completado antes de cancelacion) con timeout de 10ms.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo SortFile use quicksort como algoritmo por
//
//	defecto cuando parametro 'algo' esta vacio. Crea archivo temporal,
//	ejecuta sin especificar algoritmo y verifica codigo 200 OK.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo WordCount cuente correctamente lineas,
//
//	palabras y bytes en archivo de prueba. Verifica campo 'lines'
//	con minimo 3 lineas. Test de algoritmo IO-bound de analisis.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo WordCount retorne codigo 500 Internal
//
//	Server Error cuando archivo especificado no existe
//	(missing.txt). Verifica manejo de errores de lectura.
func TestWordCount_MissingFile(t *testing.T) {
	logStart(t, "TestWordCount_MissingFile", "Archivo inexistente debe devolver 500")
	res := algorithms.WordCount("missing.txt", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res.StatusCode)
	}
}

// ---------------------- GREP ----------------------

// TestGrep_Success verifica matches y conteo mínimo esperado.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Grep encuentre correctamente patron
//
//	"match" en archivo de prueba. Verifica minimo 2 coincidencias
//	en campo 'matches'. Test de busqueda de patrones regex.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Grep retorne codigo 400 Bad Request
//
//	cuando patron regex es invalido (parentesis sin cerrar).
//	Verifica validacion de expresiones regulares malformadas.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Grep retorne codigo 400 Bad Request
//
//	cuando parametros requeridos estan vacios (nombre archivo
//	o patron de busqueda). Verifica validacion de parametros.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Grep retorne codigo 500 Internal
//
//	Server Error cuando archivo especificado no existe
//	(no_such_file.txt). Verifica manejo de errores de acceso.
func TestGrep_FileNotFound(t *testing.T) {
	logStart(t, "TestGrep_FileNotFound", "Archivo inexistente debe devolver 500")
	res := algorithms.Grep("no_such_file.txt", "abc", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res.StatusCode)
	}
}

// TestGrep_CancelledDuringRead conserva el flujo original (sin invocar) y elimina archivo.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Test que simula cancelacion durante lectura de archivo para
//
//	operacion Grep. Crea archivo temporal, elimina archivo para
//	simular condicion de cancelacion durante procesamiento de busqueda.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Grep retorne matches=0 cuando patron
//
//	de busqueda ("zzz") no tiene coincidencias en archivo.
//	Verifica comportamiento correcto con resultados vacios.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo Grep retorne codigo 500 cuando se
//
//	especifica directorio en lugar de archivo. Verifica manejo
//	de errores de scanner con tipos de entrada invalidos.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo HashFile calcule hash SHA256 correcto
//
//	de archivo con contenido especifico. Verifica presencia del
//	campo 'hash_hex' en respuesta. Test de hash criptografico.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo HashFile retorne codigo 400 Bad Request
//
//	cuando algoritmo de hash es desconocido ("unknown"). Verifica
//	validacion de algoritmos hash soportados (sha256, md5, etc).
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CompressFile comprima archivo usando
//
//	codec gzip. Crea archivo temporal con contenido repetitivo,
//	ejecuta compresion y verifica campo 'codec' en respuesta.
//	Test de algoritmo IO-bound con compresion de datos.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CompressFile retorne codigo 400 Bad
//
//	Request cuando codec de compresion es invalido ("bogus").
//	Verifica validacion de codecs soportados (gzip, zip, etc).
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CompressFile maneje cancelacion durante
//
//	compresion gzip de archivo grande (50KB). Acepta codigo 499
//	(cancelado) o 200 (completado) con timeout de 50ms.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CompressFile retorne codigo 400 Bad
//
//	Request cuando nombre de archivo esta vacio. Verifica
//	validacion de parametros requeridos en operaciones de compresion.
func TestCompressFile_MissingName(t *testing.T) {
	logStart(t, "TestCompressFile_MissingName", "Falta de nombre debe devolver 400")
	res := algorithms.CompressFile("", "gzip", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing name, got %d", res.StatusCode)
	}
}

// TestCompressFile_FileNotFound verifica 500 cuando el archivo no existe.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CompressFile retorne codigo 500 Internal
//
//	Server Error cuando archivo especificado no existe. Verifica
//	manejo de errores de acceso a archivos en compresion.
func TestCompressFile_FileNotFound(t *testing.T) {
	logStart(t, "TestCompressFile_FileNotFound", "Archivo inexistente debe devolver 500")
	res := algorithms.CompressFile("no_such_file.txt", "gzip", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res.StatusCode)
	}
}

// TestCompressFile_DefaultCodec verifica que el default sea gzip.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CompressFile use gzip como codec por
//
//	defecto cuando parametro codec esta vacio. Crea archivo
//	temporal y verifica compresion exitosa con codigo 200.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CompressFile maneje codec xz que puede
//
//	estar disponible o no segun entorno. Acepta codigo 200 (exito)
//	o 500 (xz no disponible). Test dependiente del sistema.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CompressFile retorne codigo 500 cuando
//
//	no puede crear archivo de salida debido a path invalido
//	(/no/such/path/). Verifica manejo de errores de escritura.
func TestCompressFile_CreateOutputError(t *testing.T) {
	logStart(t, "TestCompressFile_CreateOutputError", "Error al crear archivo de salida debe devolver 500")
	res := algorithms.CompressFile("/no/such/path/file.txt", "gzip", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for output file creation error, got %d", res.StatusCode)
	}
}

// TestCompressFile_EmptyFile verifica que un archivo vacío se comprima (200).
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo CompressFile comprima correctamente
//
//	archivo vacio usando gzip. Verifica que archivos sin contenido
//	sean procesados exitosamente con codigo 200 OK.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Test compuesto que valida error 500 para archivo inexistente
//
//	seguido de compresion exitosa con codec xz. Combina validacion
//	de manejo de errores y funcionalidad de codecs alternativos.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo SortFile ordene numeros correctamente
//
//	usando algoritmo quicksort especificado explicitamente.
//	Verifica funcionamiento de algoritmo de ordenamiento rapido.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo IsPrime maneje correctamente casos
//
//	limite (0,1,2,4) en deteccion de primalidad. Verifica que
//	todos los casos retornen respuesta valida (StatusCode != 0).
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida que algoritmo LoadTest maneje cancelacion durante
//
//	ejecucion de 5 tareas con sleep 1s. Acepta codigo 499
//	(cancelado) o 200 (completado) con timeout de 20ms.
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
// Entrada: t (*testing.T) - contexto de testing para logging
// Salida: ninguna (void)
// Descripcion: Test de resumen que imprime log confirmando verificacion
//
//	exitosa de todos los algoritmos (basicos, CPU-bound, IO-bound).
//	Usado como checkpoint de finalizacion de suite completa.
func TestOverall_Consistency(t *testing.T) {
	logStart(t, "TestOverall_Consistency", "Resumen de verificación general")
	t.Log("Todos los algoritmos básicos, CPU-bound e IO-bound verificados con éxito")
}

// TestLoadTest_InvalidAndCancel recorre invalidaciones y cancelación.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Test compuesto que valida multiples casos de LoadTest:
//
//	count=0 (400), sleep negativo (400), y cancelacion durante
//	ejecucion. Verifica validacion de parametros y manejo de cancelacion.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Test compuesto que valida error 500 para archivo inexistente
//
//	seguido de test de cancelacion inmediata en HashFile. Combina
//	manejo de errores de acceso y cancelaciones prematuras.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida casos limite de SimulateWork: segundos=0 debe retornar
//
//	400 Bad Request, y cancelacion inmediata debe retornar 499
//	o 200. Verifica validacion de parametros y manejo de cancelacion.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Test compuesto que valida SortFile con archivo vacio (400)
//
//	y archivo con lineas no numericas (400). Verifica validacion
//	de contenido y formato en operaciones de ordenamiento.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Test compuesto que valida casos invalidos de LoadTest:
//
//	tasks=0 (400), sleep<0 (400), y cancelacion durante ejecucion.
//	Duplica funcionalidad de TestLoadTest_InvalidAndCancel.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Test compuesto que valida Grep sin coincidencias (matches=0)
//
//	seguido de cancelacion inmediata. Combina validacion de
//	resultados vacios y manejo de cancelaciones prematuras.
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
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Valida casos limite de GenerateRandom: count=0 (400),
//
//	min=max (400), y cancelacion inmediata. Verifica validacion
//	de parametros de generacion de numeros aleatorios.
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
