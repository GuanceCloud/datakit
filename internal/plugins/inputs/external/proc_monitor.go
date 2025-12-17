// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package external

import (
	"os"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/metrics"
	p8s "github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/process"
)

var (
	procMonitorLog = logger.DefaultSLogger("external_proc_monitor")

	// Prometheus metrics for external processes.
	externalProcCPUPercent = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Name: "datakit_external_process_cpu_percent",
			Help: "CPU usage percentage of external processes (eBPF, etc.)",
		},
		[]string{"name"},
	)

	externalProcMemoryRSS = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Name: "datakit_external_process_memory_rss_bytes",
			Help: "Resident Set Size (RSS) memory of external processes in bytes",
		},
		[]string{"name"},
	)

	externalProcMemoryVMS = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Name: "datakit_external_process_memory_vms_bytes",
			Help: "Virtual Memory Size (VMS) of external processes in bytes",
		},
		[]string{"name"},
	)

	externalProcStatus = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Name: "datakit_external_process_status",
			Help: "Status of external processes (1=running, 0=not running)",
		},
		[]string{"name", "status"},
	)

	externalProcThreads = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Name: "datakit_external_process_threads",
			Help: "Number of threads of external processes",
		},
		[]string{"name"},
	)

	externalProcOpenFiles = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Name: "datakit_external_process_open_files",
			Help: "Number of open files of external processes",
		},
		[]string{"name"},
	)

	externalProcIOReadBytes = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Name: "datakit_external_process_io_read_bytes",
			Help: "IO read bytes of external processes",
		},
		[]string{"name"},
	)

	externalProcIOWriteBytes = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Name: "datakit_external_process_io_write_bytes",
			Help: "IO write bytes of external processes",
		},
		[]string{"name"},
	)

	externalProcCtxSwitchesVoluntary = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Name: "datakit_external_process_ctx_switches_voluntary",
			Help: "Voluntary context switches of external processes",
		},
		[]string{"name"},
	)

	externalProcCtxSwitchesInvoluntary = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Name: "datakit_external_process_ctx_switches_involuntary",
			Help: "Involuntary context switches of external processes",
		},
		[]string{"name"},
	)
)

// ProcessMonitor manages external process monitoring.
type ProcessMonitor struct {
	mu        sync.RWMutex
	processes map[string]*ProcessInfo
}

// ProcessInfo holds information about a monitored process.
type ProcessInfo struct {
	Name    string
	Process *os.Process
	Pid     int32
}

var (
	globalMonitor *ProcessMonitor
	monitorOnce   sync.Once
)

// GetProcessMonitor returns the global process monitor instance.
func GetProcessMonitor() *ProcessMonitor {
	monitorOnce.Do(func() {
		globalMonitor = &ProcessMonitor{
			processes: make(map[string]*ProcessInfo),
		}

		// Register Prometheus metrics
		metrics.MustRegister(
			externalProcCPUPercent,
			externalProcMemoryRSS,
			externalProcMemoryVMS,
			externalProcStatus,
			externalProcThreads,
			externalProcOpenFiles,
			externalProcIOReadBytes,
			externalProcIOWriteBytes,
			externalProcCtxSwitchesVoluntary,
			externalProcCtxSwitchesInvoluntary,
		)

		procMonitorLog.Info("External process monitor initialized")
	})
	return globalMonitor
}

// RegisterProcess registers a process for monitoring.
func (pm *ProcessMonitor) RegisterProcess(name string, proc *os.Process) {
	if proc == nil {
		return
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.processes[name] = &ProcessInfo{
		Name:    name,
		Process: proc,
		Pid:     int32(proc.Pid),
	}

	procMonitorLog.Infof("Registered external process for monitoring: %s (PID: %d)", name, proc.Pid)
}

// UnregisterProcess unregisters a process from monitoring.
func (pm *ProcessMonitor) UnregisterProcess(name string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, ok := pm.processes[name]; ok {
		// Clean up metrics for this process
		externalProcCPUPercent.DeleteLabelValues(name)
		externalProcMemoryRSS.DeleteLabelValues(name)
		externalProcMemoryVMS.DeleteLabelValues(name)
		externalProcThreads.DeleteLabelValues(name)
		externalProcOpenFiles.DeleteLabelValues(name)
		externalProcIOReadBytes.DeleteLabelValues(name)
		externalProcIOWriteBytes.DeleteLabelValues(name)
		externalProcCtxSwitchesVoluntary.DeleteLabelValues(name)
		externalProcCtxSwitchesInvoluntary.DeleteLabelValues(name)

		delete(pm.processes, name)
		procMonitorLog.Infof("Unregistered external process: %s", name)
	}
}

// CollectMetrics collects metrics from all registered processes.
func (pm *ProcessMonitor) CollectMetrics() {
	pm.mu.RLock()
	processes := make(map[string]*ProcessInfo)
	for k, v := range pm.processes {
		processes[k] = v
	}
	pm.mu.RUnlock()

	for name, info := range processes {
		pm.collectProcessMetrics(name, info)
	}
}

func (pm *ProcessMonitor) collectProcessMetrics(name string, info *ProcessInfo) {
	proc, err := process.NewProcess(info.Pid)
	if err != nil {
		procMonitorLog.Debugf("Failed to get process %s (PID: %d): %s", name, info.Pid, err)
		externalProcStatus.WithLabelValues(name, "not_found").Set(0)
		return
	}

	// Check if process is running
	isRunning, err := proc.IsRunning()
	if err != nil || !isRunning {
		externalProcStatus.WithLabelValues(name, "stopped").Set(0)
		return
	}

	if status, _ := proc.Status(); len(status) > 0 {
		externalProcStatus.WithLabelValues(name, status[0]).Set(1)
	} else {
		externalProcStatus.WithLabelValues(name, "running").Set(1)
	}

	// CPU usage
	if cpuPercent, err := proc.CPUPercent(); err == nil {
		externalProcCPUPercent.WithLabelValues(name).Set(cpuPercent)
	}

	// Memory info
	if memInfo, err := proc.MemoryInfo(); err == nil {
		externalProcMemoryRSS.WithLabelValues(name).Set(float64(memInfo.RSS))
		externalProcMemoryVMS.WithLabelValues(name).Set(float64(memInfo.VMS))
	}

	// Thread count
	if numThreads, err := proc.NumThreads(); err == nil {
		externalProcThreads.WithLabelValues(name).Set(float64(numThreads))
	}

	// Open files count
	if openFiles, err := proc.OpenFiles(); err == nil {
		externalProcOpenFiles.WithLabelValues(name).Set(float64(len(openFiles)))
	}

	// IO counters
	if ioCounters, err := proc.IOCounters(); err == nil {
		externalProcIOReadBytes.WithLabelValues(name).Set(float64(ioCounters.ReadBytes))
		externalProcIOWriteBytes.WithLabelValues(name).Set(float64(ioCounters.WriteBytes))
	}

	// Context switches
	if ctxSwitches, err := proc.NumCtxSwitches(); err == nil {
		externalProcCtxSwitchesVoluntary.WithLabelValues(name).Set(float64(ctxSwitches.Voluntary))
		externalProcCtxSwitchesInvoluntary.WithLabelValues(name).Set(float64(ctxSwitches.Involuntary))
	}
}

// StartMonitoring starts the monitoring loop.
func (pm *ProcessMonitor) StartMonitoring(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	procMonitorLog.Infof("Starting external process monitoring with interval: %s", interval)

	for range ticker.C {
		pm.CollectMetrics()
	}
}
