// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package testutils

import (
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

var (
	metricFile string
	host       string
	mtx        sync.Mutex

	hostname string
)

type CaseStatus int

const (
	CaseStatusUnknown CaseStatus = iota
	CasePassed
	CaseFailed
	CaseSkipped
)

func (c CaseStatus) String() string {
	switch c {
	case CasePassed:
		return "pass"
	case CaseFailed:
		return "fail"
	case CaseSkipped:
		return "skip"

	default:
		return "unknown"
	}
}

type CaseResult struct {
	Name string
	Case string

	FailedMessage string

	Status CaseStatus
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
		"cost": int64(cr.Cost),
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

	p, err := point.NewPoint("testing", tags, fields, nil)
	if err != nil {
		// return commeted info
		return fmt.Sprintf("# invalid result: %s, CaseResult: %+#v\n", err.Error(), cr)
	}

	return p.String() + "\n"
}

func (cr *CaseResult) Flush() error {
	mtx.Lock()
	defer mtx.Unlock()

	first := false

	if metricFile == "" {
		if v := os.Getenv("TESTING_METRIC_PATH"); v != "" {
			metricFile = v
		} else {
			metricFile = "./testing_metrics"
		}

		if err := os.MkdirAll(path.Dir(metricFile), os.ModePerm); err != nil {
			return err
		}

		first = true
	}

	f, err := os.OpenFile(metricFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	defer f.Close()

	if first {
		if _, err := f.WriteString(fmt.Sprintf("# metrics for %s\n", cr.Name)); err != nil {
			return err
		}
	}

	if _, err := f.WriteString(cr.LineProtocol()); err != nil {
		return err
	}

	return nil
}

func (cr *CaseResult) Post() error {
	// TODO: post to some datakit://v1/write/metrics
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
