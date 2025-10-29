/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: integration_test.go
Descripcion: Suite de pruebas de integracion completas que valida
servidor embebido, endpoints HTTP, jobs asincronos y pools de workers.
*/

package tests

// ============================================================
// TEST SUITE — PRUEBAS DE INTEGRACIÓN DEL SERVIDOR HTTP
// Proyecto: PSO_PY01b — Servidor HTTP concurrente
// Descripción:
//   Este archivo contiene las pruebas de integración completas del
//   servidor embebido. Se lanza un servidor HTTP real en background
//   y se validan endpoints, jobs y concurrencia.
// ============================================================

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/EngSteven/pso-http-server/internal/handlers"
	"github.com/EngSteven/pso-http-server/internal/jobs"
	"github.com/EngSteven/pso-http-server/internal/router"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

var (
	baseURL       = "http://localhost:8080"
	serverStarted = false
	testServer    *server.Server
)

// ============================================================
// CONFIGURACIÓN DEL SERVIDOR EMBEBIDO
// ============================================================

// setupIntegration levanta el servidor embebido una sola vez.
// Entrada: t (*testing.T) - contexto de testing para logging y control
// Salida: ninguna (void)
// Descripcion: Inicializa servidor embebido HTTP si no esta activo.
//
//	Evita reinicios multiples. Inicia servidor en background
//	y espera que este disponible antes de continuar con tests.
//	Usado por todas las pruebas de integracion.
func setupIntegration(t *testing.T) {
	// === VERIFICACION DE ESTADO GLOBAL ===
	// Evitar reinicialización múltiple del servidor
	if serverStarted {
		return // evitar reinicios
	}

	t.Log("\n============================================================")
	t.Log("SETUP: Iniciando servidor embebido para pruebas de integración")
	t.Log("============================================================")

	// === ARRANQUE EN BACKGROUND ===
	// Lanzar servidor en goroutine separada
	go startEmbeddedServer()
	// === ESPERA HASTA DISPONIBILIDAD ===
	// Bloquear hasta que servidor responda
	waitForServer(t)
	t.Log("Servidor embebido disponible en:", baseURL)
}

// startEmbeddedServer inicializa rutas, pools y job manager.
// Entrada: ninguna
// Salida: ninguna (void)
// Descripcion: Configura servidor HTTP completo para testing. Inicializa
//
//	router, pools de workers para todos los algoritmos, job manager
//	con persistencia, y registra todos los endpoints HTTP.
//	Servidor queda listo para recibir requests de pruebas.
func startEmbeddedServer() {
	// === INICIALIZACION DEL ROUTER ===
	// Crear router para registro de endpoints
	r := router.NewRouter()

	// === CONFIGURACION DE POOLS DE WORKERS ===
	// Inicializar pools especializados para cada algoritmo
	// Pools básicos (mayor capacidad)
	workers.InitPool("fibonacci", 2, 5)
	workers.InitPool("reverse", 2, 5)
	workers.InitPool("toupper", 2, 5)
	workers.InitPool("sleep", 2, 5)
	workers.InitPool("simulate", 2, 5)
	workers.InitPool("loadtest", 2, 3)
	workers.InitPool("createfile", 2, 5)
	workers.InitPool("deletefile", 2, 5)
	workers.InitPool("hash", 2, 5)
	// Pools intensivos (menor capacidad para CPU/IO)
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

	// === CONFIGURACION DEL JOB MANAGER ===
	// Inicializar sistema de jobs con persistencia
	os.MkdirAll("data", 0755)
	jobMgr, err := jobs.NewJobManager("data/jobs_journal_test.jsonl", 20, 50)
	if err != nil {
		fmt.Println("Error al iniciar JobManager:", err)
		return
	}
	handlers.InitializeJobManager(jobMgr)

	// === REGISTRO DE ENDPOINTS HTTP ===
	// Registrar todos los handlers de algoritmos y jobs
	r.Handle("/reverse", handlers.ReverseHandler)
	r.Handle("/toupper", handlers.ToUpperHandler)
	r.Handle("/status", handlers.StatusHandler)
	r.Handle("/metrics", handlers.MetricsHandler)
	r.Handle("/help", handlers.HelpHandler)
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
	r.Handle("/sortfile", handlers.SortFileHandler)
	r.Handle("/wordcount", handlers.WordCountHandler)
	r.Handle("/grep", handlers.GrepHandler)
	r.Handle("/hashfile", handlers.HashFileHandler)
	r.Handle("/compress", handlers.CompressHandler)
	r.Handle("/jobs/submit", handlers.JobsSubmitHandler)
	r.Handle("/jobs/status", handlers.JobsStatusHandler)
	r.Handle("/jobs/result", handlers.JobsResultHandler)
	r.Handle("/jobs/cancel", handlers.JobsCancelHandler)

	// === CREACION Y ARRANQUE DEL SERVIDOR ===
	// Configurar servidor con router y iniciar en puerto 8080
	s := server.NewServer(":8080")
	s.Router = r
	testServer = s

	go func() {
		if err := s.Start(); err != nil {
			fmt.Println("Error al iniciar servidor:", err)
		}
	}()
}

// waitForServer bloquea hasta que /status responda o agote timeout.
// Entrada: t (*testing.T) - contexto de testing para errores fatales
// Salida: ninguna (void)
// Descripcion: Espera hasta que servidor responda en /status con codigo 200
//
//	o agote timeout de 5 segundos. Marca serverStarted como true
//	cuando servidor esta listo. Bloquea ejecucion hasta confirmar
//	que servidor acepta connections HTTP.
func waitForServer(t *testing.T) {
	// === CONFIGURACION DE TIMEOUT ===
	// Establecer límite de tiempo para evitar espera infinita
	deadline := time.Now().Add(5 * time.Second)
	for {
		// === HEALTH CHECK DEL SERVIDOR ===
		// Intentar conectar al endpoint /status
		resp, err := http.Get(baseURL + "/status")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			// === MARCADO COMO DISPONIBLE ===
			// Servidor respondió correctamente
			serverStarted = true
			return
		}
		// === VERIFICACION DE TIMEOUT ===
		// Fallar test si servidor no responde en tiempo límite
		if time.Now().After(deadline) {
			t.Fatal("El servidor no respondió en el tiempo esperado")
		}
		// === ESPERA ENTRE REINTENTOS ===
		time.Sleep(200 * time.Millisecond)
	}
}

// decodeJSONResp convierte una respuesta HTTP a JSON.
// Entrada: t (*testing.T) - contexto de testing para errores fatales
//
//	resp (*http.Response) - respuesta HTTP a decodificar
//
// Salida: map[string]interface{} - datos JSON parseados
// Descripcion: Lee body de respuesta HTTP y lo parsea como JSON.
//
//	Retorna mapa con datos decodificados. Error fatal si
//	JSON es invalido. Usado por tests para validar respuestas.
func decodeJSONResp(t *testing.T, resp *http.Response) map[string]interface{} {
	defer resp.Body.Close()
	// === LECTURA COMPLETA DEL BODY ===
	// Leer todo el contenido de la respuesta
	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	// === PARSING Y VALIDACION JSON ===
	// Decodificar JSON y fallar test si es inválido
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	return data
}

// ============================================================
// BLOQUE — PRUEBAS DE INTEGRACIÓN HTTP
// ============================================================

// TestServer_ReverseEndpoint valida la ruta /reverse con un caso básico.
func TestServer_ReverseEndpoint(t *testing.T) {
	setupIntegration(t)

	// === REQUEST AL ENDPOINT REVERSE ===
	// Probar funcionalidad básica de reversión de texto
	resp, err := http.Get(baseURL + "/reverse?text=abcd")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	// === VALIDACION DE RESPUESTA ===
	// Verificar que respuesta JSON contiene resultado esperado
	data := decodeJSONResp(t, resp)
	if data["output"] != "dcba" {
		t.Errorf("expected dcba, got %v", data["output"])
	}
}

// TestServer_ToUpper valida la ruta /toupper con texto simple.
func TestServer_ToUpper(t *testing.T) {
	setupIntegration(t)

	// === REQUEST AL ENDPOINT TOUPPER ===
	// Probar conversión a mayúsculas
	resp, err := http.Get(baseURL + "/toupper?text=hola")
	if err != nil {
		t.Fatalf("toupper request failed: %v", err)
	}
	// === VALIDACION DE CONVERSION ===
	// Verificar transformación correcta a mayúsculas
	data := decodeJSONResp(t, resp)
	if data["output"] != "HOLA" {
		t.Errorf("toupper failed: %v", data)
	}
}

// TestServer_StatusMetricsHelp valida /status, /metrics y /help.
func TestServer_StatusMetricsHelp(t *testing.T) {
	// === ENDPOINTS DE MONITOREO Y AYUDA ===
	// Verificar disponibilidad de endpoints informativos
	endpoints := []string{"/status", "/metrics", "/help"}
	for _, ep := range endpoints {
		// === REQUEST A CADA ENDPOINT ===
		resp, err := http.Get(baseURL + ep)
		if err != nil {
			t.Fatalf("failed %s: %v", ep, err)
		}
		// === VALIDACION DE CODIGO 200 ===
		// Todos los endpoints deben responder exitosamente
		if resp.StatusCode != 200 {
			t.Errorf("%s returned %d", ep, resp.StatusCode)
		}
	}
}

// TestServer_CreateDeleteFile crea y luego elimina un archivo.
func TestServer_CreateDeleteFile(t *testing.T) {
	name := "itest.txt"
	http.Get(baseURL + "/createfile?name=" + name + "&content=hi&repeat=1")
	resp, _ := http.Get(baseURL + "/deletefile?name=" + name)
	data := decodeJSONResp(t, resp)
	if data["message"] != "file deleted successfully" {
		t.Errorf("delete failed: %v", data)
	}
}

// TestServer_FibonacciIntegration valida la ejecución del endpoint Fibonacci.
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

// TestServer_InvalidPath verifica que rutas inexistentes devuelvan 404 o 400.
func TestServer_InvalidPath(t *testing.T) {
	resp, _ := http.Get(baseURL + "/notfound")
	if resp.StatusCode != 404 && resp.StatusCode != 400 {
		t.Errorf("expected 404 or 400, got %d", resp.StatusCode)
	}
}

// ============================================================
// BLOQUE — JOBS (FLOW: SUBMIT → STATUS → RESULT → CANCEL)
// ============================================================

// TestJobs_SubmitStatusResultFlow valida flujo completo de un Job.
func TestJobs_SubmitStatusResultFlow(t *testing.T) {
	// === ENVIO DE JOB ASINCRONO ===
	// Crear job fibonacci y obtener ID
	resp, err := http.Get(baseURL + "/jobs/submit?task=fibonacci&num=10")
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	submitData := decodeJSONResp(t, resp)
	jobID := submitData["job_id"].(string)

	// === ESPERA PARA PROCESAMIENTO ===
	// Dar tiempo al job para completarse
	time.Sleep(1 * time.Second)

	// === VERIFICACION DE ESTADO ===
	// Consultar estado del job
	resp2, _ := http.Get(baseURL + "/jobs/status?id=" + jobID)
	statusData := decodeJSONResp(t, resp2)
	if statusData["status"] != "done" {
		t.Errorf("expected job done, got %v", statusData["status"])
	}

	// === OBTENCION DE RESULTADO ===
	// Recuperar resultado final del job
	resp3, _ := http.Get(baseURL + "/jobs/result?id=" + jobID)
	resultData := decodeJSONResp(t, resp3)
	if int(resultData["n"].(float64)) != 10 {
		t.Errorf("expected fibonacci n=10, got %v", resultData)
	}
}

// TestJobs_Cancel prueba cancelación manual de un job en ejecución.
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
		t.Errorf("unexpected job status: %s", status)
	}
}

// ============================================================
// BLOQUE — CONCURRENCIA Y POOLS
// ============================================================

// TestServer_ConcurrentRequests ejecuta múltiples solicitudes simultáneas
// y acepta tanto 200 (éxito) como 503 (backpressure esperado).
func TestServer_ConcurrentRequests(t *testing.T) {
	setupIntegration(t)

	const N = 10
	errCh := make(chan error, N)

	t.Logf("\n--- [RUNNING] %d concurrent requests to /reverse ---", N)

	// === LANZAMIENTO DE REQUESTS CONCURRENTES ===
	// Crear N goroutines simultáneas
	for i := 0; i < N; i++ {
		go func(i int) {
			// === REQUEST CONCURRENTE ===
			// Cada goroutine hace request independiente
			resp, err := http.Get(baseURL + fmt.Sprintf("/reverse?text=req%d", i))
			if err != nil {
				errCh <- fmt.Errorf("client %d failed: %v", i, err)
				return
			}

			// === VALIDACION DE RESPUESTA ===
			// Aceptar tanto éxito como backpressure
			switch resp.StatusCode {
			case 200, 503:
				// === CODIGOS ACEPTABLES ===
				// 200: éxito, 503: cola llena (backpressure)
			default:
				errCh <- fmt.Errorf("client %d unexpected code: %d", i, resp.StatusCode)
			}

			resp.Body.Close()
			errCh <- nil
		}(i)
	}

	// === RECOPILACION DE RESULTADOS ===
	// Contar errores inesperados
	var badCount int
	for i := 0; i < N; i++ {
		if err := <-errCh; err != nil {
			badCount++
			t.Error(err)
		}
	}

	// === EVALUACION FINAL ===
	if badCount == 0 {
		t.Log("[OK] All concurrent requests handled correctly (200/503)")
	} else {
		t.Logf("[WARN] %d requests had unexpected responses", badCount)
	}
}

// TestPool_GetPoolInfo_And_DefaultTimeout valida operaciones sobre pools.
func TestPool_GetPoolInfo_And_DefaultTimeout(t *testing.T) {
	pool := workers.InitPool("testpool", 1, 2)
	if pool == nil {
		t.Fatalf("expected valid pool instance")
	}

	info, err := workers.GetPoolInfo("testpool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "testpool" || info.Workers != 1 {
		t.Errorf("unexpected pool info %+v", info)
	}

	if _, err := workers.GetPoolInfo("no_such_pool"); err == nil {
		t.Errorf("expected error for unknown pool")
	}

	if v := workers.DefaultTimeoutFor("pi"); v <= 0 {
		t.Errorf("expected timeout > 0 for 'pi'")
	}
	if v := workers.DefaultTimeoutFor("unknown"); v != 5000 {
		t.Errorf("expected fallback 5000, got %d", v)
	}

	all := workers.GetAllPools()
	if _, ok := all["testpool"]; !ok {
		t.Errorf("pool 'testpool' not found in map")
	}
}

// TestPool_SubmitAndWait_SuccessAndTimeout cubre éxito y timeout en submit.
func TestPool_SubmitAndWait_SuccessAndTimeout(t *testing.T) {
	pool := workers.InitPool("waitpool", 1, 2)

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

	fnTimeout := func(cancel <-chan struct{}) *types.Response {
		time.Sleep(31 * time.Second)
		return &types.Response{StatusCode: 200}
	}
	_, err = pool.SubmitAndWait(fnTimeout, workers.PriorityNormal)
	if err == nil {
		t.Errorf("expected ErrTimeout, got nil")
	}
}

// TestPool_InitPool_Duplicate asegura que InitPool devuelva la misma instancia.
func TestPool_InitPool_Duplicate(t *testing.T) {
	p1 := workers.InitPool("dupPool", 1, 1)
	p2 := workers.InitPool("dupPool", 2, 5)
	if p1 != p2 {
		t.Errorf("expected same pool instance on duplicate init")
	}
}

// TestSetTimeout_ModifiesDefault valida que SetTimeout actualice valores globales.
func TestSetTimeout_ModifiesDefault(t *testing.T) {
	workers.SetTimeout("customalgo", 1234)
	v := workers.DefaultTimeoutFor("customalgo")
	if v != 1234 {
		t.Errorf("expected timeout 1234, got %d", v)
	}
}
