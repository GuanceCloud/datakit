//go:build linux
// +build linux

package l7flow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
)

func TestPerfLostWarningLimiterAggregatesWithinWindow(t *testing.T) {
	now := time.Unix(1700000000, 0)
	limiter := newPerfLostWarningLimiter(func() time.Time { return now })

	msg := limiter.format(3, 10)
	if msg != "lost 10 events on cpu 3" {
		t.Fatalf("unexpected first warning: %q", msg)
	}

	if msg = limiter.format(5, 20); msg != "" {
		t.Fatalf("expected warning suppression within window, got %q", msg)
	}

	now = now.Add(apiflowPerfLostLogInterval + time.Second)
	msg = limiter.format(7, 30)
	want := "lost 30 events on cpu 7 (aggregated 20 additional lost events over 10s)"
	if msg != want {
		t.Fatalf("unexpected aggregated warning: %q != %q", msg, want)
	}
}

func TestPerfLostWarningLimiterSkipsZeroCount(t *testing.T) {
	limiter := newPerfLostWarningLimiter(time.Now)
	if msg := limiter.format(1, 0); msg != "" {
		t.Fatalf("expected zero-count updates to stay silent, got %q", msg)
	}
}

func TestAPIFlowTracerRunRejectsUninitializedTracer(t *testing.T) {
	var tracer *APIFlowTracer
	if err := tracer.Run(context.Background(), nil, nil, false, time.Second); err == nil {
		t.Fatal("expected uninitialized tracer error")
	}
}

func TestAPIFlowTracerStopCancelsInternalContext(t *testing.T) {
	tracer := NewAPIFlowTracer(context.Background())
	defer tracer.stop()

	tracer.stop()
	select {
	case <-tracer.tracer.connWatcher.stopCh:
	case <-time.After(time.Second):
		t.Fatal("internal tracer context was not canceled")
	}
}

func TestFilterUnavailableAPIFlowKernelProbesDropsMissingKernelProbe(t *testing.T) {
	probes := []*bpfutil.HookSpec{
		{ID: bpfutil.HookID{Program: apiflowSchedGetAffinityProgram, UID: apiflowSchedGetAffinityUID}},
		{ID: bpfutil.HookID{Program: "tracepoint__sys_enter_read"}},
		{ID: bpfutil.HookID{Program: "kprobe__tcp_close", UID: "tcp_close_apiflow"}},
	}

	got := filterUnavailableAPIFlowKernelProbesWithLookup(probes, func(symbol string) (bool, error) {
		return symbol == "tcp_close", nil
	})

	if hasProbeProgram(got, apiflowSchedGetAffinityProgram) {
		t.Fatal("expected missing sched_getaffinity probe to be filtered out")
	}
	if !hasProbeProgram(got, "tracepoint__sys_enter_read") {
		t.Fatal("expected tracepoint probe to be kept")
	}
	if !hasProbeProgram(got, "kprobe__tcp_close") {
		t.Fatal("expected available tcp_close probe to be kept")
	}
}

func TestFilterUnavailableAPIFlowKernelProbesKeepsProbeOnLookupError(t *testing.T) {
	probes := []*bpfutil.HookSpec{
		{ID: bpfutil.HookID{Program: apiflowSchedGetAffinityProgram, UID: apiflowSchedGetAffinityUID}},
	}

	got := filterUnavailableAPIFlowKernelProbesWithLookup(probes, func(string) (bool, error) {
		return false, errors.New("kallsyms unavailable")
	})

	if !hasProbeProgram(got, apiflowSchedGetAffinityProgram) {
		t.Fatal("expected probe to be kept when symbol detection is unavailable")
	}
}

func TestFilterUnavailableAPIFlowTracepointsDropsMissingEvent(t *testing.T) {
	probes := []*bpfutil.HookSpec{
		{ID: bpfutil.HookID{Program: "tracepoint__sys_enter_read"}},
		{ID: bpfutil.HookID{Program: "tracepoint__sys_exit_write"}},
		{ID: bpfutil.HookID{Program: "kprobe__tcp_close", UID: "tcp_close_apiflow"}},
	}

	got := filterUnavailableAPIFlowTracepointsWithLookup(probes, func(group, event string) (bool, error) {
		return group == "syscalls" && event == "sys_enter_read", nil
	})

	if !hasProbeProgram(got, "tracepoint__sys_enter_read") {
		t.Fatal("expected available tracepoint to be kept")
	}
	if hasProbeProgram(got, "tracepoint__sys_exit_write") {
		t.Fatal("expected missing tracepoint to be filtered out")
	}
	if !hasProbeProgram(got, "kprobe__tcp_close") {
		t.Fatal("expected non-tracepoint probe to be kept")
	}
}

func TestFilterUnavailableAPIFlowTracepointsKeepsProbeOnLookupError(t *testing.T) {
	probes := []*bpfutil.HookSpec{
		{ID: bpfutil.HookID{Program: "tracepoint__sys_enter_read"}},
	}

	got := filterUnavailableAPIFlowTracepointsWithLookup(probes, func(string, string) (bool, error) {
		return false, errors.New("tracefs unavailable")
	})

	if !hasProbeProgram(got, "tracepoint__sys_enter_read") {
		t.Fatal("expected tracepoint probe to be kept when tracefs detection is unavailable")
	}
}

func TestTracepointEventExistsInRoots(t *testing.T) {
	t.Run("missing root is unknown", func(t *testing.T) {
		got, err := tracepointEventExistsInRoots(
			[]string{filepath.Join(t.TempDir(), "missing")},
			"syscalls",
			"sys_enter_read",
		)
		if err == nil {
			t.Fatal("expected missing tracepoint root to be reported as unknown")
		}
		if got {
			t.Fatal("missing tracepoint root should not report event as found")
		}
	})

	t.Run("missing event is absent when root exists", func(t *testing.T) {
		got, err := tracepointEventExistsInRoots(
			[]string{t.TempDir()},
			"syscalls",
			"sys_enter_read",
		)
		if err != nil {
			t.Fatal(err)
		}
		if got {
			t.Fatal("missing tracepoint event should not be reported as found")
		}
	})

	t.Run("existing event is found", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, "syscalls", "sys_enter_read"), 0o755); err != nil {
			t.Fatal(err)
		}
		got, err := tracepointEventExistsInRoots(
			[]string{root},
			"syscalls",
			"sys_enter_read",
		)
		if err != nil {
			t.Fatal(err)
		}
		if !got {
			t.Fatal("expected tracepoint event to be found")
		}
	})
}

func TestHasAPIFlowSyscallTracepoint(t *testing.T) {
	if hasAPIFlowSyscallTracepoint([]*bpfutil.HookSpec{
		{ID: bpfutil.HookID{Program: "kprobe__tcp_close"}},
		{ID: apiflowKpFlushHookID},
	}) {
		t.Fatal("did not expect kprobes alone to satisfy apiflow syscall capture")
	}

	if !hasAPIFlowSyscallTracepoint([]*bpfutil.HookSpec{
		{ID: bpfutil.HookID{Program: "kprobe__tcp_close"}},
		{ID: bpfutil.HookID{Program: "tracepoint__sys_enter_read"}},
	}) {
		t.Fatal("expected syscall tracepoint to satisfy apiflow capture")
	}
}

func hasProbeProgram(probes []*bpfutil.HookSpec, program string) bool {
	for _, probe := range probes {
		if probe != nil && probe.ID.Program == program {
			return true
		}
	}
	return false
}
