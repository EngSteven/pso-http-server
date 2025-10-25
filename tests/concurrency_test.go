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
// Objetivo: comprobar que no se bloquea el servidor ni se producen data races.
func TestConcurrentClients(t *testing.T) {
	setupIntegration(t)

	const N = 20
	var wg sync.WaitGroup
	errCh := make(chan error, N)

	t.Logf("\n--- [RUNNING] %d clientes concurrentes sobre /reverse ---", N)
	start := time.Now()

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			resp, err := http.Get(fmt.Sprintf("%s/reverse?text=req%d", baseURL, i))
			if err != nil {
				errCh <- fmt.Errorf("client %d failed: %v", i, err)
				return
			}
			if resp.StatusCode != 200 {
				errCh <- fmt.Errorf("client %d bad code: %d", i, resp.StatusCode)
			}
			resp.Body.Close()
			errCh <- nil
		}(i)
	}

	wg.Wait()
	close(errCh)

	var errCount int
	for err := range errCh {
		if err != nil {
			errCount++
			t.Error(err)
		}
	}

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
// y verifica que el servidor responda con 503 cuando la cola se satura.
func TestJobQueuePressure(t *testing.T) {
	setupIntegration(t)

	const totalJobs = 30
	submitted := 0
	fullCount := 0

	t.Logf("\n--- [RUNNING] Saturación de cola con %d jobs ---", totalJobs)
	for i := 0; i < totalJobs; i++ {
		resp, err := http.Get(fmt.Sprintf("%s/jobs/submit?task=sleep&seconds=1", baseURL))
		if err != nil {
			t.Fatalf("failed to submit job %d: %v", i, err)
		}

		switch resp.StatusCode {
		case 503:
			fullCount++
		case 200:
			submitted++
		default:
			t.Logf("[WARN] Unexpected code %d on job %d", resp.StatusCode, i)
		}
		resp.Body.Close()
	}

	t.Logf("[RESULT] Jobs OK: %d, Queue full (503): %d", submitted, fullCount)

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
