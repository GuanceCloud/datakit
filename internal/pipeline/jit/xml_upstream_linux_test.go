// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"os"
	"testing"
)

func TestXMLPipelineGoUpstreamAndBoundaryCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 24)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	entry := `<entry>
        <fieldx>valuex</fieldx>
        <fieldy>...</fieldy>
        <fieldz>...</fieldz>
        <fieldarray>
            <fielda>element_a_1</fielda>
            <fielda>element_a_2</fielda>
        </fieldarray>
    </entry>`
	order := ` <OrderEvent actionCode = "5">
 <OrderNumber>ORD12345</OrderNumber>
 <VendorNumber>V11111</VendorNumber>
 </OrderEvent>`
	cases := []struct {
		name     string
		source   string
		fields   map[string]any
		target   string
		expected any
	}{
		{"upstream-text", `xml(_,"/entry/fieldx/text()",new_name); add_key(after,true)`, map[string]any{"message": entry}, "new_name", "valuex"},
		{"upstream-array-first", `xml(_,"/entry/fieldarray//fielda[1]/text()",field_a_1); add_key(after,true)`, map[string]any{"message": entry}, "field_a_1", "element_a_1"},
		{"upstream-attribute", `xml(_,"/OrderEvent/@actionCode",action_code); add_key(after,true)`, map[string]any{"message": order}, "action_code", "5"},
		{"upstream-order-number", `xml(_,"/OrderEvent/OrderNumber/text()",OrderNumber); add_key(after,true)`, map[string]any{"message": order}, "OrderNumber", "ORD12345"},
		{"upstream-vendor-number", `xml(_,"/OrderEvent/VendorNumber/text()",VendorNumber); add_key(after,true)`, map[string]any{"message": order}, "VendorNumber", "V11111"},
		{"upstream-invalid-xml", `xml(_,"/OrderEvent/VendorNumber/text()",VendorNumber); add_key(after,true)`, map[string]any{"message": "Not a valid XML"}, "VendorNumber", nil},
		{"upstream-invalid-xpath", `xml(_,"invalid xpath expr",VendorNumber); add_key(after,true)`, map[string]any{"message": order}, "VendorNumber", nil},
		{"no-matching-node", `xml(_,"/entry/missing/text()",result); add_key(after,true)`, map[string]any{"message": entry}, "result", nil},
		{"missing-source", `xml(payload,"/entry/fieldx/text()",result); add_key(after,true)`, map[string]any{}, "result", nil},
		{"non-string-source", `xml(payload,"/entry/fieldx/text()",result); add_key(after,true)`, map[string]any{"payload": int64(7)}, "result", nil},
		{"string-target", `xml(_,"/entry/fieldx/text()","renamed"); add_key(after,true)`, map[string]any{"message": entry}, "renamed", "valuex"},
		{"attribute-target", `xml(_,"/entry/fieldx/text()",a.result); add_key(after,true)`, map[string]any{"message": entry}, "a.result", "valuex"},
	}

	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		t.Run("batch-"+itoa(size), func(t *testing.T) {
			for _, tc := range cases {
				check := runner.Check(tc.source)
				if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" ||
					check.Capabilities.RequiredHostFlags != 0 || len(check.Capabilities.HostCalls) != 0 {
					t.Fatalf("%s route=%+v", tc.name, check)
				}
				points := make([]Point, size)
				for i := range points {
					fields := cloneFields(tc.fields)
					fields["sequence"] = int64(i)
					points[i] = Point{Version: 1, Category: "logging", Measurement: tc.name, Fields: fields}
				}
				input, err := EncodeFlatPoints(points)
				if err != nil {
					t.Fatal(err)
				}
				batch, err := runner.Process(tc.source, input)
				if err != nil || len(batch.Records) != size {
					t.Fatalf("%s records=%d error=%v", tc.name, len(batch.Records), err)
				}
				for i, record := range batch.Records {
					if record.Status != TerminalOK {
						t.Fatalf("%s record %d terminal=%d error=%s", tc.name, i, record.Status, record.Error)
					}
					applyHostCompatRecord(t, batch, i, &points[i])
					if tc.expected == nil {
						if _, ok := points[i].Fields[tc.target]; ok {
							t.Fatalf("%s record %d unexpectedly set %q: %#v", tc.name, i, tc.target, points[i])
						}
					} else if got := points[i].Fields[tc.target]; got != tc.expected {
						t.Fatalf("%s record %d %s=%#v want %#v", tc.name, i, tc.target, got, tc.expected)
					}
					if points[i].Fields["after"] != true || points[i].Fields["sequence"] != int64(i) {
						t.Fatalf("%s record %d lost continuation/input: %#v", tc.name, i, points[i])
					}
				}
			}
		})
	}
}
