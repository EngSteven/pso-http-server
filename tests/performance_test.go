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

// PerformanceConfig permite escalar los parámetros de prueba según el nivel.
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

// selectPerformanceConfig devuelve la configuración según PERF_LEVEL.
func selectPerformanceConfig() PerformanceConfig {
	level := os.Getenv("PERF_LEVEL")
	if strings.ToLower(level) == "heavy" {
		fmt.Println("[INFO] PERF_LEVEL=heavy → pruebas intensivas activadas")
		return PerformanceConfig{
			PiDigits:     100000,
			MatrixSize:   1000,
			MandelWidth:  3840,
			MandelHeight: 2160,
			MandelIters:  1200,
			SortLines:    10_000_000,
			CompressRep:  10_000_000,
			HashLines:    20_000_000,
		}
	}
	// Default (modo ligero / CI)
	return PerformanceConfig{
		PiDigits:     8000,
		MatrixSize:   300,
		MandelWidth:  1000,
		MandelHeight: 800,
		MandelIters:  300,
		SortLines:    5_000_000,
		CompressRep:  1_000_000,
		HashLines:    5_500_000,
	}
}

// ============================================================
// BLOQUE — FUNCIONES AUXILIARES DE MEDICIÓN
// ============================================================

// measureHeavy mide la duración de una tarea pesada y registra advertencias.
func measureHeavy(t *testing.T, name string, fn func()) {
	start := time.Now()
	t.Logf("\n--- [RUNNING] %s ---", name)
	fn()
	elapsed := time.Since(start)
	t.Logf("[DONE] %s completado en %v", name, elapsed)

	if elapsed < 2*time.Second {
		t.Logf("[WARN] %s se ejecutó demasiado rápido (posible entorno limitado)", name)
	}
}

// generateBigFile crea un archivo con 'lines' números aleatorios.
func generateBigFile(path string, lines int) {
	f, _ := os.Create(path)
	defer f.Close()
	for i := 0; i < lines; i++ {
		fmt.Fprintf(f, "%d\n", rand.Intn(1_000_000))
	}
}

// ============================================================
// BLOQUE — CPU-BOUND: Pi, MatrixMul, Mandelbrot
// ============================================================

// TestPerformance_PiHeavy mide el cálculo de PI con miles de dígitos.
func TestPerformance_PiHeavy(t *testing.T) {
	setupIntegration(t)

	name := fmt.Sprintf("Pi(%d dígitos)", perf.PiDigits)
	measureHeavy(t, name, func() {
		url := fmt.Sprintf("%s/pi?digits=%d", baseURL, perf.PiDigits)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("pi heavy request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	})
}

// TestPerformance_MatrixMulHeavy mide multiplicación de matrices grandes.
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
func TestPerformance_SortFile(t *testing.T) {
	setupIntegration(t)

	name := "big_sort_dataset.txt"
	generateBigFile(name, perf.SortLines)

	label := fmt.Sprintf("SortFile(≈%d líneas)", perf.SortLines)
	measureHeavy(t, label, func() {
		resp, err := http.Get(baseURL + "/sortfile?name=" + name + "&algo=merge")
		if err != nil {
			t.Fatalf("sortfile heavy failed: %v", err)
		}
		if resp.StatusCode != 200 && resp.StatusCode != 503 {
			t.Fatalf("unexpected status %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	os.Remove(name)
	os.Remove(name + ".sorted")
}

// TestPerformance_Compress mide compresión gzip de archivo de texto grande.
func TestPerformance_Compress(t *testing.T) {
	setupIntegration(t)

	name := "big_compress_dataset.txt"
	data := strings.Repeat("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789\n", perf.CompressRep)
	os.WriteFile(name, []byte(data), 0644)

	label := fmt.Sprintf("Compress(gzip, ≈%d repeticiones)", perf.CompressRep)
	measureHeavy(t, label, func() {
		resp, err := http.Get(baseURL + "/compress?name=" + name + "&codec=gzip")
		if err != nil {
			t.Fatalf("compress heavy failed: %v", err)
		}
		if resp.StatusCode != 200 && resp.StatusCode != 503 {
			t.Fatalf("unexpected status %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	os.Remove(name)
	os.Remove(name + ".gz")
}

// TestPerformance_HashFile mide cálculo de hash SHA256 sobre archivo grande.
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
func TestPerformance_MetricsAfterHeavyLoad(t *testing.T) {
	setupIntegration(t)

	resp, err := http.Get(baseURL + "/metrics")
	if err != nil {
		t.Fatalf("metrics after heavy load failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	t.Log("[OK] /metrics respondió correctamente tras pruebas de carga pesada (CPU & IO)")
}
