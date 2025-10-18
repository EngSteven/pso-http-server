package workers

import "github.com/EngSteven/pso-http-server/internal/server"
import "github.com/EngSteven/pso-http-server/internal/types"

// HandlePoolSubmit ejecuta un job en el pool indicado y devuelve una respuesta HTTP estándar.
func HandlePoolSubmit(poolName string, job JobFunc, priority int) *types.Response {
	pool := GetPool(poolName)
	if pool == nil {
		// fallback: ejecuta inline si no hay pool disponible
		return job(nil)
	}

	resp, err := pool.SubmitAndWait(job, priority)
	if err != nil {
		switch err {
		case ErrQueueFull:
			return server.NewResponse(503, "Service Unavailable", "application/json",
				[]byte(`{"error":"queue full"}`))
		case ErrTimeout:
			return server.NewResponse(500, "Internal Server Error", "application/json",
				[]byte(`{"error":"job timeout"}`))
		default:
			return server.NewResponse(500, "Internal Server Error", "application/json",
				[]byte(`{"error":"unknown error"}`))
		}
	}

	if resp == nil {
		return server.NewResponse(500, "Internal Server Error", "application/json",
			[]byte(`{"error":"empty job result"}`))
	}

	return resp
}
