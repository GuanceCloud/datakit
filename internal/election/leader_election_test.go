// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type fakeElectionClock struct {
	now time.Time
}

func (c *fakeElectionClock) Now() time.Time { return c.now }

func (c *fakeElectionClock) Advance(duration time.Duration) { c.now = c.now.Add(duration) }

type lifecycleInput struct {
	pauseCalls  int
	resumeCalls int
	onPause     func()
	onResume    func()
}

type datawayRegressionPuller struct {
	campaignBody []byte
	heartbeatErr error
}

func (p *datawayRegressionPuller) Election(_, _ string, _ io.Reader) ([]byte, error) {
	return p.campaignBody, nil
}

func (p *datawayRegressionPuller) ElectionHeartbeat(_, _ string, _ io.Reader) ([]byte, error) {
	return nil, p.heartbeatErr
}

func (*lifecycleInput) ElectionEnabled() bool { return true }

func (i *lifecycleInput) Pause() error {
	i.pauseCalls++
	if i.onPause != nil {
		i.onPause()
	}
	return nil
}

func (i *lifecycleInput) Resume() error {
	i.resumeCalls++
	if i.onResume != nil {
		i.onResume()
	}
	return nil
}

func newOperatorLeaderElection(t *testing.T, serverURL string, clock Clock, plugin inputs.ElectionInput) *leaderElection {
	t.Helper()
	return newOperatorLeaderElectionForID(t, serverURL, clock, "datakit-a", plugin)
}

func newOperatorLeaderElectionForID(
	t *testing.T,
	serverURL string,
	clock Clock,
	id string,
	plugin inputs.ElectionInput,
) *leaderElection {
	t.Helper()
	preserveCurrentElected(t)
	puller, err := NewOperatorPuller(serverURL, "tkn-secret", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("NewOperatorPuller() error = %v", err)
	}
	return newLeaderElection(&option{
		id:        id,
		namespace: "production",
		provider:  ProviderOperator,
		puller:    puller,
		clock:     clock,
	}, map[string][]inputs.ElectionInput{"container": {plugin}})
}

func writeElectionResponse(t *testing.T, w http.ResponseWriter, status, holder string, epoch int64) {
	t.Helper()
	writeElectionResponseForID(t, w, status, "datakit-a", holder, epoch)
}

func writeElectionResponseForID(t *testing.T, w http.ResponseWriter, status, id, holder string, epoch int64) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"content": map[string]any{
			"status":         status,
			"namespace":      "production",
			"id":             id,
			"incumbency_id":  holder,
			"interval":       3,
			"lease_duration": 30,
			"epoch":          epoch,
		},
	}); err != nil {
		t.Errorf("encode response: %v", err)
	}
}

func preserveCurrentElected(t *testing.T) {
	t.Helper()
	previous := CurrentElected
	t.Cleanup(func() { CurrentElected = previous })
}

func TestOperatorLeaderExpiresLocallyWithoutFallback(t *testing.T) {
	var mu sync.Mutex
	campaignCalls := 0
	heartbeatCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch req.URL.Path {
		case "/v1/dk-election":
			campaignCalls++
			if campaignCalls == 2 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			writeElectionResponse(t, w, StatusSuccess.String(), "datakit-a", int64(campaignCalls))
		case "/v1/dk-election/heartbeat":
			heartbeatCalls++
			if heartbeatCalls > 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			writeElectionResponse(t, w, StatusSuccess.String(), "datakit-a", 1)
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(server.Close)

	clock := &fakeElectionClock{now: time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)}
	plugin := &lifecycleInput{}
	election := newOperatorLeaderElection(t, server.URL, clock, plugin)
	election.pausePlugins()

	if _, err := election.runOnce(); err != nil {
		t.Fatalf("initial campaign error = %v", err)
	}
	if election.status != StatusSuccess || plugin.resumeCalls != 1 || plugin.pauseCalls != 1 {
		t.Fatalf("after campaign: status=%s pauses=%d resumes=%d", election.status, plugin.pauseCalls, plugin.resumeCalls)
	}

	clock.Advance(3 * time.Second)
	if _, err := election.runOnce(); err != nil {
		t.Fatalf("heartbeat error = %v", err)
	}
	if plugin.resumeCalls != 1 || plugin.pauseCalls != 1 {
		t.Fatalf("duplicate heartbeat changed lifecycle: pauses=%d resumes=%d", plugin.pauseCalls, plugin.resumeCalls)
	}

	clock.Advance(20 * time.Second)
	if _, err := election.runOnce(); err == nil {
		t.Fatal("expected temporary Operator failure")
	}
	if election.status != StatusSuccess || plugin.pauseCalls != 1 {
		t.Fatalf("leader stopped inside safe lease window: status=%s pauses=%d", election.status, plugin.pauseCalls)
	}

	clock.Advance(8 * time.Second)
	if _, err := election.runOnce(); err == nil {
		t.Fatal("expected Operator failure at local deadline")
	}
	if election.status != StatusFail || plugin.pauseCalls != 2 {
		t.Fatalf("expired leader: status=%s pauses=%d", election.status, plugin.pauseCalls)
	}

	clock.Advance(3 * time.Second)
	if _, err := election.runOnce(); err != nil {
		t.Fatalf("reacquire campaign error = %v", err)
	}
	if election.status != StatusSuccess || plugin.pauseCalls != 2 || plugin.resumeCalls != 2 {
		t.Fatalf("after reacquire: status=%s pauses=%d resumes=%d", election.status, plugin.pauseCalls, plugin.resumeCalls)
	}

	mu.Lock()
	defer mu.Unlock()
	if campaignCalls != 3 || heartbeatCalls != 2 {
		t.Fatalf("calls: campaign=%d heartbeat=%d", campaignCalls, heartbeatCalls)
	}
}

func TestOperatorEpochChangesRestartLifecycleOnce(t *testing.T) {
	heartbeatCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		epoch := int64(7)
		if req.URL.Path == "/v1/dk-election/heartbeat" {
			heartbeatCalls++
			if heartbeatCalls >= 2 {
				epoch = 8
			}
		}
		writeElectionResponse(t, w, StatusSuccess.String(), "datakit-a", epoch)
	}))
	t.Cleanup(server.Close)

	clock := &fakeElectionClock{now: time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)}
	plugin := &lifecycleInput{}
	election := newOperatorLeaderElection(t, server.URL, clock, plugin)
	election.pausePlugins()

	for cycle := 0; cycle < 4; cycle++ {
		if _, err := election.runOnce(); err != nil {
			t.Fatalf("cycle %d error = %v", cycle, err)
		}
		clock.Advance(3 * time.Second)
	}

	if election.status != StatusSuccess || election.epoch != 8 {
		t.Fatalf("status=%s epoch=%d", election.status, election.epoch)
	}
	if plugin.pauseCalls != 2 || plugin.resumeCalls != 2 {
		t.Fatalf("epoch lifecycle: pauses=%d resumes=%d", plugin.pauseCalls, plugin.resumeCalls)
	}
}

func TestOperatorEpochChangeLeaseStartsAtResponseTime(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		writeElectionResponse(t, w, StatusSuccess.String(), "datakit-a", int64(calls))
	}))
	t.Cleanup(server.Close)

	startedAt := time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)
	clock := &fakeElectionClock{now: startedAt}
	plugin := &lifecycleInput{}
	election := newOperatorLeaderElection(t, server.URL, clock, plugin)
	election.pausePlugins()
	if _, err := election.runOnce(); err != nil {
		t.Fatalf("campaign error = %v", err)
	}

	clock.Advance(3 * time.Second)
	plugin.onPause = func() { clock.Advance(5 * time.Second) }
	plugin.onResume = func() { clock.Advance(5 * time.Second) }
	if _, err := election.runOnce(); err != nil {
		t.Fatalf("heartbeat error = %v", err)
	}

	wantDeadline := startedAt.Add(30 * time.Second)
	if !election.leaseDeadline.Equal(wantDeadline) {
		t.Fatalf("lease deadline = %s, want response-based %s", election.leaseDeadline, wantDeadline)
	}
}

func TestOperatorHolderChangeStopsOldLeaderImmediately(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		holder := "datakit-a"
		if calls > 1 {
			holder = "datakit-b"
		}
		writeElectionResponse(t, w, StatusSuccess.String(), holder, 7)
	}))
	t.Cleanup(server.Close)

	clock := &fakeElectionClock{now: time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)}
	plugin := &lifecycleInput{}
	election := newOperatorLeaderElection(t, server.URL, clock, plugin)
	election.pausePlugins()
	if _, err := election.runOnce(); err != nil {
		t.Fatalf("campaign error = %v", err)
	}

	clock.Advance(3 * time.Second)
	if _, err := election.runOnce(); err == nil {
		t.Fatal("expected inconsistent holder error")
	}
	if election.status != StatusFail || plugin.pauseCalls != 2 || plugin.resumeCalls != 1 {
		t.Fatalf("holder change: status=%s pauses=%d resumes=%d", election.status, plugin.pauseCalls, plugin.resumeCalls)
	}
}

func TestOperatorInvalidScopeDoesNotStopLeaderInsideLease(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			writeElectionResponse(t, w, StatusSuccess.String(), "datakit-a", 7)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"content": map[string]any{
				"status":         StatusSuccess.String(),
				"namespace":      "another-namespace",
				"id":             "datakit-a",
				"incumbency_id":  "datakit-b",
				"interval":       3,
				"lease_duration": 30,
				"epoch":          8,
			},
		}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	clock := &fakeElectionClock{now: time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)}
	plugin := &lifecycleInput{}
	election := newOperatorLeaderElection(t, server.URL, clock, plugin)
	election.pausePlugins()
	if _, err := election.runOnce(); err != nil {
		t.Fatalf("campaign error = %v", err)
	}

	clock.Advance(3 * time.Second)
	if _, err := election.runOnce(); !errors.Is(err, errInvalidElectionResponse) {
		t.Fatalf("heartbeat error = %v, want invalid response", err)
	}
	if election.status != StatusSuccess || plugin.pauseCalls != 1 || plugin.resumeCalls != 1 {
		t.Fatalf("invalid response changed lifecycle: status=%s pauses=%d resumes=%d",
			election.status, plugin.pauseCalls, plugin.resumeCalls)
	}
}

func TestOperatorDefeatPausesLeaderOnlyOnce(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			writeElectionResponse(t, w, StatusSuccess.String(), "datakit-a", 1)
			return
		}
		writeElectionResponse(t, w, StatusFail.String(), "datakit-b", 1)
	}))
	t.Cleanup(server.Close)

	clock := &fakeElectionClock{now: time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)}
	plugin := &lifecycleInput{}
	election := newOperatorLeaderElection(t, server.URL, clock, plugin)
	election.pausePlugins()

	for cycle := 0; cycle < 3; cycle++ {
		if _, err := election.runOnce(); err != nil {
			t.Fatalf("cycle %d error = %v", cycle, err)
		}
		clock.Advance(3 * time.Second)
	}
	if election.status != StatusFail || plugin.pauseCalls != 2 || plugin.resumeCalls != 1 {
		t.Fatalf("defeat lifecycle: status=%s pauses=%d resumes=%d", election.status, plugin.pauseCalls, plugin.resumeCalls)
	}
}

func TestOperatorCandidateRemainsPausedOnDefeat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeElectionResponseForID(t, w, StatusFail.String(), "datakit-b", "datakit-a", 1)
	}))
	t.Cleanup(server.Close)

	clock := &fakeElectionClock{now: time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)}
	plugin := &lifecycleInput{}
	election := newOperatorLeaderElectionForID(t, server.URL, clock, "datakit-b", plugin)
	election.pausePlugins()

	for cycle := 0; cycle < 2; cycle++ {
		if _, err := election.runOnce(); err != nil {
			t.Fatalf("cycle %d error = %v", cycle, err)
		}
		clock.Advance(3 * time.Second)
	}
	if election.status != StatusFail || CurrentElected != "datakit-a" {
		t.Fatalf("candidate status=%s holder=%q", election.status, CurrentElected)
	}
	if plugin.pauseCalls != 1 || plugin.resumeCalls != 0 {
		t.Fatalf("candidate lifecycle: pauses=%d resumes=%d", plugin.pauseCalls, plugin.resumeCalls)
	}
}

func TestOperatorRejectsPollingIntervalAtOrBelowRequestTimeout(t *testing.T) {
	clock := &fakeElectionClock{now: time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)}
	election := newLeaderElection(&option{
		id:        "datakit-a",
		namespace: "production",
		provider:  ProviderOperator,
		clock:     clock,
	}, nil)
	result := &leaderElectionResult{}
	result.Content.Status = StatusSuccess.String()
	result.Content.Namespace = election.namespace
	result.Content.ID = election.id
	result.Content.IncumbencyID = election.id
	result.Content.Interval = int(operatorRequestTimeout / time.Second)
	result.Content.LeaseDuration = 30
	result.Content.Epoch = 1

	if err := election.validateOperatorResult(result); !errors.Is(err, errInvalidElectionResponse) {
		t.Fatalf("validateOperatorResult() error = %v, want invalid response", err)
	}
}

func TestOperatorDurationSecondsRejectsOverflow(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("int cannot represent a duration-overflowing number of seconds")
	}
	overflowSeconds := (1<<63-1)/int64(time.Second) + 1
	if _, ok := durationFromSeconds(int(overflowSeconds)); ok {
		t.Fatalf("durationFromSeconds(%d) succeeded", overflowSeconds)
	}
}

func TestOperatorNextCheckDoesNotOvershootSafeLeaseDeadline(t *testing.T) {
	clock := &fakeElectionClock{now: time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)}
	election := newLeaderElection(&option{
		id:        "datakit-a",
		namespace: "production",
		provider:  ProviderOperator,
		clock:     clock,
	}, nil)
	election.status = StatusSuccess
	election.leaseDeadline = clock.Now().Add(250 * time.Millisecond)

	if got := election.nextCheckInterval(3); got != 250*time.Millisecond {
		t.Fatalf("nextCheckInterval() = %s, want 250ms", got)
	}
}

func TestOperatorDoesNotStartHeartbeatInsideFinalPollingWindow(t *testing.T) {
	puller := &recordingPuller{}
	clock := &fakeElectionClock{now: time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)}
	election := newLeaderElection(&option{
		id:        "datakit-a",
		namespace: "production",
		provider:  ProviderOperator,
		puller:    puller,
		clock:     clock,
	}, nil)
	election.status = StatusSuccess
	election.pollInterval = 3 * time.Second
	election.leaseDeadline = clock.Now().Add(election.pollInterval - time.Nanosecond)

	interval, err := election.runOnce()
	if err != nil {
		t.Fatalf("runOnce() error = %v", err)
	}
	if puller.heartbeatCalls != 0 {
		t.Fatalf("heartbeat calls = %d, want 0", puller.heartbeatCalls)
	}
	if interval != 3 || election.status != StatusSuccess {
		t.Fatalf("interval=%d status=%s", interval, election.status)
	}
}

func TestOperatorElectionMetricsExposeLeaseAndBoundedErrors(t *testing.T) {
	resetOperatorElectionMetrics()
	t.Cleanup(resetOperatorElectionMetrics)

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls > 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writeElectionResponse(t, w, StatusSuccess.String(), "datakit-a", 11)
	}))
	t.Cleanup(server.Close)

	clock := &fakeElectionClock{now: time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)}
	plugin := &lifecycleInput{}
	election := newOperatorLeaderElection(t, server.URL, clock, plugin)
	if _, err := election.runOnce(); err != nil {
		t.Fatalf("campaign error = %v", err)
	}
	clock.Advance(3 * time.Second)
	if _, err := election.runOnce(); err == nil {
		t.Fatal("expected heartbeat error")
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		electionProviderInfoVec,
		electionLastSuccessVec,
		electionLeaseRemainingVec,
		electionEpochVec,
		electionRequestErrorsVec,
		electionTransitionsVec,
	)
	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	assertMetricValue(t, metrics, "datakit_election_provider_info", map[string]string{
		"namespace": "production", "provider": "operator",
	}, 1)
	assertMetricValue(t, metrics, "datakit_election_last_success_timestamp_seconds", map[string]string{
		"namespace": "production", "provider": "operator",
	}, float64(clock.now.Add(-3*time.Second).Unix()))
	assertMetricValue(t, metrics, "datakit_election_lease_remaining_seconds", map[string]string{
		"namespace": "production", "provider": "operator",
	}, 24)
	assertMetricValue(t, metrics, "datakit_election_epoch", map[string]string{
		"namespace": "production", "provider": "operator",
	}, 11)
	assertMetricValue(t, metrics, "datakit_election_request_errors_total", map[string]string{
		"namespace": "production", "provider": "operator", "operation": "heartbeat", "reason": "http_5xx",
	}, 1)
	assertMetricValue(t, metrics, "datakit_election_transitions_total", map[string]string{
		"namespace": "production", "provider": "operator", "from": "defeat", "to": "success", "reason": "acquired",
	}, 1)
}

func TestDatawayLeaderKeepsExistingErrorSemantics(t *testing.T) {
	preserveCurrentElected(t)
	clock := &fakeElectionClock{now: time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)}
	plugin := &lifecycleInput{}
	puller := &datawayRegressionPuller{
		campaignBody: []byte(`{"content":{"status":"success","namespace":"production","id":"datakit-a","incumbency_id":"datakit-a","interval":3}}`),
		heartbeatErr: errors.New("temporary DataWay failure"),
	}
	election := newLeaderElection(&option{
		id:        "datakit-a",
		namespace: "production",
		provider:  ProviderDataway,
		puller:    puller,
		clock:     clock,
	}, map[string][]inputs.ElectionInput{"container": {plugin}})
	election.pausePlugins()

	if _, err := election.runOnce(); err != nil {
		t.Fatalf("campaign error = %v", err)
	}
	clock.Advance(24 * time.Hour)
	if _, err := election.runOnce(); err == nil {
		t.Fatal("expected DataWay heartbeat error")
	}
	if election.status != StatusSuccess || plugin.pauseCalls != 1 || plugin.resumeCalls != 1 {
		t.Fatalf("DataWay behavior changed: status=%s pauses=%d resumes=%d",
			election.status, plugin.pauseCalls, plugin.resumeCalls)
	}
}

func TestDatawayCampaignKeepsMalformedResponseSemantics(t *testing.T) {
	preserveCurrentElected(t)
	clock := &fakeElectionClock{now: time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)}
	election := newLeaderElection(&option{
		id:        "datakit-a",
		namespace: "production",
		provider:  ProviderDataway,
		puller: &datawayRegressionPuller{
			campaignBody: []byte(`not-json`),
		},
		clock: clock,
	}, nil)

	interval, err := election.runOnce()
	if err != nil {
		t.Fatalf("DataWay malformed campaign error = %v", err)
	}
	if interval != datawayElectionIntervalDefault || election.status != StatusFail {
		t.Fatalf("interval=%d status=%s", interval, election.status)
	}
}

func resetOperatorElectionMetrics() {
	electionProviderInfoVec.Reset()
	electionLastSuccessVec.Reset()
	electionLeaseRemainingVec.Reset()
	electionEpochVec.Reset()
	electionRequestErrorsVec.Reset()
	electionTransitionsVec.Reset()
}

func assertMetricValue(t *testing.T, families []*dto.MetricFamily, name string, labels map[string]string, want float64) {
	t.Helper()
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.Metric {
			if !metricHasLabels(metric, labels) {
				continue
			}
			got := metric.GetGauge().GetValue()
			if metric.Counter != nil {
				got = metric.GetCounter().GetValue()
			}
			if got != want {
				t.Fatalf("metric %s = %v, want %v", name, got, want)
			}
			return
		}
	}
	t.Fatalf("metric %s with labels %v not found", name, labels)
}

func metricHasLabels(metric *dto.Metric, want map[string]string) bool {
	got := make(map[string]string, len(metric.Label))
	for _, label := range metric.Label {
		got[label.GetName()] = label.GetValue()
	}
	for name, value := range want {
		if got[name] != value {
			return false
		}
	}
	return true
}
