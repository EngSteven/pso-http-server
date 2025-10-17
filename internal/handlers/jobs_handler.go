package handlers

import (
	"encoding/json"
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

// Submit handler: GET /jobs/submit?task=TASK&...params...&priority=high|normal|low
func JobsSubmitHandler(req *types.Request) *types.Response {
	task := req.Query.Get("task")
	if task == "" {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"missing task parameter"}`))
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

	// remove task and priority from params
	delete(params, "task")
	delete(params, "priority")

	jobID, err := globalJobMgr.Submit(task, params, pr)
	if err == jobs.ErrJobQueueFull {
		return server.NewResponse(503, "Service Unavailable", "application/json", []byte(`{"error":"queue full","retry_after_ms":1000}`))
	}
	if err != nil {
		msg := `{"error":"` + err.Error() + `"}`
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}

	resp := map[string]interface{}{
		"job_id": jobID,
		"status": "queued",
	}
	b, _ := json.MarshalIndent(resp, "", "  ")
	return server.NewResponse(200, "OK", "application/json", b)
}

// Status handler: GET /jobs/status?id=JOBID
func JobsStatusHandler(req *types.Request) *types.Response {
	id := req.Query.Get("id")
	if id == "" {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"missing id parameter"}`))
	}
	meta, err := globalJobMgr.GetMeta(id)
	if err == jobs.ErrJobNotFound {
		return server.NewResponse(404, "Not Found", "application/json", []byte(`{"error":"job not found"}`))
	}
	b, _ := json.MarshalIndent(meta, "", "  ")
	return server.NewResponse(200, "OK", "application/json", b)
}

// Result: GET /jobs/result?id=JOBID
func JobsResultHandler(req *types.Request) *types.Response {
	id := req.Query.Get("id")
	if id == "" {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"missing id parameter"}`))
	}
	meta, err := globalJobMgr.GetMeta(id)
	if err == jobs.ErrJobNotFound {
		return server.NewResponse(404, "Not Found", "application/json", []byte(`{"error":"job not found"}`))
	}
	if meta.Status != jobs.StatusDone {
		return server.NewResponse(409, "Conflict", "application/json", []byte(`{"error":"result not ready","status":"`+string(meta.Status)+`"}`))
	}
	// meta.Result already JSON string
	return server.NewResponse(200, "OK", "application/json", []byte(meta.Result))
}

// Cancel: GET /jobs/cancel?id=JOBID
func JobsCancelHandler(req *types.Request) *types.Response {
	id := req.Query.Get("id")
	if id == "" {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"missing id parameter"}`))
	}
	err := globalJobMgr.Cancel(id)
	if err == jobs.ErrJobNotFound {
		return server.NewResponse(404, "Not Found", "application/json", []byte(`{"error":"job not found"}`))
	}
	if err == jobs.ErrJobCancelled {
		return server.NewResponse(409, "Conflict", "application/json", []byte(`{"error":"not cancelable"}`))
	}
	if err != nil {
		msg := `{"error":"` + err.Error() + `"}`
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}
	b, _ := json.MarshalIndent(map[string]string{"status": "canceled"}, "", "  ")
	return server.NewResponse(200, "OK", "application/json", b)
}
