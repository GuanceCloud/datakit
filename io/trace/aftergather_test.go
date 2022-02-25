package trace

import (
	"fmt"
	"testing"
	"time"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestAfterGather(t *testing.T) {
	dkioFeed = func(name, category string, pts []*dkio.Point, opt *dkio.Option) error { return nil }

	afterGather := NewAfterGather()
	afterGather.AppendCalculator(StatTracingInfo)
	closer := &CloseResource{}
	closer.UpdateIgnResList(map[string][]string{
		_services[0]: _resources,
	})
	afterGather.AppendFilter(closer.Close)
	keeper := &KeepRareResource{
		Open:     true,
		Duration: time.Second,
	}
	afterGather.AppendFilter(keeper.Keep)

	for i := 0; i < 100; i++ {
		trace := randDatakitTrace(t, 10)
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
