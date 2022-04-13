package sinkm3db

import (
	"reflect"
	"testing"
	"time"
)

func Test_makeSeries(t *testing.T) {
	type args struct {
		tags     map[string]string
		key      string
		i        interface{}
		dataTime time.Time
	}
	tests := []struct {
		name string
		args args
		want []*TimeSeries
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeSeries(tt.args.tags, tt.args.key, tt.args.i, tt.args.dataTime); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeSeries() = %v, want %v", got, tt.want)
			}
		})
	}
}
