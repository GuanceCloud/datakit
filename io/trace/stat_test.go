package trace

import (
	"testing"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestStatTracingInfo(t *testing.T) {
	ioFeed = func(name, category string, pts []*dkio.Point, opt *dkio.Option) error { return nil }
}
