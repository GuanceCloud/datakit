package zipkin

import (
	"testing"

	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
)

func TestZipkinAgent(t *testing.T) {
	afterGatherRun = itrace.AfterGatherFunc(func(inputName string, dktrace itrace.DatakitTrace, strikMod bool) {})

	testHTTPServerV1(t)
	testHTTPServerV2(t)
}

func testHTTPServerV1(t *testing.T) {
	t.Helper()
}

func testHTTPServerV2(t *testing.T) {
	t.Helper()
}
