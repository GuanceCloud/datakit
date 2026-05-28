//go:build linux
// +build linux

// Package bashhistory collects bash history
package bashhistory

//go:generate go run ../c/genlayout -target bashhistory

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os/user"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"

	"golang.org/x/sys/unix"
)

const bashEventSize = int(unsafe.Sizeof(BashEventC{})) //nolint:gosec

const (
	srcNameM      = "bash"
	inputNameBash = "ebpf-bash"
)

var l = logger.DefaultSLogger(srcNameM)

func SetLogger(nl *logger.Logger) {
	l = nl
}

func NewBashRuntime(bashReadlineEventHandler bpfutil.PerfHandler) (*bpfutil.Runtime, error) {
	perfRingBufferSize := bpfutil.SmallPerfRingBufferSize()
	runtime := &bpfutil.Runtime{
		Probes: []*bpfutil.HookSpec{
			{
				ID: bpfutil.HookID{
					Program: "uretprobe__readline",
				},
				KProbeMaxActive: 128,
				BinaryPath:      "/bin/bash",
			},
		},
		Streams: []*bpfutil.PerfStream{
			{
				Map: bpfutil.Map{
					Name: "bpfmap_bash_readline",
				},
				PerfStreamOptions: bpfutil.PerfStreamOptions{
					PerfRingBufferSize: perfRingBufferSize,
					Watermark:          bpfutil.PerfWatermark(perfRingBufferSize),
					DataHandler:        bashReadlineEventHandler,
					LostHandler: func(cpu int, count uint64, stream *bpfutil.PerfStream, runtime *bpfutil.Runtime) {
						exporter.AddPerfLost("bashhistory", stream.Name, count)
					},
					ErrorHandler: func(err error, stream *bpfutil.PerfStream, runtime *bpfutil.Runtime) {
						exporter.IncPerfReadError("bashhistory", stream.Name)
						l.Warnf("bashhistory perf stream stopped: %v", err)
					},
				},
			},
		},
	}
	loadSpec := bpfutil.LoadSpec{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
	}
	if buf, err := dkebpf.BashHistoryBin(); err != nil {
		return nil, fmt.Errorf("bash_history.o: %w", err)
	} else if err := runtime.LoadFromReader((bytes.NewReader(buf)), loadSpec); err != nil {
		return nil, err
	}

	return runtime, nil
}

type BashTracer struct {
	ch        chan *point.Point
	stopCh    chan struct{}
	stopOnce  sync.Once
	gTags     map[string]string
	userMu    sync.RWMutex
	userCache map[uint32]string
}

const bashUserCacheLimit = 4096

func NewBashTracer() *BashTracer {
	return &BashTracer{
		ch:        make(chan *point.Point, 32),
		stopCh:    make(chan struct{}),
		userCache: make(map[uint32]string),
	}
}

func (tracer *BashTracer) lookupUser(uid uint32) string {
	if tracer == nil {
		return ""
	}
	tracer.userMu.RLock()
	name, ok := tracer.userCache[uid]
	tracer.userMu.RUnlock()
	if ok {
		return name
	}

	u, err := user.LookupId(strconv.FormatUint(uint64(uid), 10))
	if err == nil && u != nil {
		name = u.Name
	} else {
		l.Debugf("lookup user for uid %d failed: %v", uid, err)
	}

	tracer.cacheUser(uid, name)
	return name
}

func (tracer *BashTracer) cacheUser(uid uint32, name string) {
	tracer.userMu.Lock()
	defer tracer.userMu.Unlock()
	if tracer.userCache == nil {
		tracer.userCache = make(map[uint32]string)
	}
	if _, ok := tracer.userCache[uid]; ok {
		return
	}
	if len(tracer.userCache) >= bashUserCacheLimit {
		tracer.userCache = make(map[uint32]string)
	}
	tracer.userCache[uid] = name
}

func (tracer *BashTracer) readlineCallBack(cpu int, data []byte,
	stream *bpfutil.PerfStream, runtime *bpfutil.Runtime,
) {
	if len(data) < bashEventSize {
		l.Debugf("drop short bash event: got %d want >= %d", len(data), bashEventSize)
		return
	}

	eventC := (*BashEventC)(unsafe.Pointer(&data[0])) //nolint:gosec

	mTags := map[string]string{}

	for k, v := range tracer.gTags {
		if _, ok := mTags[k]; !ok {
			mTags[k] = v
		}
	}

	mFields := map[string]interface{}{}

	lineChar := eventC.line
	mFields["cmd"] = unix.ByteSliceToString(lineChar[:])

	uid := uint32(eventC.uid_gid >> 32)
	userName := tracer.lookupUser(uid)
	if userName == "" {
		userName = strconv.FormatUint(uint64(uid), 10)
	}
	mFields["user"] = userName
	mFields["pid"] = fmt.Sprintf("%d", int(eventC.pid_tgid>>32))

	mFields["message"] = fmt.Sprintf("%s pid:`%s` user:`%s` cmd:`%s`",
		ntp.Now().Format(time.RFC3339), mFields["pid"], mFields["user"], mFields["cmd"])

	kvs := point.NewTags(mTags)
	kvs = append(kvs, point.NewKVs(mFields)...)
	pt := point.NewPoint(srcNameM, kvs, point.CommonLoggingOptions()...)

	select {
	case <-tracer.stopCh:
	case tracer.ch <- pt:
	default:
		l.Debug("drop bash event: queue full")
	}
}

func (tracer *BashTracer) feedHandler(ctx context.Context, interval time.Duration) {
	if tracer == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	cache := []*point.Point{}
	for {
		select {
		case <-ticker.C:
			if len(cache) > 0 {
				if err := exporter.FeedPoint(inputNameBash, point.Logging, cache); err != nil {
					l.Error(err)
				}
				cache = make([]*point.Point, 0)
			}
		case pt := <-tracer.ch:
			if pt == nil {
				continue
			}
			cache = append(cache, pt)
			if len(cache) > 128 {
				if err := exporter.FeedPoint(inputNameBash, point.Logging, cache); err != nil {
					l.Error(err)
				}
				cache = make([]*point.Point, 0)
			}
		case <-tracer.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (tracer *BashTracer) Run(ctx context.Context, gTags map[string]string,
	interval time.Duration,
) error {
	if tracer == nil {
		return fmt.Errorf("bash tracer is not initialized")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if interval <= 0 {
		interval = time.Minute
	}
	if tracer.ch == nil {
		tracer.ch = make(chan *point.Point, 32)
	}
	if tracer.stopCh == nil {
		tracer.stopCh = make(chan struct{})
	}
	tracer.gTags = gTags

	runtime, err := NewBashRuntime(tracer.readlineCallBack)
	if err != nil {
		l.Error(err)
		return err
	}
	if err := runtime.StartRuntime(); err != nil {
		l.Error(err)
		_ = runtime.Shutdown()
		return err
	}

	go tracer.feedHandler(ctx, interval)
	go func() {
		<-ctx.Done()
		_ = runtime.Shutdown()
		tracer.stopOnce.Do(func() {
			close(tracer.stopCh)
		})
	}()
	return nil
}
