package handlers

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/EngSteven/pso-http-server/internal/jobs"
	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

var globalJobMgr *jobs.JobManager

// InitializeJobManager should be called from main with config
func InitializeJobManager(jm *jobs.JobManager) {
	globalJobMgr = jm
}

// helper to convert url.Values / types.Request.Query to map[string]string
func queryToMap(values url.Values) map[string]string {
	out := make(map[string]string)
	for k, vv := range values {
		if len(vv) > 0 {
			out[k] = vv[0]
		}
	}
	return out
}

// ------------------------------------------------------------
// /jobs/submit?task=TASK&priority=high|normal|low
// ------------------------------------------------------------
func JobsSubmitHandler(req *types.Request) *types.Response {
	task := req.Query.Get("task")
	if task == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing task parameter"}`))
	}

	priorityStr := req.Query.Get("priority")
	if priorityStr == "" {
		priorityStr = "normal"
	}

	var pr jobs.Priority
	switch priorityStr {
	case "high":
		pr = jobs.PriorityHigh
	case "low":
		pr = jobs.PriorityLow
	default:
		pr = jobs.PriorityNormal
	}

	params := queryToMap(req.Query)
	delete(params, "task")
	delete(params, "priority")

	jobID, err := globalJobMgr.Submit(task, params, pr)
	if err == jobs.ErrJobQueueFull {
		return server.NewResponse(503, "Service Unavailable", "application/json",
			[]byte(`{"error":"queue full","retry_after_ms":1000}`))
	}
	if err != nil {
		msg := fmt.Sprintf(`{"error":"%s"}`, err.Error())
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}

	resp := map[string]interface{}{
		"job_id": jobID,
		"status": "queued",
	}
	b, _ := json.MarshalIndent(resp, "", "  ")
	return server.NewResponse(200, "OK", "application/json", b)
}

// ------------------------------------------------------------
// /jobs/status?id=JOBID
// ------------------------------------------------------------
func JobsStatusHandler(req *types.Request) *types.Response {
	id := req.Query.Get("id")
	if id == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing id parameter"}`))
	}

	meta, err := globalJobMgr.GetMeta(id)
	if err == jobs.ErrJobNotFound {
		return server.NewResponse(404, "Not Found", "application/json",
			[]byte(`{"error":"job not found"}`))
	}

	// Calcular progreso simple según estado
	progress := 0
	switch meta.Status {
	case jobs.StatusQueued:
		progress = 0
	case jobs.StatusRunning:
		progress = 50
	case jobs.StatusDone:
		progress = 100
	default:
		progress = 0
	}

	statusResp := map[string]interface{}{
		"id":       meta.ID,
		"status":   meta.Status,
		"progress": progress,
		"eta_ms":   0,
	}

	b, _ := json.MarshalIndent(statusResp, "", "  ")
	return server.NewResponse(200, "OK", "application/json", b)
}

// ------------------------------------------------------------
// /jobs/result?id=JOBID
// ------------------------------------------------------------
func JobsResultHandler(req *types.Request) *types.Response {
	id := req.Query.Get("id")
	if id == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing id parameter"}`))
	}

	meta, err := globalJobMgr.GetMeta(id)
	if err == jobs.ErrJobNotFound {
		return server.NewResponse(404, "Not Found", "application/json",
			[]byte(`{"error":"job not found"}`))
	}

	if meta.Status != jobs.StatusDone {
		msg := fmt.Sprintf(`{"error":"result not ready","status":"%s"}`, meta.Status)
		return server.NewResponse(409, "Conflict", "application/json", []byte(msg))
	}

	// Decodificar el types.Response guardado en meta.Result
	var res types.Response
	if err := json.Unmarshal([]byte(meta.Result), &res); err != nil {
		return server.NewResponse(500, "Internal Server Error", "application/json",
			[]byte(`{"error":"invalid result format"}`))
	}

	// Decodificar el body (que contiene el JSON real del algoritmo)
	var body map[string]interface{}
	if err := json.Unmarshal(res.Body, &body); err == nil {
		b, _ := json.MarshalIndent(body, "", "  ")
		return server.NewResponse(200, "OK", "application/json", b)
	}

	// Si no era JSON, devolver cuerpo literal
	return server.NewResponse(200, "OK", res.Headers["Content-Type"], res.Body)
}

// ------------------------------------------------------------
// /jobs/cancel?id=JOBID
// ------------------------------------------------------------
func JobsCancelHandler(req *types.Request) *types.Response {
	id := req.Query.Get("id")
	if id == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing id parameter"}`))
	}

	err := globalJobMgr.Cancel(id)
	switch err {
	case nil:
		resp := map[string]string{"status": "canceled"}
		b, _ := json.MarshalIndent(resp, "", "  ")
		return server.NewResponse(200, "OK", "application/json", b)
	case jobs.ErrJobNotFound:
		return server.NewResponse(404, "Not Found", "application/json",
			[]byte(`{"error":"job not found"}`))
	case jobs.ErrJobCancelled:
		return server.NewResponse(409, "Conflict", "application/json",
			[]byte(`{"error":"not cancelable"}`))
	default:
		msg := fmt.Sprintf(`{"error":"%s"}`, err.Error())
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}
}
