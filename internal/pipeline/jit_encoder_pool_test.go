// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"testing"

	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestProjectedEncoderReleaseDropsPointReferences(t *testing.T) {
	entries := []pljit.FlatPointEntry{{Key: "message", Value: "large input string"}}
	records := []pljit.FlatPointRecord{{Measurement: "previous source", Entries: entries}}
	encoder := &projectedJITEncoder{records: records, entries: entries, buffer: make([]byte, 16, 128)}
	releaseProjectedJITEncoder(encoder)
	// Check the backing storage directly: merely resetting slice lengths would
	// keep input strings and dynamic field values alive in the pool.
	if entries[0].Key != "" || entries[0].Value != nil || records[0].Measurement != "" || records[0].Entries != nil {
		t.Fatal("pooled scratch retained references to the previous Point")
	}
}
