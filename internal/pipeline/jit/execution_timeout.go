// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"context"
	"time"
)

type recordTimeoutKey struct{}

// WithRecordTimeout sets a cooperative execution budget for each record. It
// allocates no timer and does not change ctx.Done, ctx.Err or ctx.Deadline.
// A nonpositive budget expires immediately. Use WithTimeout instead when the
// whole request, including waiting and all records, needs a shared deadline.
func WithRecordTimeout(ctx context.Context, timeout time.Duration) context.Context {
	return context.WithValue(ctx, recordTimeoutKey{}, timeout)
}

func RecordTimeout(ctx context.Context) (time.Duration, bool) {
	timeout, ok := ctx.Value(recordTimeoutKey{}).(time.Duration)
	return timeout, ok
}

func HasExecutionControl(ctx context.Context) bool {
	if ctx.Done() != nil {
		return true
	}
	_, limited := RecordTimeout(ctx)
	return limited
}
