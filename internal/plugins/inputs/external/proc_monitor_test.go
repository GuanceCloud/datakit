// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package external

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestProcessMonitor(t *testing.T) {
	monitor := GetProcessMonitor()

	// Test registering a process (using current process for testing)
	currentProc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find current process: %v", err)
	}

	// Register the process
	monitor.RegisterProcess("test_process", currentProc)

	// Check if the process was registered
	monitor.mu.RLock()
	_, exists := monitor.processes["test_process"]
	monitor.mu.RUnlock()

	if !exists {
		t.Error("Process was not registered")
	}

	// Collect metrics
	monitor.CollectMetrics()

	// Wait a bit for metrics collection
	time.Sleep(100 * time.Millisecond)

	// Unregister the process
	monitor.UnregisterProcess("test_process")

	// Check if the process was unregistered
	monitor.mu.RLock()
	_, exists = monitor.processes["test_process"]
	monitor.mu.RUnlock()

	if exists {
		t.Error("Process was not unregistered")
	}
}

func TestProcessMonitorMultiple(t *testing.T) {
	monitor := GetProcessMonitor()

	currentProc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find current process: %v", err)
	}

	// Register multiple instances
	monitor.RegisterProcess("test_proc_1", currentProc)
	monitor.RegisterProcess("test_proc_2", currentProc)

	monitor.mu.RLock()
	count := len(monitor.processes)
	monitor.mu.RUnlock()

	if count < 2 {
		t.Errorf("Expected at least 2 processes, got %d", count)
	}

	// Collect metrics for all
	monitor.CollectMetrics()

	// Clean up
	monitor.UnregisterProcess("test_proc_1")
	monitor.UnregisterProcess("test_proc_2")
}

func TestProcessMonitorNilProcess(t *testing.T) {
	monitor := GetProcessMonitor()

	// Test registering nil process (should be ignored)
	monitor.RegisterProcess("test_nil", nil)

	monitor.mu.RLock()
	_, exists := monitor.processes["test_nil"]
	monitor.mu.RUnlock()

	if exists {
		t.Error("Nil process should not be registered")
	}
}

func TestProcessMonitorMetricsCollection(t *testing.T) {
	monitor := GetProcessMonitor()

	currentProc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find current process: %v", err)
	}

	// Register process
	processName := "test_metrics_collection"
	monitor.RegisterProcess(processName, currentProc)
	defer monitor.UnregisterProcess(processName)

	// Collect metrics
	monitor.CollectMetrics()

	// Wait for metrics to be collected
	time.Sleep(200 * time.Millisecond)

	// Verify process info is stored correctly
	monitor.mu.RLock()
	info, exists := monitor.processes[processName]
	monitor.mu.RUnlock()

	if !exists {
		t.Error("Process info should exist after registration")
	}

	if info.Name != processName {
		t.Errorf("Expected process name %s, got %s", processName, info.Name)
	}

	if info.Pid != int32(os.Getpid()) {
		t.Errorf("Expected PID %d, got %d", os.Getpid(), info.Pid)
	}
}

func TestProcessMonitorDuplicateRegistration(t *testing.T) {
	monitor := GetProcessMonitor()

	currentProc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find current process: %v", err)
	}

	processName := "test_duplicate"

	// Register same process twice
	monitor.RegisterProcess(processName, currentProc)
	monitor.RegisterProcess(processName, currentProc)

	monitor.mu.RLock()
	count := 0
	for name := range monitor.processes {
		if name == processName {
			count++
		}
	}
	monitor.mu.RUnlock()

	if count != 1 {
		t.Errorf("Expected 1 instance of process, got %d", count)
	}

	// Clean up
	monitor.UnregisterProcess(processName)
}

func TestProcessMonitorUnregisterNonExistent(t *testing.T) {
	monitor := GetProcessMonitor()

	// Unregister a process that was never registered (should not panic)
	monitor.UnregisterProcess("non_existent_process")

	// Test passes if no panic occurs
}

func TestProcessMonitorConcurrentAccess(t *testing.T) {
	monitor := GetProcessMonitor()

	currentProc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find current process: %v", err)
	}

	// Test concurrent registration and unregistration
	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func(id int) {
			processName := fmt.Sprintf("test_concurrent_%d", id)
			monitor.RegisterProcess(processName, currentProc)
			time.Sleep(10 * time.Millisecond)
			monitor.UnregisterProcess(processName)
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		go func() {
			monitor.CollectMetrics()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
