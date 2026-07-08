// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCheckEndpointPreservesNotExist(t *testing.T) {
	endpoint := "unix://" + filepath.Join(t.TempDir(), "containerd.sock")
	if err := checkEndpoint(endpoint); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("checkEndpoint error: got %v, want os.ErrNotExist", err)
	}
}

func TestRuntimeConnectionErrorClassification(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "wrapped unavailable", err: fmt.Errorf("create collector: %w", status.Error(codes.Unavailable, "refused")), want: true},
		{name: "deadline", err: context.DeadlineExceeded, want: true},
		{name: "configuration", err: fmt.Errorf("invalid log filter"), want: false},
		{name: "permission", err: status.Error(codes.PermissionDenied, "denied"), want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isRuntimeConnectionError(test.err); got != test.want {
				t.Fatalf("isRuntimeConnectionError(%v): got %t, want %t", test.err, got, test.want)
			}
		})
	}
}

type testCollector struct {
	started chan struct{}
}

func (c *testCollector) StartCollect() {
	close(c.started)
}

func TestRuntimeWaitingCollectorStartsCollector(t *testing.T) {
	started := make(chan struct{})
	stop := make(chan interface{})
	var attempts atomic.Int32

	c := &runtimeWaitingCollector{
		endpoints: []string{"unix:///run/containerd/containerd.sock"},
		interval:  time.Millisecond,
		stop:      stop,
		create: func(string) (Collector, error) {
			if attempts.Add(1) < 3 {
				return nil, status.Error(codes.Unavailable, "runtime unavailable")
			}
			return &testCollector{started: started}, nil
		},
	}

	go c.StartCollect()

	select {
	case <-started:
		if got := attempts.Load(); got != 3 {
			t.Fatalf("unexpected attempts: got %d, want 3", got)
		}
	case <-time.After(time.Second):
		t.Fatal("collector did not recover")
	}
}

func TestRuntimeWaitingCollectorStopsOnNonConnectionError(t *testing.T) {
	finished := make(chan struct{})
	stop := make(chan interface{})
	var attempts atomic.Int32

	c := &runtimeWaitingCollector{
		endpoints: []string{"unix:///run/containerd/containerd.sock"},
		interval:  time.Millisecond,
		stop:      stop,
		create: func(string) (Collector, error) {
			attempts.Add(1)
			return nil, fmt.Errorf("runtime configuration is invalid")
		},
	}

	go func() {
		c.StartCollect()
		close(finished)
	}()

	select {
	case <-finished:
		if got := attempts.Load(); got != 1 {
			t.Fatalf("unexpected attempts: got %d, want 1", got)
		}
	case <-time.After(time.Second):
		t.Fatal("collector did not stop after non-connection error")
	}
}
