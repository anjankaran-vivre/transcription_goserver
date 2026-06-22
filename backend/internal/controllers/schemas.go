package controllers

// Transcription
type TranscriptionRequest struct {
	CallID string `form:"callId" json:"callId" binding:"required"`
	RecURL string `form:"recUrl" json:"recUrl" binding:"required"`
}

type TranscriptionResponse struct {
	Status        string `json:"status"`
	CallID        string `json:"callId"`
	QueuePosition *int   `json:"queuePosition,omitempty"`
}

// Status & Metrics
type StatusResponse struct {
	Status    string  `json:"status"`
	Uptime    string  `json:"uptime"`
	Workers   int     `json:"workers"`
	QueueSize int     `json:"queue_size"`
	MemoryMB  float64 `json:"memory_mb"`
	CPUPercent float64 `json:"cpu_percent"`
	PID       int     `json:"pid"`
}

type MetricsResponse struct {
	Realtime  interface{} `json:"realtime"`
	Historical interface{} `json:"historical"`
	Today     interface{} `json:"today"`
	RateLimit interface{} `json:"rate_limit"`
}

// Admin
type AdminResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Cleared *int   `json:"cleared,omitempty"`
}

// DB Models (for API)
type CallLogRecord struct {
	Timestamp        string `json:"timestamp"`
	CallID           string `json:"call_id"`
	WorkerID         int    `json:"worker_id"`
	Status           string `json:"status"`
	DurationSec      float64 `json:"duration_sec"`
	WordCount        int    `json:"word_count"`
	AudioQuality     string `json:"audio_quality"`
	SummaryGenerated bool   `json:"summary_generated"`
	APICalls         int    `json:"api_calls"`
	Transcription    string `json:"transcription"`
	Summary          string `json:"summary"`
	ErrorMessage     string `json:"error_message"`
}

type SystemLogRecord struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Component string `json:"component"`
	Message   string `json:"message"`
	ThreadID  string `json:"thread_id"`
}
