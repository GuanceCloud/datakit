// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package telemetry

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/model"
)

func TestEventSerialization(t *testing.T) {
	// Define the expected Event structure after unmarshaling JSON.
	expected := Event{
		Time: time.Date(2022, 10, 12, 0, 1, 15, 0, time.UTC),
		Record: &PlatformInitRuntimeDone{
			Phase:              model.InitPhaseInit,
			InitializationType: model.InitTypeOnDemand,
			Status:             model.StatusSuccess,
			Spans: []model.Span{
				{
					Name:       "someTimeSpan",
					Start:      time.Date(2022, 6, 2, 12, 2, 33, 913000000, time.UTC),
					DurationMs: 70.5,
				},
			},
		},
	}

	// Define the JSON bytes to be used for testing.
	jsonData := json.RawMessage(`{
		"time": "2022-10-12T00:01:15.000Z",
		"type": "platform.initRuntimeDone",
		"record": {
			"phase": "init",
			"initializationType": "on-demand",
			"status": "success",
			"spans": [
				{
					"name": "someTimeSpan",
					"start": "2022-06-02T12:02:33.913Z",
					"durationMs": 70.5
				}
			]
		}
	}`)

	// Create a new Event instance to unmarshal JSON into.
	var event Event

	// Unmarshal JSON into the event struct.
	err := json.Unmarshal(jsonData, &event)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Check if the unmarshaled event matches the expected structure.
	if !reflect.DeepEqual(event, expected) {
		t.Errorf("unmarshaled event does not match expected:\nExpected: %+v\nGot: %+v", expected, event)
	}

	// Marshal the event back to JSON.
	_, err = json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event to JSON: %v", err)
	}
}

// TestSliceDeserialization tests the deserialization of a slice of Event records.
func TestSliceDeserialization(t *testing.T) {
	// Sample JSON input simulating the event data you provided.
	inputJSON := `[
		{
			"time": "2024-07-01T08:54:53.872Z",
			"type": "platform.initStart",
			"record": {
				"initializationType": "on-demand",
				"phase": "init",
				"runtimeVersion": "nodejs:18.v1",
				"runtimeVersionArn": "arn:aws-cn:lambda:cn-north-1::runtime:bc3d43e62df6d336b08ddf8f0e487c938063ba810aa9031bf6bd42d8f0245e0a"
			}
		},
		{
			"time": "2024-07-01T08:54:53.932Z",
			"type": "platform.telemetrySubscription",
			"record": {
				"name": "go-example-extension",
				"state": "Subscribed",
				"types": ["platform"]
			}
		},
		{
			"time": "2024-07-01T08:54:54.106Z",
			"type": "platform.initRuntimeDone",
			"record": {
				"initializationType": "on-demand",
				"phase": "init",
				"status": "success"
			}
		},
		{
			"time": "2024-07-01T08:54:54.106Z",
			"type": "platform.extension",
			"record": {
				"name": "go-example-extension",
				"state": "Ready",
				"events": ["INVOKE", "SHUTDOWN"]
			}
		},
		{
			"time": "2024-07-01T08:54:54.107Z",
			"type": "platform.initReport",
			"record": {
				"initializationType": "on-demand",
				"status": "success",
				"phase": "init",
				"metrics": {"durationMs": 234.766}
			}
		},
		{
			"time": "2024-07-01T08:54:54.108Z",
			"type": "platform.start",
			"record": {
				"requestId": "b2de8fdb-259d-476f-9546-0b4822dce67a",
				"version": "$LATEST"
			}
		},
		{
			"time": "2024-07-01T08:54:54.125Z",
			"type": "platform.runtimeDone",
			"record": {
				"requestId": "b2de8fdb-259d-476f-9546-0b4822dce67a",
				"metrics": {"durationMs": 17.12, "producedBytes": 50},
				"spans": [
					{"durationMs": 1.613, "name": "responseLatency", "start": "2024-07-01T08:54:54.108Z"},
					{"durationMs": 0.064, "name": "responseDuration", "start": "2024-07-01T08:54:54.109Z"},
					{"durationMs": 15.246, "name": "runtimeOverhead", "start": "2024-07-01T08:54:54.110Z"}
				],
				"status": "success"
			}
		},
		{
			"time": "2024-07-01T08:54:54.127Z",
			"type": "platform.report",
			"record": {
				"requestId": "b2de8fdb-259d-476f-9546-0b4822dce67a",
				"metrics": {
					"durationMs": 17.493,
					"billedDurationMs": 18,
					"initDurationMs": 235.086,
					"maxMemoryUsedMB": 75,
					"memorySizeMB": 128
				},
				"status": "success"
			}
		},
		{
			"time": "2022-10-12T00:03:50.000Z",
			"type": "function",
			"record": "[INFO] Hello world, I am a function!"
		},
		{
			"time": "2022-10-12T00:03:50.000Z",
			"type": "function",
			"record": {
				"timestamp": "2022-10-12T00:03:50.000Z",
				"level": "INFO",
				"requestId": "79b4f56e-95b1-4643-9700-2807f4e68189",
				"message": "Hello world, I am a function!"
			}
		},
		{
			"time": "2022-10-12T00:00:15.064Z",
			"type": "platform.start",
			"record": {
				"requestId": "6d68ca91-49c9-448d-89b8-7ca3e6dc66aa",
				"version": "$LATEST",
				"tracing": {
					"spanId": "54565fb41ac79632",
					"type": "X-Amzn-Trace-Id",
					"value": "Root=1-62e900b2-710d76f009d6e7785905449a;Parent=0efbd19962d95b05;Sampled=1"
				}
			}
		}
	]`

	var events []*Event
	err := json.Unmarshal([]byte(inputJSON), &events)
	assert.NoError(t, err)
	assert.Len(t, events, 11)

	// Check specific fields to verify the correct deserialization.
	if initStart, ok := events[0].Record.(*PlatformInitStart); ok {
		assert.Equal(t, "on-demand", string(initStart.InitializationType))
		assert.Equal(t, "init", string(initStart.Phase))
		assert.Equal(t, "nodejs:18.v1", initStart.RuntimeVersion)
	} else {
		t.Errorf("Record is not of type PlatformInitStart")
	}

	if telemetrySub, ok := events[1].Record.(*PlatformTelemetrySubscription); ok {
		assert.Equal(t, "go-example-extension", telemetrySub.Name)
		assert.Equal(t, "Subscribed", telemetrySub.State)
		assert.Contains(t, telemetrySub.Types, model.SubscriptionEventPlatform)
	} else {
		t.Errorf("Record is not of type PlatformTelemetrySubscription")
	}

	if report, ok := events[7].Record.(*PlatformReport); ok {
		assert.Equal(t, "b2de8fdb-259d-476f-9546-0b4822dce67a", report.RequestID)
		assert.Equal(t, model.StatusSuccess, report.Status)
		assert.Equal(t, 17.493, report.Metrics.DurationMs)
		assert.Equal(t, uint64(18), report.Metrics.BilledDurationMs)
	} else {
		t.Errorf("Record is not of type PlatformReport")
	}

	if report, ok := events[8].Record.(*FunctionLog); ok {
		assert.Equal(t, "[INFO] Hello world, I am a function!", report.Fields["message"])
	} else {
		t.Errorf("Record is not of type PlatformReport")
	}

	if report, ok := events[9].Record.(*FunctionLog); ok {
		assert.Equal(t, map[string]any{
			"timestamp": "2022-10-12T00:03:50.000Z",
			"level":     "INFO",
			"requestId": "79b4f56e-95b1-4643-9700-2807f4e68189",
			"message":   "Hello world, I am a function!",
		}, report.Fields)
	} else {
		t.Errorf("Record is not of type PlatformReport")
	}

	if report, ok := events[10].Record.(*PlatformStart); ok {
		assert.Equal(t, "6d68ca91-49c9-448d-89b8-7ca3e6dc66aa", report.RequestID)
		assert.Equal(t, "$LATEST", report.Version)
		assert.Equal(t, &model.TracingContext{
			SpanID: "54565fb41ac79632",
			Tracing: model.Tracing{
				Type: "X-Amzn-Trace-Id",
				Value: model.TraceInfo{
					Root:    "1-62e900b2-710d76f009d6e7785905449a",
					Parent:  "0efbd19962d95b05",
					Sampled: "1",
				},
			},
		}, report.Tracing)
	} else {
		t.Errorf("Record is not of type PlatformReport")
	}
}

func TestSeparateEvents(t *testing.T) {
	event1 := &Event{
		Time: time.Now(),
		Record: &FunctionLog{
			LambdaLog{Fields: map[string]any{"1": "2"}},
		},
	}
	event2 := &Event{
		Time: time.Now(),
		Record: &PlatformReport{
			RequestID: "114514",
		},
	}
	event3 := &Event{
		Time: time.Now(),
		Record: &ExtensionLog{
			LambdaLog{Fields: map[string]any{"1": "2"}},
		},
	}

	events := []*Event{event1, event2, event3}
	expectedDelData := []*LogEvent{
		{
			Time: event1.Time,
			Record: &FunctionLog{
				LambdaLog{Fields: map[string]any{"1": "2"}},
			},
		},
		{
			Time: event3.Time,
			Record: &ExtensionLog{
				LambdaLog{Fields: map[string]any{"1": "2"}},
			},
		},
	}

	events, delData := SeparateEvents(events)

	if len(delData) != len(expectedDelData) {
		t.Fatalf("expected %d deleted log events, but got %d", len(expectedDelData), len(delData))
	}

	for i, logEvent := range delData {
		assert.Equal(t, logEvent, expectedDelData[i], "expected log event %v at index %d, but got %v", expectedDelData[i], i, logEvent)
	}

	expectedRemainingEvents := []*Event{event2}
	if len(events) != len(expectedRemainingEvents) {
		t.Fatalf("expected %d remaining events, but got %d", len(expectedRemainingEvents), len(events))
	}

	for i, event := range events {
		if event != expectedRemainingEvents[i] && event.Record.GetType() == expectedRemainingEvents[i].Record.GetType() {
			t.Errorf("expected event %v at index %d, but got %v", expectedRemainingEvents[i], i, event)
		}
	}
}
