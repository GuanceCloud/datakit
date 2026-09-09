// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestBoundedJITCheckDetail(t *testing.T) {
	if got := boundedJITCheckDetail("compile failed\nline 2"); got != "compile failed\nline 2" {
		t.Fatalf("short check detail changed: %q", got)
	}

	got := boundedJITCheckDetail(strings.Repeat("界", maximumJITCheckDetailBytes))
	if !utf8.ValidString(got) {
		t.Fatalf("bounded check detail split UTF-8: %q", got)
	}
	if len(got) > maximumJITCheckDetailBytes+len("...") {
		t.Fatalf("bounded check detail is %d bytes, maximum is %d", len(got), maximumJITCheckDetailBytes+len("..."))
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("truncated check detail has no marker: %q", got)
	}

	got = boundedJITCheckDetail("invalid:\xff")
	if !utf8.ValidString(got) {
		t.Fatalf("invalid UTF-8 check detail was not sanitized: %q", got)
	}
}
