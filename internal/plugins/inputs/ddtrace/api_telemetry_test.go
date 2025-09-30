// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ddtrace ddtrace apm telemetry
package ddtrace

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseTelemetryRequest(t *testing.T) {
	om := Manager{
		OBS:    map[string]*jvmTelemetry{},
		OBChan: make(chan *jvmTelemetry, 10),
	}
	header := http.Header{
		"Dd-Telemetry-Request-Type": []string{"no_use"},
		"Dd-Telemetry-Api-Version":  []string{"v2"},
	}
	tests := []struct {
		name               string
		requestType        RequestType
		requestTypePayload []byte
		object             interface{}
	}{
		{
			name:        "app-started",
			requestType: RequestTypeAppStarted,
			object: AppStarted{
				Configuration: []Configuration{
					{Name: "dd.trace.enable", Value: true, Origin: "default"},
				},
				Products: Products{
					AppSec: ProductDetails{Enabled: true},
				},
			},
		},
		{
			name:        "generate-metrics",
			requestType: RequestTypeGenerateMetrics,
			object: Metrics{
				Namespace: "none",
				Series: []Series{
					{
						Metric: "spans_created",
						Points: [][2]float64{
							{11, 1}, {12, 3},
						},
						Tags:      []string{"type:datadog"},
						Common:    false,
						Namespace: "",
					},
					{
						Metric: "spans_finished",
						Points: [][2]float64{
							{11, 1}, {12, 3},
						},
						Tags:      []string{"type:http"},
						Common:    false,
						Namespace: "",
					},
				},
			},
		},
		{
			name:        "app-integrations-change",
			requestType: RequestTypeAppIntegrationsChange,
			object: &IntegrationsChange{
				Integrations: []Integration{
					{Name: "mysql", Enabled: true},
				},
			},
		},
	}
	commonBody := &Body{
		APIVersion: "v2",
		TracerTime: 0,
		RuntimeID:  "a-b-c-d",
		SeqID:      0,
		Debug:      false,
		Payload:    nil,
		Application: Application{
			ServiceName:     "tmall",
			Env:             "-",
			ServiceVersion:  "v1.0.1",
			TracerVersion:   "1.30.1-guance",
			LanguageName:    "jvm",
			LanguageVersion: "jdk-1.8",
			RuntimeName:     "",
			RuntimeVersion:  "",
			RuntimePatches:  "",
		},
		Host: Host{
			Hostname:   "test_hostname",
			OS:         "linux",
			KernelName: "linux",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commonBody.RequestType = tt.requestType
			commonBody.Payload = tt.object
			bts, _ := json.Marshal(commonBody)
			om.parseTelemetryRequest(header, bts)
		})
	}

	// check object.
	isbreak, mastBreak := 0, 0
	for {
		select {
		case ob := <-om.OBChan:
			t.Logf("get ob fields len=%d", len(ob.fields))
			isbreak++
		default:
			t.Logf("none")
			mastBreak++
		}
		if isbreak >= len(tests) || mastBreak > 3 {
			break
		}
	}

	if ob, ok := om.OBS[commonBody.Application.ServiceName+commonBody.RuntimeID]; ok && ob != nil {
		assert.Equal(t, ob.tags["hostname"], commonBody.Host.Hostname)
		assert.Equal(t, ob.tags["os"], commonBody.Host.OS)
		assert.Equal(t, ob.tags["kernel_name"], commonBody.Host.KernelName)
		assert.Equal(t, ob.tags["service"], commonBody.Application.ServiceName)
		assert.Equal(t, ob.tags["service_version"], commonBody.Application.ServiceVersion)

		assert.NotEmpty(t, ob.fields["app_started"])
		assert.NotEmpty(t, ob.fields["app_integrations_change"])

		assert.Equal(t, ob.fields["spans_created_type_datadog"], float64(4))
		assert.Equal(t, ob.fields["spans_finished_type_http"], float64(4))
	} else {
		t.Errorf("can find telemetry")
	}
}
