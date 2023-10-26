//go:build linux
// +build linux

// Package conntrack place probes on kernel functions
// `__nf_conntrack_hash_insert` and `nf_ct_delete`
package conntrack

import (
	"bytes"
	"fmt"
	"math"

	manager "github.com/DataDog/ebpf-manager"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
	"golang.org/x/sys/unix"
)

func NewConntrackManger(constEditor []manager.ConstantEditor) (*manager.Manager, error) {
	m := &manager.Manager{
		Probes: []*manager.Probe{
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe___nf_conntrack_hash_insert",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__nf_ct_delete",
				},
			},
		},
	}
	mOpts := manager.Options{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
		ConstantEditors: constEditor,
	}

	buf, err := dkebpf.ConntrackBin()
	if err != nil {
		return nil, fmt.Errorf("conntrack.o: %w", err)
	}

	if err := m.InitWithOptions((bytes.NewReader(buf)), mOpts); err != nil {
		return nil, fmt.Errorf("init conntrack tracer: %w", err)
	}
	return m, nil
}
