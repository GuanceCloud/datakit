package kubernetes

import (
	"encoding/json"
	"strings"
)

func addMapToFields(key string, value map[string]string, fields map[string]interface{}) {
	if fields == nil {
		return
	}

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
	if fields == nil {
		return
	}
	fields[key] = strings.Join(value, ",")
}

func addMessageToFields(tags map[string]string, fields map[string]interface{}) {
	if fields == nil {
		return
	}

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

func addLabelToFields(labels map[string]string, fields map[string]interface{}) {
	if fields == nil {
		return
	}

	// empty array
	labelsString := "[]"

	if len(labels) != 0 {
		var lb []string
		for k, v := range labels {
			lb = append(lb, k+":"+v)
		}

		b, err := json.Marshal(lb)
		if err == nil {
			labelsString = string(b)
		}
	}

	// http://gitlab.jiagouyun.com/cloudcare-tools/kodo/-/issues/61#note_11580
	fields["df_label"] = labelsString
	fields["df_label_permission"] = "read_only"
	fields["df_label_source"] = "datakit"
}
