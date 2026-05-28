//go:build linux
// +build linux

package l7flow

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

const (
	kpFlushTriggerCPUsEnv      = "DK_EBPF_L7FLOW_KP_FLUSH_TRIGGER_CPUS"
	defaultKpFlushTriggerCPUs  = 4
	maxKpFlushTriggerCPUs      = 64
	kpFlushTriggerInterval     = 5 * time.Second
	minKpFlushTriggerTotalCPUs = 1
	maxAffinityCPUSetBits      = 1024
)

func SetAffinity(cpuID int) error {
	var newMask unix.CPUSet
	newMask.Set(cpuID)

	err := unix.SchedSetaffinity(0, &newMask)
	if err != nil {
		return fmt.Errorf("set affinity error: %w", err)
	}
	return nil
}

// newKpFlushTrigger periodically calls sched_getaffinity so the apiflow
// kprobe can clear stale per-CPU network event scratch space.
func newKpFlushTrigger(ctx context.Context) {
	if ctx == nil {
		return
	}
	cpus := allowedCPUs(runtime.NumCPU())
	workers := kpFlushTriggerCPUs(len(cpus))
	if workers <= 0 {
		log.Infof("l7flow kernel flush trigger disabled")
		return
	}
	if workers > len(cpus) {
		workers = len(cpus)
	}
	log.Infof("l7flow kernel flush trigger workers=%d allowed_cpus=%d interval=%s",
		workers, len(cpus), kpFlushTriggerInterval)

	for workerID := 0; workerID < workers; workerID++ {
		assignedCPUs := kpFlushTriggerCPUsForWorker(cpus, workerID, workers)
		go func(assignedCPUs []int) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			var originalMask unix.CPUSet
			restoreMask := unix.SchedGetaffinity(0, &originalMask) == nil
			if restoreMask {
				defer func() {
					if err := unix.SchedSetaffinity(0, &originalMask); err != nil {
						log.Warnf("restore affinity failed: %s", err.Error())
					}
				}()
			}

			ticker := time.NewTicker(kpFlushTriggerInterval)
			defer ticker.Stop()

			cpuSet := unix.CPUSet{}
			for {
				select {
				case <-ticker.C:
					for _, cpuID := range assignedCPUs {
						if err := SetAffinity(cpuID); err != nil {
							log.Errorf("set affinity error: %s, cpu id: %d", err.Error(), cpuID)
							return
						}
						_ = unix.SchedGetaffinity(0, &cpuSet)
					}
				case <-ctx.Done():
					return
				}
			}
		}(assignedCPUs)
	}
}

func kpFlushTriggerCPUsForWorker(cpus []int, workerID, workers int) []int {
	if len(cpus) == 0 {
		return nil
	}
	if workers <= 0 {
		workers = 1
	}
	if workerID < 0 {
		workerID = 0
	}
	if workerID >= workers {
		return nil
	}

	assigned := make([]int, 0, (len(cpus)+workers-1)/workers)
	for i := workerID; i < len(cpus); i += workers {
		assigned = append(assigned, cpus[i])
	}
	return assigned
}

func kpFlushTriggerCPUs(totalCPUs int) int {
	if totalCPUs < minKpFlushTriggerTotalCPUs {
		return 0
	}
	defaultCPUs := defaultKpFlushTriggerCPUs
	if totalCPUs < defaultCPUs {
		defaultCPUs = totalCPUs
	}

	raw := strings.TrimSpace(os.Getenv(kpFlushTriggerCPUsEnv))
	if raw == "" {
		return defaultCPUs
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		log.Warnf("invalid %s=%q, use default %d", kpFlushTriggerCPUsEnv, raw, defaultCPUs)
		return defaultCPUs
	}
	if n <= 0 {
		return 0
	}
	if n > totalCPUs {
		n = totalCPUs
	}
	if n > maxKpFlushTriggerCPUs {
		n = maxKpFlushTriggerCPUs
	}
	return n
}

func allowedCPUs(totalCPUs int) []int {
	if totalCPUs < minKpFlushTriggerTotalCPUs {
		return nil
	}
	if totalCPUs > maxAffinityCPUSetBits {
		totalCPUs = maxAffinityCPUSetBits
	}
	var mask unix.CPUSet
	if err := unix.SchedGetaffinity(0, &mask); err != nil {
		cpus := make([]int, 0, totalCPUs)
		for i := 0; i < totalCPUs; i++ {
			cpus = append(cpus, i)
		}
		return cpus
	}
	cpus := make([]int, 0, totalCPUs)
	for i := 0; i < totalCPUs; i++ {
		if mask.IsSet(i) {
			cpus = append(cpus, i)
		}
	}
	if len(cpus) == 0 {
		for i := 0; i < totalCPUs; i++ {
			cpus = append(cpus, i)
		}
	}
	return cpus
}
