package kubernetes

import (
	"encoding/json"
	"strings"

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

func addMapToFields(key string, value map[string]string, fields map[string]interface{}) {
	if value == nil || len(value) == 0 {
		// 如果该 map 为空，则对应值为空字符串，否则在 json 序列化为 "null"
		fields[key] = defaultStringValue
		return
	}
	b, err := json.Marshal(value)
	if err != nil {
		return
	}
	fields[key] = string(b)
}

func addSliceToFields(key string, value []string, fields map[string]interface{}) {
	fields[key] = strings.Join(value, ",")
}

func addMessageToFields(tags map[string]string, fields map[string]interface{}) {
	var temp = make(map[string]interface{})
	for k, v := range tags {
		temp[k] = v
	}
	for k, v := range fields {
		fields[k] = v
	}
	b, err := json.Marshal(temp)
	if err != nil {
		return
	}
	fields["message"] = string(b)
}
