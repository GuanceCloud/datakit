// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package windowsremote

import (
	"testing"
	"time"
)

func Test_winServer_collectLog(t *testing.T) {
	server := newServer("", "", "")
	ticker := time.NewTicker(20 * time.Second)
	for range ticker.C {
		pts := server.collectLog()
		if len(pts) != 0 {
			for _, pt := range pts {
				t.Logf("log event point string= %s \n", pt.LineProto())
			}
			return
		}
	}
}

func Test_winServer_collectMetric(t *testing.T) {
	server := newServer("", "", "")
	ticker := time.NewTicker(20 * time.Second)
	for range ticker.C {
		pts := server.collectMetric(time.Now())
		if len(pts) != 0 {
			t.Log("collect metric: ")
			for _, pt := range pts {
				t.Logf(" %s \n", pt.LineProto())
			}
			t.Log("out \n")
			return
		}
	}
}

func TestCollectObject(t *testing.T) {
	server := newServer("", "", "")

	server.beginCollectObject()
	pts := server.toObjectPoints()
	for _, pt := range pts {
		t.Log(pt.LineProto())
	}
}

func TestTime(t *testing.T) {
	timenano := 1744371941425000000
	ts := time.Unix(0, int64(timenano))
	t.Log(ts)
}
