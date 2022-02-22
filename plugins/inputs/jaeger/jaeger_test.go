package jaeger

import (
	"testing"

	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
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
