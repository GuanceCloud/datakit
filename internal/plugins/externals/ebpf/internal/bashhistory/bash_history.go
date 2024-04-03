//go:build linux
// +build linux

// Package bashhistory collects bash history
package bashhistory

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"os/user"
	"time"
	"unsafe"

	manager "github.com/DataDog/ebpf-manager"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"

	"golang.org/x/sys/unix"
)

// #include "../c/bash_history/bash_history.h"
import "C"

type BashEventC C.struct_bash_event

const srcNameM = "bash"

var (
	l = logger.DefaultSLogger(srcNameM)
)

func SetLogger(nl *logger.Logger) {
	l = nl
}

func NewBashManger(bashReadlineEventHandler func(cpu int, data []byte,
	perfmap *manager.PerfMap, manager *manager.Manager)) (*manager.Manager, error) {
	m := &manager.Manager{
		Probes: []*manager.Probe{
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "uretprobe__readline",
				},
				KProbeMaxActive: 128,
				BinaryPath:      "/bin/bash",
			},
		},
		PerfMaps: []*manager.PerfMap{
			{
				Map: manager.Map{
					Name: "bpfmap_bash_readline",
				},
				PerfMapOptions: manager.PerfMapOptions{
					PerfRingBufferSize: 32 * os.Getpagesize(),
					DataHandler:        bashReadlineEventHandler,
				},
			},
		},
	}
	mOpts := manager.Options{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
	}
	if buf, err := dkebpf.BashHistoryBin(); err != nil {
		return nil, fmt.Errorf("bash_history.o: %w", err)
	} else if err := m.InitWithOptions((bytes.NewReader(buf)), mOpts); err != nil {
		return nil, err
	}

	return m, nil
}

type BashTracer struct {
	ch     chan *point.Point
	stopCh chan struct{}
	gTags  map[string]string
}

func NewBashTracer() *BashTracer {
	return &BashTracer{
		ch:     make(chan *point.Point, 32),
		stopCh: make(chan struct{}),
	}
}

func (tracer *BashTracer) readlineCallBack(cpu int, data []byte,
	perfmap *manager.PerfMap, manager *manager.Manager) {
	eventC := (*BashEventC)(unsafe.Pointer(&data[0])) //nolint:gosec

	mTags := map[string]string{}

	for k, v := range tracer.gTags {
		if _, ok := mTags[k]; !ok {
			mTags[k] = v
		}
	}

	mFields := map[string]interface{}{}

	lineChar := (*(*[128]byte)(unsafe.Pointer(&eventC.line)))
	mFields["cmd"] = unix.ByteSliceToString(lineChar[:])

	u, err := user.LookupId(fmt.Sprintf("%d", int(eventC.uid_gid>>32)))
	if err != nil {
		l.Error(err)
	} else {
		mFields["user"] = u.Name
	}
	mFields["pid"] = fmt.Sprintf("%d", int(eventC.pid_tgid>>32))

	mFields["message"] = fmt.Sprintf("%s pid:`%s` user:`%s` cmd:`%s`",
		time.Now().Format(time.RFC3339), mFields["pid"], mFields["user"], mFields["cmd"])

	kvs := point.NewTags(mTags)
	kvs = append(kvs, point.NewKVs(mFields)...)
	pt := point.NewPointV2(srcNameM, kvs, point.CommonLoggingOptions()...)

	select {
	case <-tracer.stopCh:
	case tracer.ch <- pt:
	}
}

func (tracer *BashTracer) feedHandler(ctx context.Context, datakitPostURL string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	cache := []*point.Point{}
	for {
		select {
		case <-ticker.C:
			if len(cache) > 0 {
				if err := exporter.FeedPoint(datakitPostURL, cache, false); err != nil {
					l.Error(err)
				}
				cache = make([]*point.Point, 0)
			}
		case pt := <-tracer.ch:
			cache = append(cache, pt)
			if len(cache) > 128 {
				if err := exporter.FeedPoint(datakitPostURL, cache, false); err != nil {
					l.Error(err)
				}
				cache = make([]*point.Point, 0)
			}
		case <-tracer.stopCh:
			return
		}
	}
}

func (tracer *BashTracer) Run(ctx context.Context, gTags map[string]string,
	datakitPostURL string, interval time.Duration) error {
	tracer.gTags = gTags

	go tracer.feedHandler(ctx, datakitPostURL, interval)

	bpfManger, err := NewBashManger(tracer.readlineCallBack)
	if err != nil {
		l.Error(err)
		return err
	}
	if err := bpfManger.Start(); err != nil {
		l.Error(err)
		return err
	}

	go func() {
		<-ctx.Done()
		close(tracer.stopCh)
	}()
	return nil
}
