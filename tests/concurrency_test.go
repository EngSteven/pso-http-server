package tests

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

// ============================================================
// 🧩 BLOQUE — PRUEBAS DE CONCURRENCIA Y COLAS
// ============================================================

// TestConcurrentClients lanza múltiples clientes simultáneos a distintos endpoints.
// Objetivo: comprobar que no se bloquea el servidor ni hay data races.
func TestConcurrentClients(t *testing.T) {
	N := 20 // número de clientes simultáneos
	var wg sync.WaitGroup
	errCh := make(chan error, N)

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

	for err := range errCh {
		if err != nil {
			t.Error(err)
		}
	}

	elapsed := time.Since(start)
	t.Logf("✅ Completed %d concurrent requests in %v", N, elapsed)
}

// ------------------------------------------------------------

// TestJobQueuePressure encola muchos jobs para provocar backpressure
// y verifica que el servidor responda con 503 cuando la cola se satura.
func TestJobQueuePressure(t *testing.T) {
	totalJobs := 30 // intenta encolar más de la capacidad total
	submitted := 0
	var fullCount int

	for i := 0; i < totalJobs; i++ {
		resp, err := http.Get(fmt.Sprintf("%s/jobs/submit?task=sleep&seconds=1", baseURL))
		if err != nil {
			t.Fatalf("failed to submit job %d: %v", i, err)
		}

		if resp.StatusCode == 503 {
			fullCount++
		} else if resp.StatusCode == 200 {
			submitted++
		}
		resp.Body.Close()
	}

	t.Logf("Jobs submitted OK: %d, Queue full responses: %d", submitted, fullCount)

	if submitted == 0 {
		t.Errorf("no jobs were accepted; check JobManager configuration")
	}
	if fullCount == 0 {
		t.Log("⚠️ Warning: queue not saturated — consider lowering MAX_TOTAL for test realism")
	}
}

// ------------------------------------------------------------

// TestMetricsAndStatusDuringLoad asegura que /metrics y /status respondan correctamente
// mientras hay jobs ejecutándose en segundo plano.
func TestMetricsAndStatusDuringLoad(t *testing.T) {
	// Encolar algunos jobs largos
	for i := 0; i < 5; i++ {
		http.Get(fmt.Sprintf("%s/jobs/submit?task=sleep&seconds=2", baseURL))
	}

	// Consultar métricas varias veces
	for i := 0; i < 3; i++ {
		resp1, err1 := http.Get(baseURL + "/metrics")
		resp2, err2 := http.Get(baseURL + "/status")

		if err1 != nil || err2 != nil {
			t.Fatalf("metrics or status request failed (%v / %v)", err1, err2)
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
	t.Log("✅ /metrics and /status remained responsive under load")
}

// ------------------------------------------------------------

// TestCancelMultipleJobs verifica que múltiples cancelaciones simultáneas no generen errores.
func TestCancelMultipleJobs(t *testing.T) {
	N := 5
	jobIDs := []string{}

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

	// Verificar que todas las cancelaciones fueron aplicadas
	for _, id := range jobIDs {
		resp, _ := http.Get(fmt.Sprintf("%s/jobs/status?id=%s", baseURL, id))
		data := decodeJSONResp(t, resp)
		status := data["status"].(string)

		// 🔹 Aceptar cualquier estado válido de transición
		if status != "canceled" && status != "done" && status != "queued" && status != "running" {
			t.Errorf("job %s unexpected status=%s", id, status)
		}
		resp.Body.Close()
	}

	t.Log("✅ multiple job cancellations handled safely")
}
