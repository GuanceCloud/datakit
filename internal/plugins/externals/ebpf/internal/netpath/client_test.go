//go:build linux
// +build linux

package netpath

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerPostsBatch(t *testing.T) {
	gotReq := make(chan request, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get(tokenHeader) != "secret" {
			t.Fatalf("missing token header")
		}
		defer req.Body.Close() //nolint:errcheck
		var payload request
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		gotReq <- payload
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	scheduler := NewScheduler(Config{
		API:           server.URL,
		Token:         "secret",
		BatchSize:     2,
		FlushInterval: time.Hour,
		Timeout:       time.Second,
		Host:          "node-1",
		Tags: map[string]string{
			"env": "testing",
		},
	})
	defer scheduler.Close()

	if !scheduler.Schedule(Candidate{Hostname: "example.com", TargetIP: "203.0.113.10", Port: 443, Protocol: "tcp"}) {
		t.Fatal("first schedule failed")
	}
	if !scheduler.Schedule(Candidate{TargetIP: "203.0.113.11", Port: 443, Protocol: "tcp"}) {
		t.Fatal("second schedule failed")
	}

	select {
	case payload := <-gotReq:
		if payload.Source != "ebpf_netflow" {
			t.Fatalf("unexpected source %q", payload.Source)
		}
		if payload.Host != "node-1" {
			t.Fatalf("unexpected host %q", payload.Host)
		}
		if payload.Tags["env"] != "testing" {
			t.Fatalf("unexpected tags %#v", payload.Tags)
		}
		if len(payload.Tests) != 2 {
			t.Fatalf("expected 2 tests, got %d", len(payload.Tests))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for batch")
	}
}

func TestScheduleRejectsEmptyTarget(t *testing.T) {
	scheduler := &Scheduler{ch: make(chan Candidate, 1)}
	if scheduler.Schedule(Candidate{}) {
		t.Fatal("empty candidate should be rejected")
	}
}

func TestSchedulerRetriesFailedBatchWithoutLosingCandidates(t *testing.T) {
	requests := make(chan request, 2)
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close() //nolint:errcheck
		var payload request
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Errorf("decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		requests <- payload
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	scheduler := NewScheduler(Config{
		API:             server.URL,
		BatchSize:       1,
		FlushInterval:   time.Hour,
		Timeout:         time.Second,
		retryBackoff:    10 * time.Millisecond,
		maxPostAttempts: 3,
	})
	defer scheduler.Close()

	candidate := Candidate{Hostname: "example.com", Port: 443, Protocol: "tcp"}
	if !scheduler.Schedule(candidate) {
		t.Fatal("schedule failed")
	}

	for requestNumber := 1; requestNumber <= 2; requestNumber++ {
		select {
		case payload := <-requests:
			if len(payload.Tests) != 1 || payload.Tests[0].Hostname != candidate.Hostname {
				t.Fatalf("attempt %d changed batch: %#v", requestNumber, payload.Tests)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for request %d", requestNumber)
		}
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("attempts = %d, want 2", got)
	}
}

func TestSchedulerDoesNotRetryPermanentHTTPError(t *testing.T) {
	requests := make(chan struct{}, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close() //nolint:errcheck
		requests <- struct{}{}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	scheduler := NewScheduler(Config{
		API:             server.URL,
		BatchSize:       1,
		FlushInterval:   time.Hour,
		Timeout:         time.Second,
		retryBackoff:    10 * time.Millisecond,
		maxPostAttempts: 3,
	})
	defer scheduler.Close()
	if !scheduler.Schedule(Candidate{Hostname: "example.com", Port: 443, Protocol: "tcp"}) {
		t.Fatal("schedule failed")
	}

	select {
	case <-requests:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for request")
	}
	select {
	case <-requests:
		t.Fatal("permanent HTTP error was retried")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestSchedulerStopsAfterMaximumPostAttempts(t *testing.T) {
	requests := make(chan request, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close() //nolint:errcheck
		var payload request
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Errorf("decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		requests <- payload
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	scheduler := NewScheduler(Config{
		API:             server.URL,
		BatchSize:       1,
		FlushInterval:   time.Hour,
		Timeout:         time.Second,
		retryBackoff:    5 * time.Millisecond,
		maxPostAttempts: 2,
	})
	defer scheduler.Close()

	candidate := Candidate{Hostname: "example.com", Port: 443, Protocol: "tcp"}
	if !scheduler.Schedule(candidate) {
		t.Fatal("schedule failed")
	}
	for attempt := 1; attempt <= 2; attempt++ {
		select {
		case payload := <-requests:
			if len(payload.Tests) != 1 || payload.Tests[0].Hostname != candidate.Hostname {
				t.Fatalf("attempt %d changed batch: %#v", attempt, payload.Tests)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for attempt %d", attempt)
		}
	}
	select {
	case <-requests:
		t.Fatal("retry limit was exceeded")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestPostRetryDelayIsBounded(t *testing.T) {
	if got := postRetryDelay(time.Second, 3); got != 4*time.Second {
		t.Fatalf("delay = %s, want 4s", got)
	}
	if got := postRetryDelay(time.Second, 100); got != maxRetryBackoff {
		t.Fatalf("bounded delay = %s, want %s", got, maxRetryBackoff)
	}
	if got := postRetryDelay(time.Hour, 1); got != maxRetryBackoff {
		t.Fatalf("large base delay = %s, want %s", got, maxRetryBackoff)
	}
}
