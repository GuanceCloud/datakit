package dialtesting

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func StartCollect() error {
	// l = logger.SLogger("dialtesting_io")

	x.DatawayHost = datakit.Cfg.MainCfg.DataWay.URL

	x.MaxCacheCnt = 200
	if datakit.Cfg.MainCfg.DataWay.Timeout != "" {
		du, err := time.ParseDuration(datakit.Cfg.MainCfg.DataWay.Timeout)
		if err != nil {
			l.Warnf("parse dataway timeout failed: %s, default 30s", err.Error())
		} else {
			x.HTTPTimeout = du
		}
	}

	if datakit.OutputFile != "" {
		x.OutputFile = datakit.OutputFile
	}

	if datakit.Cfg.MainCfg.StrictMode {
		x.StrictMode = true
	}

	x.FlushInterval = datakit.IntervalDuration

	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		x.StartIO(true)
	}()

	l.Debugf("io: %+#v", x)

	return nil
}

func Feed(name, category string, pt *io.Point, opt *io.Option) error {

	pts := []*io.Point{}
	pts = append(pts, pt)

	return x.DoFeed(pts, category, name, opt)
}
