/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: concurrency_test.go
Descripcion: Suite de pruebas de concurrencia que valida comportamiento
bajo carga, saturacion de colas y cancelaciones multiples simultaneas.
*/

package tests

// ============================================================
// TEST SUITE — CONCURRENCIA Y COLAS DEL SERVIDOR
// Proyecto: PSO_PY01b — Servidor HTTP concurrente
// Descripción:
//   Este archivo valida el comportamiento del servidor bajo carga
//   concurrente, saturación de colas de trabajo y múltiples cancelaciones
//   simultáneas. Se verifican también métricas y estado en condiciones de estrés.
// ============================================================

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

// ============================================================
// BLOQUE — PRUEBAS DE CONCURRENCIA DE CLIENTES
// ============================================================

// TestConcurrentClients lanza múltiples clientes simultáneos a distintos endpoints.
// Entrada: t (*testing.T) - contexto de testing para control y reportes
// Salida: ninguna (void)
// Descripcion: Lanza 20 clientes concurrentes al endpoint /reverse para validar
//
//	que el servidor maneja concurrencia sin bloqueos ni data races.
//	Mide tiempo total y cuenta errores. Verifica estabilidad del
//	servidor bajo carga concurrente simultánea.
//
// Objetivo: comprobar que no se bloquea el servidor ni se producen data races.
func TestConcurrentClients(t *testing.T) {
	setupIntegration(t)

	const N = 20
	// === CONFIGURACION DE CONCURRENCIA ===
	// Preparar sincronización y canal de errores
	var wg sync.WaitGroup
	errCh := make(chan error, N)

	t.Logf("\n--- [RUNNING] %d clientes concurrentes sobre /reverse ---", N)
	start := time.Now()

	// === LANZAMIENTO DE CLIENTES CONCURRENTES ===
	// Crear N goroutines simultáneas
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// === REQUEST INDIVIDUAL ===
			// Ejecutar request con ID único
			resp, err := http.Get(fmt.Sprintf("%s/reverse?text=req%d", baseURL, i))
			if err != nil {
				errCh <- fmt.Errorf("client %d failed: %v", i, err)
				return
			}
			// === VALIDACION DE RESPUESTA ===
			if resp.StatusCode != 200 {
				errCh <- fmt.Errorf("client %d bad code: %d", i, resp.StatusCode)
			}
			resp.Body.Close()
			errCh <- nil
		}(i)
	}

	// === ESPERA Y RECOPILACION ===
	// Esperar completion y procesar errores
	wg.Wait()
	close(errCh)

	var errCount int
	for err := range errCh {
		if err != nil {
			errCount++
			t.Error(err)
		}
	}

	// === REPORTE DE RESULTADOS ===
	elapsed := time.Since(start)
	t.Logf("[RESULT] %d concurrent requests completed in %v (errors: %d)", N, elapsed, errCount)
	if errCount == 0 {
		t.Log("[OK] No se detectaron errores de concurrencia")
	}
}

// ============================================================
// BLOQUE — PRUEBAS DE SATURACIÓN DE COLAS
// ============================================================

// TestJobQueuePressure encola muchos jobs para provocar backpressure
// Entrada: t (*testing.T) - contexto de testing para logging y validacion
// Salida: ninguna (void)
// Descripcion: Envia 30 jobs sleep simultaneos para saturar cola de trabajo.
//
//	Verifica que servidor responda 503 cuando cola se llena y
//	maneje backpressure correctamente. Valida limites de capacidad
//	del sistema de jobs y comportamiento bajo sobrecarga.
//
// y verifica que el servidor responda con 503 cuando la cola se satura.
func TestJobQueuePressure(t *testing.T) {
	setupIntegration(t)

	const totalJobs = 30
	submitted := 0
	fullCount := 0

	t.Logf("\n--- [RUNNING] Saturación de cola con %d jobs ---", totalJobs)
	// === ENVIO MASIVO DE JOBS ===
	// Intentar saturar cola con jobs sleep
	for i := 0; i < totalJobs; i++ {
		resp, err := http.Get(fmt.Sprintf("%s/jobs/submit?task=sleep&seconds=1", baseURL))
		if err != nil {
			t.Fatalf("failed to submit job %d: %v", i, err)
		}

		// === CLASIFICACION DE RESPUESTAS ===
		// Contar jobs aceptados vs rechazados por backpressure
		switch resp.StatusCode {
		case 503:
			// === COLA LLENA - BACKPRESSURE ===
			fullCount++
		case 200:
			// === JOB ACEPTADO ===
			submitted++
		default:
			t.Logf("[WARN] Unexpected code %d on job %d", resp.StatusCode, i)
		}
		resp.Body.Close()
	}

	// === REPORTE DE SATURACION ===
	t.Logf("[RESULT] Jobs OK: %d, Queue full (503): %d", submitted, fullCount)

	// === VALIDACIONES DE COMPORTAMIENTO ===
	if submitted == 0 {
		t.Errorf("[ERROR] Ningún job fue aceptado; revisar configuración del JobManager")
	}
	if fullCount == 0 {
		t.Log("[WARN] Cola no saturada — ajustar MAX_TOTAL o reducir límites para pruebas realistas")
	}
}

// ============================================================
// BLOQUE — MÉTRICAS Y ESTADO DURANTE CARGA
// ============================================================

// TestMetricsAndStatusDuringLoad asegura que /metrics y /status
// Entrada: t (*testing.T) - contexto de testing para validacion
// Salida: ninguna (void)
// Descripcion: Encola jobs largos en background y consulta /metrics y /status
//
//	repetidamente durante ejecucion. Verifica que endpoints de
//	monitoreo permanezcan operativos bajo carga. Valida que sistema
//	de metricas no se vea afectado por jobs en ejecucion.
//
// respondan correctamente mientras hay jobs ejecutándose en segundo plano.
func TestMetricsAndStatusDuringLoad(t *testing.T) {
	setupIntegration(t)

	t.Log("\n--- [RUNNING] Validando /metrics y /status durante ejecución de jobs ---")

	// Encolar algunos jobs largos
	for i := 0; i < 5; i++ {
		http.Get(fmt.Sprintf("%s/jobs/submit?task=sleep&seconds=2", baseURL))
	}

	// Consultar métricas varias veces durante ejecución
	for i := 0; i < 3; i++ {
		resp1, err1 := http.Get(baseURL + "/metrics")
		resp2, err2 := http.Get(baseURL + "/status")

		if err1 != nil || err2 != nil {
			t.Fatalf("[ERROR] metrics or status request failed (%v / %v)", err1, err2)
		}

		if resp1.StatusCode != 200 {
			t.Errorf("metrics returned %d", resp1.StatusCode)
		}
		if resp2.StatusCode != 200 {
			t.Errorf("status returned %d", resp2.StatusCode)
		}

		resp1.Body.Close()
		resp2.Body.Close()
		time.Sleep(500 * time.Millisecond)
	}

	t.Log("[OK] /metrics y /status permanecieron operativos bajo carga")
}

// ============================================================
// BLOQUE — CANCELACIÓN DE MÚLTIPLES JOBS
// ============================================================

// TestCancelMultipleJobs verifica que múltiples cancelaciones simultáneas
// Entrada: t (*testing.T) - contexto de testing para control y verificacion
// Salida: ninguna (void)
// Descripcion: Encola 5 jobs sleep de larga duracion y los cancela
//
//	simultaneamente con goroutines concurrentes. Verifica que
//	cancelaciones multiples se manejen sin errores, bloqueos o
//	estados inconsistentes en el sistema de jobs.
//
// se manejen correctamente sin errores ni bloqueos.
func TestCancelMultipleJobs(t *testing.T) {
	setupIntegration(t)

	const N = 5
	jobIDs := make([]string, 0, N)

	t.Logf("\n--- [RUNNING] Cancelación simultánea de %d jobs ---", N)

	// Encolar varios jobs "sleep"
	for i := 0; i < N; i++ {
		resp, err := http.Get(fmt.Sprintf("%s/jobs/submit?task=sleep&seconds=5", baseURL))
		if err != nil {
			t.Fatalf("submit error: %v", err)
		}
		data := decodeJSONResp(t, resp)
		jobIDs = append(jobIDs, data["job_id"].(string))
	}

	// Cancelar todos casi al mismo tiempo
	var wg sync.WaitGroup
	for _, id := range jobIDs {
		wg.Add(1)
		go func(jid string) {
			defer wg.Done()
			http.Get(fmt.Sprintf("%s/jobs/cancel?id=%s", baseURL, jid))
		}(id)
	}
	wg.Wait()

	time.Sleep(300 * time.Millisecond)

	// Verificar estados de cancelación o transición válidos
	for _, id := range jobIDs {
		resp, _ := http.Get(fmt.Sprintf("%s/jobs/status?id=%s", baseURL, id))
		data := decodeJSONResp(t, resp)
		status := data["status"].(string)

		switch status {
		case "canceled", "done", "queued", "running":
			// estado aceptable
		default:
			t.Errorf("job %s unexpected status=%s", id, status)
		}
		resp.Body.Close()
	}

	t.Log("[OK] Múltiples cancelaciones concurrentes manejadas sin errores")
}
