// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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
	timeNow := time.Now()
	tests := []struct {
		name string
		args args
		want []*TimeSeries
	}{
		{
			name: "case_int",
			args: args{
				tags:     map[string]string{"host": "testHostName"},
				key:      "tot",
				i:        int64(10),
				dataTime: timeNow,
			},
			want: []*TimeSeries{
				&TimeSeries{
					Labels: []Label{
						{
							Name:  "host",
							Value: "testHostName",
						},
						{
							Name:  "__name__",
							Value: "tot",
						},
					},
					Datapoint: Datapoint{
						Timestamp: timeNow,
						Value:     10,
					},
				},
			},
		},
		{
			name: "case_uint64",
			args: args{
				tags:     map[string]string{"host": "testHostName"},
				key:      "tot",
				i:        int64(10),
				dataTime: timeNow,
			},
			want: []*TimeSeries{
				&TimeSeries{
					Labels: []Label{
						{
							Name:  "host",
							Value: "testHostName",
						},
						{
							Name:  "__name__",
							Value: "tot",
						},
					},
					Datapoint: Datapoint{
						Timestamp: timeNow,
						Value:     10,
					},
				},
			},
		},
		{
			name: "case_float64",
			args: args{
				tags:     map[string]string{"host": "testHostName"},
				key:      "tot",
				i:        float64(10),
				dataTime: timeNow,
			},
			want: []*TimeSeries{
				&TimeSeries{
					Labels: []Label{
						{
							Name:  "host",
							Value: "testHostName",
						},
						{
							Name:  "__name__",
							Value: "tot",
						},
					},
					Datapoint: Datapoint{
						Timestamp: timeNow,
						Value:     10,
					},
				},
			},
		},
		{
			name: "case_map",
			args: args{
				tags:     map[string]string{"host": "testHostName"},
				key:      "tot",
				i:        map[string]int64{"cpu": 10, "mem": 20},
				dataTime: timeNow,
			},
			want: []*TimeSeries{
				&TimeSeries{
					Labels: []Label{
						{
							Name:  "host",
							Value: "testHostName",
						},
						{
							Name:  "__name__",
							Value: "cpu",
						},
					},
					Datapoint: Datapoint{
						Timestamp: timeNow,
						Value:     10,
					},
				},
				&TimeSeries{
					Labels: []Label{
						{
							Name:  "host",
							Value: "testHostName",
						},
						{
							Name:  "__name__",
							Value: "mem",
						},
					},
					Datapoint: Datapoint{
						Timestamp: timeNow,
						Value:     20,
					},
				},
			},
		},
		{
			name: "case_array",
			args: args{
				tags:     map[string]string{"host": "testHostName"},
				key:      "tot",
				i:        []int64{10, 20},
				dataTime: timeNow,
			},
			want: []*TimeSeries{
				&TimeSeries{
					Labels: []Label{
						{
							Name:  "host",
							Value: "testHostName",
						},
						{
							Name:  "__name__",
							Value: "tot",
						},
					},
					Datapoint: Datapoint{
						Timestamp: timeNow,
						Value:     10,
					},
				},
				&TimeSeries{
					Labels: []Label{
						{
							Name:  "host",
							Value: "testHostName",
						},
						{
							Name:  "__name__",
							Value: "tot",
						},
					},
					Datapoint: Datapoint{
						Timestamp: timeNow,
						Value:     20,
					},
				},
			},
		},
		{
			name: "case_others",
			args: args{
				tags:     map[string]string{"host": "testHostName"},
				key:      "bool",
				i:        true,
				dataTime: timeNow,
			},
			want: []*TimeSeries{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeSeries(tt.args.tags, tt.args.key, tt.args.i, tt.args.dataTime); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeSeries() = %v, want %v", got, tt.want)
			}
		})
	}
}
