package tests

// ============================================================
// TEST SUITE — BENCHMARK DE LATENCIA Y THROUGHPUT
// Proyecto: PSO_PY01b — Servidor HTTP concurrente
// Descripción:
//   Este archivo ejecuta perfiles de carga para medir latencia y throughput
//   en distintos endpoints del servidor HTTP. Incluye perfiles low/medium/high
//   y calcula percentiles p50, p95 y p99 de tiempo de respuesta.
// ============================================================

import (
	"net/http"
	"sort"
	"sync"
	"testing"
	"time"
)

// ============================================================
// BLOQUE — PERFILES DE CARGA
// ============================================================

// loadProfile define un perfil de prueba de carga con nombre,
// cantidad total de solicitudes, concurrencia y endpoint.
type loadProfile struct {
	name       string
	requests   int
	concurrent int
	endpoint   string
}

// Perfiles de carga predeterminados
var loadProfiles = []loadProfile{
	{"low", 10, 2, "/pi?digits=1000"},
	{"medium", 50, 5, "/matrixmul?size=200"},
	{"high", 100, 10, "/sortfile?name=data_big.txt&algo=merge"},
}

// ============================================================
// BLOQUE — FUNCIONES AUXILIARES
// ============================================================

// percentile calcula el percentil p (0.0–1.0) de un conjunto de latencias.
func percentile(latencies []float64, p float64) float64 {
	if len(latencies) == 0 {
		return 0
	}
	sort.Float64s(latencies)
	k := int(float64(len(latencies)-1) * p)
	return latencies[k]
}

// runLoadTest ejecuta un perfil de carga completo y calcula métricas.
func runLoadTest(t *testing.T, profile loadProfile) {
	setupIntegration(t)

	t.Logf("\n--- [RUNNING] Perfil de carga: %s (%d solicitudes, %d concurrentes) ---",
		profile.name, profile.requests, profile.concurrent)

	var wg sync.WaitGroup
	results := make(chan float64, profile.requests)
	semaphore := make(chan struct{}, profile.concurrent)

	start := time.Now()

	for i := 0; i < profile.requests; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(i int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			begin := time.Now()
			resp, err := http.Get(baseURL + profile.endpoint)
			if err == nil {
				resp.Body.Close()
				results <- float64(time.Since(begin).Milliseconds())
			} else {
				results <- -1 // marcar error
			}
		}(i)
	}

	wg.Wait()
	close(results)

	elapsed := time.Since(start).Seconds()
	var latencies []float64
	var errors int

	for l := range results {
		if l >= 0 {
			latencies = append(latencies, l)
		} else {
			errors++
		}
	}

	total := len(latencies)
	if total == 0 {
		t.Fatalf("[ERROR] No hubo respuestas exitosas en el perfil %s", profile.name)
	}

	// Cálculo de métricas
	p50 := percentile(latencies, 0.50)
	p95 := percentile(latencies, 0.95)
	p99 := percentile(latencies, 0.99)
	throughput := float64(total) / elapsed

	t.Logf(`
[RESULTADOS] Perfil: %s
  Requests totales: %d (Errores: %d)
  Duración total:   %.2fs
  Throughput:       %.2f req/s
  Latencia p50:     %.2f ms
  Latencia p95:     %.2f ms
  Latencia p99:     %.2f ms
`, profile.name, total, errors, elapsed, throughput, p50, p95, p99)

	t.Logf("[OK] Perfil '%s' completado correctamente", profile.name)
}

// ============================================================
// BLOQUE — EJECUCIÓN DE TODOS LOS PERFILES
// ============================================================

// TestBenchmark_Profiles ejecuta todos los perfiles definidos en secuencia.
func TestBenchmark_Profiles(t *testing.T) {
	setupIntegration(t)

	t.Log("\n============================================================")
	t.Log("BENCHMARK DE CARGA — Iniciando perfiles (low, medium, high)")
	t.Log("============================================================")

	for _, prof := range loadProfiles {
		runLoadTest(t, prof)
	}

	t.Log("\n[OK] Todos los perfiles de carga completados exitosamente.")
}
