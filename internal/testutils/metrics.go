package testutils

import "time"

var (
	Path = "testing.metrics"
	Host = "http://localhost:9529"
)

type CaseResult struct {
	Name          string
	Case          string
	Passed        bool
	FailedMessage string
	Cost          time.Duration
}

func (cr *CaseResult) Flush() error {
	// TODO: write to file
	return nil
}

func (cr *CaseResult) Post() error {
	// TODO: post to some datakit://v1/write/metrics
	return nil
}

func init() {
	// TODO: init file-path and HTTP remote host
}
