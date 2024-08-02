// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package metrics

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ErrCountVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Name:      "error_total",
			Help:      "Total errors, only count on error source, not include error message",
		},
		[]string{
			"source",
			"category",
		},
	)

	LastErrVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: DatakitLastError,
			Help: "Datakit errors(when error occurred), these errors come from inputs or any sub modules",
		},
		[]string{
			"input",
			"source",
			"category",
			"error",
		},
	)
)

type LastError struct {
	Input, Source string
	Categories    []point.Category
}

type LastErrorOption func(*LastError)

const defaultInputSource = "not-set"

func NewLastError() *LastError {
	return &LastError{
		Input:  defaultInputSource,
		Source: defaultInputSource,
	}
}

func WithLastErrorInput(input string) LastErrorOption {
	return func(le *LastError) {
		le.Input = input
		if len(le.Source) == 0 || le.Source == defaultInputSource { // If Source is empty, filling with Input.
			le.Source = input
		}
	}
}

func WithLastErrorSource(source string) LastErrorOption {
	return func(le *LastError) {
		le.Source = source
		if len(le.Input) == 0 || le.Input == defaultInputSource { // If Input is empty, filling with Source.
			le.Input = source
		}
	}
}

func WithLastErrorCategory(cats ...point.Category) LastErrorOption {
	return func(le *LastError) {
		le.Categories = cats
	}
}

// FeedLastError feed some error message(*unblocking*) to inputs stats
// we can see the error in monitor.
//
// NOTE: the error may be skipped if there is too many error.
//
// Deprecated: should use DefaultFeeder to get global default feeder.
func FeedLastError(source, err string, cat ...point.Category) {
	le := NewLastError()

	catStr := point.SUnknownCategory
	if len(cat) == 1 {
		catStr = le.Categories[0].String()
	}

	ErrCountVec.WithLabelValues(le.Source, catStr).Inc()
	LastErrVec.WithLabelValues(le.Input, le.Source, catStr, err).Set(float64(time.Now().Unix()))
}
