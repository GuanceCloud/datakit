// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package external

import (
	"os"
	"sync"

	p8s "github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/process"
)

var (

	// Prometheus metrics for external processes.

	externalProcCPUPercentDesc = p8s.NewDesc(
		"datakit_external_process_cpu_percent",
		"CPU usage percentage of external processes (eBPF, etc.)",
		[]string{"name"}, nil,
	)

	externalProcMemoryRSSDesc = p8s.NewDesc(
		"datakit_external_process_memory_rss_bytes",
		"Resident Set Size (RSS) memory of external processes in bytes",
		[]string{"name"}, nil,
	)

	externalProcMemoryVMSDesc = p8s.NewDesc(
		"datakit_external_process_memory_vms_bytes",
		"Virtual Memory Size (VMS) of external processes in bytes",
		[]string{"name"}, nil,
	)

	externalProcStatusDesc = p8s.NewDesc(
		"datakit_external_process_status",
		"Status of external processes (1=running, 0=not running)",
		[]string{"name", "status"}, nil,
	)

	externalProcThreadsDesc = p8s.NewDesc(
		"datakit_external_process_threads",
		"Number of threads of external processes",
		[]string{"name"}, nil,
	)

	externalProcOpenFilesDesc = p8s.NewDesc(
		"datakit_external_process_open_files",
		"Number of open files of external processes",
		[]string{"name"}, nil,
	)

	externalProcIOReadBytesDesc = p8s.NewDesc(
		"datakit_external_process_io_read_bytes",
		"IO read bytes of external processes",
		[]string{"name"}, nil,
	)

	externalProcIOWriteBytesDesc = p8s.NewDesc(
		"datakit_external_process_io_write_bytes",
		"IO write bytes of external processes",
		[]string{"name"}, nil,
	)

	externalProcCtxSwitchesVoluntaryDesc = p8s.NewDesc(
		"datakit_external_process_ctx_switches_voluntary",
		"Voluntary context switches of external processes",
		[]string{"name"}, nil,
	)

	externalProcCtxSwitchesInvoluntaryDesc = p8s.NewDesc(
		"datakit_external_process_ctx_switches_involuntary",
		"Involuntary context switches of external processes",
		[]string{"name"}, nil,
	)
)

type metricCollector struct{}

func (c metricCollector) Describe(ch chan<- *p8s.Desc) {
	p8s.DescribeByCollect(c, ch)
}

func (c metricCollector) Collect(ch chan<- p8s.Metric) {
	globalMonitor.CollectMetrics(ch)
}

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

var globalMonitor = &ProcessMonitor{
	processes: make(map[string]*ProcessInfo),
}

// GetProcessMonitor returns the global process monitor instance.
func GetProcessMonitor() *ProcessMonitor {
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

	log.Infof("Registered external process for monitoring: %s (PID: %d)", name, proc.Pid)
}

// UnregisterProcess unregisters a process from monitoring.
func (pm *ProcessMonitor) UnregisterProcess(name string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, ok := pm.processes[name]; ok {
		// Clean up metrics for this process
		delete(pm.processes, name)
		log.Infof("Unregistered external process: %s", name)
	}
}

// CollectMetrics collects metrics from all registered processes.
func (pm *ProcessMonitor) CollectMetrics(ch chan<- p8s.Metric) {
	pm.mu.RLock()
	processes := make(map[string]*ProcessInfo)
	for k, v := range pm.processes {
		processes[k] = v
	}
	pm.mu.RUnlock()

	for name, info := range processes {
		pm.collectProcessMetrics(name, info, ch)
	}
}

func (pm *ProcessMonitor) collectProcessMetrics(name string, info *ProcessInfo, ch chan<- p8s.Metric) {
	proc, err := process.NewProcess(info.Pid)
	if err != nil {
		log.Debugf("Failed to get process %s (PID: %d): %s", name, info.Pid, err)
		ch <- p8s.MustNewConstMetric(externalProcStatusDesc, p8s.GaugeValue, 0, name, "not_found")
		return
	}

	// Check if process is running
	isRunning, err := proc.IsRunning()
	if err != nil || !isRunning {
		ch <- p8s.MustNewConstMetric(externalProcStatusDesc, p8s.GaugeValue, 0, name, "stopped")
		return
	}

	if status, _ := proc.Status(); len(status) > 0 {
		ch <- p8s.MustNewConstMetric(externalProcStatusDesc, p8s.GaugeValue, 1, name, status[0])
	} else {
		ch <- p8s.MustNewConstMetric(externalProcStatusDesc, p8s.GaugeValue, 1, name, "running")
	}

	// CPU usage
	if cpuPercent, err := proc.CPUPercent(); err == nil {
		ch <- p8s.MustNewConstMetric(externalProcCPUPercentDesc, p8s.GaugeValue, cpuPercent, name)
	}

	// Memory info
	if memInfo, err := proc.MemoryInfo(); err == nil {
		ch <- p8s.MustNewConstMetric(externalProcMemoryRSSDesc, p8s.GaugeValue, float64(memInfo.RSS), name)
		ch <- p8s.MustNewConstMetric(externalProcMemoryVMSDesc, p8s.GaugeValue, float64(memInfo.VMS), name)
	}

	// Thread count
	if numThreads, err := proc.NumThreads(); err == nil {
		ch <- p8s.MustNewConstMetric(externalProcThreadsDesc, p8s.GaugeValue, float64(numThreads), name)
	}

	// Open files count
	if openFiles, err := proc.OpenFiles(); err == nil {
		ch <- p8s.MustNewConstMetric(externalProcOpenFilesDesc, p8s.GaugeValue, float64(len(openFiles)), name)
	}

	// IO counters
	if ioCounters, err := proc.IOCounters(); err == nil {
		ch <- p8s.MustNewConstMetric(externalProcIOReadBytesDesc, p8s.GaugeValue, float64(ioCounters.ReadBytes), name)
		ch <- p8s.MustNewConstMetric(externalProcIOWriteBytesDesc, p8s.GaugeValue, float64(ioCounters.WriteBytes), name)
	}

	// Context switches
	if ctxSwitches, err := proc.NumCtxSwitches(); err == nil {
		ch <- p8s.MustNewConstMetric(externalProcCtxSwitchesVoluntaryDesc, p8s.GaugeValue, float64(ctxSwitches.Voluntary), name)
		ch <- p8s.MustNewConstMetric(externalProcCtxSwitchesInvoluntaryDesc, p8s.GaugeValue, float64(ctxSwitches.Involuntary), name)
	}
}
