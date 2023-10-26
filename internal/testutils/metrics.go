// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package testutils

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

var (
	metricFile string
	DatawayURL = "-"

	mtx sync.Mutex

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
	TestID        string        `json:"test_id"`
	Name          string        `json:"name"`
	OS            string        `json:"os"`
	Arch          string        `json:"arch"`
	GoVersion     string        `json:"go_version"`
	Branch        string        `json:"branch"`
	Cost          time.Duration `json:"cost"`
	Status        TestStatus    `json:"status"`
	Coverage      float64       `json:"coverage"`
	NoTest        bool          `json:"no_test"`
	FailedMessage string        `json:"failed_message"`
	Message       string        `json:"message"`
}

func (mr *ModuleResult) LineProtocol() string {
	tags := map[string]string{
		"arch":    mr.Arch,
		"branch":  mr.Branch,
		"go":      mr.GoVersion,
		"host":    hostname,
		"name":    mr.Name,
		"os":      mr.OS,
		"status":  mr.Status.String(),
		"test_id": mr.TestID,
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

	p := point.NewPointV2("testing_module", append(point.NewTags(tags), point.NewKVs(fields)...))

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

	p := point.NewPointV2("testing", append(point.NewTags(tags), point.NewKVs(fields)...))

	return p.LineProto() + "\n"
}

// Flush write line-protocol data to file or remote HTTP server.
func Flush(m TestingMetric) error {
	mtx.Lock()
	defer mtx.Unlock()

	lp := m.LineProtocol()

	if err := flushToFile([]byte(lp)); err != nil {
		return err
	}

	if err := flushToDataway([]byte(lp)); err != nil {
		return err
	}

	return nil
}

func flushToDataway(data []byte) error {
	if len(DatawayURL) > 10 {
		log.Printf("write to dataway %s**********", DatawayURL[:len(DatawayURL)-10])
	} else {
		log.Printf("write to dataway %s", DatawayURL)
	}

	if !strings.HasPrefix(DatawayURL, "http") {
		return nil
	}

	// nolint: gosec
	resp, err := http.Post(DatawayURL, "", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	defer resp.Body.Close() // nolint: errcheck

	switch resp.StatusCode / 100 {
	case 2:
		return nil
	default:
		return fmt.Errorf("post dataway %s", resp.Status)
	}
}

func flushToFile(data []byte) error {
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

	f, err := os.OpenFile(filepath.Clean(metricFile), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	defer f.Close() //nolint:errcheck,gosec

	if _, err := f.Write(data); err != nil {
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
