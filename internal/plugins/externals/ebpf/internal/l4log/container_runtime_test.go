//go:build linux
// +build linux

package l4log

import "testing"

type fakeContainerRuntime struct {
	closed bool
}

func (r *fakeContainerRuntime) ListContainers() ([]containerSnapshot, error) {
	return nil, nil
}

func (r *fakeContainerRuntime) Close() error {
	r.closed = true
	return nil
}

func TestCloseContainerRuntimes(t *testing.T) {
	rt := &fakeContainerRuntime{}

	closeContainerRuntimes([]containerRuntime{nil, rt})

	if !rt.closed {
		t.Fatal("expected closeContainerRuntimes to close owned runtime")
	}
}
