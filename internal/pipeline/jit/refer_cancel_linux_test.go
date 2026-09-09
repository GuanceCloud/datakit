// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/pipeline-go/ptinput/refertable"
)

func TestReferPullCancellationInterruptsHTTP(t *testing.T) {
	entered, aborted, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var enterOnce, abortOnce sync.Once
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enterOnce.Do(func() { close(entered) })
		select {
		case <-r.Context().Done():
			abortOnce.Do(func() { close(aborted) })
		case <-release:
		}
	}))
	defer server.Close()
	defer close(release)
	table, err := refertable.NewReferTable(refertable.RefTbCfg{URL: server.URL, Interval: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { defer close(done); table.PullWorker(ctx) }()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("pull did not reach HTTP server")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("canceled worker remained blocked in HTTP")
	}
	select {
	case <-aborted:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not observe request cancellation")
	}
}
