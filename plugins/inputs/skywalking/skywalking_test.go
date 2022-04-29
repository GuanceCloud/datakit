package skywalking

import (
	"testing"

	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
)

func TestSkyWalkingAgent(t *testing.T) {
	afterGatherRun = itrace.AfterGatherFunc(func(inputName string, dktrace itrace.DatakitTrace, strikMod bool) {})
}
