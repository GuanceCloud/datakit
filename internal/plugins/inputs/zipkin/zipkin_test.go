// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package zipkin

import (
	"testing"

	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func TestZipkinAgent(t *testing.T) {
	afterGatherRun = itrace.AfterGatherFunc(func(inputName string, dktraces itrace.DatakitTraces) {})

	testHTTPServerV1(t)
	testHTTPServerV2(t)
}

func testHTTPServerV1(t *testing.T) {
	t.Helper()
}

func testHTTPServerV2(t *testing.T) {
	t.Helper()
}
