// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import "testing"

func TestV(t *testing.T) {
	anys := []interface{}{
		"1",
		"strings",
		float64(1.3435),
		int64(200),
		true,
		false,
	}
	for _, v := range anys {
		t.Logf("log v:%v", v)
	}
}
