package io

import "testing"

type debugLogFilterMock struct{}

func (*debugLogFilterMock) GetLogFilter() ([]byte, error) {
	return []byte(`
{
	"content": ["{source == mongodb}"]
}
`), nil
}

func TestLogFilter(t *testing.T) {
	defLogFilterMock = &debugLogFilterMock{}
}
