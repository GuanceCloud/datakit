// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"reflect"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestInput_AvailableArchs(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "case all arch",
			want: datakit.AllArch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := &Input{}
			if got := in.AvailableArchs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AvailableArchs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInput_Catalog(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "test catalog 1",
			want: inputName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Input{}
			if got := i.Catalog(); got != tt.want {
				t.Errorf("Catalog() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInput_RegHTTPHandler(t *testing.T) {
	type fields struct {
		OHTTPc *otlpHTTPCollector
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "case1",
			fields: fields{OHTTPc: &otlpHTTPCollector{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Input{
				OHTTPc: tt.fields.OHTTPc,
			}
			i.RegHTTPHandler()
		})
	}
}

func TestInput_Run(t *testing.T) {
	type fields struct {
		Ogrpc               *otlpGrpcCollector
		OHTTPc              *otlpHTTPCollector
		CloseResource       map[string][]string
		Sampler             *itrace.Sampler
		IgnoreAttributeKeys []string
		Tags                map[string]string
		inputName           string
		semStop             *cliutils.Sem
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "case1",
			fields: fields{
				Ogrpc: &otlpGrpcCollector{
					TraceEnable: true, Addr: "127.0.0.1:4317",
				},
				OHTTPc: &otlpHTTPCollector{
					Enable:          true,
					HTTPStatusOK:    200,
					ExpectedHeaders: nil,
				},
				CloseResource:       map[string][]string{"service1": {"name1", "name2"}},
				Sampler:             &itrace.Sampler{SamplingRateGlobal: 0},
				IgnoreAttributeKeys: []string{"os_*"},
				Tags:                map[string]string{},
				inputName:           "opentelemetry",
				semStop:             cliutils.NewSem(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Input{
				Ogrpc:               tt.fields.Ogrpc,
				OHTTPc:              tt.fields.OHTTPc,
				CloseResource:       tt.fields.CloseResource,
				Sampler:             tt.fields.Sampler,
				IgnoreAttributeKeys: tt.fields.IgnoreAttributeKeys,
				Tags:                tt.fields.Tags,
				inputName:           tt.fields.inputName,
				semStop:             tt.fields.semStop,
			}
			go i.Run()
			stop := make(chan int)
			go func() {
				time.Sleep(time.Second * 4)
				if i.OHTTPc.storage == nil {
					t.Errorf("storage == nil")
				}
				i.semStop.Close() // stop
				stop <- 1
			}()
			<-stop
		})
	}
}

func TestInput_SampleConfig(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "test catalog 1",
			want: sampleConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Input{}
			if got := i.SampleConfig(); got != tt.want {
				t.Errorf("Catalog() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInput_SampleMeasurement(t *testing.T) {
	tests := []struct {
		name string
		want []inputs.Measurement
	}{
		{
			name: "name1",
			want: []inputs.Measurement{&itrace.TraceMeasurement{Name: inputName}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := &Input{}
			if got := in.SampleMeasurement(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SampleMeasurement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInput_exit(t *testing.T) {
	type fields struct {
		Ogrpc *otlpGrpcCollector
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "case1_stopFunc_is_nil",
			fields: fields{Ogrpc: &otlpGrpcCollector{
				TraceEnable:  false,
				MetricEnable: false,
				Addr:         "",
			}},
		},
		{
			name: "case2_not_fil",
			fields: fields{Ogrpc: &otlpGrpcCollector{
				TraceEnable:  false,
				MetricEnable: false,
				Addr:         "",
				stopFunc: func() {
					t.Log("to stop")
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Input{
				Ogrpc: tt.fields.Ogrpc,
			}
			i.exit()
		})
	}
}
