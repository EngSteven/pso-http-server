package algorithms

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// LoadTest ejecuta múltiples tareas simuladas en paralelo (para medir concurrencia).
func LoadTest(taskCount, sleepSeconds int, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if taskCount <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: tasks must be > 0"}`))
	}
	if sleepSeconds < 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameter: sleep must be >= 0"}`))
	}

	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"loadtest cancelled before start"}`))
	default:
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]string, 0, taskCount)

	for i := 1; i <= taskCount; i++ {
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(fmt.Sprintf(`{"error":"loadtest cancelled after %d/%d tasks"}`, i-1, taskCount)))
		default:
			wg.Add(1)
			go func(taskID int) {
				defer wg.Done()

				select {
				case <-cancelCh:
					mu.Lock()
					results = append(results, fmt.Sprintf("task-%d: cancelled", taskID))
					mu.Unlock()
					return
				default:
					time.Sleep(time.Duration(sleepSeconds) * time.Second)
					mu.Lock()
					results = append(results, fmt.Sprintf("task-%d: done", taskID))
					mu.Unlock()
				}
			}(i)
		}
	}

	wg.Wait()

	data, _ := json.MarshalIndent(map[string]interface{}{
		"tasks":       taskCount,
		"sleep_s":     sleepSeconds,
		"results":     results,
		"elapsed_ms":  time.Since(start).Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}
