package jobs

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/util"
	"github.com/EngSteven/pso-http-server/internal/workers"
	"github.com/EngSteven/pso-http-server/internal/algorithms"
)

var (
	ErrJobNotFound  = errors.New("job not found")
	ErrJobQueueFull = errors.New("job manager queue full")
	ErrJobCancelled = errors.New("job cancelled")
)

// Default timeouts per command (ms)
var defaultTimeouts = map[string]int{
	"isprime":     5000,
	"factor":      8000,
	"pi":          15000,
	"matrixmul":   7000,
	"mandelbrot":  20000,
	"fibonacci":   3000,
	"createfile":  2000,
}

func timeoutForCommand(cmd string) int {
	if v, ok := defaultTimeouts[cmd]; ok {
		return v
	}
	return 5000
}

// JobManager manages job queues (priority) and dispatch to pools
type JobManager struct {
	mu sync.Mutex

	// queues per priority
	highQ, normalQ, lowQ chan *JobMeta

	// metadata
	store         map[string]*JobMeta
	resChMap      map[string]chan *types.Response
	cancelChMap   map[string]chan struct{}
	journalPath   string
	journalFile   *os.File
	stop          chan struct{}
	wg            sync.WaitGroup
	maxQueueTotal int
}

// NewJobManager creates and starts dispatcher
func NewJobManager(journalPath string, qDepthPerPriority int, maxQueueTotal int) (*JobManager, error) {
	j := &JobManager{
		highQ:         make(chan *JobMeta, qDepthPerPriority),
		normalQ:       make(chan *JobMeta, qDepthPerPriority),
		lowQ:          make(chan *JobMeta, qDepthPerPriority),
		store:         make(map[string]*JobMeta),
		resChMap:      make(map[string]chan *types.Response),
		cancelChMap:   make(map[string]chan struct{}),
		journalPath:   journalPath,
		stop:          make(chan struct{}),
		maxQueueTotal: maxQueueTotal,
	}

	f, err := os.OpenFile(journalPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("open journal: %w", err)
	}
	j.journalFile = f

	if err := j.rehydrate(); err != nil {
		fmt.Printf("warning: unable to rehydrate jobs: %v\n", err)
	}

	j.wg.Add(1)
	go j.dispatcher()
	return j, nil
}

func (j *JobManager) rehydrate() error {
	f := j.journalFile
	_, err := f.Seek(0, 0)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var meta JobMeta
		if err := json.Unmarshal(scanner.Bytes(), &meta); err == nil {
			j.store[meta.ID] = &meta
		}
	}
	return nil
}

func (j *JobManager) appendToJournal(meta *JobMeta) {
	if j.journalFile == nil {
		return
	}
	line, _ := json.Marshal(meta)
	j.journalFile.Write(append(line, '\n'))
	j.journalFile.Sync()
}

// Submit creates a job meta and enqueues it respecting priority
func (j *JobManager) Submit(command string, params map[string]string, priority Priority) (string, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	total := len(j.highQ) + len(j.normalQ) + len(j.lowQ)
	if total >= j.maxQueueTotal {
		// backpressure → reject and ask client to retry
		retryAfter := workers.DefaultTimeoutFor(command)
		return "", fmt.Errorf("queue full: retry_after_ms=%d", retryAfter)
	}

	id := util.NewRequestID()
	meta := &JobMeta{
		ID:         id,
		Command:    command,
		Params:     params,
		Priority:   priority,
		Status:     StatusQueued,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		TimeoutMs:  timeoutForCommand(command),
	}
	j.store[id] = meta
	j.appendToJournal(meta)

	// enqueue respecting priority
	switch priority {
	case PriorityHigh:
		select {
		case j.highQ <- meta:
		default:
			select {
			case j.normalQ <- meta:
			default:
				select {
				case j.lowQ <- meta:
				default:
					delete(j.store, id)
					return "", ErrJobQueueFull
				}
			}
		}
	case PriorityNormal:
		select {
		case j.normalQ <- meta:
		default:
			select {
			case j.lowQ <- meta:
			default:
				select {
				case j.highQ <- meta:
				default:
					delete(j.store, id)
					return "", ErrJobQueueFull
				}
			}
		}
	default:
		select {
		case j.lowQ <- meta:
		default:
			select {
			case j.normalQ <- meta:
			default:
				delete(j.store, id)
				return "", ErrJobQueueFull
			}
		}
	}
	return id, nil
}

// dispatcher consumes queues and dispatches to worker pools
func (j *JobManager) dispatcher() {
	defer j.wg.Done()
	for {
		select {
		case <-j.stop:
			return
		default:
			var meta *JobMeta
			// adaptive priority selection
			switch {
			case rand.Intn(100) < 50:
				select {
				case meta = <-j.highQ:
				default:
					select {
					case meta = <-j.normalQ:
					default:
						select {
						case meta = <-j.lowQ:
						default:
							time.Sleep(50 * time.Millisecond)
							continue
						}
					}
				}
			default:
				select {
				case meta = <-j.normalQ:
				default:
					select {
					case meta = <-j.highQ:
					default:
						select {
						case meta = <-j.lowQ:
						default:
							time.Sleep(50 * time.Millisecond)
							continue
						}
					}
				}
			}

			if meta == nil {
				continue
			}

			j.mu.Lock()
			meta.Status = StatusRunning
			meta.UpdatedAt = time.Now()
			j.appendToJournal(meta)
			j.mu.Unlock()

			pool := workers.GetPool(meta.Command)
			if pool == nil {
				res := j.executeCommandInline(meta)
				j.updateJobResult(meta, res)
				continue
			}

			jobFn := j.wrapJob(meta)

			jobID, pResCh, cancelCh, err := pool.Enqueue(jobFn, workers.PriorityNormal)
			if err != nil {
				j.mu.Lock()
				meta.Status = StatusQueued
				meta.UpdatedAt = time.Now()
				meta.Error = "pool full"
				j.appendToJournal(meta)
				j.mu.Unlock()
				go func(m *JobMeta) {
					time.Sleep(200 * time.Millisecond)
					select { case j.normalQ <- m: default: }
				}(meta)
				continue
			}

			j.mu.Lock()
			j.resChMap[jobID] = pResCh
			j.cancelChMap[jobID] = cancelCh
			j.mu.Unlock()

			go j.waitForResult(meta, jobID, pResCh, cancelCh)
		}
	}
}

func (j *JobManager) waitForResult(meta *JobMeta, id string, pch chan *types.Response, cancelCh chan struct{}) {
	timeout := time.Duration(meta.TimeoutMs) * time.Millisecond
	select {
	case res := <-pch:
		j.updateJobResult(meta, res)
	case <-time.After(timeout):
		close(cancelCh)
		j.mu.Lock()
		meta.Status = StatusTimeout
		meta.Error = fmt.Sprintf("timed out after %d ms", meta.TimeoutMs)
		meta.UpdatedAt = time.Now()
		j.appendToJournal(meta)
		delete(j.resChMap, id)
		delete(j.cancelChMap, id)
		j.mu.Unlock()
	}
}

/*Algoritmos de los jobs*/

func (j *JobManager) wrapJob(meta *JobMeta) workers.JobFunc {
	return func(cancelCh <-chan struct{}) *types.Response {
		switch meta.Command {
		case "fibonacci":
			n, _ := strconv.Atoi(meta.Params["num"])
			return algorithms.CalculateFibonacci(n, cancelCh)
		case "createfile":
			name := meta.Params["name"]
			content := meta.Params["content"]
			repeat := 1
			if rr, ok := meta.Params["repeat"]; ok {
				if v, err := strconv.Atoi(rr); err == nil && v > 0 {
					repeat = v
				}
			}
			full := strings.Repeat(content+"\n", repeat)
			_ = os.WriteFile(name, []byte(full), 0644)
			data, _ := json.Marshal(map[string]string{"file": name, "message": "file created"})
			return j.newResponse(200, "OK", "application/json", data)
		default:
			return j.newResponse(400, "Bad Request", "application/json", []byte(`{"error":"unknown command"}`))
		}
	}
}

func (j *JobManager) updateJobResult(meta *JobMeta, res *types.Response) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if res != nil {
		b, _ := json.Marshal(res)
		meta.Result = string(b)
		meta.Status = StatusDone
	} else {
		meta.Error = "nil response"
		meta.Status = StatusError
	}
	meta.UpdatedAt = time.Now()
	j.appendToJournal(meta)
	delete(j.resChMap, meta.ID)
	delete(j.cancelChMap, meta.ID)
}

func (j *JobManager) newResponse(statusCode int, status, ctype string, body []byte) *types.Response {
	return &types.Response{
		StatusCode: statusCode,
		StatusText: status,
		Headers:    map[string]string{"Content-Type": ctype},
		Body:       body,
	}
}

func (j *JobManager) executeCommandInline(meta *JobMeta) *types.Response {
	switch meta.Command {
	case "fibonacci":
		n, _ := strconv.Atoi(meta.Params["num"])
		series := make([]int, n)
		if n > 0 {
			series[0] = 0
		}
		if n > 1 {
			series[1] = 1
			for i := 2; i < n; i++ {
				series[i] = series[i-1] + series[i-2]
			}
		}
		data, _ := json.Marshal(map[string]interface{}{"n": n, "series": series})
		return j.newResponse(200, "OK", "application/json", data)
	default:
		return j.newResponse(400, "Bad Request", "application/json", []byte(`{"error":"unknown command"}`))
	}
}

func (j *JobManager) GetMeta(id string) (*JobMeta, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if meta, ok := j.store[id]; ok {
		c := *meta
		return &c, nil
	}
	return nil, ErrJobNotFound
}

func (j *JobManager) Cancel(id string) error {
	j.mu.Lock()
	meta, ok := j.store[id]
	if !ok {
		j.mu.Unlock()
		return ErrJobNotFound
	}
	if meta.Status == StatusDone || meta.Status == StatusError || meta.Status == StatusCanceled {
		j.mu.Unlock()
		return ErrJobCancelled
	}
	if meta.Status == StatusQueued {
		meta.Status = StatusCanceled
		meta.UpdatedAt = time.Now()
		meta.Error = "canceled before dispatch"
		j.appendToJournal(meta)
		j.mu.Unlock()
		return nil
	}
	if cancelCh, ok := j.cancelChMap[id]; ok {
		close(cancelCh)
		meta.Status = StatusCanceled
		meta.UpdatedAt = time.Now()
		j.appendToJournal(meta)
		j.mu.Unlock()
		return nil
	}
	j.mu.Unlock()
	return ErrJobCancelled
}
