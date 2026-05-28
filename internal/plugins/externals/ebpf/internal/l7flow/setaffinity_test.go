//go:build linux
// +build linux

package l7flow

import "testing"

func TestKpFlushTriggerCPUsDefaultIsBounded(t *testing.T) {
	t.Setenv(kpFlushTriggerCPUsEnv, "")

	if got := kpFlushTriggerCPUs(2); got != 2 {
		t.Fatalf("kpFlushTriggerCPUs(2) = %d, want 2", got)
	}
	if got := kpFlushTriggerCPUs(128); got != defaultKpFlushTriggerCPUs {
		t.Fatalf("kpFlushTriggerCPUs(128) = %d, want %d", got, defaultKpFlushTriggerCPUs)
	}
}

func TestKpFlushTriggerCPUsEnv(t *testing.T) {
	t.Setenv(kpFlushTriggerCPUsEnv, "0")
	if got := kpFlushTriggerCPUs(128); got != 0 {
		t.Fatalf("disabled kpFlushTriggerCPUs = %d, want 0", got)
	}

	t.Setenv(kpFlushTriggerCPUsEnv, "999")
	if got := kpFlushTriggerCPUs(128); got != maxKpFlushTriggerCPUs {
		t.Fatalf("clamped kpFlushTriggerCPUs = %d, want %d", got, maxKpFlushTriggerCPUs)
	}

	t.Setenv(kpFlushTriggerCPUsEnv, "bad")
	if got := kpFlushTriggerCPUs(8); got != defaultKpFlushTriggerCPUs {
		t.Fatalf("invalid kpFlushTriggerCPUs = %d, want %d", got, defaultKpFlushTriggerCPUs)
	}
}

func TestKpFlushTriggerCPUsForWorkerCoversAllCPUsEachInterval(t *testing.T) {
	cpus := []int{10, 11, 12, 13, 14, 15, 16, 17}
	workers := 4
	seen := map[int]struct{}{}

	for worker := 0; worker < workers; worker++ {
		for _, cpu := range kpFlushTriggerCPUsForWorker(cpus, worker, workers) {
			if _, ok := seen[cpu]; ok {
				t.Fatalf("cpu %d assigned more than once", cpu)
			}
			seen[cpu] = struct{}{}
		}
	}

	if len(seen) != len(cpus) {
		t.Fatalf("covered cpus = %v, want all %v", seen, cpus)
	}
}

func TestKpFlushTriggerCPUsForWorkerBounds(t *testing.T) {
	cpus := []int{1, 2, 3}

	if got := kpFlushTriggerCPUsForWorker(cpus, 3, 3); len(got) != 0 {
		t.Fatalf("out-of-range worker assigned cpus = %v", got)
	}

	got := kpFlushTriggerCPUsForWorker(cpus, -1, 0)
	if len(got) != len(cpus) {
		t.Fatalf("single worker assigned cpus = %v, want %v", got, cpus)
	}
}
