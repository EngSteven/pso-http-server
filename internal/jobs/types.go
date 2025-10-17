package jobs

import "time"

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

type JobStatus string

const (
	StatusQueued   JobStatus = "queued"
	StatusRunning  JobStatus = "running"
	StatusDone     JobStatus = "done"
	StatusError    JobStatus = "error"
	StatusCanceled JobStatus = "canceled"
)

type JobMeta struct {
	ID        string            `json:"id"`
	Command   string            `json:"command"`
	Params    map[string]string `json:"params"`
	Priority  Priority          `json:"priority"`
	Status    JobStatus         `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Result    string            `json:"result,omitempty"` // JSON serialized result
	Error     string            `json:"error,omitempty"`
}
