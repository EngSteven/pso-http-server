package jobs

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/EngSteven/pso-http-server/internal/util"
	"github.com/EngSteven/pso-http-server/internal/workers"
	"github.com/EngSteven/pso-http-server/internal/types"
)

var (
	ErrJobNotFound   = errors.New("job not found")
	ErrJobQueueFull  = errors.New("job manager queue full")
	ErrJobCancelled  = errors.New("job cancelled")
)

// JobManager manages job queues (priority) and dispatch to pools
type JobManager struct {
	mu sync.Mutex
	// queues per priority (buffered channels)
	highQ   chan *JobMeta
	normalQ chan *JobMeta
	lowQ    chan *JobMeta

	// job store (metadata)
	store map[string]*JobMeta

	// map jobID -> result channel/cancel channel references
	resChMap    map[string]chan *types.Response
	cancelChMap map[string]chan struct{}

	// persistence journal
	journalPath string
	journalFile *os.File

	// dispatcher control
	stop chan struct{}
	wg   sync.WaitGroup

	// config
	maxQueueTotal int // total max across all priorities
}

// NewJobManager creates and starts dispatcher. queueDepthPerPriority is used to size each channel.
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

	// open journal file for append (create if needed)
	f, err := os.OpenFile(journalPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("open journal: %w", err)
	}
	j.journalFile = f

	// rehydrate metadata if exists
	if err := j.rehydrate(); err != nil {
		// non-fatal but warn
		fmt.Printf("warning: unable to rehydrate jobs: %v\n", err)
	}

	// start dispatcher
	j.wg.Add(1)
	go j.dispatcher()
	return j, nil
}

// rehydrate reads journal and rebuilds store (metadata only)
func (j *JobManager) rehydrate() error {
	f := j.journalFile
	_, err := f.Seek(0, 0)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var meta JobMeta
		if err := json.Unmarshal(scanner.Bytes(), &meta); err != nil {
			continue
		}
		// keep the last known state
		j.store[meta.ID] = &meta
	}
	return nil
}

// appendToJournal writes job meta as JSON line
func (j *JobManager) appendToJournal(meta *JobMeta) {
	if j.journalFile == nil {
		return
	}
	line, _ := json.Marshal(meta)
	j.journalFile.Write(append(line, '\n'))
	j.journalFile.Sync()
}

// Submit creates a JobMeta and enqueues it to priority queue (non-blocking).
// Returns jobID or error if queues full.
func (j *JobManager) Submit(command string, params map[string]string, priority Priority) (string, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	// check total capacity
	total := len(j.highQ) + len(j.normalQ) + len(j.lowQ)
	if total >= j.maxQueueTotal {
		return "", ErrJobQueueFull
	}

	id := util.NewRequestID()
	meta := &JobMeta{
		ID:        id,
		Command:   command,
		Params:    params,
		Priority:  priority,
		Status:    StatusQueued,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	j.store[id] = meta
	j.appendToJournal(meta)

	// enqueue into appropriate priority queue (non-blocking)
	switch priority {
	case PriorityHigh:
		select {
		case j.highQ <- meta:
		default:
			// fallback: try normal, low, otherwise fail
			select {
			case j.normalQ <- meta:
			default:
				select {
				case j.lowQ <- meta:
				default:
					// cannot enqueue
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
	case PriorityLow:
		select {
		case j.lowQ <- meta:
		default:
			// try normal/high as fallback
			select {
			case j.normalQ <- meta:
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
		// unknown priority => normal
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

	// appended & enqueued successfully
	return id, nil
}

// dispatcher pulls jobs from queues honoring priority and dispatches them to corresponding pool
func (j *JobManager) dispatcher() {
	defer j.wg.Done()
	for {
		select {
		case <-j.stop:
			return
		default:
			var meta *JobMeta
			// priority: high -> normal -> low (non-blocking selection)
			select {
			case meta = <-j.highQ:
			default:
				select {
				case meta = <-j.normalQ:
				default:
					select {
					case meta = <-j.lowQ:
					default:
						// nothing to do, sleep briefly
						time.Sleep(50 * time.Millisecond)
						continue
					}
				}
			}

			// dispatch meta to pool
			if meta == nil {
				continue
			}

			// update meta status to running
			j.mu.Lock()
			meta.Status = StatusRunning
			meta.UpdatedAt = time.Now()
			j.appendToJournal(meta)
			j.mu.Unlock()

			// prepare a job function that calls the actual command handler.
			// Mapping from command->pool name (for now same)
			poolName := meta.Command

			p := workers.GetPool(poolName)
			if p == nil {
				// no pool for this command -> execute inline (blocking)
				res := j.executeCommandInline(meta)
				j.mu.Lock()
				if res != nil {
					data, _ := json.Marshal(res)
					meta.Result = string(data)
				} else {
					meta.Error = "nil result"
				}
				meta.Status = StatusDone
				meta.UpdatedAt = time.Now()
				j.appendToJournal(meta)
				j.mu.Unlock()
				continue
			}

			// create job function that uses params and returns *types.Response
			jobFn := func(cancelCh <-chan struct{}) *types.Response {
				// map to call existing handlers: for commands we support, call the same logic as handlers do.
				switch meta.Command {
				case "fibonacci":
					nStr := meta.Params["num"]
					n, _ := strconv.Atoi(nStr)
					// compute
					series := make([]int, n)
					if n > 0 {
						series[0] = 0
					}
					if n > 1 {
						series[1] = 1
						for i := 2; i < n; i++ {
							select {
							case <-cancelCh:
								return j.newResponse(500, "Canceled", "application/json", []byte(`{"error":"cancelled"}`))
							default:
							}
							series[i] = series[i-1] + series[i-2]
						}
					}
					respObj := map[string]interface{}{"n": n, "series": series}
					data, _ := json.Marshal(respObj)
					return j.newResponse(200, "OK", "application/json", data)
				case "createfile":
					// perform create file operation
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
					respObj := map[string]string{"file": name, "message": "file created"}
					data, _ := json.Marshal(respObj)
					return j.newResponse(200, "OK", "application/json", data)
				default:
					// unknown command
					return j.newResponse(400, "Bad Request", "application/json", []byte(`{"error":"unknown command"}`))
				}
			}

			// Try to enqueue into pool (non-blocking)
			_, pResCh, cancelCh, err := p.Enqueue(jobFn, workers.PriorityNormal)
			if err != nil {
				// pool queue full => mark meta back to queued and re-enqueue with small sleep (or drop)
				j.mu.Lock()
				meta.Status = StatusQueued
				meta.UpdatedAt = time.Now()
				j.appendToJournal(meta)
				j.mu.Unlock()
				// requeue with slight delay to avoid busy loop
				go func(m *JobMeta) {
					time.Sleep(100 * time.Millisecond)
					// best-effort re-enqueue into normalQ
					select {
					case j.normalQ <- m:
					default:
						// drop if cannot requeue
						j.mu.Lock()
						m.Status = StatusError
						m.Error = "unable to enqueue to pool"
						j.appendToJournal(m)
						j.mu.Unlock()
					}
				}(meta)
				continue
			}

			// record mapping to collect result and allow cancellation (use meta.ID as key)
			j.mu.Lock()
			j.resChMap[meta.ID] = pResCh
			j.cancelChMap[meta.ID] = cancelCh
			// update meta to running with pool-specific job id if needed
			meta.UpdatedAt = time.Now()
			j.appendToJournal(meta)
			j.mu.Unlock()

			// wait asynchronously for result and update meta
			// capture meta.ID and meta pointer to avoid closure issues
			go func(m *JobMeta, metaID string, pch chan *types.Response) {
				select {
				case res := <-pch:
					j.mu.Lock()
					if res != nil {
						b, _ := json.Marshal(res)
						m.Result = string(b)
						m.Status = StatusDone
					} else {
						m.Error = "nil response"
						m.Status = StatusError
					}
					m.UpdatedAt = time.Now()
					j.appendToJournal(m)
					// cleanup maps
					delete(j.resChMap, metaID)
					delete(j.cancelChMap, metaID)
					j.mu.Unlock()
				}
			}(meta, meta.ID, pResCh)
		}
	}
}

// newResponse creates a new types.Response
func (j *JobManager) newResponse(statusCode int, status string, contentType string, body []byte) *types.Response {
	return &types.Response{
		StatusCode: statusCode,
		StatusText: status,
		Headers:    map[string]string{"Content-Type": contentType},
		Body:       body,
	}
}

func (j *JobManager) executeCommandInline(meta *JobMeta) *types.Response {
	switch meta.Command {
	case "fibonacci":
		nStr := meta.Params["num"]
		n, _ := strconv.Atoi(nStr)
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
		respObj := map[string]interface{}{"n": n, "series": series}
		data, _ := json.Marshal(respObj)
		return j.newResponse(200, "OK", "application/json", data)
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
		respObj := map[string]string{"file": name, "message": "file created"}
		data, _ := json.Marshal(respObj)
		return j.newResponse(200, "OK", "application/json", data)
	default:
		return j.newResponse(400, "Bad Request", "application/json", []byte(`{"error":"unknown command"}`))
	}
}

// GetMeta returns a copy of job meta
func (j *JobManager) GetMeta(id string) (*JobMeta, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if meta, ok := j.store[id]; ok {
		// return a shallow copy
		copy := *meta
		return &copy, nil
	}
	return nil, ErrJobNotFound
}

// Cancel attempts to cancel a job. If job is queued, remove from queue if possible.
// If running, try to close cancelCh and set status to canceled; if not cancelable, return ErrJobCancelled.
func (j *JobManager) Cancel(id string) error {
	j.mu.Lock()
	meta, ok := j.store[id]
	if !ok {
		j.mu.Unlock()
		return ErrJobNotFound
	}
	// If done or error, cannot cancel
	if meta.Status == StatusDone || meta.Status == StatusError || meta.Status == StatusCanceled {
		j.mu.Unlock()
		return ErrJobCancelled
	}
	// If queued: mark canceled (we won't remove from channel trivially), but mark status and journalize.
	if meta.Status == StatusQueued {
		meta.Status = StatusCanceled
		meta.UpdatedAt = time.Now()
		meta.Error = "canceled before dispatch"
		j.appendToJournal(meta)
		j.mu.Unlock()
		return nil
	}
	// If running: find cancelCh by job id
	if cancelCh, ok := j.cancelChMap[id]; ok {
		// close cancel channel (non-blocking)
		close(cancelCh)
		meta.Status = StatusCanceled
		meta.UpdatedAt = time.Now()
		j.appendToJournal(meta)
		j.mu.Unlock()
		return nil
	}
	j.mu.Unlock()
	// If no cancelCh, cannot cancel
	return ErrJobCancelled
}
