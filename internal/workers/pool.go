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
    "isprime":     5000,  // 5 segundos
    "factor":      8000,
    "pi":          15000,
    "matrixmul":   7000,
    "mandelbrot":  20000,
    "fibonacci":   3000,
    "createfile":  2000,
}

var (
	ErrQueueFull = errors.New("queue full")
	ErrTimeout   = errors.New("timeout waiting for job result")
)

// JobFunc es la función ejecutable de un job.
// Recibe un canal de cancelación y devuelve una respuesta HTTP.
type JobFunc func(cancelCh <-chan struct{}) *types.Response

// Prioridades de los jobs (para compatibilidad futura)
const (
	PriorityLow = iota
	PriorityNormal
	PriorityHigh
)

// Estructura interna del job encolado
type job struct {
	id       string
	fn       JobFunc
	resCh    chan *types.Response
	cancelCh chan struct{}
	priority int
}

// Pool mantiene el conjunto de workers y su cola de trabajo
type Pool struct {
	name     string
	workers  int
	queue    chan *job
	busy     int32
	metrics  *metrics.PoolMetrics
	stopChan chan struct{}
}

var pools = make(map[string]*Pool)

// InitPool crea o devuelve un pool existente
func InitPool(name string, workersCount, queueDepth int) *Pool {
	if p, ok := pools[name]; ok {
		return p
	}
	p := &Pool{
		name:     name,
		workers:  workersCount,
		queue:    make(chan *job, queueDepth),
		metrics:  metrics.NewPoolMetrics(1000),
		stopChan: make(chan struct{}),
	}
	pools[name] = p
	p.start()
	return p
}

// start inicia los workers en goroutines
func (p *Pool) start() {
	for i := 0; i < p.workers; i++ {
		go func(workerID int) {
			for {
				select {
				case jb := <-p.queue:
					atomic.AddInt32(&p.busy, 1)
					start := time.Now()

					resp := jb.fn(jb.cancelCh)

					// agregar identificador del worker al header
					if resp != nil {
							if resp.Headers == nil {
									resp.Headers = map[string]string{}
							}
							resp.Headers["X-Worker-Id"] = fmt.Sprintf("%s-%d", p.name, workerID)
					}

					p.metrics.Record(time.Since(start))

					select {
					case jb.resCh <- resp:
					default:
					}

					atomic.AddInt32(&p.busy, -1)
				case <-p.stopChan:
					return
				}
			}
		}(i)
	}
}

// Enqueue agrega un job a la cola (sin bloquear)
func (p *Pool) Enqueue(fn JobFunc, priority int) (jobID string, resCh chan *types.Response, cancelCh chan struct{}, err error) {
	jb := &job{
		id:       util.NewRequestID(),
		fn:       fn,
		resCh:    make(chan *types.Response, 1),
		cancelCh: make(chan struct{}),
		priority: priority,
	}
	select {
	case p.queue <- jb:
		return jb.id, jb.resCh, jb.cancelCh, nil
	default:
		return "", nil, nil, ErrQueueFull
	}
}

// SubmitAndWait es la interfaz estándar usada por los handlers.
// Por compatibilidad, el segundo parámetro se interpreta como prioridad (no timeout).
func (p *Pool) SubmitAndWait(fn JobFunc, priority int) (*types.Response, error) {
	id, resCh, _, err := p.Enqueue(fn, priority)
	if err != nil {
		return nil, ErrQueueFull
	}
	_ = id

	select {
	case resp := <-resCh:
		return resp, nil
	case <-time.After(30 * time.Second):
		return nil, ErrTimeout
	}
}

// GetPool devuelve un pool existente o nil si no existe
func GetPool(name string) *Pool {
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

func (p *Pool) Info() PoolInfo {
	return PoolInfo{
		Name:           p.name,
		Workers:        p.workers,
		BusyWorkers:    atomic.LoadInt32(&p.busy),
		QueueLength:    len(p.queue),
		TotalProcessed: p.metrics.TotalProcessed,
		AvgLatencyMs:   p.metrics.AvgLatencyMs(),
		P50Ms:          p.metrics.Percentile(50),
		P95Ms:          p.metrics.Percentile(95),
	}
}

func GetPoolInfo(name string) (*PoolInfo, error) {
	p := GetPool(name)
	if p == nil {
		return nil, fmt.Errorf("pool not found")
	}
	info := p.Info()
	return &info, nil
}

func DefaultTimeoutFor(name string) int {
    if v, ok := defaultTimeouts[name]; ok {
        return v
    }
    return 5000 // fallback
}

func GetAllPools() map[string]*Pool {
    return pools
}
