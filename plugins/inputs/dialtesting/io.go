package dialtesting

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

// dataway (todo)
func StartCollect() error {
	// x.DatawayHost = datakit.Cfg.MainCfg.DataWay.URL

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
