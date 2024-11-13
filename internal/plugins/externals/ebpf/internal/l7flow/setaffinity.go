//go:build linux
// +build linux

package l7flow

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"golang.org/x/sys/unix"
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

func newKpFlushTrigger(ctx context.Context) {
	for i := 0; i < runtime.NumCPU(); i++ {
		go func(i int) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			if err := SetAffinity(i); err != nil {
				log.Errorf("set affinity error: %s, cpu id: %d", err.Error(), i)
			} else {
				log.Debugf("set affinity, thread id: %d, cpu id: %d", unix.Gettid(), i)
			}

			ticker := time.NewTicker(time.Second * 5)

			cpuSet := unix.CPUSet{}
			for {
				select {
				case <-ticker.C:
					_ = unix.SchedGetaffinity(0, &cpuSet)
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}
}
