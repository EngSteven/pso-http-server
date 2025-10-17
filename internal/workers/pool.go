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

var (
	ErrQueueFull = errors.New("queue full")
	ErrTimeout   = errors.New("timeout waiting for job result")
)

type JobFunc func(cancelCh <-chan struct{}) *types.Response

type job struct {
	id    string
	fn    JobFunc
	resCh chan *types.Response
	// cancel channel per job: workers will not close it, job function should observe it
	cancelCh chan struct{}
}

// Pool keeps workers and a FIFO queue
type Pool struct {
	name     string
	workers  int
	queue    chan *job
	busy     int32
	metrics  *metrics.PoolMetrics
	stopChan chan struct{}
}

var pools = make(map[string]*Pool)

// InitPool creates and starts a pool (idempotent)
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

func (p *Pool) start() {
	for i := 0; i < p.workers; i++ {
		go func(workerID int) {
			for {
				select {
				case jb := <-p.queue:
					atomic.AddInt32(&p.busy, 1)
					start := time.Now()
					// execute job
					resp := jb.fn(jb.cancelCh)
					// record metrics
					p.metrics.Record(time.Since(start))
					// deliver result (non-blocking)
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

// Enqueue attempts to put a job in the pool queue without blocking.
// If the queue is full returns ErrQueueFull.
func (p *Pool) Enqueue(fn JobFunc) (jobID string, resCh chan *types.Response, cancelCh chan struct{}, err error) {
	jb := &job{
		id:       util.NewRequestID(),
		fn:       fn,
		resCh:    make(chan *types.Response, 1),
		cancelCh: make(chan struct{}),
	}
	select {
	case p.queue <- jb:
		return jb.id, jb.resCh, jb.cancelCh, nil
	default:
		return "", nil, nil, ErrQueueFull
	}
}

// SubmitAndWait keeps backward compatibility: enqueue and wait (with optional timeout)
func (p *Pool) SubmitAndWait(fn JobFunc, timeoutMs int) (*types.Response, error) {
	id, resCh, _, err := p.Enqueue(fn)
	if err != nil {
		return nil, ErrQueueFull
	}
	_ = id // id unused here
	if timeoutMs <= 0 {
		resp := <-resCh
		return resp, nil
	}
	select {
	case resp := <-resCh:
		return resp, nil
	case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
		return nil, ErrTimeout
	}
}

// GetPool returns existing pool or nil
func GetPool(name string) *Pool {
	return pools[name]
}

// Info, PoolInfo, GetPoolInfo remain same as before...
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
