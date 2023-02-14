// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package testutils

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

var (
	metricFile string
	mtx        sync.Mutex

	hostname string
)

type TestStatus int

const (
	TestStatusUnknown TestStatus = iota
	TestPassed
	TestFailed
	TestSkipped
)

func (c TestStatus) String() string {
	switch c {
	case TestPassed:
		return "pass"
	case TestFailed:
		return "fail"
	case TestSkipped:
		return "skip"
	case TestStatusUnknown:
		return "unknown"

	default:
		return "unknown"
	}
}

type TestingMetric interface {
	LineProtocol() string
}

// ModuleResult collect `go test` metrics for single go module.
type ModuleResult struct {
	Name     string
	Cost     time.Duration
	Status   TestStatus
	Coverage float64
	NoTest   bool

	FailedMessage string
	Message       string
}

func (mr *ModuleResult) LineProtocol() string {
	tags := map[string]string{
		"name":   mr.Name,
		"status": mr.Status.String(),
		"host":   hostname,
	}

	// Is the module have test or not?
	if mr.NoTest {
		tags["notest"] = "T"
	} else {
		tags["notest"] = "F"
	}

	fields := map[string]any{
		"cost":           int64(mr.Cost),
		"coverage":       mr.Coverage,
		"message":        mr.Message,
		"failed_message": mr.FailedMessage,
	}

	p := point.NewPointV2([]byte("testing_module"), append(point.NewTags(tags), point.NewKVs(fields)...))

	return p.LineProto() + "\n"
}

// CaseResult collect `go test -run` metrics for single case within go module.
type CaseResult struct {
	Name string
	Case string

	FailedMessage string
	Message       string

	Status TestStatus
	Cost   time.Duration

	ExtraTags   map[string]string
	ExtraFields map[string]any
}

func (cr *CaseResult) AddField(k string, v any) {
	if cr.ExtraFields == nil {
		cr.ExtraFields = map[string]any{}
	}

	cr.ExtraFields[k] = v
}

func (cr *CaseResult) AddTag(k, v string) {
	if cr.ExtraTags == nil {
		cr.ExtraTags = map[string]string{}
	}

	cr.ExtraTags[k] = v
}

func (cr *CaseResult) LineProtocol() string {
	tags := map[string]string{
		"name":   cr.Name,
		"case":   cr.Case,
		"status": cr.Status.String(),
		"host":   hostname,
	}

	fields := map[string]any{
		"cost":           int64(cr.Cost),
		"message":        cr.Message,
		"failed_message": cr.FailedMessage,
	}

	for k, v := range cr.ExtraTags {
		switch k {
		case "name", "case", "status": // ingored
		default:
			tags[k] = v
		}
	}

	for k, v := range cr.ExtraFields {
		switch k {
		case "cost", "time": // ingored
		default:
			fields[k] = v
		}
	}

	p := point.NewPointV2([]byte("testing"), append(point.NewTags(tags), point.NewKVs(fields)...))

	return p.LineProto() + "\n"
}

func Flush(m TestingMetric) error {
	mtx.Lock()
	defer mtx.Unlock()

	if metricFile == "" {
		if v := os.Getenv("TESTING_METRIC_PATH"); v != "" {
			metricFile = v
		} else {
			metricFile = "./testing_metrics"
		}

		if err := os.MkdirAll(path.Dir(metricFile), os.ModePerm); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(filepath.Clean(metricFile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	defer f.Close() //nolint:errcheck,gosec

	lp := m.LineProtocol()
	log.Printf("write %s...", lp)
	if _, err := f.WriteString(lp); err != nil {
		return err
	}

	return nil
}

func init() { //nolint:gochecknoinits
	x, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	} else {
		hostname = x
	}
}
