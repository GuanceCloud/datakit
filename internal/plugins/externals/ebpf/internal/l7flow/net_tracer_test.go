//go:build linux
// +build linux

package l7flow

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/protodec"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/procwatch"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/spanid"
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

type cleanupCountingAgg struct {
	proto       protodec.L7Protocol
	exportCalls int32
	cleanupCall int32
	onExport    func()
}

func (a *cleanupCountingAgg) Proto() protodec.L7Protocol {
	return a.proto
}

func (a *cleanupCountingAgg) Obs(_ *comm.ConnectionInfo, _ *protodec.ProtoData) {}

func (a *cleanupCountingAgg) Export(_ map[string]string, _ *cli.K8sInfo) []*point.Point {
	atomic.AddInt32(&a.exportCalls, 1)
	if a.onExport != nil {
		a.onExport()
	}
	return nil
}

func (a *cleanupCountingAgg) Cleanup() {
	atomic.AddInt32(&a.cleanupCall, 1)
}

func (a *cleanupCountingAgg) Len() int {
	return 0
}

func TestTracerStartDoesNotCleanupAfterExport(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	agg := &cleanupCountingAgg{
		proto:    protodec.ProtoHTTP,
		onExport: cancel,
	}
	tracer := &Tracer{
		aggPool: map[protodec.L7Protocol]protodec.AggPool{
			protodec.ProtoHTTP: agg,
		},
	}

	done := make(chan struct{})
	go func() {
		tracer.Start(ctx, time.Millisecond)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("tracer start did not stop after context cancellation")
	}
	if got := atomic.LoadInt32(&agg.exportCalls); got == 0 {
		t.Fatal("expected aggregator export to run")
	}
	if got := atomic.LoadInt32(&agg.cleanupCall); got != 0 {
		t.Fatalf("unexpected cleanup call after export: %d", got)
	}
}

func TestTracerStartHandlesNonPositiveInterval(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tracer := &Tracer{
		aggPool: map[protodec.L7Protocol]protodec.AggPool{},
	}
	done := make(chan struct{})
	go func() {
		tracer.Start(ctx, 0)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("tracer start did not stop with canceled context")
	}
}

func TestConnWatcherProcessTaskDropsInvalidWatcher(t *testing.T) {
	watcher := &ConnWatcher{}

	done := make(chan struct{})
	go func() {
		watcher.processTask(connWatcherTask{
			netdata: getNetwrkData(0),
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("processTask blocked for invalid watcher")
	}
}

func TestGenPtsSkipsNilInputs(t *testing.T) {
	if got := genPts([]*protodec.ProtoData{{}}, nil); len(got) != 0 {
		t.Fatalf("genPts with nil conn returned %d points, want 0", len(got))
	}

	conn := &comm.ConnectionInfo{
		Saddr: [4]uint32{0, 0, 0, 0x0100000A},
		Daddr: [4]uint32{0, 0, 0, 0x0200000A},
		Sport: 12345,
		Dport: 80,
		Meta:  0,
	}
	if got := genPts([]*protodec.ProtoData{
		nil,
		{
			KVs:       point.KVs{},
			Direction: comm.DOut,
			Time:      time.Now().UnixNano(),
			Duration:  time.Millisecond.Nanoseconds(),
		},
	}, conn); len(got) != 1 {
		t.Fatalf("genPts returned %d points, want 1", len(got))
	}
}

func TestGenPtsUsesNumericFieldsForInnerID(t *testing.T) {
	const (
		pid = int32(1234)
		tid = int32(5678)
	)

	conn := &comm.ConnectionInfo{
		Saddr: [4]uint32{0, 0, 0, 0x0100000A},
		Daddr: [4]uint32{0, 0, 0, 0x0200000A},
		Sport: 12345,
		Dport: 80,
		Pid:   uint32(pid),
		Meta:  0,
	}
	pts := genPts([]*protodec.ProtoData{
		{
			KVs:       point.KVs{},
			Direction: comm.DOut,
			Time:      time.Now().UnixNano(),
			Duration:  time.Millisecond.Nanoseconds(),
			KTime:     uint64(200),
			Meta: protodec.ProtoMeta{
				Threads: [2][2]int32{{tid, 0}},
			},
		},
	}, conn)
	if len(pts) != 1 {
		t.Fatalf("genPts returned %d points, want 1", len(pts))
	}
	if got := pts[0].Get(comm.FieldPid); got != int64(pid) {
		t.Fatalf("pid field = %#v, want int64(%d)", got, pid)
	}
	if got := pts[0].Get(comm.FieldKernelThread); got != int64(tid) {
		t.Fatalf("kernel thread field = %#v, want int64(%d)", got, tid)
	}

	var threads comm.ThreadTrace
	want := threads.Insert(comm.DIn, pid, [2]int32{tid, 0}, 100)
	setInnerID(pts[0], &threads)
	if got := pts[0].Get(spanid.ThrTraceID); got != want {
		t.Fatalf("inner trace id = %#v, want %d", got, want)
	}
}

func resolveProcessIntoCatalog(t *testing.T, catalog *procwatch.Catalog, pid int) {
	t.Helper()

	if _, _, err := catalog.Resolve(pid); err != nil {
		t.Fatalf("resolve process %d: %v", pid, err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if _, ok := catalog.Lookup(pid); ok {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("process %d was not visible in catalog cache", pid)
}

func TestPopulateProcessInfoQueuesResolveOnCacheMiss(t *testing.T) {
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
	if conn.ProcessName != "" || conn.ServiceName != "" {
		t.Fatalf("expected cache miss to avoid synchronous process resolution, got %+v", conn)
	}
}

func TestPopulateProcessInfoUsesCachedProcessInfo(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	const pid = 4321
	writeFakeProc(t, procRoot, pid, "python3", 1, []string{"python3", "/opt/app/server.py"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracer := &Tracer{
		catalog: procwatch.NewCatalog(ctx, procwatch.WithSelfPID(-1)),
	}
	resolveProcessIntoCatalog(t, tracer.catalog, pid)

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
	resolveProcessIntoCatalog(t, tracer.catalog, pid)

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

func TestTracePointBufferSize(t *testing.T) {
	t.Run("unset", func(t *testing.T) {
		t.Setenv(tracePointBufferSizeEnv, "")

		assert.Equal(t, defaultTracePointBufferSize, tracePointBufferSize())
	})

	t.Run("valid", func(t *testing.T) {
		t.Setenv(tracePointBufferSizeEnv, "3")

		assert.Equal(t, 3, tracePointBufferSize())
	})

	t.Run("invalid", func(t *testing.T) {
		t.Setenv(tracePointBufferSizeEnv, "0")

		assert.Equal(t, defaultTracePointBufferSize, tracePointBufferSize())
	})

	t.Run("clamped", func(t *testing.T) {
		t.Setenv(tracePointBufferSizeEnv, strconv.Itoa(maxTracePointBufferSize+1))

		assert.Equal(t, maxTracePointBufferSize, tracePointBufferSize())
	})
}

func TestConnWatcherQueueSizeIsClamped(t *testing.T) {
	t.Setenv(connWatcherQueueSizeEnv, strconv.Itoa(maxConnWatcherQueueSize+1))

	assert.Equal(t, maxConnWatcherQueueSize, connWatcherQueueSize())
}

func TestConnWatcherAppendTracePointsCapsBuffer(t *testing.T) {
	watcher := &ConnWatcher{
		trace:           &NetTrace{},
		tracePointLimit: 2,
	}

	watcher.appendTracePoints([]*point.Point{nil, nil, nil})
	if got := len(watcher.trace.ptsCur); got != 2 {
		t.Fatalf("expected trace point buffer to be capped at 2, got %d", got)
	}

	watcher.appendTracePoints([]*point.Point{nil})
	if got := len(watcher.trace.ptsCur); got != 2 {
		t.Fatalf("expected full trace point buffer to remain capped, got %d", got)
	}
}

func TestConnWatcherHandleDropsWhenAsyncQueueFull(t *testing.T) {
	queue := make(chan connWatcherTask, 1)
	queue <- connWatcherTask{}
	watcher := &ConnWatcher{
		eventQueues: []chan connWatcherTask{queue},
	}

	data := getNetwrkData(16)
	data.Payload = append(data.Payload, []byte("0123456789abcdef")...)
	watcher.handle(time.Now().UnixNano(), CUniID{}, data)

	if len(data.Payload) != 0 {
		t.Fatal("expected dropped network data to be reset before returning to the pool")
	}
	if got := len(queue); got != 1 {
		t.Fatalf("expected full queue length to remain unchanged, got %d", got)
	}
}

func TestConnWatcherHandleDropsAfterStop(t *testing.T) {
	stopCh := make(chan struct{})
	close(stopCh)
	queue := make(chan connWatcherTask, 1)
	watcher := &ConnWatcher{
		eventQueues: []chan connWatcherTask{queue},
		stopCh:      stopCh,
	}

	data := getNetwrkData(16)
	data.Payload = append(data.Payload, []byte("0123456789abcdef")...)
	data.SockPtr = 1234
	watcher.handle(time.Now().UnixNano(), CUniID{}, data)

	if len(data.Payload) != 0 || data.SockPtr != 0 {
		t.Fatal("expected stopped watcher to reset dropped network data before returning it to the pool")
	}
	if got := len(queue); got != 0 {
		t.Fatalf("expected stopped watcher to avoid enqueueing work, got queue length %d", got)
	}
}

func TestDrainConnWatcherQueueReturnsNetworkData(t *testing.T) {
	queue := make(chan connWatcherTask, 2)
	data := getNetwrkData(16)
	data.Payload = append(data.Payload, []byte("0123456789abcdef")...)
	data.SockPtr = 1234
	queue <- connWatcherTask{netdata: data}

	drainConnWatcherQueue(queue)

	if got := len(queue); got != 0 {
		t.Fatalf("expected queue to be drained, got %d items", got)
	}
	if len(data.Payload) != 0 || data.SockPtr != 0 {
		t.Fatal("expected drained network data to be reset before returning to the pool")
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

func TestProtoKernelFilterSetFnNonBlocking(t *testing.T) {
	filter := &protoKernelFilter{
		fn: make(chan func(uint64), 1),
	}
	filter.setFn(func(uint64) {
		t.Fatal("stale function should be replaced")
	})

	var got uint64
	filter.setFn(func(v uint64) {
		got = v
	})

	select {
	case fn := <-filter.fn:
		fn(42)
	default:
		t.Fatal("expected pending filter function")
	}
	if got != 42 {
		t.Fatalf("unexpected function value %d", got)
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

func TestStreamHandleDropsNewPipeAtConnMapLimit(t *testing.T) {
	trace := &NetTrace{connShardLimit: 1}
	uniID := CUniID{}
	data := getNetwrkData(16)
	data.Payload = append(data.Payload, []byte("first")...)

	trace.StreamHandle(time.Now().UnixNano(), uniID, data)

	shard := trace.shardFor(uniID)
	if got := len(shard.open); got != 1 {
		t.Fatalf("open conn map entries = %d, want 1", got)
	}

	otherID := sameConnShardTestID(uniID)
	dropped := getNetwrkData(16)
	dropped.Payload = append(dropped.Payload, []byte("second")...)
	dropped.SockPtr = 1234

	trace.StreamHandle(time.Now().UnixNano(), otherID, dropped)

	if got := len(shard.open); got != 1 {
		t.Fatalf("open conn map entries after limit = %d, want 1", got)
	}
	if len(dropped.Payload) != 0 || dropped.SockPtr != 0 {
		t.Fatal("expected dropped network data to be reset before returning to the pool")
	}

	for _, pipe := range shard.open {
		pipe.releaseQueuedData("test_cleanup")
	}
}

func TestStreamHandleSkipsUnknownClosePipe(t *testing.T) {
	trace := &NetTrace{connShardLimit: 1}
	uniID := CUniID{}
	data := getNetwrkData(16)
	data.Payload = append(data.Payload, []byte("close")...)
	data.SockPtr = 1234
	data.Fn = comm.FnSysClose

	trace.StreamHandle(time.Now().UnixNano(), uniID, data)

	shard := trace.shardFor(uniID)
	if got := len(shard.open) + len(shard.closed); got != 0 {
		t.Fatalf("conn map entries = %d, want 0", got)
	}
	if len(data.Payload) != 0 || data.SockPtr != 0 {
		t.Fatal("expected close event network data to be reset before returning to the pool")
	}
}

func TestConnMapLimitEnv(t *testing.T) {
	t.Setenv(connMapLimitEnv, "bad")
	if got := connMapLimit(); got != defaultConnMapLimit {
		t.Fatalf("invalid conn map limit = %d, want %d", got, defaultConnMapLimit)
	}

	t.Setenv(connMapLimitEnv, "1")
	if got := connMapLimit(); got != connMapShardCount {
		t.Fatalf("small conn map limit = %d, want %d", got, connMapShardCount)
	}

	t.Setenv(connMapLimitEnv, "9999999")
	if got := connMapLimit(); got != maxConnMapLimit {
		t.Fatalf("clamped conn map limit = %d, want %d", got, maxConnMapLimit)
	}
}

func TestEnabledProtoListKeepsStableOrder(t *testing.T) {
	got := enabledProtoList(map[protodec.L7Protocol]struct{}{
		protodec.ProtoRedis: {},
		protodec.ProtoHTTP:  {},
		protodec.ProtoMySQL: {},
	})
	want := []protodec.L7Protocol{protodec.ProtoHTTP, protodec.ProtoMySQL, protodec.ProtoRedis}
	if !assert.Equal(t, want, got) {
		t.Fatalf("unexpected protocol order: %+v", got)
	}

	got = enabledProtoList(nil)
	assert.Equal(t, []protodec.L7Protocol{protodec.ProtoHTTP}, got)
}

func sameConnShardTestID(base CUniID) CUniID {
	target := connMapShardIndex(base)
	for i := uint32(1); ; i++ {
		candidate := CUniID{id: i}
		if connMapShardIndex(candidate) == target {
			return candidate
		}
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

func TestProcessPipeReleasesQueuedDataAfterProtocolGiveup(t *testing.T) {
	queued := getNetwrkData(16)
	queued.Payload = append(queued.Payload, []byte("queued-payload")...)
	queued.SockPtr = 1
	current := getNetwrkData(16)
	current.Payload = append(current.Payload, []byte("current-payload")...)
	current.SockPtr = 2

	pipe := &FlowPipe{
		detecTimes: maxDetec,
		sort: dataQueue{
			li: []*comm.NetwrkData{queued},
		},
	}
	trace := &NetTrace{}

	result, connClose := trace.processPipe(time.Now().UnixNano(), pipe, current)

	if result != nil || connClose {
		t.Fatalf("expected protocol giveup to discard data without a result, got result=%v connClose=%v", result, connClose)
	}
	if len(pipe.sort.li) != 0 {
		t.Fatalf("expected queued data to be released, got %d items", len(pipe.sort.li))
	}
	if len(queued.Payload) != 0 || queued.SockPtr != 0 {
		t.Fatal("expected queued network data to be reset before returning to the pool")
	}
	if len(current.Payload) != 0 || current.SockPtr != 0 {
		t.Fatal("expected current network data to be reset before returning to the pool")
	}
}

func TestFinalizePipeCloseReleasesQueuedData(t *testing.T) {
	data := getNetwrkData(16)
	data.Payload = append(data.Payload, []byte("queued-payload")...)
	data.SockPtr = 1234

	var uniID CUniID
	pipe := &FlowPipe{
		sort: dataQueue{
			li: []*comm.NetwrkData{data},
		},
	}
	trace := &NetTrace{}
	shard := trace.shardFor(uniID)
	shard.ensureMaps()
	shard.open[uniID] = pipe

	trace.finalizePipeClose(uniID, pipe, false)

	if len(shard.open) != 0 {
		t.Fatalf("expected pipe to be removed from open map, got %d entries", len(shard.open))
	}
	if len(pipe.sort.li) != 0 {
		t.Fatalf("expected queued data to be released, got %d items", len(pipe.sort.li))
	}
	if len(data.Payload) != 0 || data.SockPtr != 0 {
		t.Fatal("expected queued network data to be reset before returning to the pool")
	}
}

func TestSweepExpiredConnMapsReleasesQueuedData(t *testing.T) {
	data := getNetwrkData(16)
	data.Payload = append(data.Payload, []byte("queued-payload")...)
	data.SockPtr = 1234

	var uniID CUniID
	pipe := &FlowPipe{
		sort: dataQueue{
			li: []*comm.NetwrkData{data},
		},
	}
	trace := &NetTrace{}
	shard := trace.shardFor(uniID)
	shard.ensureMaps()
	shard.open[uniID] = pipe

	openLen, closedLen := trace.sweepExpiredConnMaps(time.Now().UnixNano())

	if openLen != 0 || closedLen != 0 {
		t.Fatalf("expected expired conn maps to be empty, got open=%d closed=%d", openLen, closedLen)
	}
	if len(pipe.sort.li) != 0 {
		t.Fatalf("expected queued data to be released, got %d items", len(pipe.sort.li))
	}
	if len(data.Payload) != 0 || data.SockPtr != 0 {
		t.Fatal("expected queued network data to be reset before returning to the pool")
	}
}

func TestApiflowMinCaptureSizePatch(t *testing.T) {
	t.Run("unset", func(t *testing.T) {
		t.Setenv(apiflowMinCaptureSizeEnv, "")

		patch, ok := apiflowMinCaptureSizePatch()
		if !ok {
			t.Fatal("expected zero patch")
		}
		want := bpfutil.ConstantPatch{Name: "apiflow_min_capture_size", Value: uint64(0)}
		assert.Equal(t, want, patch)
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
		if !ok {
			t.Fatal("expected zero patch")
		}
		want := bpfutil.ConstantPatch{Name: "apiflow_min_capture_size", Value: uint64(0)}
		assert.Equal(t, want, patch)
	})
}

func TestApiflowMapMaxEntriesOverrideDefaults(t *testing.T) {
	t.Setenv(apiflowArgMapMaxEntriesEnv, "")
	t.Setenv(apiflowSKMapMaxEntriesEnv, "")
	t.Setenv(apiflowProtocolFilterMapMaxEntriesEnv, "")

	got := apiflowMapMaxEntriesOverride()
	for _, name := range []string{
		mapAPIFlowSyscallRWArg,
		mapAPIFlowSyscallRWVArg,
		mapAPIFlowSSLReadArgs,
		mapAPIFlowBIONewSocketArgs,
		mapAPIFlowSSLContextSockFD,
		mapAPIFlowSSLBIOFD,
		mapAPIFlowSSLPIDTgidContext,
		mapAPIFlowSyscallSendfileArg,
	} {
		assert.Equal(t, uint32(defaultApiflowArgMapMaxEntries), got[name], name)
	}
	assert.Equal(t, uint32(defaultApiflowSKMapMaxEntries), got[mapAPIFlowSKInfo])
	assert.Equal(t, uint32(defaultApiflowProtocolFilterEntries), got[mapAPIFlowProtocolFilter])
}

func TestApiflowMapMaxEntriesOverrideEnv(t *testing.T) {
	t.Setenv(apiflowArgMapMaxEntriesEnv, "64")
	t.Setenv(apiflowSKMapMaxEntriesEnv, "2048")
	t.Setenv(apiflowProtocolFilterMapMaxEntriesEnv, "99999999")

	got := apiflowMapMaxEntriesOverride()
	assert.Equal(t, uint32(minApiflowArgMapMaxEntries), got[mapAPIFlowSyscallRWArg])
	assert.Equal(t, uint32(2048), got[mapAPIFlowSKInfo])
	assert.Equal(t, uint32(maxApiflowMapMaxEntries), got[mapAPIFlowProtocolFilter])
}

func TestApiflowMapMaxEntriesFromEnvInvalidFallsBack(t *testing.T) {
	t.Setenv(apiflowArgMapMaxEntriesEnv, "bad")

	got := apiflowMapMaxEntriesFromEnv(
		apiflowArgMapMaxEntriesEnv,
		defaultApiflowArgMapMaxEntries,
		minApiflowArgMapMaxEntries,
		maxApiflowMapMaxEntries,
	)
	assert.Equal(t, uint32(defaultApiflowArgMapMaxEntries), got)
}

func TestApiflowMapMaxEntriesFromEnvHugeValueClamps(t *testing.T) {
	t.Setenv(apiflowArgMapMaxEntriesEnv, "4294967296")

	got := apiflowMapMaxEntriesFromEnv(
		apiflowArgMapMaxEntriesEnv,
		defaultApiflowArgMapMaxEntries,
		minApiflowArgMapMaxEntries,
		maxApiflowMapMaxEntries,
	)
	assert.Equal(t, uint32(maxApiflowMapMaxEntries), got)
}

func TestDefaultApiflowPerfBufferPages(t *testing.T) {
	assert.Equal(t, 1024, defaultApiflowPerfBufferPages(1, 4096))
	assert.Equal(t, 512, defaultApiflowPerfBufferPages(64, 4096))
	assert.Equal(t, 32, defaultApiflowPerfBufferPages(1024, 4096))
	assert.Equal(t, 8, defaultApiflowPerfBufferPages(256, 64*1024))
	assert.Equal(t, 1024, defaultApiflowPerfBufferPages(0, 4096))
}

func TestApiflowPerfRingBufferSize(t *testing.T) {
	t.Setenv(apiflowPerfBufferPagesEnv, "32")

	assert.Equal(t, 32*os.Getpagesize(), apiflowPerfRingBufferSize())
}

func TestApiflowPerfRingBufferPagesIsClampedToMemoryBudget(t *testing.T) {
	t.Setenv(apiflowPerfBufferPagesEnv, "999999")

	assert.Equal(t, 512, apiflowPerfRingBufferPages(64, 4096))
	assert.Equal(t, 32, apiflowPerfRingBufferPages(1024, 4096))
}

func TestApiflowPerfWatermark(t *testing.T) {
	pageSize := os.Getpagesize()

	assert.Equal(t, 0, apiflowPerfWatermark(pageSize))
	assert.Equal(t, pageSize, apiflowPerfWatermark(pageSize*2))
}
