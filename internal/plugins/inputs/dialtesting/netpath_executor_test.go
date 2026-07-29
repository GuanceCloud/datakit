// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	netpathinput "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netpath"
)

func newTestNetPathTask(t *testing.T, raw string) *dt.NetPathTask {
	t.Helper()
	task, err := dt.NewTask(raw, &dt.NetPathTask{})
	require.NoError(t, err)
	netPathTask, ok := task.(*dt.NetPathTask)
	require.True(t, ok)
	return netPathTask
}

func TestNetPathExecutorRunRespectsCancelledContext(t *testing.T) {
	raw := `{
		"external_id":"netpath-cancel",
		"name":"cancel probe",
		"status":"OK",
		"frequency":"1m",
		"post_url":"https://openway.example.com?token=tkn_test",
		"protocol":"icmp",
		"host":"netpath.invalid",
		"advance_options":{"timeout":"1s","max_ttl":3,"e2e_queries":1,"traceroute_queries":1},
		"success_when":[{"e2e_status":[{"op":"eq","target":"reached"}]}],
		"success_when_logic":"and"
	}`
	netPathTask := newTestNetPathTask(t, raw)
	executor := newNetPathExecutor(netPathTask)

	// A pre-canceled context must short-circuit the probe instead of
	// blocking until the ~5m internal deadline. The fix derives runCtx from
	// the caller's ctx so the cancellation propagates to RunDialProbe.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := dt.NetPathProbeConfig{
		Host: "netpath.invalid", Protocol: "icmp",
		Timeout: time.Second, MaxTTL: 3,
		TracerouteQueries: 1, E2EQueries: 1,
	}
	start := time.Now()
	err := executor.Run(ctx, cfg)
	elapsed := time.Since(start)

	// Run never returns an error (it records failures in the result); the
	// important guarantees are prompt return and a failed result.
	assert.NoError(t, err)
	assert.Less(t, elapsed, time.Second)

	_, success := executor.CheckResult()
	assert.False(t, success)
}

func TestNetPathExecutorRunRespectsStopChannel(t *testing.T) {
	raw := `{
		"external_id":"netpath-stop",
		"name":"stop probe",
		"status":"OK",
		"frequency":"1m",
		"post_url":"https://openway.example.com?token=tkn_test",
		"protocol":"icmp",
		"host":"netpath.invalid",
		"advance_options":{"timeout":"1s","max_ttl":3,"e2e_queries":1,"traceroute_queries":1},
		"success_when":[{"e2e_status":[{"op":"eq","target":"reached"}]}],
		"success_when_logic":"and"
	}`
	netPathTask := newTestNetPathTask(t, raw)

	// A closed stop channel (e.g. input terminated while the dialer is
	// blocked inside Run) must cancel the in-flight probe promptly instead
	// of letting it run until its internal deadline.
	stop := make(chan interface{})
	executor := newNetPathExecutor(netPathTask, stop)
	probeStarted := make(chan struct{})
	executor.runProbe = func(
		ctx context.Context,
		_ dt.NetPathProbeConfig,
		_ func(net.IP) error,
	) netpathinput.DialProbeResult {
		close(probeStarted)
		<-ctx.Done()
		return netpathinput.DialProbeResult{
			Tags:   map[string]string{},
			Fields: map[string]interface{}{},
			Err:    ctx.Err(),
		}
	}

	cfg := dt.NetPathProbeConfig{
		Host: "netpath.invalid", Protocol: "icmp",
		Timeout: 30 * time.Second, MaxTTL: 3,
		TracerouteQueries: 1, E2EQueries: 1,
	}
	runDone := make(chan error, 1)
	go func() {
		runDone <- executor.Run(context.Background(), cfg)
	}()

	select {
	case <-probeStarted:
	case <-time.After(time.Second):
		t.Fatal("probe did not start")
	}
	close(stop)

	select {
	case err := <-runDone:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("probe was not canceled by the stop channel")
	}

	_, success := executor.CheckResult()
	assert.False(t, success)
}

func TestNetPathExecutorRunRespectsParentContext(t *testing.T) {
	netPathTask := newTestNetPathTask(t, testNetPathTaskJSON)
	parentCtx, cancelParent := context.WithCancel(context.Background())
	executor := newNetPathExecutor(netPathTask).withContext(parentCtx)
	started := make(chan struct{})
	executor.runProbe = func(
		ctx context.Context,
		_ dt.NetPathProbeConfig,
		_ func(net.IP) error,
	) netpathinput.DialProbeResult {
		close(started)
		<-ctx.Done()
		now := time.Now()
		return netpathinput.DialProbeResult{
			Tags:        map[string]string{},
			Fields:      map[string]interface{}{},
			ScheduledAt: now,
			StartedAt:   now,
			FinishedAt:  now,
			Err:         ctx.Err(),
		}
	}

	runDone := make(chan error, 1)
	go func() {
		runDone <- executor.Run(context.Background(), dt.NetPathProbeConfig{})
	}()
	<-started
	cancelParent()

	select {
	case err := <-runDone:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("probe was not canceled by the parent context")
	}
}

func TestNetPathExecutorAppliesResolvedIPValidator(t *testing.T) {
	netPathTask := newTestNetPathTask(t, testNetPathTaskJSON)
	validationErr := errors.New("private destination")
	executor := newNetPathExecutor(netPathTask).withResolvedIPValidator(func(ip net.IP) error {
		assert.Equal(t, "10.0.0.1", ip.String())
		return validationErr
	})
	executor.runProbe = func(
		_ context.Context,
		_ dt.NetPathProbeConfig,
		validateResolvedIP func(net.IP) error,
	) netpathinput.DialProbeResult {
		require.NotNil(t, validateResolvedIP)
		assert.ErrorIs(t, validateResolvedIP(net.ParseIP("10.0.0.1")), validationErr)
		now := time.Now()
		return netpathinput.DialProbeResult{
			Tags: map[string]string{"e2e_status": "reached"},
			Fields: map[string]interface{}{
				"e2e_rtt_avg": float64(1),
			},
			ScheduledAt: now,
			StartedAt:   now,
			FinishedAt:  now,
		}
	}

	require.NoError(t, executor.Run(context.Background(), dt.NetPathProbeConfig{}))
}

func TestNetPathExecutorConcurrencyWaitIsCancelable(t *testing.T) {
	netPathTask := newTestNetPathTask(t, testNetPathTaskJSON)
	concurrency := make(chan struct{}, 1)
	concurrency <- struct{}{}
	parentCtx, cancelParent := context.WithCancel(context.Background())
	executor := newNetPathExecutor(netPathTask).
		withContext(parentCtx).
		withConcurrency(concurrency)
	executor.runProbe = func(
		context.Context,
		dt.NetPathProbeConfig,
		func(net.IP) error,
	) netpathinput.DialProbeResult {
		t.Fatal("probe must not run without a concurrency slot")
		return netpathinput.DialProbeResult{}
	}

	runDone := make(chan error, 1)
	go func() {
		runDone <- executor.Run(context.Background(), dt.NetPathProbeConfig{})
	}()
	cancelParent()

	select {
	case err := <-runDone:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("concurrency wait did not observe cancellation")
	}
}

func TestNetPathExecutorEvaluateDeterministicOrder(t *testing.T) {
	raw := `{
		"external_id":"netpath-order",
		"name":"order probe",
		"status":"OK",
		"frequency":"1m",
		"post_url":"https://openway.example.com?token=tkn_test",
		"protocol":"icmp",
		"host":"netpath.invalid",
		"advance_options":{"timeout":"1s","max_ttl":3,"e2e_queries":1,"traceroute_queries":1},
		"success_when":[{
			"e2e_status":[{"op":"eq","target":"reached"}],
			"hop_count":[{"op":"leq","target":30}],
			"e2e_rtt_avg":[{"op":"leq","target":"100ms"}]
		}],
		"success_when_logic":"and"
	}`
	netPathTask := newTestNetPathTask(t, raw)
	executor := newNetPathExecutor(netPathTask)

	// All assertions fail against an empty result; the failure list must
	// follow the fixed field declaration order (e2e_rtt_avg before
	// hop_count before e2e_status) on every invocation.
	expected := []string{
		"e2e_rtt_avg is missing",
		"hop_count is missing",
		"e2e_status is missing",
	}
	for i := 0; i < 20; i++ {
		reasons, success := executor.evaluate(netpathinput.DialProbeResult{
			Tags:   map[string]string{},
			Fields: map[string]interface{}{},
		})
		assert.False(t, success)
		assert.Equal(t, expected, reasons, "iteration %d", i)
	}
}

func TestNetPathStatusConditionRejectsNonEqualityOperators(t *testing.T) {
	result := netpathinput.DialProbeResult{
		Tags: map[string]string{
			"e2e_status":        "reached",
			"traceroute_status": "reached",
		},
	}

	for _, field := range []string{"e2e_status", "traceroute_status"} {
		for _, op := range []string{"", "ne", "lt"} {
			t.Run(field+"/"+op, func(t *testing.T) {
				err := checkNetPathCondition(field, &dt.NetPathCondition{
					Op:     op,
					Target: json.RawMessage(`"reached"`),
				}, result)
				assert.EqualError(t, err, field+" only supports eq")
			})
		}
	}
}

func TestNetPathNumberConditionRejectsNullTarget(t *testing.T) {
	err := checkNetPathCondition("hop_count", &dt.NetPathCondition{
		Op:     "eq",
		Target: json.RawMessage(`null`),
	}, netpathinput.DialProbeResult{
		Fields: map[string]interface{}{"hop_count": float64(0)},
	})
	assert.EqualError(t, err, "hop_count: target must be numeric")
}

func TestNetPathLogIdentifiersExcludeAccessKey(t *testing.T) {
	task := newTestNetPathTask(t, `{
		"external_id":"netpath-log",
		"access_key":"ak-secret",
		"status":"OK",
		"frequency":"1m",
		"protocol":"icmp",
		"host":"netpath.invalid",
		"advance_options":{"timeout":"1s"},
		"success_when":[{"e2e_status":[{"op":"eq","target":"reached"}]}]
	}`)

	var logs bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(&logs),
		zap.DebugLevel,
	)
	previousLogger := l.Sugar()
	l.SetSugar(zap.New(core).Sugar())
	t.Cleanup(func() {
		l.SetSugar(previousLogger)
	})

	d := &dialer{task: task, class: dt.ClassNetPath}
	l.Debugf("dialer: %+#v", d.logSafeValue())
	l.Infof("task %s", taskLogID(task))

	assert.Contains(t, logs.String(), "netpath-log")
	assert.NotContains(t, logs.String(), "ak-secret")
}

func TestRedactNetPathTaskError(t *testing.T) {
	payload := `{"access_key":"ak-secret","post_url":"https://openway.example.com?token=tkn_secret","protocol":"icmp"}`
	err := errors.New("json.Unmarshal failed: invalid character 'x', task json: " + payload)

	redacted := redactNetPathTaskError(err, payload)
	assert.NotContains(t, redacted, "ak-secret")
	assert.NotContains(t, redacted, "tkn_secret")
	assert.Contains(t, redacted, "<redacted>")
	assert.Contains(t, redacted, "json.Unmarshal failed")

	// Errors that never embedded the payload pass through unchanged.
	plain := errors.New("port must be between 1 and 65535")
	assert.Equal(t, plain.Error(), redactNetPathTaskError(plain, payload))
	assert.Equal(t, "", redactNetPathTaskError(nil, payload))
}

func TestNetPathSecureValuesAreRedactedFromErrorsAndResults(t *testing.T) {
	const secret = "secure-netpath-sentinel"
	task := newTestNetPathTask(t, `{
		"external_id":"netpath-redact",
		"name":"redact probe",
		"status":"OK",
		"frequency":"1m",
		"protocol":"icmp",
		"host":"netpath.invalid",
		"config_vars":[{"name":"secret","value":"secure-netpath-sentinel","secure":true}],
		"advance_options":{"timeout":"1s"},
		"success_when":[{"e2e_status":[{"op":"eq","target":"reached"}]}]
	}`)
	executor := newNetPathExecutor(task)

	logText := taskErrorForLog(task, errors.New("invalid value "+secret))
	assert.NotContains(t, logText, secret)
	assert.Contains(t, logText, "<redacted>")

	executor.SetError("invalid value " + secret)
	_, fields := executor.GetResults()
	assert.NotContains(t, fields["fail_reason"], secret)
	assert.NotContains(t, fields["message"], secret)
	assert.Contains(t, fields["fail_reason"], "<redacted>")
}

func TestProtectedRunRedactsRenderedNetPathError(t *testing.T) {
	const secret = "secure-rendered-sentinel"
	task := newTestNetPathTask(t, `{
		"external_id":"netpath-render-redact",
		"name":"redact rendered error",
		"status":"OK",
		"frequency":"1m",
		"protocol":"{{secret}}",
		"host":"netpath.invalid",
		"config_vars":[{"name":"secret","value":"secure-rendered-sentinel","secure":true}],
		"advance_options":{"timeout":"1s"},
		"success_when":[{"e2e_status":[{"op":"eq","target":"reached"}]}]
	}`)
	d := newDialer(task, defaultInput())

	var logs bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(&logs),
		zap.DebugLevel,
	)
	previousLogger := l.Sugar()
	l.SetSugar(zap.New(core).Sugar())
	t.Cleanup(func() {
		l.SetSugar(previousLogger)
	})

	protectedRun(d)

	assert.NotContains(t, logs.String(), secret)
	assert.Contains(t, logs.String(), "<redacted>")
	assert.Contains(t, logs.String(), "netpath-render-redact")
}

const testNetPathTaskJSON = `{
	"external_id":"netpath-dispatch",
	"name":"dispatch probe",
	"status":"OK",
	"frequency":"1m",
	"post_url":"https://openway.example.com?token=tkn_test",
	"protocol":"icmp",
	"host":"netpath.invalid",
	"advance_options":{"timeout":"1s","max_ttl":3,"e2e_queries":1,"traceroute_queries":1},
	"success_when":[{"e2e_status":[{"op":"eq","target":"reached"}]}],
	"success_when_logic":"and"
}`

func TestNewTaskRunRejectsUnsupportedNetPathPlatform(t *testing.T) {
	old := netPathPlatformCheck
	defer func() { netPathPlatformCheck = old }()
	netPathPlatformCheck = func(string) error {
		return errors.New("netpath dial testing is unsupported on test-os")
	}

	// The rejection must happen before any dialer is created so unsupported
	// tasks never enter the schedule loop.
	ipt := &Input{}
	d, err := ipt.newTaskRun(newTestNetPathTask(t, testNetPathTaskJSON))
	assert.Nil(t, d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestDoUpdateTaskRejectsUnsupportedRenderedNetPathPlatform(t *testing.T) {
	old := netPathPlatformCheck
	defer func() { netPathPlatformCheck = old }()
	netPathPlatformCheck = func(protocol string) error {
		if strings.EqualFold(protocol, "tcp") {
			return errors.New("unsupported netpath protocol " + protocol + " on darwin")
		}
		return nil
	}

	current := newTestNetPathTask(t, testNetPathTaskJSON)
	updated := newTestNetPathTask(t, strings.Replace(testNetPathTaskJSON,
		`"protocol":"icmp"`,
		`"protocol":"{{protocol}}","port":"443","config_vars":[{"name":"protocol","value":"tcp","secure":true}]`,
		1))
	d := newDialer(current, defaultInput())

	err := d.doUpdateTask(updated)
	require.Error(t, err)
	assert.ErrorIs(t, err, errNetPathPlatformUpdate)
	assert.Contains(t, err.Error(), "unsupported")
	assert.Equal(t, "tcp", updated.Protocol)
	assert.NotContains(t, taskErrorForLog(updated, err), "tcp")
	assert.Contains(t, taskErrorForLog(updated, err), "<redacted>")
	assert.Same(t, current, d.task)
}

func TestNetPathPlatformCheckWrapper(t *testing.T) {
	// Templated protocols always defer to probe-time validation, even on
	// platforms where NETPATH is unsupported.
	assert.NoError(t, netPathPlatformCheck("{{protocol}}"))

	// Windows is not a supported NETPATH platform; the real check must
	// reject tasks there (this machine included).
	if runtime.GOOS == "windows" {
		assert.Error(t, netPathPlatformCheck("icmp"))
	}
}

func TestDialerRunTaskNetPathConcurrency(t *testing.T) {
	t.Run("netpath tasks respect max concurrency", func(t *testing.T) {
		ipt := defaultInput()
		ipt.netPathConcurrency = make(chan struct{}, 1)

		var (
			mu     sync.Mutex
			active int
			max    int
		)
		runFn := func() error {
			mu.Lock()
			active++
			if active > max {
				max = active
			}
			mu.Unlock()

			time.Sleep(30 * time.Millisecond)

			mu.Lock()
			active--
			mu.Unlock()
			return nil
		}

		d1 := &dialer{
			task:   &runTaskStub{class: dt.ClassNetPath, runFn: runFn},
			ipt:    ipt,
			done:   ipt.semStop.Wait(),
			stopCh: make(chan interface{}),
		}
		d2 := &dialer{
			task:   &runTaskStub{class: dt.ClassNetPath, runFn: runFn},
			ipt:    ipt,
			done:   ipt.semStop.Wait(),
			stopCh: make(chan interface{}),
		}

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			assert.NoError(t, d1.runTask())
		}()
		go func() {
			defer wg.Done()
			assert.NoError(t, d2.runTask())
		}()
		wg.Wait()

		assert.Equal(t, 1, max)
	})

	t.Run("netpath task stopped while waiting is skipped", func(t *testing.T) {
		ipt := defaultInput()
		ipt.netPathConcurrency = make(chan struct{}, 1)
		ipt.netPathConcurrency <- struct{}{}

		ran := false
		d := &dialer{
			task: &runTaskStub{
				class: dt.ClassNetPath,
				runFn: func() error {
					ran = true
					return nil
				},
			},
			ipt:    ipt,
			done:   ipt.semStop.Wait(),
			stopCh: make(chan interface{}),
		}

		errCh := make(chan error, 1)
		go func() {
			errCh <- d.runTask()
		}()

		close(d.stopCh)

		err := <-errCh
		assert.ErrorIs(t, err, errTaskRunSkipped)
		assert.False(t, ran)
	})

	t.Run("other tasks ignore the netpath limiter", func(t *testing.T) {
		ipt := defaultInput()
		ipt.netPathConcurrency = make(chan struct{}, 1)
		ipt.netPathConcurrency <- struct{}{}

		ran := false
		d := &dialer{
			task: &runTaskStub{
				class: dt.ClassHTTP,
				runFn: func() error {
					ran = true
					return nil
				},
			},
			ipt:    ipt,
			done:   ipt.semStop.Wait(),
			stopCh: make(chan interface{}),
		}

		assert.NoError(t, d.runTask())
		assert.True(t, ran)
	})
}

func TestDialerNetPathStopCancelsActiveAndFutureProbe(t *testing.T) {
	newProbe := func(t *testing.T) (*dt.NetPathTask, *dialer, *netPathExecutor, chan struct{}) {
		t.Helper()
		task := newTestNetPathTask(t, testNetPathTaskJSON)
		require.NoError(t, task.RenderTemplateAndInit(nil))
		d := newDialer(task, defaultInput())
		executor := newNetPathExecutor(task, d.stopCh)
		started := make(chan struct{})
		executor.runProbe = func(
			ctx context.Context,
			_ dt.NetPathProbeConfig,
			_ func(net.IP) error,
		) netpathinput.DialProbeResult {
			close(started)
			<-ctx.Done()
			return netpathinput.DialProbeResult{
				Tags:   map[string]string{},
				Fields: map[string]interface{}{},
				Err:    ctx.Err(),
			}
		}
		task.SetExecutor(executor)
		return task, d, executor, started
	}

	t.Run("status stop cancels active probe", func(t *testing.T) {
		_, d, _, started := newProbe(t)
		runDone := make(chan error, 1)
		go func() {
			runDone <- d.runTask()
		}()
		<-started

		require.NoError(t, d.updateTask(&runTaskStub{class: dt.ClassNetPath, status: dt.StatusStop}))
		select {
		case err := <-runDone:
			require.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("active probe was not canceled")
		}
		assert.NotPanics(t, d.exit)
	})

	t.Run("stop before task run remains observable", func(t *testing.T) {
		task, d, _, started := newProbe(t)
		d.stop()

		runDone := make(chan error, 1)
		go func() {
			runDone <- task.Run()
		}()
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("probe did not observe the persistent stop signal")
		}
		select {
		case err := <-runDone:
			require.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("probe started after stop but was not canceled")
		}
	})
}
