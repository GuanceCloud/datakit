// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import "testing"

// Ordinary json() uses Go's GJSONGet path. Deleting extraction has a separate
// DOM implementation; its behavior must not be assumed for ordinary projection.
func TestJITJSONProjectionSemanticsOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"unselected_overflow", `json(message,value,result)`, `{"value":42,"unused":1e400}`, "result", float64(42), false},
		{"unselected_surrogate", `json(message,value,result)`, `{"value":42,"unused":"\ud800"}`, "result", float64(42), false},
		{"selected_surrogate", `json(message,value,result)`, `{"value":"\ud800"}`, "result", "�", false},
		{"duplicate_scalar", `json(message,value,result)`, `{"value":1,"value":2}`, "result", float64(2), false},
		{"duplicate_parent", `json(message,value.child,result)`, `{"value":{"other":1},"value":{"child":2}}`, "result", float64(2), false},
		{"invalid_tail", `json(message,value,result)`, `{"value":42,"unused":}`, "result", nil, false},
	})
}
