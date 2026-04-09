// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package collect

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFeedPoint(t *testing.T) {
	tests := []struct {
		name       string
		points     []*point.Point
		serverResp string
		wantErr    bool
	}{
		{
			name:       "empty points",
			points:     []*point.Point{},
			serverResp: "OK",
			wantErr:    false,
		},
		{
			name: "single point",
			points: []*point.Point{
				point.NewPoint("test", point.KVs{
					point.NewKV("message", "test message"),
					point.NewKV("status", "info"),
				}),
			},
			serverResp: "OK",
			wantErr:    false,
		},
		{
			name: "multiple points",
			points: []*point.Point{
				point.NewPoint("test", point.KVs{
					point.NewKV("message", "message 1"),
					point.NewKV("status", "info"),
				}),
				point.NewPoint("test", point.KVs{
					point.NewKV("message", "message 2"),
					point.NewKV("status", "warning"),
				}),
			},
			serverResp: "OK",
			wantErr:    false,
		},
		{
			name: "server error",
			points: []*point.Point{
				point.NewPoint("test", point.KVs{
					point.NewKV("message", "test message"),
					point.NewKV("status", "info"),
				}),
			},
			serverResp: "Internal Server Error",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldTransport := http.DefaultClient.Transport
			t.Cleanup(func() {
				http.DefaultClient.Transport = oldTransport
			})

			http.DefaultClient.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
				statusCode := http.StatusOK
				if tt.serverResp == "Internal Server Error" {
					statusCode = http.StatusInternalServerError
				}

				return &http.Response{
					StatusCode: statusCode,
					Body:       io.NopCloser(strings.NewReader(tt.serverResp)),
					Header:     make(http.Header),
				}, nil
			})

			ipt := &Input{
				dkURLPath: "http://datakit.test/v1/write/logging",
			}

			err := ipt.feedPoint(tt.points)

			if (err != nil) != tt.wantErr {
				t.Errorf("feedPoint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriteData(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		serverResp string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful write",
			data:       []byte("test data"),
			serverResp: "OK",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty data",
			data:       []byte{},
			serverResp: "OK",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "server error",
			data:       []byte("test data"),
			serverResp: "Internal Server Error",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
		{
			name:       "bad request",
			data:       []byte("invalid data"),
			serverResp: "Bad Request",
			statusCode: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldTransport := http.DefaultClient.Transport
			t.Cleanup(func() {
				http.DefaultClient.Transport = oldTransport
			})

			http.DefaultClient.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(strings.NewReader(tt.serverResp)),
					Header:     make(http.Header),
				}, nil
			})

			ipt := &Input{
				dkURLPath: "http://datakit.test/v1/write/logging",
			}

			err := ipt.writeData(tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("writeData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInput_NoURLPath(t *testing.T) {
	points := []*point.Point{
		point.NewPoint("test", point.KVs{
			point.NewKV("message", "test message"),
		}),
	}

	ipt := &Input{
		dkURLPath: "",
	}

	err := ipt.feedPoint(points)
	if err != nil {
		t.Errorf("feedPoint() with empty URL should not return error, got %v", err)
	}
}

func TestInput_ContextTimeout(t *testing.T) {
	oldTransport := http.DefaultClient.Transport
	t.Cleanup(func() {
		http.DefaultClient.Transport = oldTransport
	})

	http.DefaultClient.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("OK")),
			Header:     make(http.Header),
		}, nil
	})

	ipt := &Input{
		dkURLPath: "http://datakit.test/v1/write/logging",
	}

	data := []byte("test data")
	err := ipt.writeData(data)
	if err != nil {
		t.Errorf("writeData() unexpected error: %v", err)
	}
}
