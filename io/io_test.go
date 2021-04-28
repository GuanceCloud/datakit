package io

import (
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var x *IO

func StartCollect() error {
	// l = logger.SLogger("dialtesting_io")

	x.DatawayHost = "https://openway.dataflux.cn?token=tkn_76d2d1efd3ff43db984497bfb4f3c25a;https://openway.dataflux.cn/v1/write/metrics?token=tkn_a5cbdacf23214966aa382ae0182e972b"

	// x.MaxCacheCnt = 200
	// if datakit.Cfg.MainCfg.DataWay.Timeout != "" {
	// 	du, err := time.ParseDuration(datakit.Cfg.MainCfg.DataWay.Timeout)
	// 	if err != nil {
	// 		l.Warnf("parse dataway timeout failed: %s, default 30s", err.Error())
	// 	} else {
	// 		x.HTTPTimeout = du
	// 	}
	// }

	if datakit.OutputFile != "" {
		x.OutputFile = datakit.OutputFile
	}

	x.FlushInterval = 1 * time.Second

	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		x.StartIO(true)
	}()

	l.Debugf("io: %+#v", x)

	return nil
}

func FeedX(name, category string, pt *Point, opt *Option) error {
	pts := []*Point{}
	pts = append(pts, pt)

	return x.DoFeed(pts, category, name, opt)
}

func TestFlush(t *testing.T) {
	x = NewIO()
	StartCollect()

	for i := 0; i < 100; i++ {
		l.Debug("loop index", i)
		time.Sleep(1 * time.Second)
		tags := map[string]string{
			"abc": "123",
		}

		fields := map[string]interface{}{
			"value1": 123,
			"value2": 234,
		}

		data, err := MakePoint("test", tags, fields, time.Now())
		if err != nil {
			l.Warnf("make metric failed: %s", err.Error)
		}

		FeedX("test", Metric, data, &Option{})
	}
}
