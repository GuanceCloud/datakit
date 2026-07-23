// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	climetrics "github.com/GuanceCloud/cliutils/metrics"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/client_golang/prometheus"
	clientmodel "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func TestLegacyPromConfigCompatibility(t *testing.T) {
	protectMode := config.Cfg.ProtectMode
	config.Cfg.ProtectMode = true
	t.Cleanup(func() { config.Cfg.ProtectMode = protectMode })
	pod := &apicorev1.Pod{ObjectMeta: metav1.ObjectMeta{UID: "legacy", Name: "pod-a", Namespace: "market",
		Labels: map[string]string{"team": "observe"}, OwnerReferences: []metav1.OwnerReference{{Kind: "Job", Name: "job-a"}}},
		Spec: apicorev1.PodSpec{NodeName: "node-a"}, Status: apicorev1.PodStatus{PodIP: "10.0.0.1"}}
	raw := "[[inputs.prom]]\nurls=['http://$IP:9100/metrics']\nurl='http://$IP:9200/metrics'\ninterval='1s'\ntimeout='9s'\n" +
		"removed_or_unknown_field=true\n[inputs.prom.tags]\nnamespace='$NAMESPACE'\npod='$PODNAME'\nnode='$NODENAME'\n"
	configs, err := parsePromConfigs(completePromConfig(pod, raw))
	require.NoError(t, err)
	require.Equal(t, []string{"http://10.0.0.1:9100/metrics", "http://10.0.0.1:9200/metrics"}, configs[0].URLs)
	require.Equal(t, 10*time.Second, configs[0].Interval)
	require.Equal(t, 9*time.Second, configs[0].Timeout)
	require.Equal(t, map[string]string{"namespace": "market", "pod": "pod-a", "node": "node-a"}, configs[0].Tags)
	runners, err := newPromRunnersForPod(pod, configs, &Config{LabelAsTagsForMetric: LabelsOption{Keys: []string{"team"}}})
	require.NoError(t, err)
	t.Cleanup(func() { stopPromRunners(runners) })
	require.Equal(t, "market/job-a", runners[0].conf.Source)
	require.Equal(t, "observe", runners[0].conf.Tags["team"])
}

func TestPromRunnerEmitsSelectedPodLabelsWithoutConfiguredTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintln(w, "sample_metric 1")
	}))
	t.Cleanup(server.Close)

	configs, err := parsePromConfigs(fmt.Sprintf(
		"[[inputs.prom]]\nurl=%q\nsource='test'\ninterval='30s'\n",
		server.URL,
	))
	require.NoError(t, err)
	feeder := &changeTestFeeder{}
	pod := &apicorev1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"team": "observe"}}}
	runners, err := newPromRunnersForPod(pod, configs, &Config{
		Feeder:               feeder,
		LabelAsTagsForMetric: LabelsOption{Keys: []string{"team"}},
	})
	require.NoError(t, err)
	t.Cleanup(func() { stopPromRunners(runners) })

	require.NoError(t, runners[0].scrape(context.Background(), time.Unix(1, 0), time.Second))
	require.Len(t, feeder.points, 1)
	require.Equal(t, "observe", feeder.points[0].GetTag("team"))
}

func TestPromManagerRefreshesEffectivePodState(t *testing.T) {
	manager := newPromTaskManager(&Config{
		LabelAsTagsForMetric: LabelsOption{Keys: []string{"team"}},
	})
	t.Cleanup(func() {
		manager.removeAllPods()
		manager.cancel()
	})

	raw := "[[inputs.prom]]\nurl='http://$IP:9100/metrics'\ninterval='30s'\n"
	first := &apicorev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       types.UID("same-uid"),
			Name:      "pod-a",
			Namespace: "default",
			Labels:    map[string]string{"team": "old-team", "ignored": "old-ignored"},
			OwnerReferences: []metav1.OwnerReference{{
				Kind: "Job", Name: "old-owner",
			}},
		},
		Status: apicorev1.PodStatus{Phase: apicorev1.PodRunning, PodIP: "10.0.0.1"},
	}
	manager.applySnapshot([]promPodCandidate{newPromPodCandidate(first, raw)}, time.Unix(1, 0))
	firstState := manager.pods[first.UID]

	irrelevant := first.DeepCopy()
	irrelevant.Labels["ignored"] = "new-ignored"
	manager.applySnapshot([]promPodCandidate{newPromPodCandidate(irrelevant, raw)}, time.Unix(2, 0))
	require.Same(t, firstState, manager.pods[first.UID])

	updated := irrelevant.DeepCopy()
	updated.Status.PodIP = "10.0.0.2"
	updated.Labels["team"] = "new-team"
	updated.OwnerReferences[0].Name = "new-owner"
	manager.applySnapshot([]promPodCandidate{newPromPodCandidate(updated, raw)}, time.Unix(3, 0))

	updatedState := manager.pods[first.UID]
	require.NotSame(t, firstState, updatedState)
	runner := updatedState.tasks[0].runner
	require.Equal(t, []string{"http://10.0.0.2:9100/metrics"}, runner.conf.URLs)
	require.Equal(t, "new-team", runner.conf.Tags["team"])
	require.Equal(t, "default/new-owner", runner.conf.Source)

	delete(updated.Labels, "team")
	manager.applySnapshot([]promPodCandidate{newPromPodCandidate(updated, raw)}, time.Unix(4, 0))
	require.NotSame(t, updatedState, manager.pods[first.UID])
	require.NotContains(t, manager.pods[first.UID].tasks[0].runner.conf.Tags, "team")
}

func TestPromScheduleKeepsUndispatchedTasksDue(t *testing.T) {
	manager := newPromTaskManager(&Config{})
	manager.cancel()
	manager.workerCount = 4
	manager.jobs = make(chan promJob, manager.workerCount)
	base := time.Unix(1_700_000_000, 0)
	manager.tasks = make([]*promTask, 5)
	for idx := range manager.tasks {
		manager.tasks[idx] = &promTask{
			interval: time.Minute,
			nextDue:  base,
			ctx:      context.Background(),
		}
	}

	manager.schedule(base)

	require.Len(t, manager.jobs, 4)
	require.Equal(t, base, manager.tasks[4].nextDue)
	require.False(t, manager.tasks[4].busy)
}

func TestPromWorkersDrainAllDueTasksWithoutAnotherTick(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests.Add(1)
	}))
	t.Cleanup(server.Close)

	base := time.Unix(1_700_000_000, 0)
	var nowNano atomic.Int64
	nowNano.Store(base.UnixNano())
	manager, ticks := startPromManager(t, &Config{NodeLocal: true}, 4, time.Second,
		func() time.Time { return time.Unix(0, nowNano.Load()) })
	raw := strings.Repeat(fmt.Sprintf(
		"[[inputs.prom]]\nurl=%q\ninterval='10s'\n", server.URL,
	), 10)
	manager.publishPromPods([]promPodCandidate{promCandidate("ten-tasks", raw)})
	require.Eventually(t, func() bool {
		return promValue(t, podAnnotationPromActiveTasks) == 10
	}, time.Second, time.Millisecond)

	nowNano.Store(base.Add(10 * time.Second).UnixNano())
	ticks <- base.Add(10 * time.Second)
	require.Eventually(t, func() bool {
		return requests.Load() == 10
	}, time.Second, time.Millisecond)
}

func TestPromPodRevisionTracksOnlyEffectiveInputs(t *testing.T) {
	labelKeys := []string{"team"}
	raw := "[[inputs.prom]]\nurl='http://$IP:9100/metrics'\ninterval='30s'\n"
	basePod := &apicorev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       types.UID("same-uid"),
			Name:      "pod-a",
			Namespace: "default",
			Labels:    map[string]string{"team": "old-team", "ignored": "old-ignored"},
			OwnerReferences: []metav1.OwnerReference{{
				Kind: "Job", Name: "old-owner",
			}},
		},
		Status: apicorev1.PodStatus{Phase: apicorev1.PodRunning, PodIP: "10.0.0.1"},
	}
	base, err := preparePromPod(newPromPodCandidate(basePod, raw), labelKeys)
	require.NoError(t, err)

	tests := []struct {
		name   string
		mutate func(*apicorev1.Pod)
		equal  bool
	}{
		{
			name: "pod IP",
			mutate: func(pod *apicorev1.Pod) {
				pod.Status.PodIP = "10.0.0.2"
			},
		},
		{
			name: "owner",
			mutate: func(pod *apicorev1.Pod) {
				pod.OwnerReferences[0].Name = "new-owner"
			},
		},
		{
			name: "selected label",
			mutate: func(pod *apicorev1.Pod) {
				pod.Labels["team"] = "new-team"
			},
		},
		{
			name: "unselected label",
			mutate: func(pod *apicorev1.Pod) {
				pod.Labels["ignored"] = "new-ignored"
			},
			equal: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pod := basePod.DeepCopy()
			test.mutate(pod)
			prepared, err := preparePromPod(newPromPodCandidate(pod, raw), labelKeys)
			require.NoError(t, err)
			if test.equal {
				require.Equal(t, base.revision, prepared.revision)
			} else {
				require.NotEqual(t, base.revision, prepared.revision)
			}
		})
	}

	explicitRaw := "[[inputs.prom]]\nurl='http://static:9100/metrics'\nsource='explicit'\ninterval='30s'\n" +
		"[inputs.prom.tags]\nteam='explicit'\n"
	explicitBase, err := preparePromPod(newPromPodCandidate(basePod, explicitRaw), labelKeys)
	require.NoError(t, err)
	irrelevant := basePod.DeepCopy()
	irrelevant.Status.PodIP = "10.0.0.2"
	irrelevant.OwnerReferences[0].Name = "new-owner"
	irrelevant.Labels["team"] = "new-team"
	explicitUpdated, err := preparePromPod(newPromPodCandidate(irrelevant, explicitRaw), labelKeys)
	require.NoError(t, err)
	require.Equal(t, explicitBase.revision, explicitUpdated.revision)
}

func TestPromRunnerReloadsBearerTokenForEveryRequest(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(tokenFile, []byte("old-token"), 0o600))

	var authorizations []string
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		authorizations = append(authorizations, req.Header.Get("Authorization"))
	}))
	t.Cleanup(server.Close)

	raw := fmt.Sprintf(
		"[[inputs.prom]]\nurl=%q\ninterval='30s'\nbearer_token_file=%q\n",
		server.URL, tokenFile,
	)
	configs, err := parsePromConfigs(raw)
	require.NoError(t, err)
	configs[0].HTTPHeaders = map[string]string{"Authorization": "Bearer static-token"}
	runners, err := newPromRunnersForPod(&apicorev1.Pod{}, configs, &Config{})
	require.NoError(t, err)
	t.Cleanup(func() { stopPromRunners(runners) })

	require.NoError(t, runners[0].scrape(context.Background(), time.Unix(1, 0), time.Second))
	require.NoError(t, os.WriteFile(tokenFile, []byte("new-token\n"), 0o600))
	require.NoError(t, runners[0].scrape(context.Background(), time.Unix(31, 0), time.Second))
	require.NoError(t, os.Remove(tokenFile))
	require.Error(t, runners[0].scrape(context.Background(), time.Unix(61, 0), time.Second))
	require.NoError(t, os.WriteFile(tokenFile, []byte("recovered-token"), 0o600))
	require.NoError(t, runners[0].scrape(context.Background(), time.Unix(91, 0), time.Second))
	require.Equal(t,
		[]string{"Bearer old-token", "Bearer new-token", "Bearer recovered-token"},
		authorizations)
}

func TestPromRunnerRejectsInvalidAuthWithoutPanic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	t.Cleanup(server.Close)
	configs, err := parsePromConfigs(fmt.Sprintf(
		"[[inputs.prom]]\nurl=%q\ninterval='30s'\n[inputs.prom.auth]\ntoken='secret'\n",
		server.URL,
	))
	require.NoError(t, err)
	runners, err := newPromRunnersForPod(&apicorev1.Pod{}, configs, &Config{})
	require.NoError(t, err)
	t.Cleanup(func() { stopPromRunners(runners) })

	require.Error(t, runners[0].scrape(context.Background(), time.Unix(1, 0), time.Second))
}

type panicPromFeeder struct {
	panicOnFeed bool
}

func (f *panicPromFeeder) Feed(_ point.Category, _ []*point.Point, _ ...dkio.FeedOption) error {
	if f.panicOnFeed {
		panic("test feeder panic")
	}
	return nil
}

func (*panicPromFeeder) FeedLastError(string, ...metrics.LastErrorOption) {}

func TestPromRunnerCleansInflightAfterCallbackPanic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintln(w, "sample_metric 1")
	}))
	t.Cleanup(server.Close)
	configs, err := parsePromConfigs(fmt.Sprintf(
		"[[inputs.prom]]\nurl=%q\ninterval='30s'\n", server.URL,
	))
	require.NoError(t, err)
	feeder := &panicPromFeeder{panicOnFeed: true}
	runners, err := newPromRunnersForPod(&apicorev1.Pod{}, configs, &Config{Feeder: feeder})
	require.NoError(t, err)
	t.Cleanup(func() { stopPromRunners(runners) })

	inflightBefore := promValue(t, podAnnotationPromInflightScrapes)
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_ = runners[0].scrape(context.Background(), time.Unix(1, 0), time.Second)
	}()
	require.NotNil(t, recovered)
	require.Equal(t, inflightBefore, promValue(t, podAnnotationPromInflightScrapes))

	feeder.panicOnFeed = false
	require.NoError(t, runners[0].scrape(context.Background(), time.Unix(31, 0), time.Second))
}

func TestPromRunnerReloadsTLSFilesBeforeRequest(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.StartTLS()
	t.Cleanup(server.Close)
	wrongCAPEM := newTestCAPEM(t)

	caFile := filepath.Join(t.TempDir(), "ca.pem")
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	require.NoError(t, os.WriteFile(caFile, caPEM, 0o600))

	raw := fmt.Sprintf(
		"[[inputs.prom]]\nurl=%q\ninterval='30s'\ntls_open=true\ntls_ca=%q\n",
		server.URL, caFile,
	)
	configs, err := parsePromConfigs(raw)
	require.NoError(t, err)
	runners, err := newPromRunnersForPod(&apicorev1.Pod{}, configs, &Config{})
	require.NoError(t, err)
	t.Cleanup(func() { stopPromRunners(runners) })

	runner := runners[0]
	initialProm := runner.pm
	require.NoError(t, runner.scrape(context.Background(), time.Unix(1, 0), time.Second))
	require.Same(t, initialProm, runner.pm)
	require.Empty(t, runner.promSource)

	require.NoError(t, os.WriteFile(caFile, wrongCAPEM, 0o600))
	require.Error(t, runner.scrape(context.Background(), time.Unix(31, 0), time.Second))
	wrongProm := runner.pm
	require.NotSame(t, initialProm, wrongProm)

	require.NoError(t, os.WriteFile(caFile, caPEM, 0o600))
	require.NoError(t, runner.scrape(context.Background(), time.Unix(61, 0), time.Second))
	require.NotSame(t, wrongProm, runner.pm)
	require.Empty(t, runner.promSource)
}

func TestPromRunnerReloadsTLSClientCertificateBeforeRequest(t *testing.T) {
	clientCAPEM, clientCA, clientCAKey := newTestCA(t, "client-ca")
	clientCAPool := x509.NewCertPool()
	require.True(t, clientCAPool.AppendCertsFromPEM(clientCAPEM))

	clientSerials := make(chan int64, 2)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		if len(req.TLS.PeerCertificates) == 0 {
			clientSerials <- 0
			return
		}
		clientSerials <- req.TLS.PeerCertificates[0].SerialNumber.Int64()
	}))
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.TLS = &tls.Config{
		MinVersion: tls.VersionTLS12,
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  clientCAPool,
	}
	server.StartTLS()
	t.Cleanup(server.Close)

	directory := t.TempDir()
	caFile := filepath.Join(directory, "server-ca.pem")
	certFile := filepath.Join(directory, "client.pem")
	keyFile := filepath.Join(directory, "client-key.pem")
	serverCAPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	require.NoError(t, os.WriteFile(caFile, serverCAPEM, 0o600))
	certPEM, keyPEM := newTestClientCertificate(t, clientCA, clientCAKey, 1)
	require.NoError(t, os.WriteFile(certFile, certPEM, 0o600))
	require.NoError(t, os.WriteFile(keyFile, keyPEM, 0o600))

	raw := fmt.Sprintf(
		"[[inputs.prom]]\nurl=%q\ninterval='30s'\ntls_open=true\ntls_ca=%q\ntls_cert=%q\ntls_key=%q\n",
		server.URL, caFile, certFile, keyFile,
	)
	configs, err := parsePromConfigs(raw)
	require.NoError(t, err)
	runners, err := newPromRunnersForPod(&apicorev1.Pod{}, configs, &Config{})
	require.NoError(t, err)
	t.Cleanup(func() { stopPromRunners(runners) })

	runner := runners[0]
	initialProm := runner.pm
	require.NoError(t, runner.scrape(context.Background(), time.Unix(1, 0), time.Second))
	require.Equal(t, int64(1), <-clientSerials)

	certPEM, keyPEM = newTestClientCertificate(t, clientCA, clientCAKey, 2)
	require.NoError(t, os.WriteFile(certFile, certPEM, 0o600))
	require.NoError(t, os.WriteFile(keyFile, keyPEM, 0o600))
	require.NoError(t, runner.scrape(context.Background(), time.Unix(31, 0), time.Second))
	require.Equal(t, int64(2), <-clientSerials)
	require.NotSame(t, initialProm, runner.pm)
}

func newTestCAPEM(t *testing.T) []byte {
	t.Helper()
	caPEM, _, _ := newTestCA(t, "unrelated-test-ca")
	return caPEM
}

func newTestCA(t *testing.T, name string) ([]byte, *x509.Certificate, ed25519.PrivateKey) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: name},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	certificateDER, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	require.NoError(t, err)
	certificate, err := x509.ParseCertificate(certificateDER)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateDER}), certificate, privateKey
}

func newTestClientCertificate(
	t *testing.T,
	ca *x509.Certificate,
	caKey ed25519.PrivateKey,
	serial int64,
) ([]byte, []byte) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject:      pkix.Name{CommonName: fmt.Sprintf("client-%d", serial)},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	certificate, err := x509.CreateCertificate(rand.Reader, template, ca, publicKey, caKey)
	require.NoError(t, err)
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER})
}

func TestPromTLSFileRevisionTracksEveryCredentialFile(t *testing.T) {
	directory := t.TempDir()
	paths := []string{
		filepath.Join(directory, "ca.pem"),
		filepath.Join(directory, "client.pem"),
		filepath.Join(directory, "client-key.pem"),
	}
	for idx, path := range paths {
		require.NoError(t, os.WriteFile(path, []byte(fmt.Sprintf("initial-%d", idx)), 0o600))
	}
	config := &promConfig{
		TLSOpen:    true,
		CacertFile: paths[0],
		CertFile:   paths[1],
		KeyFile:    paths[2],
	}
	initial, watched, err := promTLSFileRevision(config)
	require.NoError(t, err)
	require.True(t, watched)

	for idx, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			require.NoError(t, os.WriteFile(path, []byte(fmt.Sprintf("rotated-%d", idx)), 0o600))
			rotated, watched, err := promTLSFileRevision(config)
			require.NoError(t, err)
			require.True(t, watched)
			require.NotEqual(t, initial, rotated)
			require.NoError(t, os.WriteFile(path, []byte(fmt.Sprintf("initial-%d", idx)), 0o600))
		})
	}
}

func TestPromManagerAdmissionAndRawHash(t *testing.T) {
	manager, _ := startPromManager(t, &Config{NodeLocal: true}, 2, time.Second, time.Now)
	large := promCandidate("large", strings.Repeat("[[inputs.prom]]\nurl='http://127.0.0.1'\ninterval='1h'\n", 1001))
	small := promCandidate("small", "[[inputs.prom]]\nurl='http://127.0.0.1'\ninterval='1h'\n")
	atLimit := promCandidate("limit", strings.Repeat(small.rawConfig, defaultPromTaskLimit))
	manager.publishPromPods([]promPodCandidate{atLimit})
	require.Eventually(t, func() bool { return promValue(t, podAnnotationPromActiveTasks) == 1000 }, time.Second, time.Millisecond)
	pods := make([]promPodCandidate, defaultPromPodLimit+1)
	for i := range pods {
		pods[i] = promCandidate(fmt.Sprint(i), small.rawConfig)
	}
	manager.publishPromPods(pods)
	require.Eventually(t, func() bool { return promValue(t, podAnnotationPromActiveTasks) == 500 }, time.Second, time.Millisecond)
	manager.publishPromPods([]promPodCandidate{large, small})
	require.Eventually(t, func() bool { return promValue(t, podAnnotationPromActiveTasks) == 1 }, time.Second, time.Millisecond)

	tokenFile := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(tokenFile, []byte("token"), 0o600))
	block := fmt.Sprintf("[[inputs.prom]]\nurl='http://127.0.0.1'\nbearer_token_file=%q\n", tokenFile)
	raw := block + block
	tokenPod := promCandidate("token", raw)
	manager.publishPromPods([]promPodCandidate{tokenPod})
	require.Eventually(t, func() bool { return promValue(t, podAnnotationPromActiveTasks) == 2 }, time.Second, time.Millisecond)
	require.NoError(t, os.Remove(tokenFile))
	manager.publishPromPods([]promPodCandidate{tokenPod, small})
	require.Eventually(t, func() bool { return promValue(t, podAnnotationPromActiveTasks) == 3 }, time.Second, time.Millisecond)
	manager.publishPromPods([]promPodCandidate{promCandidate("token", raw+"# changed\n"), small})
	require.Eventually(t, func() bool { return promValue(t, podAnnotationPromActiveTasks) == 1 }, time.Second, time.Millisecond)
}

func TestPromManagerBoundsSkipsAndCancelsScrapes(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-req.Context().Done()
	}))
	t.Cleanup(server.Close)
	base := time.Unix(1_700_000_000, 0)
	var nowOffset atomic.Int64
	manager, ticks := startPromManager(t, &Config{}, 2, 300*time.Millisecond,
		func() time.Time { return base.Add(time.Duration(nowOffset.Load())) })
	timeoutBefore := promValue(t, podAnnotationPromScrapesTotal.WithLabelValues("timeout"))
	canceledBefore := promValue(t, podAnnotationPromScrapesTotal.WithLabelValues("canceled"))
	raw := strings.Repeat(fmt.Sprintf("[[inputs.prom]]\nurl=%q\ninterval='10s'\ntimeout='100ms'\n", server.URL), 3)
	manager.publishPromPods([]promPodCandidate{promCandidate("slow", raw)})
	manager.setActive(true)
	require.Zero(t, promValue(t, podAnnotationPromActiveTasks))
	manager.publishPromPods([]promPodCandidate{promCandidate("slow", raw)})
	require.Eventually(t, func() bool { return promValue(t, podAnnotationPromActiveTasks) == 3 }, time.Second, time.Millisecond)
	ticks <- base.Add(10 * time.Second)
	require.Eventually(t, func() bool { return requests.Load() == 2 }, time.Second, time.Millisecond)
	ticks <- base.Add(20 * time.Second)
	nowOffset.Store(int64(21 * time.Second))
	time.Sleep(30 * time.Millisecond)
	require.EqualValues(t, 2, requests.Load())
	require.Eventually(t, func() bool {
		return requests.Load() == 3 &&
			promValue(t, podAnnotationPromScrapesTotal.WithLabelValues("timeout")) == timeoutBefore+3 &&
			promValue(t, podAnnotationPromInflightScrapes) == 0
	}, 2*time.Second, 5*time.Millisecond)

	nextTick := 30
	require.Eventually(t, func() bool {
		ticks <- base.Add(time.Duration(nextTick) * time.Second)
		nextTick++
		return requests.Load() == 5
	}, time.Second, time.Millisecond)
	manager.setActive(false)
	require.Eventually(t, func() bool {
		return promValue(t, podAnnotationPromScrapesTotal.WithLabelValues("canceled")) == canceledBefore+2 &&
			promValue(t, podAnnotationPromInflightScrapes) == 0
	}, time.Second, 5*time.Millisecond)
	require.Zero(t, promValue(t, podAnnotationPromActiveTasks))
	manager.setActive(true)
	require.Zero(t, promValue(t, podAnnotationPromActiveTasks))
}

func TestMaxPromRequestTimeoutHonorsConfiguredUpperBound(t *testing.T) {
	require.Equal(t, 2*time.Minute, maxPromRequestTimeout(defaultPromRequestTimeout, 2*time.Minute))
	require.Equal(t, defaultPromRequestTimeout, maxPromRequestTimeout(defaultPromRequestTimeout, 9*time.Second))
}

func TestPromSnapshotRequiresEveryPage(t *testing.T) {
	client := &pageFailClient{}
	p := newPod(client, &Config{promPodPublisher: client})
	p.gatherObject(context.Background(), "", false)
	require.Equal(t, 2, client.calls)
	require.False(t, client.published)
}

func TestPromMetricsHaveBoundedLabels(t *testing.T) {
	families, err := climetrics.Gather()
	require.NoError(t, err)
	byName := make(map[string]*clientmodel.MetricFamily, len(families))
	for _, family := range families {
		byName[family.GetName()] = family
	}
	prefix := "datakit_input_container_kubernetes_pod_annotation_prom_"
	for _, name := range []string{prefix + "active_tasks", prefix + "inflight_scrapes"} {
		require.Len(t, byName[name].GetMetric(), 1)
		require.Empty(t, byName[name].GetMetric()[0].GetLabel())
	}
	scrapes := byName[prefix+"scrapes_total"].GetMetric()
	require.Len(t, scrapes, 4)
	for _, metric := range scrapes {
		require.Len(t, metric.GetLabel(), 1)
		require.Equal(t, "result", metric.GetLabel()[0].GetName())
	}
}

func startPromManager(t *testing.T, cfg *Config, workers int, timeout time.Duration, now func() time.Time) (*promTaskManager, chan time.Time) {
	t.Helper()
	ticks := make(chan time.Time)
	manager := newPromTaskManager(cfg)
	manager.workerCount, manager.requestTimeout, manager.tickC = workers, timeout, ticks
	manager.now = now
	go manager.run()
	t.Cleanup(manager.close)
	return manager, ticks
}

func promCandidate(uid, raw string) promPodCandidate {
	pod := &apicorev1.Pod{ObjectMeta: metav1.ObjectMeta{UID: types.UID(uid), Name: uid, Namespace: "default"},
		Status: apicorev1.PodStatus{Phase: apicorev1.PodRunning, PodIP: "127.0.0.1"}}
	return newPromPodCandidate(pod, raw)
}

func promValue(t *testing.T, source prometheus.Metric) float64 {
	t.Helper()
	metric := &clientmodel.Metric{}
	require.NoError(t, source.Write(metric))
	return metric.GetGauge().GetValue() + metric.GetCounter().GetValue()
}

type pageFailClient struct {
	k8sClient
	corev1.PodInterface
	calls     int
	published bool
}

func (p *pageFailClient) GetPods(string) corev1.PodInterface { return p }
func (p *pageFailClient) publishPromPods([]promPodCandidate) { p.published = true }

func (p *pageFailClient) List(context.Context, metav1.ListOptions) (*apicorev1.PodList, error) {
	p.calls++
	if p.calls == 1 {
		return &apicorev1.PodList{ListMeta: metav1.ListMeta{Continue: "next"}}, nil
	}
	return nil, fmt.Errorf("page 2 failed")
}
