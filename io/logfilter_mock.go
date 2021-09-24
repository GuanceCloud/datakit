package io

// import "gitlab.jiagouyun.com/cloudcare-tools/datakit"

var defLogFilterMock logFilterMock = &prodLogFilterMock{}

type logFilterMock interface {
	getLogFilter() ([]byte, error)
	preparePoints(pts []*Point) []*Point
}

type prodLogFilterMock struct{}

func (*prodLogFilterMock) getLogFilter() ([]byte, error) {
	return defaultIO.dw.GetLogFilter()
}

func (*prodLogFilterMock) preparePoints(pts []*Point) []*Point {
	return pts
}
