package evidence

type RunStatus string

const (
	StatusOK   RunStatus = "OK"
	StatusFail RunStatus = "FAIL"
	StatusSkip RunStatus = "SKIP"
)

type ErrorInfo struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Stack   string `json:"stack,omitempty"`
}

type StepResult struct {
	Seq          int                 `json:"seq"`
	Name         string              `json:"name"`
	Action       string              `json:"action,omitempty"`
	Selector     string              `json:"selector,omitempty"`
	InputDisplay string              `json:"input_display,omitempty"`
	ValueFrom    string              `json:"value_from,omitempty"`
	Expected     string              `json:"expected,omitempty"`
	TimeoutMS    int                 `json:"timeout_ms,omitempty"`
	Auth         bool                `json:"auth,omitempty"`
	Status       RunStatus           `json:"status"`
	StartedAt    string              `json:"started_at,omitempty"`
	EndedAt      string              `json:"ended_at,omitempty"`
	DurationUS   int64               `json:"duration_us"`
	URL          string              `json:"url,omitempty"`
	Title        string              `json:"title,omitempty"`
	Performance  *PerformanceMetrics `json:"performance,omitempty"`
	Screenshot   string              `json:"screenshot,omitempty"`
	SkipReason   string              `json:"skip_reason,omitempty"`
	Error        *ErrorInfo          `json:"error,omitempty"`
}

type RetryRecord struct {
	Attempt     int       `json:"attempt"`
	StartedAt   string    `json:"started_at,omitempty"`
	EndedAt     string    `json:"ended_at,omitempty"`
	DurationUS  int64     `json:"duration_us,omitempty"`
	Status      RunStatus `json:"status"`
	Success     bool      `json:"success"`
	FailedStep  int       `json:"failed_step,omitempty"`
	FailReason  string    `json:"fail_reason,omitempty"`
	FailureType string    `json:"failure_type,omitempty"`
	Message     string    `json:"message,omitempty"`
}

type ConsoleEvent struct {
	Seq       int    `json:"seq"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Text      string `json:"text"`
	Location  any    `json:"location,omitempty"`
}

type NetworkEvent struct {
	Seq          int    `json:"seq"`
	Timestamp    string `json:"timestamp"`
	Event        string `json:"event"`
	URL          string `json:"url"`
	Method       string `json:"method,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
	TraceID      string `json:"trace_id,omitempty"`
	Status       int64  `json:"status,omitempty"`
	Failure      string `json:"failure,omitempty"`
}

type DomSnapshot struct {
	CapturedAt string `json:"captured_at"`
	URL        string `json:"url,omitempty"`
	Title      string `json:"title,omitempty"`
	Text       string `json:"text,omitempty"`
	HTML       string `json:"html,omitempty"`
	Error      string `json:"error,omitempty"`
}

type PerformanceMetrics struct {
	TTFBMS             int64   `json:"ttfb_ms,omitempty"`
	LoadingTimeMS      int64   `json:"loading_time_ms,omitempty"`
	LCPMS              int64   `json:"lcp_ms,omitempty"`
	CLS                float64 `json:"cls,omitempty"`
	DOMContentLoadedMS int64   `json:"dom_content_loaded_ms,omitempty"`
	LoadEventEndMS     int64   `json:"load_event_end_ms,omitempty"`
}
