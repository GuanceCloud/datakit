package kubernetes

import (
	"testing"
)

func TestAddLabelToFields(t *testing.T) {
	cases := []struct {
		labels map[string]string
		fiedls map[string]interface{}
	}{
		{
			map[string]string{
				"label_key1": "label_value1",
				"label_key2": "label_value2",
			},
			map[string]interface{}{
				"fields_key1": "fields_value1",
			},
		},
		{
			map[string]string{
				"label_key1": "label_value1",
				"label_key2": "label_value2",
			},
			map[string]interface{}{},
		},
		{
			map[string]string{
				"label_key1": "label_value1",
				"label_key2": "label_value2",
				"label_key3": "label_value3",
			},
			nil,
		},
		{
			map[string]string{},
			map[string]interface{}{
				"fields_key1": "fields_value1",
			},
		},
	}

	for idx, tc := range cases {
		addLabelToFields(tc.labels, tc.fiedls)
		t.Logf("[%d] %v\n", idx, tc.fiedls)
	}
}
