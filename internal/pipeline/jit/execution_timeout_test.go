// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"context"
	"testing"
	"time"
)

func TestRecordTimeoutDoesNotCreateContextCancellation(t *testing.T) {
	ctx := WithRecordTimeout(context.Background(), time.Millisecond)
	if ctx.Done() != nil || ctx.Err() != nil {
		t.Fatal("clock budget created cancellation")
	}
	if _, ok := ctx.Deadline(); ok {
		t.Fatal("record budget became request deadline")
	}
	if limit, ok := RecordTimeout(ctx); !ok || limit != time.Millisecond {
		t.Fatal("budget missing")
	}
	parent, cancel := context.WithCancel(context.Background())
	child := WithRecordTimeout(parent, time.Second)
	cancel()
	if child.Err() != context.Canceled {
		t.Fatal("lost parent cancellation")
	}
}
