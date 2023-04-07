// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cgroup

import (
	"os"
	"reflect"
	"runtime"
	"syscall"
	"testing"

	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func TestCgroup_makeLinuxResource(t *testing.T) {
	type fields struct {
		opt       *CgroupOptions
		cpuHigh   float64
		quotaHigh int64
		err       error
		control   cgroup1.Cgroup
		manager   *cgroup2.Manager
	}
	tests := []struct {
		name   string
		fields fields
		want   *specs.LinuxResources
	}{
		{
			name: "",
			fields: fields{
				opt: &CgroupOptions{
					Path:       "/datakit_test",
					CPUMax:     0.30,
					MemMax:     20,
					DisableOOM: false,
					Enable:     true,
				},
			},
			want: mockLinuxResource(0.30, 20),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cgroup{
				opt:       tt.fields.opt,
				cpuHigh:   tt.fields.cpuHigh,
				quotaHigh: tt.fields.quotaHigh,
				err:       tt.fields.err,
				control:   tt.fields.control,
				manager:   tt.fields.manager,
			}
			if got := c.makeLinuxResource(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeLinuxResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockLinuxResource(cpu float64, mem int64) *specs.LinuxResources {
	quotaHigh := int64(float64(period) * (cpu * float64(runtime.NumCPU()) / 100))
	memMax := mem * MB
	swap := memMax
	swappiness := uint64(0)
	oom := false
	resource := &specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Period: &period,
			Quota:  &quotaHigh,
		},
		Memory: &specs.LinuxMemory{
			Limit:            &memMax,
			Swap:             &swap,
			Swappiness:       &swappiness,
			DisableOOMKiller: &oom,
		},
	}

	return resource
}

func Test_Setup(t *testing.T) {
	if rd := syscall.Access("/sys/fs/cgroup", syscall.O_RDWR); rd != nil {
		t.Log("permission denied,return")
		return
	}

	c := &Cgroup{opt: &CgroupOptions{Path: "/datakit_test"}}
	pid := os.Getpid()
	if cgroups.Mode() == cgroups.Unified {
		t.Log("goto test cgroup V2")
		err := c.setupV2(mockLinuxResource(0.30, 100), pid)
		if err != nil || c.manager == nil {
			t.Errorf("setup cgroup V2 err=%v", err)
			return
		} else {
			_ = c.manager.Delete()
		}
	} else {
		l.Infof("goto test cgroup V1")
		err := c.setupV1(mockLinuxResource(0.30, 100), pid)
		if err != nil || c.control == nil {
			t.Errorf("setup cgroup V1 err= %v", err)
			return
		} else {
			_ = c.control.Delete()
		}
	}
}
