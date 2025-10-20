package tests

import (
	"net/http"
	"sort"
	"sync"
	"testing"
	"time"
)

// ============================================================
// 📊 BLOQUE — MÉTRICAS DE LATENCIA Y THROUGHPUT
// ============================================================

type loadProfile struct {
	name       string
	requests   int
	concurrent int
	endpoint   string
}

var loadProfiles = []loadProfile{
    {"low", 10, 2, "/pi?digits=1000"},
    {"medium", 50, 5, "/matrixmul?size=200"},
    {"high", 100, 10, "/sortfile?name=data_big.txt&algo=merge"},
}

func percentile(latencies []float64, p float64) float64 {
	if len(latencies) == 0 {
		return 0
	}
	sort.Float64s(latencies)
	k := int(float64(len(latencies)-1) * p)
	return latencies[k]
}

func runLoadTest(t *testing.T, profile loadProfile) {
	t.Logf("🚀 Starting load profile: %s (%d reqs, %d concurrent)", profile.name, profile.requests, profile.concurrent)

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

			s1 := time.Now()
			resp, err := http.Get(baseURL + profile.endpoint)
			if err == nil {
				resp.Body.Close()
				results <- float64(time.Since(s1).Milliseconds())
			} else {
				results <- -1 // marcar error
			}
		}(i)
	}

	wg.Wait()
	close(results)

	elapsed := time.Since(start).Seconds()
	latencies := []float64{}
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
		t.Fatalf("no successful responses in profile %s", profile.name)
	}

	// Métricas básicas
	p50 := percentile(latencies, 0.50)
	p95 := percentile(latencies, 0.95)
	p99 := percentile(latencies, 0.99)
	throughput := float64(total) / elapsed

	t.Logf(`
📊 Profile: %s
  Requests: %d (Errors: %d)
  Duration: %.2fs
  Throughput: %.2f req/s
  Latency (p50): %.2f ms
  Latency (p95): %.2f ms
  Latency (p99): %.2f ms
`, profile.name, total, errors, elapsed, throughput, p50, p95, p99)
}

// ------------------------------------------------------------
// 🚦 Entrypoint del benchmark
// ------------------------------------------------------------

func TestBenchmark_Profiles(t *testing.T) {
	for _, prof := range loadProfiles {
		runLoadTest(t, prof)
	}
}
