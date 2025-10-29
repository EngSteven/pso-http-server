/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: benchmark_test.go
Descripcion: Suite de pruebas de benchmark que mide latencia y throughput
en diferentes perfiles de carga con calculo de percentiles de rendimiento.
*/

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

// loadProfile define perfil de carga con requests, concurrencia y endpoint.
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
// Entrada: latencies ([]float64) - conjunto de latencias en milisegundos
//
//	p (float64) - percentil a calcular (0.0 a 1.0)
//
// Salida: float64 - valor del percentil calculado
// Descripcion: Ordena latencias y calcula percentil especificado.
//
//	Retorna 0 si conjunto vacio. Usado para metricas P50, P95, P99
//	en analisis de rendimiento de carga HTTP.
func percentile(latencies []float64, p float64) float64 {
	// === VALIDACION DE CONJUNTO VACIO ===
	// Retornar 0 si no hay datos para calcular
	if len(latencies) == 0 {
		return 0
	}
	// === ORDENAMIENTO PARA PERCENTIL ===
	// Ordenar latencias de menor a mayor
	sort.Float64s(latencies)
	// === CALCULO DEL INDICE ===
	// Convertir percentil a índice en array ordenado
	k := int(float64(len(latencies)-1) * p)
	return latencies[k]
}

// runLoadTest ejecuta un perfil de carga completo y calcula métricas.
// Entrada: t (*testing.T) - contexto de testing para logging y errores
//
//	profile (loadProfile) - configuracion de carga a ejecutar
//
// Salida: ninguna (void)
// Descripcion: Ejecuta perfil de carga con requests concurrentes al endpoint.
//
//	Mide latencias individuales, calcula percentiles P50/P95/P99,
//	throughput y maneja errores. Reporta metricas completas de
//	rendimiento para analisis de capacidad del servidor.
func runLoadTest(t *testing.T, profile loadProfile) {
	setupIntegration(t)

	t.Logf("\n--- [RUNNING] Perfil de carga: %s (%d solicitudes, %d concurrentes) ---",
		profile.name, profile.requests, profile.concurrent)

	// === CONFIGURACION DE CONCURRENCIA ===
	// Preparar sincronización y control de flujo
	var wg sync.WaitGroup
	results := make(chan float64, profile.requests)
	semaphore := make(chan struct{}, profile.concurrent)

	start := time.Now()

	// === LANZAMIENTO DE REQUESTS CONCURRENTES ===
	// Crear goroutines con límite de concurrencia
	for i := 0; i < profile.requests; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(i int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			// === MEDICION DE LATENCIA INDIVIDUAL ===
			// Medir tiempo de request específico
			begin := time.Now()
			resp, err := http.Get(baseURL + profile.endpoint)
			if err == nil {
				resp.Body.Close()
				results <- float64(time.Since(begin).Milliseconds())
			} else {
				// === MARCADO DE ERROR ===
				results <- -1 // marcar error
			}
		}(i)
	}

	// === ESPERA Y RECOPILACION ===
	// Esperar completion de todos los requests
	wg.Wait()
	close(results)

	// === PROCESAMIENTO DE RESULTADOS ===
	// Calcular métricas de rendimiento
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

	// === CALCULO DE METRICAS ESTADISTICAS ===
	// Computar percentiles y throughput
	p50 := percentile(latencies, 0.50)
	p95 := percentile(latencies, 0.95)
	p99 := percentile(latencies, 0.99)
	throughput := float64(total) / elapsed

	// === REPORTE DE RESULTADOS ===
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
// Entrada: t (*testing.T) - contexto de testing para control y logging
// Salida: ninguna (void)
// Descripcion: Ejecuta suite completa de benchmark con perfiles low/medium/high.
//
//	Cada perfil mide latencia y throughput en diferentes endpoints
//	con niveles escalados de concurrencia. Proporciona analisis
//	completo de capacidad del servidor bajo diferentes cargas.
func TestBenchmark_Profiles(t *testing.T) {
	setupIntegration(t)

	t.Log("\n============================================================")
	t.Log("BENCHMARK DE CARGA — Iniciando perfiles (low, medium, high)")
	t.Log("============================================================")

	// === EJECUCION SECUENCIAL DE PERFILES ===
	// Ejecutar cada perfil de carga definido
	for _, prof := range loadProfiles {
		runLoadTest(t, prof)
	}

	t.Log("\n[OK] Todos los perfiles de carga completados exitosamente.")
}
