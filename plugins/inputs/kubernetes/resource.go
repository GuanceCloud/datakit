package kubernetes

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type resource interface {
	Gather()
	LineProto() (*io.Point, error)
	Info() *inputs.MeasurementInfo
}

var resourceList = map[string]resource{
	"cluster": &cluster{},
	// "pod":        &pod{},
	"deployment": &deployment{},
	"replicaSet": &replicaSet{},
	"service":    &service{},
	"node":       &node{},
	"job":        &job{},
	"cronJob":    &cronJob{},
}

type exporter interface {
	Export()
	Stop()
}
