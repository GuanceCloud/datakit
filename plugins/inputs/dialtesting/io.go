package dialtesting

import (
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func Feed(name, category string, pt *io.Point, opt *io.Option) error {

	pts := []*io.Point{}
	pts = append(pts, pt)

	return io.Feed(name, category, pts, opt)
}

func LineDataFeed(data string, name, category string, opt *io.Option) error {
	pts, err := lp.ParsePoints([]byte(data), nil)
	if err != nil {
		return err
	}

	x := []*io.Point{}
	for _, pt := range pts {
		x = append(x, &io.Point{Point: pt})
	}

	return io.Feed(name, category, x, opt)
}
