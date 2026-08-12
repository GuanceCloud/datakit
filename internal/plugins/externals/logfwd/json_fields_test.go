// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package logfwd

import (
	"encoding/json"
	"testing"
)

func TestForwardJSONAsFieldsFlag(t *testing.T) {
	config := &logConfig{
		Source:       "source",
		StorageIndex: "configured",
		JSONAsFields: true,
	}

	var payload []byte
	forward := forwardFunc(config, func(data []byte) error {
		payload = append([]byte(nil), data...)
		return nil
	})
	if err := forward("/var/log/app.log", `{"value":1}`, map[string]interface{}{"line": int64(1)}); err != nil {
		t.Fatalf("forwardFunc() error = %v", err)
	}

	var got message
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !got.JSONAsFields {
		t.Fatal("message.JSONAsFields = false, want true")
	}
	if got.Log != `{"value":1}` {
		t.Fatalf("message.Log = %q, want raw JSON log", got.Log)
	}

	var legacy struct {
		Log string `json:"log"`
	}
	if err := json.Unmarshal(payload, &legacy); err != nil {
		t.Fatalf("legacy json.Unmarshal() error = %v", err)
	}
	if legacy.Log != got.Log {
		t.Fatalf("legacy log = %q, want %q", legacy.Log, got.Log)
	}
}
