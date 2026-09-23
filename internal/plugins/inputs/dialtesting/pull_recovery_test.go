// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime/pprof"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/GuanceCloud/cliutils"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestDispatchTasksNullResponse(t *testing.T) {
	for _, content := range []string{
		`{"HTTP":null}`,
		`{"HTTP":[null]}`,
		`{"region":{"name_i18n":null,"extra":null},"variables":null}`,
	} {
		t.Run(content, func(t *testing.T) {
			ipt := defaultInput()
			require.NotPanics(t, func() { require.NoError(t, ipt.dispatchTasks([]byte(`{"content":`+content+`}`))) })
		})
	}
	t.Run("valid task after null element", func(t *testing.T) {
		ipt := defaultInput()
		task := `{"external_id":"valid-after-null","owner_external_id":"workspace","status":"stop","update_time":123,"frequency":"1m","post_url":"http://localhost?token=test","url":"http://localhost","method":"GET"}`
		payload, err := json.Marshal(map[string]any{"content": map[string]any{"HTTP": []any{nil, task}}})
		require.NoError(t, err)
		require.NotPanics(t, func() { require.NoError(t, ipt.dispatchTasks(payload)) })
		require.EqualValues(t, 123, ipt.pos, "the valid task after null must still be processed")
	})
}

func TestServerTaskRecovery(t *testing.T) {
	// Isolate process-wide Exit and goroutine snapshots from legacy input tests.
	if os.Getenv("DATAKIT_TEST_PULL_RECOVERY") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestServerTaskRecovery$", "-test.timeout=30s")
		cmd.Env = append(os.Environ(), "DATAKIT_TEST_PULL_RECOVERY=1")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "%s", output)
		return
	}
	t.Run("recovers beyond seven panics with one variable worker", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Server = "http://dialtesting.test"
		ipt.RegionID = "pull-recovery"
		interval := 10 * time.Millisecond
		attempts := make([]time.Time, 0, 9)
		reachedSuccess := make(chan struct{})
		releaseSuccess := make(chan struct{})
		release := sync.OnceFunc(func() { close(releaseSuccess) })
		ipt.cli = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			attempts = append(attempts, time.Now())
			if len(attempts) <= 8 {
				if len(attempts)%2 == 0 {
					panic("injected string panic")
				}
				panic(errors.New("injected error panic"))
			}
			close(reachedSuccess)
			<-releaseSuccess
			ipt.semStop.Close()
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"content":{"region":{"name":"recovered"}}}`))}, nil
		})}
		done := make(chan struct{})
		go func() { defer close(done); ipt.runServerTasks(interval) }()
		defer func() {
			ipt.semStop.Close()
			release()
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Error("pull loop did not stop")
			}
		}()
		select {
		case <-reachedSuccess:
		case <-time.After(3 * time.Second):
			t.Fatal("pulling did not recover after eight panics")
		}
		var stacks bytes.Buffer
		require.NoError(t, pprof.Lookup("goroutine").WriteTo(&stacks, 2))
		require.Equal(t, 1, strings.Count(stacks.String(), "(*Variable).run."), "only one variable worker should run")
		for i := 1; i < len(attempts); i++ {
			require.GreaterOrEqual(t, attempts[i].Sub(attempts[i-1]), interval)
		}
		var metric dto.Metric
		require.NoError(t, taskPullPanicCounter.WithLabelValues(ipt.RegionID).Write(&metric))
		require.EqualValues(t, 8, metric.GetCounter().GetValue())
		release()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("pull loop did not exit after successful dispatch")
		}
		require.Equal(t, "recovered", ipt.regionName)
	})
	for _, signal := range []string{"input_stop", "global_exit"} {
		t.Run("stop during retry wait/"+signal, func(t *testing.T) {
			ipt := defaultInput()
			ipt.Server = "http://dialtesting.test"
			ipt.RegionID = "stop-" + signal
			attempted := make(chan struct{})
			ipt.cli = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { close(attempted); panic("injected panic") })}
			done := make(chan struct{})
			go func() { defer close(done); ipt.runServerTasks(time.Hour) }()
			defer ipt.semStop.Close()
			select {
			case <-attempted:
			case <-time.After(time.Second):
				t.Fatal("pull did not start")
			}
			if signal == "input_stop" {
				ipt.semStop.Close()
			} else {
				datakit.Exit.Close()
			}
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("shutdown waited for retry interval")
			}
			require.Eventually(t, func() bool {
				var stacks bytes.Buffer
				if err := pprof.Lookup("goroutine").WriteTo(&stacks, 2); err != nil {
					return false
				}
				return !strings.Contains(stacks.String(), "(*Variable).run.")
			}, time.Second, time.Millisecond)
		})
	}
}

func TestServerTaskPollingCadence(t *testing.T) {
	// Exit must be created inside the fake-clock bubble; isolate the global
	// replacement from other tests and their background goroutines.
	if os.Getenv("DATAKIT_TEST_POLL_CADENCE") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestServerTaskPollingCadence$", "-test.timeout=30s")
		cmd.Env = append(os.Environ(), "DATAKIT_TEST_POLL_CADENCE=1")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "%s", output)
		return
	}
	synctest.Test(t, func(t *testing.T) {
		datakit.Exit = cliutils.NewSem()
		const interval = time.Minute
		for _, tc := range []struct {
			name       string
			processing time.Duration
			panicFirst bool
			wantGap    time.Duration
		}{
			{name: "normal response", processing: 20 * time.Second, wantGap: interval},
			{name: "slow response consumes pending tick", processing: 90 * time.Second, wantGap: 90 * time.Second},
			{name: "panic discards pending tick", processing: 90 * time.Second, panicFirst: true, wantGap: 90*time.Second + interval},
		} {
			func() {
				ipt := defaultInput()
				ipt.Server = "http://dialtesting.test"
				ipt.RegionID = tc.name
				var attempts []time.Time
				ipt.cli = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
					attempts = append(attempts, time.Now())
					if len(attempts) == 1 {
						time.Sleep(tc.processing)
						if tc.panicFirst {
							panic("injected slow-round panic")
						}
					} else {
						ipt.semStop.Close()
					}
					return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"content":{}}`))}, nil
				})}
				defer ipt.semStop.Close()
				ipt.runServerTasks(interval)
				require.Len(t, attempts, 2)
				require.Equal(t, tc.wantGap, attempts[1].Sub(attempts[0]), tc.name)
			}()
		}
	})
}
