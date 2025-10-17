package jobs

import "time"

// Job status constants
const (
	StatusQueued   = "queued"
	StatusRunning  = "running"
	StatusDone     = "done"
	StatusError    = "error"
	StatusCanceled = "canceled"
	StatusTimeout  = "timeout"
)

// Job priority
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

// JobMeta represents the metadata and current state of a job.
// It is persisted to the journal for recovery after restart.
type JobMeta struct {
	ID         string            `json:"id"`
	Command    string            `json:"command"`
	Params     map[string]string `json:"params"`
	Priority   Priority          `json:"priority"`
	Status     string            `json:"status"`
	Error      string            `json:"error,omitempty"`
	Result     string            `json:"result,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	TimeoutMs  int               `json:"timeout_ms,omitempty"` // nuevo: timeout individual por job
	SubmittedAt time.Time        `json:"submitted_at,omitempty"`
}
