//go:build linux && cgo
// +build linux,cgo

package l7flow

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/procwatch"
)

func writeFakeProc(t *testing.T, procRoot string, pid int, name string, ppid int, cmdline []string) {
	t.Helper()

	pidDir := filepath.Join(procRoot, strconv.Itoa(pid))
	if err := os.MkdirAll(pidDir, 0o755); err != nil {
		t.Fatal(err)
	}

	stat := "1 (" + name + ") S " +
		strconv.Itoa(ppid) +
		" 1 1 0 -1 4194560 0 0 0 0 0 0 0 0 20 0 1 0 12345 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0\n"
	if err := os.WriteFile(filepath.Join(pidDir, "stat"), []byte(stat), 0o644); err != nil {
		t.Fatal(err)
	}

	if len(cmdline) == 0 {
		return
	}

	payload := make([]byte, 0, 64)
	for _, arg := range cmdline {
		payload = append(payload, []byte(arg)...)
		payload = append(payload, 0)
	}
	if err := os.WriteFile(filepath.Join(pidDir, "cmdline"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPerfEventHandleDropsShortRecord(t *testing.T) {
	tracer := &Tracer{}

	done := make(chan struct{})
	go func() {
		tracer.PerfEventHandle(0, nil, nil, nil)
		tracer.PerfEventHandle(0, make([]byte, eventsHdrSize-1), nil, nil)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("PerfEventHandle blocked on short record")
	}
}

func TestPopulateProcessInfoResolvesOnDemand(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	const pid = 4321
	writeFakeProc(t, procRoot, pid, "python3", 1, []string{"python3", "/opt/app/server.py"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracer := &Tracer{
		catalog: procwatch.NewCatalog(ctx, procwatch.WithSelfPID(-1)),
	}

	conn := comm.ConnectionInfo{Pid: pid}
	if ok := tracer.populateProcessInfo(&conn); !ok {
		t.Fatal("expected collectable process")
	}
	if conn.ProcessName != "python3" {
		t.Fatalf("unexpected process name: %q", conn.ProcessName)
	}
	if conn.ServiceName != "server" {
		t.Fatalf("unexpected service name: %q", conn.ServiceName)
	}
}

func TestPopulateProcessInfoDropsBlacklistedProcess(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	const pid = 1234
	writeFakeProc(t, procRoot, pid, "curl", 1, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracer := &Tracer{
		catalog: procwatch.NewCatalog(ctx,
			procwatch.WithSelfPID(-1),
			procwatch.WithTracing(true),
			procwatch.WithNameBlacklist([]string{"curl"}),
		),
	}

	conn := comm.ConnectionInfo{Pid: pid}
	if ok := tracer.populateProcessInfo(&conn); ok {
		t.Fatal("expected blacklisted process to be rejected immediately")
	}
	if conn.ProcessName != "" || conn.ServiceName != "" {
		t.Fatalf("expected blacklisted process metadata to remain empty, got %+v", conn)
	}
}

func TestSort(t *testing.T) {
	fn := func(cases []uint64, expected []uint64, start ...uint64) {
		rst := []uint64{}
		netdata := dataQueue{}
		if len(start) > 0 {
			netdata.prvDataPos = start[0]
		}
		for _, v := range cases {
			r := netdata.Queue(&comm.NetwrkData{Index: v})
			for _, d := range r {
				rst = append(rst, d.Index)
			}
		}
		assert.Equal(t, expected, rst)
	}

	t.Run("c1", func(t *testing.T) {
		li := []uint64{1, 4, 3, 5, 2}
		fn(li, []uint64{1, 2, 3, 4, 5})
	})

	t.Run("c2", func(t *testing.T) {
		li := []uint64{5, 4, 3, 2, 1}
		fn(li, []uint64{1, 2, 3, 4, 5})
	})

	t.Run("c3", func(t *testing.T) {
		li := []uint64{1, 2, 3, 4, 5}
		fn(li, []uint64{1, 2, 3, 4, 5})
	})

	t.Run("c4", func(t *testing.T) {
		li := []uint64{1, 2}
		fn(li, []uint64{1, 2})
	})

	t.Run("c5", func(t *testing.T) {
		li := []uint64{2, 1}
		fn(li, []uint64{1, 2})
	})
	t.Run("c5", func(t *testing.T) {
		li := []uint64{2, 4, 3, 1}
		fn(li, []uint64{1, 2, 3, 4})
	})

	t.Run("c6", func(t *testing.T) {
		li := []uint64{1, 6, 2, 3, 4, 5, 7, 9, 10, 8, 11, 14, 13, 12}
		fn(li, []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14})
	})

	t.Run("c7", func(t *testing.T) {
		var li []uint64
		var rst []uint64
		for i := 2; i < 2+queueWindow; i++ {
			rst = append(rst, uint64(i))
			li = append(li, uint64(i))
		}
		fn(li, rst)
	})

	t.Run("c8", func(t *testing.T) {
		var li, rst []uint64
		startPos := uint64(math.MaxUint64 - 100)
		for i := startPos + 1; i < math.MaxUint64; i++ {
			li = append(li, i)
			rst = append(rst, i)
		}

		li = append(li, 1, math.MaxUint64, 2, 0, 3)
		rst = append(rst, math.MaxUint64, 0, 1, 2, 3)
		fn(li, rst, startPos)
	})
}

func TestConnWatcherRotateTracePoints(t *testing.T) {
	watcher := &ConnWatcher{
		trace: &NetTrace{
			ptsPrv: []*point.Point{nil},
			ptsCur: []*point.Point{nil, nil},
		},
	}

	pts, threadInnerID, prevLen, currLen := watcher.rotateTracePoints()
	if len(pts) != 1 {
		t.Fatalf("unexpected rotated previous points length: %d", len(pts))
	}
	if threadInnerID == nil {
		t.Fatal("expected thread trace to be returned")
	}
	if prevLen != 1 || currLen != 2 {
		t.Fatalf("unexpected trace cache lengths: prev=%d curr=%d", prevLen, currLen)
	}
	if len(watcher.trace.ptsPrv) != 2 {
		t.Fatalf("expected current points to become previous batch, got %d", len(watcher.trace.ptsPrv))
	}
	if watcher.trace.ptsCur != nil {
		t.Fatal("expected current batch to be cleared after rotation")
	}
}

func TestProtoKernelFilterTryFilterNonBlocking(t *testing.T) {
	filter := &protoKernelFilter{
		keySk: make(chan uint64, 1),
	}
	filter.keySk <- 1

	if ok := filter.tryFilter(2); ok {
		t.Fatal("expected tryFilter to fail when the queue is full")
	}
}

func TestStreamHandleQueuesProtocolFilterOnlyOnce(t *testing.T) {
	filter := &protoKernelFilter{
		keySk: make(chan uint64, 2),
	}
	pipe := &FlowPipe{
		detecTimes: maxDetec,
	}
	var uniID CUniID
	trace := &NetTrace{
		protocolFilter: filter,
	}
	shard := trace.shardFor(uniID)
	shard.ensureMaps()
	shard.open[uniID] = pipe
	data := &comm.NetwrkData{
		SockPtr: 1234,
	}

	trace.StreamHandle(time.Now().UnixNano(), uniID, data)
	if !pipe.protocolFilterQueued {
		t.Fatal("expected protocol filter request to be queued")
	}
	if got := len(filter.keySk); got != 1 {
		t.Fatalf("expected one queued protocol filter request, got %d", got)
	}

	trace.StreamHandle(time.Now().UnixNano(), uniID, data)
	if got := len(filter.keySk); got != 1 {
		t.Fatalf("expected repeated packets to avoid requeueing the same protocol filter request, got %d", got)
	}
}

func TestConnMapShardMaybeCompact(t *testing.T) {
	openPipe := &FlowPipe{}
	closedPipe := &FlowPipe{}
	var openID CUniID
	var closedID CUniID
	trace := &NetTrace{}
	shard := trace.shardFor(openID)
	shard.ensureMaps()
	shard.open[openID] = openPipe
	shard.closed[closedID] = closedPipe
	shard.delCount = [2]int{
		connMapCompactThreshold + 1,
		connMapCompactThreshold + 1,
	}

	oldOpenMap := shard.open
	oldClosedMap := shard.closed

	shard.maybeCompact()

	if shard.delCount[0] != 0 || shard.delCount[1] != 0 {
		t.Fatalf("unexpected delete counters: %+v", shard.delCount)
	}
	if len(shard.open) != 1 || shard.open[openID] != openPipe {
		t.Fatalf("unexpected compacted open conn map: %+v", shard.open)
	}
	if len(shard.closed) != 1 || shard.closed[closedID] != closedPipe {
		t.Fatalf("unexpected compacted closed conn map: %+v", shard.closed)
	}

	delete(oldOpenMap, openID)
	if len(shard.open) != 1 {
		t.Fatal("expected open conn map to be rebuilt")
	}
	delete(oldClosedMap, closedID)
	if len(shard.closed) != 1 {
		t.Fatal("expected closed conn map to be rebuilt")
	}
}

func TestApiflowMinCaptureSizePatch(t *testing.T) {
	t.Run("unset", func(t *testing.T) {
		t.Setenv(apiflowMinCaptureSizeEnv, "")

		patch, ok := apiflowMinCaptureSizePatch()
		if ok {
			t.Fatalf("unexpected patch: %+v", patch)
		}
	})

	t.Run("valid", func(t *testing.T) {
		t.Setenv(apiflowMinCaptureSizeEnv, "32")

		patch, ok := apiflowMinCaptureSizePatch()
		if !ok {
			t.Fatal("expected patch")
		}
		want := bpfutil.ConstantPatch{Name: "apiflow_min_capture_size", Value: uint64(32)}
		assert.Equal(t, want, patch)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Setenv(apiflowMinCaptureSizeEnv, "-1")

		patch, ok := apiflowMinCaptureSizePatch()
		if ok {
			t.Fatalf("unexpected patch: %+v", patch)
		}
	})
}
