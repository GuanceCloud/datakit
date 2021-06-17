package io

import "gitlab.jiagouyun.com/cloudcare-tools/datakit"

var (
	defLogFilterMock logFilterMock = &prodLogFilterMock{}
)

type logFilterMock interface {
	GetLogFilter() ([]byte, error)
}

type prodLogFilterMock struct{}

func (*prodLogFilterMock) GetLogFilter() ([]byte, error) {
	return datakit.Cfg.DataWay.GetLogFilter()
}
