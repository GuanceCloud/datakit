// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package skywalking testing.
package skywalking

import (
	"net/http"
	"testing"
)

func TestSkyConsumer_Init(t *testing.T) {
	type fields struct {
		DKEndpoint string
		Topics     []string
		Namespace  string
		topics     []string
		client     http.RoundTripper
		dkURLs     map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "logging",
			fields: fields{
				DKEndpoint: "http://127.0.0.1:9529",
				Topics:     []string{"skywalking-logging"},
				Namespace:  "",
			},
			wantErr: false,
		},
		{
			name: "dkURLs",
			fields: fields{
				DKEndpoint: "127.0.0.1:9529",
				Topics:     []string{"skywalking-logging"},
				Namespace:  "",
			},
			wantErr: true,
		},
		{
			name: "metrics",
			fields: fields{
				DKEndpoint: "http://127.0.0.1:9529",
				Topics:     []string{"skywalking-metrics"},
				Namespace:  "",
			},
			wantErr: false,
		},
		{
			name: "namespace",
			fields: fields{
				DKEndpoint: "http://127.0.0.1:9529",
				Topics:     []string{"skywalking-logging"},
				Namespace:  "kafka_01",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sky := &SkyConsumer{
				DKEndpoint: tt.fields.DKEndpoint,
				Topics:     tt.fields.Topics,
				Namespace:  tt.fields.Namespace,
				topics:     tt.fields.topics,
				client:     tt.fields.client,
				dkURLs:     tt.fields.dkURLs,
			}
			if err := sky.Init(); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
			if sky.Namespace != "" {
				for _, topic := range sky.Topics {
					if _, ok := sky.dkURLs[sky.Namespace+"-"+topic]; !ok {
						t.Errorf("can not find topic=%s from dkURLs", topic)
					}
				}
			}
		})
	}
}
