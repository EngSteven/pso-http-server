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

type JobFunc func() *types.Response

type job struct {
	id    string
	fn    JobFunc
	resCh chan *types.Response
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
					resp := jb.fn()
					// record metrics
					p.metrics.Record(time.Since(start))
					// deliver result (non-blocking with select to avoid deadlocks)
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

// Submit tries to enqueue a job. If queue full returns ErrQueueFull.
func (p *Pool) Submit(fn JobFunc) (*types.Response, error) {
	jb := &job{
		id:    util.NewRequestID(),
		fn:    fn,
		resCh: make(chan *types.Response, 1),
	}
	select {
	case p.queue <- jb:
		// wait for result (blocking). Could add timeout at higher level.
		resp := <-jb.resCh
		return resp, nil
	default:
		return nil, ErrQueueFull
	}
}

// SubmitAndWait submits a job and waits up to timeoutMs milliseconds for result.
// If timeoutMs <= 0, wait indefinitely.
func (p *Pool) SubmitAndWait(fn JobFunc, timeoutMs int) (*types.Response, error) {
	jb := &job{
		id:    util.NewRequestID(),
		fn:    fn,
		resCh: make(chan *types.Response, 1),
	}
	select {
	case p.queue <- jb:
		// wait with timeout
		if timeoutMs <= 0 {
			resp := <-jb.resCh
			return resp, nil
		}
		select {
		case resp := <-jb.resCh:
			return resp, nil
		case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
			return nil, ErrTimeout
		}
	default:
		return nil, ErrQueueFull
	}
}

// GetPool returns existing pool or nil
func GetPool(name string) *Pool {
	return pools[name]
}

// PoolInfo holds observable info
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

// Info returns pool metrics snapshot
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

// GetPoolInfo returns info for a named pool (nil if not found)
func GetPoolInfo(name string) (*PoolInfo, error) {
	p := GetPool(name)
	if p == nil {
		return nil, fmt.Errorf("pool not found")
	}
	info := p.Info()
	return &info, nil
}
