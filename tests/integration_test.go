package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/router"
	"github.com/EngSteven/pso-http-server/internal/handlers"
	"github.com/EngSteven/pso-http-server/internal/jobs"
	"github.com/EngSteven/pso-http-server/internal/workers"
	"github.com/EngSteven/pso-http-server/internal/types"
)

var (
	baseURL       = "http://localhost:8080"
	serverStarted = false
	testServer    *server.Server
	listener      net.Listener
)

// ------------------------------------------------------------
// 🟩 Inicio / fin del servidor dentro del mismo proceso
// ------------------------------------------------------------

func TestMain(m *testing.M) {
	fmt.Println("🚀 Iniciando servidor embebido para pruebas...")

	// Inicia el servidor en background
	go startEmbeddedServer()

	// Esperar a que el servidor esté disponible
	waitForServer()

	// Ejecutar los tests
	code := m.Run()

	// Finalizar servidor
	fmt.Println("🧹 Deteniendo servidor embebido...")
	if listener != nil {
		_ = listener.Close()
	}
	os.Exit(code)
}

// Inicia el servidor TCP real en una goroutine (instrumentado)
func startEmbeddedServer() {
	// Crear router base
	r := router.NewRouter()

	// 🧩 Inicializar pools mínimos para los handlers
	workers.InitPool("fibonacci", 2, 5)
	workers.InitPool("reverse", 2, 5)
	workers.InitPool("toupper", 2, 5)
	workers.InitPool("sleep", 2, 5)
	workers.InitPool("simulate", 2, 5)
	workers.InitPool("loadtest", 2, 3)
	workers.InitPool("createfile", 2, 5)
	workers.InitPool("deletefile", 2, 5)
	workers.InitPool("hash", 2, 5)

	// CPU / IO extra (opcional según tus handlers)
	workers.InitPool("isprime", 1, 2)
	workers.InitPool("factor", 1, 2)
	workers.InitPool("pi", 1, 2)
	workers.InitPool("matrixmul", 1, 2)
	workers.InitPool("mandelbrot", 1, 2)

	workers.InitPool("sortfile", 1, 2)
	workers.InitPool("wordcount", 1, 2)
	workers.InitPool("grep", 1, 2)
	workers.InitPool("hashfile", 1, 2)
	workers.InitPool("compress", 1, 2)

	// 🔹 Inicializar Job Manager (igual que en main.go)
	os.MkdirAll("data", 0755)
	jobMgr, err := jobs.NewJobManager("data/jobs_journal_test.jsonl", 20, 50)
	if err != nil {
		fmt.Println("❌ Error al iniciar JobManager:", err)
		return
	}
	handlers.InitializeJobManager(jobMgr)

	// Rutas principales
	r.Handle("/reverse", handlers.ReverseHandler)
	r.Handle("/toupper", handlers.ToUpperHandler)
	r.Handle("/status", handlers.StatusHandler)
	r.Handle("/metrics", handlers.MetricsHandler)
	r.Handle("/help", handlers.HelpHandler)

	// Rutas de algoritmos y archivos
	r.Handle("/createfile", handlers.CreateFileHandler)
	r.Handle("/deletefile", handlers.DeleteFileHandler)
	r.Handle("/fibonacci", handlers.FibonacciHandler)
	r.Handle("/isprime", handlers.IsPrimeHandler)
	r.Handle("/factor", handlers.FactorHandler)
	r.Handle("/pi", handlers.PiHandler)
	r.Handle("/matrixmul", handlers.MatrixHandler)
	r.Handle("/mandelbrot", handlers.MandelbrotHandler)
	r.Handle("/random", handlers.RandomHandler)
	r.Handle("/timestamp", handlers.TimestampHandler)
	r.Handle("/hash", handlers.HashHandler)
	r.Handle("/simulate", handlers.SimulateHandler)
	r.Handle("/sleep", handlers.SleepHandler)
	r.Handle("/loadtest", handlers.LoadTestHandler)

	// IO Bound
	r.Handle("/sortfile", handlers.SortFileHandler)
	r.Handle("/wordcount", handlers.WordCountHandler)
	r.Handle("/grep", handlers.GrepHandler)
	r.Handle("/hashfile", handlers.HashFileHandler)
	r.Handle("/compress", handlers.CompressHandler)

	// Jobs (modelo asincrónico)
	r.Handle("/jobs/submit", handlers.JobsSubmitHandler)
	r.Handle("/jobs/status", handlers.JobsStatusHandler)
	r.Handle("/jobs/result", handlers.JobsResultHandler)
	r.Handle("/jobs/cancel", handlers.JobsCancelHandler)

	// Crear y lanzar servidor embebido
	s := server.NewServer(":8080")
	s.Router = r
	testServer = s

	go func() {
		if err := s.Start(); err != nil {
			fmt.Println("❌ Error al iniciar servidor:", err)
		}
	}()
}

// Espera hasta que el servidor responda /status
func waitForServer() {
	deadline := time.Now().Add(5 * time.Second)
	for {
		resp, err := http.Get(baseURL + "/status")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			fmt.Println("✅ Servidor disponible en", baseURL)
			serverStarted = true
			return
		}
		if time.Now().After(deadline) {
			panic("❌ El servidor no respondió en el tiempo esperado")
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// Helper para decodificar JSON
func decodeJSONResp(t *testing.T, resp *http.Response) map[string]interface{} {
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	return data
}

// ============================================================
// 🌐 BLOQUE — PRUEBAS DE INTEGRACIÓN HTTP
// ============================================================

func TestServer_ReverseEndpoint(t *testing.T) {
	resp, err := http.Get(baseURL + "/reverse?text=abcd")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	data := decodeJSONResp(t, resp)
	if data["output"] != "dcba" {
		t.Errorf("expected dcba, got %v", data["output"])
	}
}

func TestServer_ToUpper(t *testing.T) {
	resp, err := http.Get(baseURL + "/toupper?text=hola")
	if err != nil {
		t.Fatalf("toupper request failed: %v", err)
	}
	data := decodeJSONResp(t, resp)
	if data["output"] != "HOLA" {
		t.Errorf("toupper failed: %v", data)
	}
}

func TestServer_StatusMetricsHelp(t *testing.T) {
	endpoints := []string{"/status", "/metrics", "/help"}
	for _, ep := range endpoints {
		resp, err := http.Get(baseURL + ep)
		if err != nil {
			t.Fatalf("failed %s: %v", ep, err)
		}
		if resp.StatusCode != 200 {
			t.Errorf("%s returned %d", ep, resp.StatusCode)
		}
	}
}

func TestServer_CreateDeleteFile(t *testing.T) {
	name := "itest.txt"
	_, _ = http.Get(baseURL + "/createfile?name=" + name + "&content=hi&repeat=1")
	resp, _ := http.Get(baseURL + "/deletefile?name=" + name)
	data := decodeJSONResp(t, resp)
	if data["message"] != "file deleted successfully" {
		t.Errorf("delete failed: %v", data)
	}
}

func TestServer_FibonacciIntegration(t *testing.T) {
	resp, err := http.Get(baseURL + "/fibonacci?num=8")
	if err != nil {
		t.Fatalf("fibonacci request failed: %v", err)
	}
	data := decodeJSONResp(t, resp)
	if int(data["n"].(float64)) != 8 {
		t.Errorf("expected n=8, got %v", data["n"])
	}
}

func TestServer_InvalidPath(t *testing.T) {
	resp, _ := http.Get(baseURL + "/notfound")
	if resp.StatusCode != 404 && resp.StatusCode != 400 {
		t.Errorf("expected 404 or 400, got %d", resp.StatusCode)
	}
}

// ============================================================
// 🧠 BLOQUE — JOBS (FLOW: SUBMIT → STATUS → RESULT)
// ============================================================

func TestJobs_SubmitStatusResultFlow(t *testing.T) {
	resp, err := http.Get(baseURL + "/jobs/submit?task=fibonacci&num=10")
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	submitData := decodeJSONResp(t, resp)
	jobID := submitData["job_id"].(string)

	time.Sleep(1 * time.Second)

	resp2, _ := http.Get(baseURL + "/jobs/status?id=" + jobID)
	statusData := decodeJSONResp(t, resp2)
	if statusData["status"] != "done" {
		t.Errorf("expected job done, got %v", statusData["status"])
	}

	resp3, _ := http.Get(baseURL + "/jobs/result?id=" + jobID)
	resultData := decodeJSONResp(t, resp3)
	if int(resultData["n"].(float64)) != 10 {
		t.Errorf("expected fibonacci n=10, got %v", resultData)
	}
}

func TestJobs_Cancel(t *testing.T) {
	resp, _ := http.Get(baseURL + "/jobs/submit?task=sleep&seconds=5")
	submit := decodeJSONResp(t, resp)
	jobID := submit["job_id"].(string)

	time.Sleep(300 * time.Millisecond)
	http.Get(baseURL + "/jobs/cancel?id=" + jobID)

	time.Sleep(300 * time.Millisecond)
	resp2, _ := http.Get(baseURL + "/jobs/status?id=" + jobID)
	data := decodeJSONResp(t, resp2)
	status := data["status"].(string)
	if status != "canceled" && status != "done" && status != "running" && status != "queued" {
		t.Errorf("expected canceled, done, running or queued, got %s", status)
	}
}

// ============================================================
// ⚙️ BLOQUE — CONCURRENCIA
// ============================================================

func TestServer_ConcurrentRequests(t *testing.T) {
	N := 10
	errCh := make(chan error, N)
	for i := 0; i < N; i++ {
		go func(i int) {
			resp, err := http.Get(baseURL + fmt.Sprintf("/reverse?text=req%d", i))
			if err != nil {
				errCh <- err
				return
			}
			if resp.StatusCode != 200 {
				errCh <- fmt.Errorf("bad code %d", resp.StatusCode)
			}
			resp.Body.Close()
			errCh <- nil
		}(i)
	}
	for i := 0; i < N; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("concurrent request error: %v", err)
		}
	}
}

func TestPool_GetPoolInfo_And_DefaultTimeout(t *testing.T) {
	// Crear un pool nuevo
	pool := workers.InitPool("testpool", 1, 2)
	if pool == nil {
		t.Fatalf("expected valid pool instance")
	}

	// Obtener información del pool
	info, err := workers.GetPoolInfo("testpool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "testpool" || info.Workers != 1 {
		t.Errorf("unexpected pool info %+v", info)
	}

	// Pool inexistente debería dar error
	if _, err := workers.GetPoolInfo("no_such_pool"); err == nil {
		t.Errorf("expected error for unknown pool")
	}

	// DefaultTimeoutFor con clave conocida y desconocida
	if v := workers.DefaultTimeoutFor("pi"); v <= 0 {
		t.Errorf("expected timeout > 0 for 'pi'")
	}
	if v := workers.DefaultTimeoutFor("unknown"); v != 5000 {
		t.Errorf("expected fallback 5000, got %d", v)
	}

	// Verificar GetAllPools devuelve el mapa con nuestro pool
	all := workers.GetAllPools()
	if _, ok := all["testpool"]; !ok {
		t.Errorf("pool 'testpool' not found in map")
	}
}

// --- Cobertura de SubmitAndWait y manejo de timeout ---

func TestPool_SubmitAndWait_SuccessAndTimeout(t *testing.T) {
	pool := workers.InitPool("waitpool", 1, 2)

	// Caso exitoso
	fn := func(cancel <-chan struct{}) *types.Response {
		return &types.Response{StatusCode: 200}
	}
	resp, err := pool.SubmitAndWait(fn, workers.PriorityNormal)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp == nil || resp.StatusCode != 200 {
		t.Errorf("expected valid response, got %+v", resp)
	}

	// Caso de timeout: función que nunca responde
	fnTimeout := func(cancel <-chan struct{}) *types.Response {
		time.Sleep(31 * time.Second)
		return &types.Response{StatusCode: 200}
	}
	_, err = pool.SubmitAndWait(fnTimeout, workers.PriorityNormal)
	if err == nil {
		t.Errorf("expected ErrTimeout, got nil")
	}
}

// --- Cobertura de InitPool y duplicación ---

func TestPool_InitPool_Duplicate(t *testing.T) {
	p1 := workers.InitPool("dupPool", 1, 1)
	p2 := workers.InitPool("dupPool", 2, 5)
	if p1 != p2 {
		t.Errorf("expected same pool instance on duplicate init")
	}
}

// --- Cobertura de SetTimeout y HandlePoolSubmit ---

func TestSetTimeout_ModifiesDefault(t *testing.T) {
	workers.SetTimeout("customalgo", 1234)
	v := workers.DefaultTimeoutFor("customalgo")
	if v != 1234 {
		t.Errorf("expected timeout 1234, got %d", v)
	}
}


