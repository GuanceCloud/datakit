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
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Copied from pipeline-go v1.4.3 ptinput/funcs/fn_default_time_test.go.
// Duplicate inputs with the same timezone are collapsed.
var pipelineGoDefaultTimeVectors = []timestampTestVector{
	{value: "02 Dec 11:55:34.000", timezone: ""},
	{value: "02 Dec 2021 11:55:34.000", timezone: ""},
	{value: "02 Dec 2021 12:55:34.000", timezone: "Asia/Tokyo"},
	{value: "02 Dec 2021, 11:55", timezone: ""},
	{value: "02 December 2021", timezone: ""},
	{value: "02/Dec/2021:11:55:34 +0800", timezone: ""},
	{value: "12.02.2021", timezone: ""},
	{value: "12.02.21", timezone: ""},
	{value: "12.2.2021", timezone: ""},
	{value: "12/02/2021", timezone: ""},
	{value: "12/02/2021 11:55", timezone: ""},
	{value: "12/02/2021 11:55 AM", timezone: ""},
	{value: "12/02/2021 11:55:43", timezone: ""},
	{value: "12/02/2021 11:55:43 AM", timezone: ""},
	{value: "12/02/2021 11:55:43.9999999", timezone: ""},
	{value: "12/2/2021", timezone: ""},
	{value: "12/2/2021 11:55", timezone: ""},
	{value: "12/2/21", timezone: ""},
	{value: "1638417343", timezone: ""},
	{value: "1638417343001", timezone: ""},
	{value: "1638417343001002", timezone: ""},
	{value: "1638417343001002003", timezone: ""},
	{value: "2 Dec 2021", timezone: ""},
	{value: "2 Dec 2021, 11:55", timezone: ""},
	{value: "2 Dec 21", timezone: ""},
	{value: "2 December 2021", timezone: ""},
	{value: "2021", timezone: ""},
	{value: "2021-12", timezone: ""},
	{value: "2021-12-02", timezone: ""},
	{value: "2021-12-02 11:55:34.000 UTC", timezone: ""},
	{value: "2021-12-2 11:55", timezone: ""},
	{value: "2021-12-2 11:55:43", timezone: ""},
	{value: "2021-12-2 11:55:43 +0800", timezone: ""},
	{value: "2021-12-2 11:55:43 +0800 +08", timezone: ""},
	{value: "2021-12-2 11:55:43 +0800 GMT", timezone: ""},
	{value: "2021-12-2 11:55:43 +0800 GMT m=+0.000000001", timezone: ""},
	{value: "2021-12-2 11:55:43 +0800 UTC", timezone: ""},
	{value: "2021-12-2 11:55:43 +08:00", timezone: ""},
	{value: "2021-12-2 11:55:43 AM", timezone: ""},
	{value: "2021-12-2 11:55:43 CST", timezone: ""},
	{value: "2021-12-2 11:55:43 GMT", timezone: ""},
	{value: "2021-12-2 11:55:43 UTC", timezone: ""},
	{value: "2021-12-2 11:55:43+08:00", timezone: ""},
	{value: "2021-12-2 11:55:43,212", timezone: ""},
	{value: "2021-12-2 11:55:43.0001 +0800 GMT m=+0.000000001", timezone: ""},
	{value: "2021-12-2 11:55:43.1 +0800 UTC", timezone: ""},
	{value: "2021-12-2 11:55:43.12223", timezone: ""},
	{value: "2021-12-2 11:55:43.122230000", timezone: ""},
	{value: "2021-12-2 3:55:43.1 +0000 UTC", timezone: ""},
	{value: "2021-12-2T11:55:43", timezone: ""},
	{value: "2021-12-2T11:55:43+08:00", timezone: ""},
	{value: "2021-12-2T3:55:43Z", timezone: ""},
	{value: "2021-Dec-02", timezone: ""},
	{value: "2021.12", timezone: ""},
	{value: "2021.12.02", timezone: ""},
	{value: "2021/12/02 - 11:55:34", timezone: ""},
	{value: "20211202", timezone: ""},
	{value: "20211202115543", timezone: ""},
	{value: "2021:12:02", timezone: ""},
	{value: "2021:12:02 11:55", timezone: ""},
	{value: "2021:12:02 11:55:43", timezone: ""},
	{value: "2021:12:2", timezone: ""},
	{value: "2021:12:2 11:55", timezone: ""},
	{value: "2021:12:2 11:55:43", timezone: ""},
	{value: "2021:12:2 11:55:43.89555", timezone: ""},
	{value: "2021:12:2 12:55:43", timezone: "Asia/Tokyo"},
	{value: "2021年12月02日", timezone: ""},
	{value: "211202 11:55:34", timezone: ""},
	{value: "Dec 2, '21", timezone: "UTC"},
	{value: "Dec 2, 2021", timezone: ""},
	{value: "Dec 2, 2021", timezone: "+0"},
	{value: "Dec 2, 2021", timezone: "UTC"},
	{value: "Dec 2, 2021 11:55:34 AM", timezone: ""},
	{value: "Dec 2, 21", timezone: "UTC"},
	{value: "December 02, 2021 11:55:34am", timezone: ""},
	{value: "December 02, 2021 at 11:55:34am CST+08", timezone: ""},
	{value: "December 02, 2021 at 11:55am CST+08", timezone: ""},
	{value: "December 02, 2021, 11:55:34", timezone: ""},
	{value: "December 2, 2021", timezone: ""},
	{value: "December 2th, 2021", timezone: ""},
	{value: "Tue 02 Dec 2021 11:55:34 AM CST", timezone: ""},
	{value: "Tue Dec  2 11:55:34 2021", timezone: ""},
	{value: "Tue Dec  2 11:55:34 CST 2021", timezone: ""},
	{value: "Tue Dec  2 11:55:34 CST 2021", timezone: "+8"},
	{value: "Tue Dec 02 11:55:34 +0800 2021", timezone: ""},
	{value: "Tue Dec 02 2021 11:55:34 GMT+0800 (GMT Daylight Time)", timezone: ""},
	{value: "Tue Dec 2 11:55:34.000000 2021", timezone: ""},
	{value: "Tue, 02 Dec 2021 11:55:34 +0800", timezone: ""},
	{value: "Tue, 02 Dec 2021 11:55:34 +0800 (CST)", timezone: ""},
	{value: "Tue, 02 Dec 2021 11:55:34 CST", timezone: ""},
	{value: "Tue, 2 Dec 2021 11:55:34 +0800", timezone: ""},
	{value: "Tuesday, 02-Dec-21 11:55:34 CST", timezone: ""},
}

func TestPipelineGoDefaultTimeUpstreamCorpus(t *testing.T) {
	const childKey = "DATAKIT_JIT_DEFAULT_TIME_UPSTREAM_CHILD"
	if os.Getenv(childKey) != "1" {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^"+t.Name()+"$", "-test.count=1")
		// The native runtime observes the process timezone through libc, while
		// pipeline-go observes time.Local. Configure both sides before starting
		// the child so this oracle is deterministic on UTC containers as well as
		// CST developer hosts.
		cmd.Env = append(environmentWithout(os.Environ(), "TZ"), childKey+"=1", "TZ=CST-8")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("default_time upstream subprocess: %v\n%s", err, output)
		}
		return
	}
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	// Match the upstream test without leaking its process-global Local.
	time.Local = time.FixedZone("CST", 8*60*60)
	host, recorder := newObservedPipelineGoHost()
	runner, err := NewRunnerWithHost(runtimePath, "pipeline-go-1.4.3-datakit", 16, host)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	groups := map[string][]timestampTestVector{}
	for _, vector := range pipelineGoDefaultTimeVectors {
		groups[vector.timezone] = append(groups[vector.timezone], vector)
	}
	for timezone, vectors := range groups {
		t.Run("timezone-"+strconv.Quote(timezone), func(t *testing.T) {
			source := "default_time(time)\n"
			if timezone != "" {
				source = "default_time(time, " + strconv.Quote(timezone) + ")\n"
			}
			compareTimestampBatchWithPipelineGo(t, runner, recorder, source, vectors)
		})
	}
}

func environmentWithout(environment []string, key string) []string {
	prefix := key + "="
	filtered := make([]string, 0, len(environment))
	for _, entry := range environment {
		if !strings.HasPrefix(entry, prefix) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}
