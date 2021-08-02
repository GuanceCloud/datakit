package kubernetes

import (
	"encoding/json"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type resource interface {
	Gather()
	LineProto() (*io.Point, error)
	Info() *inputs.MeasurementInfo
}

var resourceList = map[string]resource{
	"cluster":    &cluster{},
	"pod":        &pod{},
	"deployment": &deployment{},
	"replicaSet": &replicaSet{},
	"service":    &service{},
	"node":       &node{},
	"job":        &job{},
	"cronJob":    &cronJob{},
}

func addJSONStringToMap(key string, value interface{}, m map[string]interface{}) {
	if value == nil {
		return
	}
	b, err := json.Marshal(value)
	if err != nil {
		return
	}
	m[key] = string(b)
}

func addMessageToFields(tags map[string]string, fields map[string]interface{}) {
	var temp = make(map[string]interface{})
	for k, v := range tags {
		temp[k] = v
	}
	for k, v := range fields {
		fields[k] = v
	}
	addJSONStringToMap("message", temp, fields)
}
