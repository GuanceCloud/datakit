// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package zipkin handle Zipkin APM traces.
package zipkin

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	zpkmodel "github.com/openzipkin/zipkin-go/model"
)

func TestTraceID(t *testing.T) {
	traceID := zpkmodel.TraceID{
		High: 0xaaaaaa,
		Low:  0xbbbbbb,
	}

	t.Logf("trace_id =%x%x", traceID.High, traceID.Low)
	t.Logf("trace_id =%016x%016x", traceID.High, traceID.Low)

	assert.Len(t, fmt.Sprintf("%016x%016x", traceID.High, traceID.Low), 32)
}
