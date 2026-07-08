// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	componenthealth "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/health"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type fakeCRIService struct {
	listErr error
	listFn  func() error
	closed  bool
}

func TestShouldReconnectCRIWrappedError(t *testing.T) {
	err := fmt.Errorf("query status: %w", status.Error(codes.Unavailable, "connection refused"))
	if !shouldReconnectCRI(err) {
		t.Fatalf("wrapped transport error was not reconnectable: %v", err)
	}
}

func (*fakeCRIService) Version(string) (*runtimeapi.VersionResponse, error) {
	return &runtimeapi.VersionResponse{
		RuntimeName:       "containerd",
		RuntimeVersion:    "1.7.0",
		RuntimeApiVersion: "v1",
	}, nil
}

func (f *fakeCRIService) ListContainers(*runtimeapi.ContainerFilter) ([]*runtimeapi.Container, error) {
	if f.listFn != nil {
		return nil, f.listFn()
	}
	return nil, f.listErr
}

func TestCRIRuntimeReconnectsOnceForConcurrentFailures(t *testing.T) {
	t.Parallel()

	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	failedService := &fakeCRIService{listFn: func() error {
		entered <- struct{}{}
		<-release
		return status.Error(codes.Unavailable, "connection refused")
	}}
	recoveredService := &fakeCRIService{}
	services := []criRuntimeService{failedService, recoveredService}
	factoryCalls := 0

	rt, err := newCRIRuntime("unix:///run/containerd/containerd.sock", "/proc", func(string, time.Duration) (criRuntimeService, error) {
		service := services[factoryCalls]
		factoryCalls++
		return service, nil
	})
	if err != nil {
		t.Fatalf("newCRIRuntime: %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, callErr := rt.ListContainers()
			errs <- callErr
		}()
	}
	<-entered
	<-entered
	close(release)
	wg.Wait()
	close(errs)

	for callErr := range errs {
		if callErr != nil {
			t.Fatalf("ListContainers: %v", callErr)
		}
	}
	if factoryCalls != 2 {
		t.Fatalf("factory calls: got %d, want 2", factoryCalls)
	}
}

func (*fakeCRIService) ContainerStatus(string, bool) (*runtimeapi.ContainerStatusResponse, error) {
	return nil, nil
}

func (*fakeCRIService) ContainerStats(string) (*runtimeapi.ContainerStats, error) {
	return nil, nil
}

func (f *fakeCRIService) Close() error {
	f.closed = true
	return nil
}

func TestCRIRuntimeReconnectsAfterTransportFailure(t *testing.T) {
	t.Parallel()

	failedService := &fakeCRIService{listErr: status.Error(codes.Unavailable, "connection refused")}
	recoveredService := &fakeCRIService{}
	services := []criRuntimeService{failedService, recoveredService}
	factoryCalls := 0

	rt, err := newCRIRuntime("unix:///run/containerd/containerd.sock", "/proc", func(string, time.Duration) (criRuntimeService, error) {
		service := services[factoryCalls]
		factoryCalls++
		return service, nil
	})
	if err != nil {
		t.Fatalf("newCRIRuntime: %v", err)
	}

	containers, err := rt.ListContainers()
	if err != nil {
		t.Fatalf("ListContainers: %v", err)
	}
	if len(containers) != 0 {
		t.Fatalf("unexpected containers: %v", containers)
	}
	if factoryCalls != 2 {
		t.Fatalf("factory calls: got %d, want 2", factoryCalls)
	}
	if !failedService.closed {
		t.Fatal("failed CRI service was not closed after reconnect")
	}
	registry := componenthealth.NewRegistry()
	reporter := registry.Register("container-runtime", "test", componenthealth.Options{})
	rt.healthReporter = reporter
	if _, err := rt.ListContainers(); err != nil {
		t.Fatalf("healthy ListContainers: %v", err)
	}
	if !registry.Snapshot().Live {
		t.Fatal("healthy runtime reported not live")
	}
}

func TestCRIRuntimeDoesNotReconnectAfterApplicationFailure(t *testing.T) {
	t.Parallel()

	service := &fakeCRIService{listErr: status.Error(codes.PermissionDenied, "denied")}
	factoryCalls := 0
	rt, err := newCRIRuntime("unix:///run/containerd/containerd.sock", "/proc", func(string, time.Duration) (criRuntimeService, error) {
		factoryCalls++
		return service, nil
	})
	if err != nil {
		t.Fatalf("newCRIRuntime: %v", err)
	}

	if _, err := rt.ListContainers(); err == nil {
		t.Fatal("ListContainers returned nil error")
	}
	if factoryCalls != 1 {
		t.Fatalf("factory calls: got %d, want 1", factoryCalls)
	}
}

func TestCRIRuntimeKeepsHealthFailedWhenReconnectedServiceStillFails(t *testing.T) {
	t.Parallel()

	failedService := &fakeCRIService{listErr: status.Error(codes.Unavailable, "connection refused")}
	recoveredService := &fakeCRIService{listErr: status.Error(codes.Unavailable, "runtime still unavailable")}
	services := []criRuntimeService{failedService, recoveredService}
	factoryCalls := 0

	rt, err := newCRIRuntime("unix:///run/containerd/containerd.sock", "/proc", func(string, time.Duration) (criRuntimeService, error) {
		service := services[factoryCalls]
		factoryCalls++
		return service, nil
	})
	if err != nil {
		t.Fatalf("newCRIRuntime: %v", err)
	}

	registry := componenthealth.NewRegistry()
	reporter := registry.Register("container-runtime", "test", componenthealth.Options{
		FailureTimeout: time.Nanosecond,
	})
	rt.healthReporter = reporter

	if _, err := rt.ListContainers(); err == nil {
		t.Fatal("ListContainers returned nil error")
	}
	time.Sleep(time.Millisecond)
	if registry.Snapshot().Live {
		t.Fatal("runtime reported live after reconnecting to a still-failing service")
	}
}

func TestCRIRuntimeCloseReleasesServiceAndHealth(t *testing.T) {
	t.Parallel()

	service := &fakeCRIService{}
	rt, err := newCRIRuntime("unix:///run/containerd/containerd.sock", "/proc", func(string, time.Duration) (criRuntimeService, error) {
		return service, nil
	})
	if err != nil {
		t.Fatalf("newCRIRuntime: %v", err)
	}
	registry := componenthealth.NewRegistry()
	reporter := registry.Register("container-runtime", "test", componenthealth.Options{FailureTimeout: time.Nanosecond})
	rt.healthReporter = reporter
	reporter.Failure()
	time.Sleep(time.Millisecond)
	if registry.Snapshot().Live {
		t.Fatal("failed runtime did not affect liveness")
	}

	if err := rt.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if !service.closed {
		t.Fatal("CRI service was not closed")
	}
	if !registry.Snapshot().Live {
		t.Fatal("closed runtime still affects liveness")
	}
}

func TestCRIRuntimeLimitsReconnectAttempts(t *testing.T) {
	t.Parallel()

	service := &fakeCRIService{listErr: status.Error(codes.Unavailable, "connection refused")}
	factoryCalls := 0
	rt, err := newCRIRuntime("unix:///run/containerd/containerd.sock", "/proc", func(string, time.Duration) (criRuntimeService, error) {
		factoryCalls++
		if factoryCalls > 1 {
			return nil, errors.New("runtime is still unavailable")
		}
		return service, nil
	})
	if err != nil {
		t.Fatalf("newCRIRuntime: %v", err)
	}

	if _, err := rt.ListContainers(); err == nil {
		t.Fatal("first ListContainers returned nil error")
	}
	if _, err := rt.ListContainers(); err == nil {
		t.Fatal("second ListContainers returned nil error")
	}
	if factoryCalls != 2 {
		t.Fatalf("factory calls: got %d, want 2", factoryCalls)
	}
	registry := componenthealth.NewRegistry()
	reporter := registry.Register("container-runtime", "test", componenthealth.Options{
		FailureTimeout: time.Nanosecond,
	})
	rt.healthReporter = reporter
	if _, err := rt.ListContainers(); err == nil {
		t.Fatal("third ListContainers returned nil error")
	}
	time.Sleep(time.Millisecond)
	if registry.Snapshot().Live {
		t.Fatal("failed runtime recovery did not fail liveness")
	}
}
