package ddtrace

import (
	"io"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var (
	defDDTraceMock ddtraceMock = &prodDDTraceMock{}
)

type ddtraceMock interface {
	unmarshalDdtraceMsgpack(body io.ReadCloser) ([][]*Span, error)
	statistic(origin [][]*Span, sampled []*dkio.Point)
}

type prodDDTraceMock struct{}

func (this *prodDDTraceMock) unmarshalDdtraceMsgpack(body io.ReadCloser) ([][]*Span, error) {
	return unmarshalDdtraceMsgpack(body)
}

func (this *prodDDTraceMock) statistic(origin [][]*Span, sampled []*dkio.Point) {}
