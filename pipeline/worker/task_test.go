package worker

import (
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func TestResult_checkFieldValLen(t *testing.T) {
	type fields struct {
		output      *parser.Output
		measurement string
		ts          time.Time
		err         string
	}
	type args struct {
		messageLen int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "case1",
			fields: fields{
				output: &parser.Output{
					Dropped: false,
					Error:   "",
					Cost:    nil,
					Tags:    nil,
					Data: map[string]interface{}{
						"msg":         "0123456789",
						"message":     "0123456789",
						"other_field": "0123456789",
					},
				},
				measurement: "logging",
				ts:          time.Now(),
				err:         "",
			},
			args: args{messageLen: 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{
				output:      tt.fields.output,
				measurement: tt.fields.measurement,
				ts:          tt.fields.ts,
				err:         tt.fields.err,
			}
			r.checkFieldValLen(tt.args.messageLen)

			for key := range r.output.Data {
				if i, err := r.GetField(key); err == nil {
					if mass, isString := i.(string); isString {
						if len(mass) > tt.args.messageLen {
							t.Errorf("key=%s val=%s over massageLen:%d", key, mass, tt.args.messageLen)
						}
					}
				}
			}
		})
	}
}
