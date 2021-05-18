package dialtesting

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func Feed(name, category string, pt *io.Point, opt *io.Option) error {

	pts := []*io.Point{}
	pts = append(pts, pt)

	return io.Feed(name, category, pts, opt)
}
