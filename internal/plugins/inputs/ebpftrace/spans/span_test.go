// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !(windows && 386)
// +build !windows !386

package spans

import (
	"testing"
	"time"
)

func TestMRTrace(t *testing.T) {
	pts := MockTraceData(1, 2, 2, time.Now(), true, true)

	t.Log(len(pts))

	tn := time.Now()

	meta, ok := spanMeta(pts)
	if !ok {
		t.Error("!ok")
	}
	mrr := MRRunner{}
	metadata := MRMetaData{
		firstHalf: meta,
	}
	mrr.connectSpans(&metadata)
	info, _ := mrr.linkAndgatherTrace(&metadata)

	for _, v := range info {
		t.Log(*v)
	}
	t.Log(time.Since(tn))

	// for _, v := range meta {
	// 	t.Error(v)
	// }
}
