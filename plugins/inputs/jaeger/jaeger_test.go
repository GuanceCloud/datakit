// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jaeger

import (
	"testing"

	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func TestJaegerAgent(t *testing.T) {
	afterGatherRun = itrace.AfterGatherFunc(func(inputName string, dktrace itrace.DatakitTrace, strikMod bool) {})

	testHTTPHandler(t)
	testUDPClient(t)
}

func testHTTPHandler(t *testing.T) {
	t.Helper()
}

func testUDPClient(t *testing.T) {
	t.Helper()
}
