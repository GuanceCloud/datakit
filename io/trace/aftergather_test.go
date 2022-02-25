package trace

import (
	"fmt"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestAfterGather(t *testing.T) {
	dkioFeed = func(name, category string, pts []*dkio.Point, opt *dkio.Option) error { return nil }

	StartTracingStatistic()

	afterGather := NewAfterGather()
	afterGather.AppendCalculator(StatTracingInfo)

	closer := &CloseResource{}
	closer.UpdateIgnResList(map[string][]string{_services[0]: _resources})
	afterGather.AppendFilter(closer.Close)

	keeper := &KeepRareResource{}
	keeper.UpdateStatus(true, time.Second)
	afterGather.AppendFilter(keeper.Keep)

	sampler := &Sampler{}
	sampler.UpdateArgs(PriorityAuto, 0.33)
	afterGather.AppendFilter(sampler.Sample)

	for i := 0; i < 100; i++ {
		trace := randDatakitTraceByService(t, 10, testutils.RandWithinStrings(_services), testutils.RandWithinStrings(_resources), "")
		parentialize(trace)
		afterGather.Run("test_after_gather", trace, false)
	}
}

func TestBuildPoint(t *testing.T) {
	for i := 0; i < 100; i++ {
		if pt, err := BuildPoint(randDatakitSpan(t), false); err != nil {
			t.Error(err.Error())
			t.FailNow()
		} else {
			fmt.Println(pt.String())
		}
	}
}

func TestBuildPointsBatch(t *testing.T) {
	for i := 0; i < 100; i++ {
		BuildPointsBatch(randDatakitTrace(t, 10), false)
	}
}
