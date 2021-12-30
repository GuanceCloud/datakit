//go:build (linux && ignore) || ebpf
// +build linux,ignore ebpf

package ebpf

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"os/user"
	"time"
	"unsafe"

	"github.com/DataDog/ebpf/manager"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/c"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/feed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/external"
	"golang.org/x/sys/unix"
)

// #include "../c/bash_history/bash_history.h"
import "C"

type BashEventC C.struct_bash_event

const srcNameM = "bash"

var l = logger.DefaultSLogger(srcNameM)

func SetLogger(nl *logger.Logger) {
	l = nl
}

type Inputs struct {
	external.ExernalInput
}

func NewBashManger(bashReadlineEventHandler func(cpu int, data []byte,
	perfmap *manager.PerfMap, manager *manager.Manager)) (*manager.Manager, error) {
	m := &manager.Manager{
		Probes: []*manager.Probe{
			{
				Section:    "uretprobe/readline",
				BinaryPath: "/bin/bash",
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
	if buf, err := dkebpf.Asset("bash_history.o"); err != nil {
		return nil, err
	} else if err := m.InitWithOptions((bytes.NewReader(buf)), mOpts); err != nil {
		return nil, err
	}

	return m, nil
}

type BashTracer struct {
	ch     chan *measurement
	stopCh chan struct{}
	gTags  map[string]string
}

func NewBashTracer() *BashTracer {
	return &BashTracer{
		ch:     make(chan *measurement, 32),
		stopCh: make(chan struct{}),
	}
}

func (tracer *BashTracer) readlineCallBack(cpu int, data []byte,
	perfmap *manager.PerfMap, manager *manager.Manager) {
	eventC := (*BashEventC)(unsafe.Pointer(&data[0])) //nolint:gosec

	m := measurement{
		ts:     time.Now(),
		name:   srcNameM,
		tags:   map[string]string{},
		fields: map[string]interface{}{},
	}

	lineChar := (*(*[128]byte)(unsafe.Pointer(&eventC.line)))
	m.fields["cmd"] = unix.ByteSliceToString(lineChar[:])

	u, err := user.LookupId(fmt.Sprintf("%d", int(eventC.uid_gid>>32)))
	if err != nil {
		l.Error(err)
	} else {
		m.fields["user"] = u.Name
	}
	m.fields["pid"] = fmt.Sprintf("%d", int(eventC.pid_tgid>>32))

	m.fields["message"] = fmt.Sprintf("%s pid:`%s` user:`%s` cmd:`%s`",
		m.ts.Format(time.RFC3339), m.fields["pid"], m.fields["user"], m.fields["cmd"])
	select {
	case <-tracer.stopCh:
	case tracer.ch <- &m:
	}
}

func (tracer *BashTracer) feedHandler(ctx context.Context, datakitPostURL string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	cache := []inputs.Measurement{}
	for {
		select {
		case <-ticker.C:
			if len(cache) > 0 {
				if err := feed.FeedMeasurement(cache, datakitPostURL); err != nil {
					l.Error(err)
				}
				cache = make([]inputs.Measurement, 0)
			}
		case m := <-tracer.ch:
			for k, v := range tracer.gTags {
				if _, ok := m.tags[k]; !ok {
					m.tags[k] = v
				}
			}
			l.Debug(m)
			cache = append(cache, m)
			if len(cache) > 128 {
				if err := feed.FeedMeasurement(cache, datakitPostURL); err != nil {
					l.Error(err)
				}
				cache = make([]inputs.Measurement, 0)
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

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *measurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *measurement) Info() *inputs.MeasurementInfo {
	return nil
}
