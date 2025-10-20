package tests

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
// ⚙️ BLOQUE — PRUEBAS PESADAS DE IO Y CPU
// ============================================================

func measureHeavy(t *testing.T, name string, fn func()) {
	start := time.Now()
	fn()
	elapsed := time.Since(start)
	t.Logf("⏱ %s completed in %v", name, elapsed)
	if elapsed < 2*time.Second {
		t.Logf("⚠️ %s ran too fast — consider increasing size for realism", name)
	}
}

// ------------------------------------------------------------
// 🧠 CPU-BOUND: Pi, MatrixMul, Mandelbrot
// ------------------------------------------------------------

// ~6–8 s en CPUs modernas
func TestPerformance_PiHeavy(t *testing.T) {
	measureHeavy(t, "Pi(100000 digits)", func() {
		resp, err := http.Get(baseURL + "/pi?digits=8000")
		if err != nil {
			t.Fatalf("pi heavy request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	})
}

// ~4–6 s (200×200 matrices)
func TestPerformance_MatrixMulHeavy(t *testing.T) {
	measureHeavy(t, "MatrixMul(size=300, seed=7)", func() {
		resp, err := http.Get(baseURL + "/matrixmul?size=300&seed=7")
		if err != nil {
			t.Fatalf("matrixmul heavy failed: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	})
}

// ~5–10 s según iteraciones
func TestPerformance_MandelbrotHeavy(t *testing.T) {
	measureHeavy(t, "Mandelbrot(1000x800, max_iter=300)", func() {
		resp, err := http.Get(baseURL + "/mandelbrot?width=1000&height=800&max_iter=300")
		if err != nil {
			t.Fatalf("mandelbrot heavy failed: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	})
}

// ------------------------------------------------------------
// 💾 IO-BOUND: SortFile, Compress, HashFile (≥50 MB)
// ------------------------------------------------------------

// Genera un archivo de 50 MB con números aleatorios
func generateBigFile(path string, lines int) {
	f, _ := os.Create(path)
	for i := 0; i < lines; i++ {
		fmt.Fprintf(f, "%d\n", rand.Intn(1000000))
	}
	f.Close()
}

func TestPerformance_SortFile_50MB(t *testing.T) {
	name := "big_sort_50mb.txt"
	lines := 5_000_000 // ≈ 55 MB
	generateBigFile(name, lines)

	measureHeavy(t, "SortFile(≈50MB)", func() {
		resp, err := http.Get(baseURL + "/sortfile?name=" + name + "&algo=merge")
		if err != nil {
			t.Fatalf("sortfile heavy failed: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	os.Remove(name)
	os.Remove(name + ".sorted")
}

// Archivo de texto plano de 60 MB para compresión
func TestPerformance_Compress_50MB(t *testing.T) {
	name := "big_compress_50mb.txt"
	data := strings.Repeat("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789\n", 1_000_000) // ~60 MB
	os.WriteFile(name, []byte(data), 0644)

	measureHeavy(t, "Compress(gzip, ≈60MB)", func() {
		resp, err := http.Get(baseURL + "/compress?name=" + name + "&codec=gzip")
		if err != nil {
			t.Fatalf("compress heavy failed: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	os.Remove(name)
	os.Remove(name + ".gz")
}

// Hash de archivo de 50 MB
func TestPerformance_HashFile_50MB(t *testing.T) {
	name := "big_hash_50mb.txt"
	data := strings.Repeat("HASHLINE\n", 5_500_000) // ~55 MB
	os.WriteFile(name, []byte(data), 0644)

	measureHeavy(t, "HashFile(sha256, ≈50MB)", func() {
		resp, err := http.Get(baseURL + "/hashfile?name=" + name + "&algo=sha256")
		if err != nil {
			t.Fatalf("hashfile heavy failed: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	os.Remove(name)
}

// ------------------------------------------------------------
// 📊 MÉTRICAS DESPUÉS DE CARGA PESADA
// ------------------------------------------------------------

func TestPerformance_MetricsAfterHeavyLoad(t *testing.T) {
	resp, err := http.Get(baseURL + "/metrics")
	if err != nil {
		t.Fatalf("metrics after heavy load failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()
	t.Log("✅ /metrics endpoint responded after 50MB+ IO & heavy CPU tests")
}
