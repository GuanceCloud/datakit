//go:build (linux && amd64 && ebpf) || (linux && arm64 && ebpf)
// +build linux,amd64,ebpf linux,arm64,ebpf

package conntrack

import (
	"bytes"
	"fmt"
	"math"

	"github.com/DataDog/ebpf/manager"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/c"
	"golang.org/x/sys/unix"
)

func NewConntrackManger(constEditor []manager.ConstantEditor) (*manager.Manager, error) {
	m := &manager.Manager{
		Probes: []*manager.Probe{
			{
				Section: "kprobe/__nf_conntrack_hash_insert",
			},
			{
				Section: "kprobe/nf_ct_delete_from_lists",
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
		return nil, fmt.Errorf("init offset conntrack guess: %w", err)
	}
	return m, nil
}
