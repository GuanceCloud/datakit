package trace

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestStatTracingInfo(t *testing.T) {
	dkioFeed = func(name, category string, pts []*dkio.Point, opt *dkio.Option) error { return nil }

	var (
		traces    DatakitTraces
		services  = []string{"login", "name_service", "logout"}
		resources = []string{"/get_user/name", "/push/data", "/check/security"}
	)
	for i := 0; i < 100; i++ {
		t := randDatakitTraceByService(t, 10, testutils.RandWithinStrings(services), testutils.RandWithinStrings(resources), "")
		traces = append(traces, t)
	}

	StartTracingStatistic()

	for i := range traces {
		StatTracingInfo(traces[i])
	}
}
