/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: pool.go
Descripcion: Pool de workers concurrentes con colas priorizadas, timeouts
configurables y metricas detalladas para procesamiento de jobs.
*/

package workers

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/EngSteven/pso-http-server/internal/metrics"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/util"
)

var defaultTimeouts = map[string]int{
	"isprime":    5000, // 5 segundos
	"factor":     8000,
	"pi":         15000,
	"matrixmul":  7000,
	"mandelbrot": 20000,
	"fibonacci":  3000,
	"createfile": 2000,
}

var (
	ErrQueueFull = errors.New("queue full")
	ErrTimeout   = errors.New("timeout waiting for job result")
)

// JobFunc define funcion ejecutable por workers que retorna respuesta HTTP.
// Recibe canal de cancelacion para manejo de timeouts y cancelaciones.
type JobFunc func(cancelCh <-chan struct{}) *types.Response

// Prioridades de los jobs (para compatibilidad futura)
const (
	PriorityLow = iota
	PriorityNormal
	PriorityHigh
)

// job estructura interna para jobs encolados con canales de comunicacion.
type job struct {
	id       string
	fn       JobFunc
	resCh    chan *types.Response
	cancelCh chan struct{}
	priority int
}

// Pool mantiene conjunto de workers y cola de trabajo con metricas.
type Pool struct {
	name     string
	workers  int
	queue    chan *job
	busy     int32
	metrics  *metrics.PoolMetrics
	stopChan chan struct{}
}

var pools = make(map[string]*Pool)

// InitPool crea o retorna pool existente con workers y cola configurados.
// Entrada: name (string) - identificador unico del pool
//
//	workersCount (int) - numero de workers concurrentes
//	queueDepth (int) - capacidad maxima de la cola de jobs
//
// Salida: *Pool - instancia del pool de workers inicializado
// Descripcion: Crea nuevo pool si no existe o retorna existente.
//
//	Inicializa workers, cola de trabajo, metricas de rendimiento
//	y sistema de cancelacion. Pool queda listo para recibir jobs.
func InitPool(name string, workersCount, queueDepth int) *Pool {
	// === VERIFICACION DE POOL EXISTENTE ===
	// Retornar instancia existente si ya fue inicializada
	if p, ok := pools[name]; ok {
		return p
	}
	// === CREACION DE NUEVA INSTANCIA ===
	// Inicializar pool con configuración especificada
	p := &Pool{
		name:     name,
		workers:  workersCount,
		queue:    make(chan *job, queueDepth),  // Cola buffereada
		metrics:  metrics.NewPoolMetrics(1000), // Métricas con 1K muestras
		stopChan: make(chan struct{}),          // Canal de shutdown
	}
	// === REGISTRO GLOBAL Y ACTIVACION ===
	// Registrar pool y iniciar workers
	pools[name] = p
	p.start()
	return p
}

// start inicia los workers en goroutines concurrentes del pool.
// Entrada: ninguna (metodo receptor)
// Salida: ninguna (void)
// Descripcion: Inicia goroutines de workers que procesan jobs de la cola.
//
//	Cada worker ejecuta jobs, actualiza metricas de rendimiento,
//	agrega headers de identificacion y maneja cancelaciones.
//	Workers corren hasta que se cierre el canal stopChan.
func (p *Pool) start() {
	// === CREACION DE WORKERS CONCURRENTES ===
	// Lanzar goroutines worker según configuración
	for i := 0; i < p.workers; i++ {
		go func(workerID int) {
			for {
				select {
				case jb := <-p.queue:
					// === MARCADO DE WORKER OCUPADO ===
					// Incrementar contador de workers activos
					atomic.AddInt32(&p.busy, 1)
					start := time.Now()

					// === EJECUCION DEL JOB ===
					// Ejecutar función con canal de cancelación
					resp := jb.fn(jb.cancelCh)

					// === INYECCION DE HEADER DE WORKER ===
					// Agregar identificador del worker a respuesta
					if resp != nil {
						if resp.Headers == nil {
							resp.Headers = map[string]string{}
						}
						resp.Headers["X-Worker-Id"] = fmt.Sprintf("%s-%d", p.name, workerID)
					}

					// === REGISTRO DE METRICAS ===
					// Actualizar estadísticas de rendimiento
					p.metrics.Record(time.Since(start))

					// === ENVIO DE RESULTADO ===
					// Entregar respuesta al canal no bloqueante
					select {
					case jb.resCh <- resp:
					default:
					}

					// === LIBERACION DE WORKER ===
					// Marcar worker como disponible
					atomic.AddInt32(&p.busy, -1)
				case <-p.stopChan:
					// === SHUTDOWN GRACEFUL ===
					// Terminar worker al recibir señal de stop
					return
				}
			}
		}(i)
	}
}

// Enqueue agrega un job a la cola del pool sin bloquear.
// Entrada: fn (JobFunc) - funcion del job a ejecutar
//
//	priority (int) - prioridad del job (sin uso actual)
//
// Salida: jobID (string) - identificador unico del job
//
//	resCh (chan *types.Response) - canal para recibir resultado
//	cancelCh (chan struct{}) - canal para cancelar job
//	err (error) - error si cola esta llena
//
// Descripcion: Agrega job a cola sin esperar. Retorna canales de comunicacion
//
//	para recibir resultado o cancelar job. Error si cola llena.
func (p *Pool) Enqueue(fn JobFunc, priority int) (jobID string, resCh chan *types.Response, cancelCh chan struct{}, err error) {
	// === CONSTRUCCION DEL JOB ===
	// Crear estructura job con canales de comunicación
	jb := &job{
		id:       util.NewRequestID(),           // ID único
		fn:       fn,                            // Función ejecutable
		resCh:    make(chan *types.Response, 1), // Canal buffereado para resultado
		cancelCh: make(chan struct{}),           // Canal de cancelación
		priority: priority,                      // Prioridad (futura expansión)
	}
	// === ENCOLADO NO BLOQUEANTE ===
	// Intentar agregar job a cola sin esperar
	select {
	case p.queue <- jb:
		// === ENCOLADO EXITOSO ===
		// Retornar canales para comunicación
		return jb.id, jb.resCh, jb.cancelCh, nil
	default:
		// === COLA LLENA - BACKPRESSURE ===
		// Rechazar job y retornar error
		return "", nil, nil, ErrQueueFull
	}
}

// SubmitAndWait envia job al pool y espera resultado con timeout.
// Entrada: fn (JobFunc) - funcion del job a ejecutar
//
//	priority (int) - prioridad del job (por compatibilidad)
//
// Salida: *types.Response - respuesta del job ejecutado
//
//	error - error de cola llena o timeout
//
// Descripcion: Interfaz estandar para handlers. Envia job al pool y espera
//
//	resultado hasta 30 segundos. Retorna error si cola llena o
//	si job no completa en tiempo limite. Usado por HTTP handlers.
func (p *Pool) SubmitAndWait(fn JobFunc, priority int) (*types.Response, error) {
	// === ENCOLADO DEL JOB ===
	// Enviar job al pool y obtener canales de comunicación
	id, resCh, _, err := p.Enqueue(fn, priority)
	if err != nil {
		return nil, ErrQueueFull
	}
	_ = id

	// === ESPERA CON TIMEOUT ===
	// Esperar resultado o timeout de 30 segundos
	select {
	case resp := <-resCh:
		// === RESULTADO RECIBIDO ===
		// Job completado exitosamente
		return resp, nil
	case <-time.After(30 * time.Second):
		// === TIMEOUT ALCANZADO ===
		// Job no completó en tiempo límite
		return nil, ErrTimeout
	}
}

// GetPool devuelve un pool existente por nombre o nil si no existe.
// Entrada: name (string) - nombre del pool a buscar
// Salida: *Pool - instancia del pool o nil si no existe
// Descripcion: Busca y retorna pool existente en registro global.
//
//	Retorna nil si pool con nombre especificado no existe.
//	Usado para acceder a pools previamente inicializados.
func GetPool(name string) *Pool {
	// === BUSQUEDA EN REGISTRO GLOBAL ===
	// Retornar pool existente o nil si no existe
	return pools[name]
}

// --- Métricas e información ---

type PoolInfo struct {
	Name           string  `json:"name"`
	Workers        int     `json:"workers"`
	BusyWorkers    int32   `json:"busy_workers"`
	QueueLength    int     `json:"queue_length"`
	TotalProcessed int64   `json:"total_processed"`
	AvgLatencyMs   float64 `json:"avg_latency_ms"`
	P50Ms          int64   `json:"p50_ms"`
	P95Ms          int64   `json:"p95_ms"`
}

// Info retorna informacion completa del estado actual del pool.
// Entrada: ninguna (metodo receptor)
// Salida: PoolInfo - estructura con metricas y estado del pool
// Descripcion: Recopila informacion del pool incluyendo workers activos,
//
//	longitud de cola, jobs procesados, latencia promedio y
//	percentiles P50/P95. Usado para monitoreo y diagnosticos.
func (p *Pool) Info() PoolInfo {
	// === RECOPILACION DE METRICAS ACTUALES ===
	// Construir snapshot del estado del pool
	return PoolInfo{
		Name:           p.name,
		Workers:        p.workers,
		BusyWorkers:    atomic.LoadInt32(&p.busy), // Lectura atómica
		QueueLength:    len(p.queue),              // Longitud actual de cola
		TotalProcessed: p.metrics.TotalProcessed,  // Jobs procesados totales
		AvgLatencyMs:   p.metrics.AvgLatencyMs(),  // Latencia promedio
		P50Ms:          p.metrics.Percentile(50),  // Percentil 50
		P95Ms:          p.metrics.Percentile(95),  // Percentil 95
	}
}

// GetPoolInfo obtiene informacion detallada de un pool por nombre.
// Entrada: name (string) - nombre del pool a consultar
// Salida: *PoolInfo - puntero a informacion del pool
//
//	error - error si pool no existe
//
// Descripcion: Busca pool por nombre y retorna su informacion completa.
//
//	Error si pool no existe. Wrapper conveniente para obtener
//	metricas de pool especifico sin acceso directo a instancia.
func GetPoolInfo(name string) (*PoolInfo, error) {
	// === BUSQUEDA DEL POOL ===
	// Verificar existencia del pool solicitado
	p := GetPool(name)
	if p == nil {
		return nil, fmt.Errorf("pool not found")
	}
	// === OBTENCION DE INFORMACION ===
	// Generar snapshot de métricas del pool
	info := p.Info()
	return &info, nil
}

// DefaultTimeoutFor retorna timeout por defecto para algoritmo especifico.
// Entrada: name (string) - nombre del algoritmo
// Salida: int - timeout en milisegundos
// Descripcion: Retorna timeout predefinido para algoritmo conocido o
//
//	5000ms como fallback. Timeouts optimizados por complejidad:
//	isprime(5s), factor(8s), pi(15s), matrixmul(7s), etc.
func DefaultTimeoutFor(name string) int {
	// === BUSQUEDA DE TIMEOUT CONFIGURADO ===
	// Retornar timeout específico o fallback de 5 segundos
	if v, ok := defaultTimeouts[name]; ok {
		return v
	}
	return 5000 // fallback
}

// GetAllPools retorna mapa con todos los pools registrados.
// Entrada: ninguna
// Salida: map[string]*Pool - mapa de nombre a instancia de pool
// Descripcion: Retorna referencia al registro global de pools.
//
//	Permite acceso a todos los pools activos para monitoreo
//	o administracion. No es copia, modificaciones afectan original.
func GetAllPools() map[string]*Pool {
	// === RETORNO DEL REGISTRO COMPLETO ===
	// Exponer mapa global de pools (no es copia)
	return pools
}
