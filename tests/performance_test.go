/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: performance_test.go
Descripcion: Suite de pruebas de rendimiento intensivo que evalua
algoritmos CPU-bound y IO-bound con configuraciones escalables.
*/

package tests

// ============================================================
// TEST SUITE — PRUEBAS DE RENDIMIENTO (IO & CPU)
// Proyecto: PSO_PY01b — Servidor HTTP concurrente
// Descripción:
//   Este archivo contiene pruebas intensivas de CPU y IO
//   diseñadas para evaluar el desempeño del servidor HTTP.
//   Incluye cálculo de Pi, Mandelbrot, multiplicación de matrices,
//   y operaciones con archivos grandes (sort, compress, hash).
// ============================================================

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// ============================================================
// BLOQUE — CONFIGURACIÓN DE MODO DE RENDIMIENTO
// ============================================================

// PerformanceConfig configura parametros de prueba segun nivel de intensidad.
type PerformanceConfig struct {
	PiDigits     int
	MatrixSize   int
	MandelWidth  int
	MandelHeight int
	MandelIters  int
	SortLines    int
	CompressRep  int
	HashLines    int
}

var perf = selectPerformanceConfig()

// selectPerformanceConfig retorna configuracion segun variable PERF_LEVEL.
// Entrada: ninguna
// Salida: PerformanceConfig - configuracion de parametros de prueba
// Descripcion: Lee variable de entorno PERF_LEVEL y retorna configuracion
//
//	apropiada. "heavy" activa pruebas intensivas con parametros
//	grandes. Por defecto usa configuracion ligera para CI.
//	Configura tamaños de matrices, iteraciones, lineas de archivo.
func selectPerformanceConfig() PerformanceConfig {
	// === LECTURA DE VARIABLE DE ENTORNO ===
	// Determinar nivel de intensidad de las pruebas
	level := os.Getenv("PERF_LEVEL")
	if strings.ToLower(level) == "heavy" {
		fmt.Println("[INFO] PERF_LEVEL=heavy → pruebas intensivas activadas")
		// === CONFIGURACION INTENSIVA ===
		// Parámetros para pruebas de carga pesada
		return PerformanceConfig{
			PiDigits:     100000,     // 100K dígitos de PI
			MatrixSize:   1000,       // Matrices 1000x1000
			MandelWidth:  3840,       // 4K de ancho
			MandelHeight: 2160,       // 4K de alto
			MandelIters:  1200,       // Más iteraciones
			SortLines:    10_000_000, // 10M líneas
			CompressRep:  10_000_000, // 10M repeticiones
			HashLines:    20_000_000, // 20M líneas
		}
	}
	// === CONFIGURACION LIGERA (DEFAULT) ===
	// Parámetros para CI y desarrollo
	return PerformanceConfig{
		PiDigits:     8000,      // 8K dígitos
		MatrixSize:   300,       // Matrices 300x300
		MandelWidth:  1000,      // HD ancho
		MandelHeight: 800,       // HD alto
		MandelIters:  300,       // Menos iteraciones
		SortLines:    5_000_000, // 5M líneas
		CompressRep:  1_000_000, // 1M repeticiones
		HashLines:    5_500_000, // 5.5M líneas
	}
}

// ============================================================
// BLOQUE — FUNCIONES AUXILIARES DE MEDICIÓN
// ============================================================

// measureHeavy mide la duración de una tarea pesada y registra advertencias.
// Entrada: t (*testing.T) - contexto de testing para logging
//
//	name (string) - nombre descriptivo de la tarea
//	fn (func()) - funcion de la tarea a medir
//
// Salida: ninguna (void)
// Descripcion: Ejecuta funcion midiendo tiempo transcurrido. Registra inicio,
//
//	completion y duracion. Genera advertencia si tarea completa
//	muy rapido (posible entorno limitado). Usado para benchmark.
func measureHeavy(t *testing.T, name string, fn func()) {
	// === INICIO DE MEDICION ===
	// Capturar timestamp de inicio
	start := time.Now()
	t.Logf("\n--- [RUNNING] %s ---", name)

	// === EJECUCION DE TAREA PESADA ===
	// Ejecutar función que puede tomar mucho tiempo
	fn()

	// === CALCULO DE DURACION ===
	// Medir tiempo transcurrido
	elapsed := time.Since(start)
	t.Logf("[DONE] %s completado en %v", name, elapsed)

	// === DETECCION DE ENTORNO LIMITADO ===
	// Advertir si ejecución fue demasiado rápida
	if elapsed < 2*time.Second {
		t.Logf("[WARN] %s se ejecutó demasiado rápido (posible entorno limitado)", name)
	}
}

// generateBigFile crea un archivo con 'lines' números aleatorios.
// Entrada: path (string) - ruta del archivo a crear
//
//	lines (int) - cantidad de lineas con numeros aleatorios
//
// Salida: ninguna (void)
// Descripcion: Genera archivo de texto con numeros aleatorios del 0 al 999,999
//
//	en lineas separadas. Usado para crear datasets de prueba para
//	algoritmos de ordenamiento y procesamiento. No maneja errores.
func generateBigFile(path string, lines int) {
	// === CREACION DEL ARCHIVO ===
	// Crear archivo para dataset de prueba
	f, _ := os.Create(path)
	defer f.Close()

	// === GENERACION DE DATOS ALEATORIOS ===
	// Escribir números aleatorios línea por línea
	for i := 0; i < lines; i++ {
		fmt.Fprintf(f, "%d\n", rand.Intn(1_000_000))
	}
}

// ============================================================
// BLOQUE — CPU-BOUND: Pi, MatrixMul, Mandelbrot
// ============================================================

// TestPerformance_PiHeavy mide el cálculo de PI con miles de dígitos.
// Entrada: t (*testing.T) - contexto de testing para control y logging
// Salida: ninguna (void)
// Descripcion: Ejecuta calculo intensivo de PI con configuracion escalable
//
//	de digitos (8K-100K segun PERF_LEVEL). Mide tiempo de respuesta
//	del algoritmo CPU-bound. Valida que servidor maneje calculos
//	matematicos intensivos sin fallos ni timeouts.
func TestPerformance_PiHeavy(t *testing.T) {
	setupIntegration(t)

	// === CONFIGURACION DINAMICA ===
	// Usar parámetros basados en PERF_LEVEL
	name := fmt.Sprintf("Pi(%d dígitos)", perf.PiDigits)
	measureHeavy(t, name, func() {
		// === REQUEST INTENSIVO CPU ===
		// Cálculo de PI con miles de dígitos
		url := fmt.Sprintf("%s/pi?digits=%d", baseURL, perf.PiDigits)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("pi heavy request failed: %v", err)
		}
		// === VALIDACION DE EXITO ===
		// Verificar que cálculo completó sin error
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	})
}

// TestPerformance_MatrixMulHeavy mide multiplicación de matrices grandes.
// Entrada: t (*testing.T) - contexto de testing para control y logging
// Salida: ninguna (void)
// Descripcion: Ejecuta multiplicacion de matrices cuadradas con tamaño
//
//	escalable (300x300 a 1000x1000 segun PERF_LEVEL). Operacion
//	CPU-intensiva que valida capacidad de computo del servidor.
//	Mide tiempo de respuesta de algoritmo matricial complejo.
func TestPerformance_MatrixMulHeavy(t *testing.T) {
	setupIntegration(t)

	name := fmt.Sprintf("MatrixMul(size=%d, seed=7)", perf.MatrixSize)
	measureHeavy(t, name, func() {
		url := fmt.Sprintf("%s/matrixmul?size=%d&seed=7", baseURL, perf.MatrixSize)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("matrixmul heavy failed: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	})
}

// TestPerformance_MandelbrotHeavy mide generación de fractal Mandelbrot.
// Entrada: t (*testing.T) - contexto de testing para control y logging
// Salida: ninguna (void)
// Descripcion: Genera fractal Mandelbrot con resolucion escalable (1000x800
//
//	a 3840x2160 segun PERF_LEVEL) y iteraciones configurables.
//	Algoritmo CPU-intensivo que valida capacidad de computo
//	matematico complejo y renderizado de fractales.
func TestPerformance_MandelbrotHeavy(t *testing.T) {
	setupIntegration(t)

	name := fmt.Sprintf("Mandelbrot(%dx%d, max_iter=%d)", perf.MandelWidth, perf.MandelHeight, perf.MandelIters)
	measureHeavy(t, name, func() {
		url := fmt.Sprintf("%s/mandelbrot?width=%d&height=%d&max_iter=%d",
			baseURL, perf.MandelWidth, perf.MandelHeight, perf.MandelIters)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("mandelbrot heavy failed: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	})
}

// ============================================================
// BLOQUE — IO-BOUND: SortFile, Compress, HashFile
// ============================================================

// TestPerformance_SortFile mide tiempo de ordenamiento de archivo grande.
// Entrada: t (*testing.T) - contexto de testing para control y logging
// Salida: ninguna (void)
// Descripcion: Genera archivo con millones de numeros aleatorios y ejecuta
//
//	algoritmo merge sort. Mide rendimiento IO-bound con operaciones
//	de lectura/escritura intensivas. Valida capacidad de manejo
//	de archivos grandes y algoritmos de ordenamiento eficientes.
func TestPerformance_SortFile(t *testing.T) {
	setupIntegration(t)

	// === GENERACION DE DATASET ===
	// Crear archivo grande con números aleatorios
	name := "big_sort_dataset.txt"
	generateBigFile(name, perf.SortLines)

	// === PRUEBA DE ORDENAMIENTO INTENSIVO ===
	// Ordenar millones de líneas con merge sort
	label := fmt.Sprintf("SortFile(≈%d líneas)", perf.SortLines)
	measureHeavy(t, label, func() {
		resp, err := http.Get(baseURL + "/sortfile?name=" + name + "&algo=merge")
		if err != nil {
			t.Fatalf("sortfile heavy failed: %v", err)
		}
		// === ACEPTAR EXITO O BACKPRESSURE ===
		// 200: éxito, 503: pool saturado
		if resp.StatusCode != 200 && resp.StatusCode != 503 {
			t.Fatalf("unexpected status %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	// === LIMPIEZA DE ARCHIVOS ===
	// Remover archivos temporales creados
	os.Remove(name)
	os.Remove(name + ".sorted")
}

// TestPerformance_Compress mide compresión gzip de archivo de texto grande.
// Entrada: t (*testing.T) - contexto de testing para control y logging
// Salida: ninguna (void)
// Descripcion: Crea archivo de texto con millones de repeticiones de patron
//
//	alfanumerico y ejecuta compresion gzip. Mide rendimiento de
//	operaciones IO-intensivas con compresion. Valida capacidad
//	de procesamiento de archivos grandes con codecs.
func TestPerformance_Compress(t *testing.T) {
	setupIntegration(t)

	// === CREACION DE DATASET COMPRESIBLE ===
	// Generar archivo con patrón repetitivo para compresión
	name := "big_compress_dataset.txt"
	data := strings.Repeat("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789\n", perf.CompressRep)
	os.WriteFile(name, []byte(data), 0644)

	// === PRUEBA DE COMPRESION INTENSIVA ===
	// Comprimir archivo grande con gzip
	label := fmt.Sprintf("Compress(gzip, ≈%d repeticiones)", perf.CompressRep)
	measureHeavy(t, label, func() {
		resp, err := http.Get(baseURL + "/compress?name=" + name + "&codec=gzip")
		if err != nil {
			t.Fatalf("compress heavy failed: %v", err)
		}
		// === VALIDACION DE RESULTADO ===
		// Aceptar éxito o backpressure del pool
		if resp.StatusCode != 200 && resp.StatusCode != 503 {
			t.Fatalf("unexpected status %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	// === LIMPIEZA POST-PRUEBA ===
	// Remover archivos originales y comprimidos
	os.Remove(name)
	os.Remove(name + ".gz")
}

// TestPerformance_HashFile mide cálculo de hash SHA256 sobre archivo grande.
// Entrada: t (*testing.T) - contexto de testing para control y logging
// Salida: ninguna (void)
// Descripcion: Genera archivo con millones de lineas de texto repetido y
//
//	calcula hash SHA256. Mide rendimiento de operaciones criptograficas
//	sobre archivos grandes. Valida capacidad de procesamiento
//	IO-intensivo combinado con calculos cryptograficos.
func TestPerformance_HashFile(t *testing.T) {
	setupIntegration(t)

	name := "big_hash_dataset.txt"
	data := strings.Repeat("HASHLINE\n", perf.HashLines)
	os.WriteFile(name, []byte(data), 0644)

	label := fmt.Sprintf("HashFile(sha256, ≈%d líneas)", perf.HashLines)
	measureHeavy(t, label, func() {
		resp, err := http.Get(baseURL + "/hashfile?name=" + name + "&algo=sha256")
		if err != nil {
			t.Fatalf("hashfile heavy failed: %v", err)
		}
		if resp.StatusCode != 200 && resp.StatusCode != 503 {
			t.Fatalf("unexpected status %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	os.Remove(name)
}

// ============================================================
// BLOQUE — MÉTRICAS POST-CARGA PESADA
// ============================================================

// TestPerformance_MetricsAfterHeavyLoad valida que /metrics siga operativo.
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Verifica que endpoint /metrics permanezca funcional despues
//
//	de ejecutar pruebas intensivas CPU y IO. Valida que sistema
//	de metricas no se vea afectado por carga pesada previa y
//	mantenga capacidad de respuesta para monitoreo.
func TestPerformance_MetricsAfterHeavyLoad(t *testing.T) {
	setupIntegration(t)

	// === VERIFICACION POST-CARGA ===
	// Verificar que sistema de métricas sigue operativo
	resp, err := http.Get(baseURL + "/metrics")
	if err != nil {
		t.Fatalf("metrics after heavy load failed: %v", err)
	}
	// === VALIDACION DE DISPONIBILIDAD ===
	// Endpoint /metrics debe seguir respondiendo
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	t.Log("[OK] /metrics respondió correctamente tras pruebas de carga pesada (CPU & IO)")
}
