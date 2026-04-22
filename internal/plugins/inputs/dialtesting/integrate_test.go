// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GuanceCloud/cliutils"
	dt "github.com/GuanceCloud/cliutils/dialtesting"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	RegionID = "regionID"
	DialAK   = "dialak"
	DialSK   = "dialsk"
)

var collectPointsCache = make([]*point.Point, 0)

type mockSender struct {
	mu   sync.Mutex
	urls []string
	pts  []*point.Point
}

func (m *mockSender) send(url string, pt *point.Point) error {
	m.mu.Lock()
	collectPointsCache = append(collectPointsCache, pt)
	m.urls = append(m.urls, url)
	m.pts = append(m.pts, pt)
	m.mu.Unlock()
	return nil
}

func (m *mockSender) checkToken(token, scheme, host string) (bool, error) {
	return true, nil
}

func (m *mockSender) URLs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	res := make([]string, len(m.urls))
	copy(res, m.urls)
	return res
}

func (m *mockSender) Points() []*point.Point {
	m.mu.Lock()
	defer m.mu.Unlock()

	res := make([]*point.Point, len(m.pts))
	copy(res, m.pts)
	return res
}

type mockDialtestingServer struct {
	mu              sync.Mutex
	pullResponse    []byte
	pullStatusCode  int
	pullCount       int
	variableUpdates [][]byte
}

func newMockDialtestingServer(t *testing.T, pullResponse []byte) *mockDialtestingServer {
	t.Helper()

	ms := &mockDialtestingServer{
		pullResponse:   pullResponse,
		pullStatusCode: http.StatusOK,
	}

	return ms
}

func (ms *mockDialtestingServer) URL() string {
	return "http://mock.local"
}

func (ms *mockDialtestingServer) PullCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.pullCount
}

func (ms *mockDialtestingServer) VariableUpdates() [][]byte {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	res := make([][]byte, 0, len(ms.variableUpdates))
	for _, item := range ms.variableUpdates {
		cp := make([]byte, len(item))
		copy(cp, item)
		res = append(res, cp)
	}
	return res
}

func snapshotCollectedPoints() []*point.Point {
	res := make([]*point.Point, len(collectPointsCache))
	copy(res, collectPointsCache)
	return res
}

func (ms *mockDialtestingServer) client(t *testing.T) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/v1/task/pull":
				ms.mu.Lock()
				ms.pullCount++
				body := ms.pullResponse
				statusCode := ms.pullStatusCode
				ms.mu.Unlock()

				return &http.Response{
					StatusCode: statusCode,
					Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
					Body:       io.NopCloser(strings.NewReader(string(body))),
					Header:     make(http.Header),
				}, nil
			case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf("/v1/variable/update/%s", RegionID):
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				_ = r.Body.Close()

				ms.mu.Lock()
				ms.variableUpdates = append(ms.variableUpdates, body)
				ms.mu.Unlock()

				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
					Header:     make(http.Header),
				}, nil
			default:
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Status:     "404 Not Found",
					Body:       io.NopCloser(strings.NewReader(`not found`)),
					Header:     make(http.Header),
				}, nil
			}
		}),
	}
}

func newPullResponse(t *testing.T, content map[string]interface{}) []byte {
	t.Helper()

	body, err := json.Marshal(map[string]interface{}{
		"content": content,
	})
	require.NoError(t, err)
	return body
}

func newHTTPTaskString(t *testing.T, name, targetURL, postURL string) string {
	t.Helper()

	task, err := json.Marshal(map[string]interface{}{
		"name":              name,
		"external_id":       name,
		"method":            "GET",
		"url":               targetURL,
		"post_url":          postURL,
		"status":            "ok",
		"frequency":         "1s",
		"region":            RegionID,
		"owner_external_id": "owner",
		"success_when": []map[string]interface{}{
			{
				"status_code": []map[string]string{
					{"is": "200"},
				},
			},
		},
	})
	require.NoError(t, err)

	return string(task)
}

func assertSelectedMeasurements(selected []string) func(pts []*point.Point) error {
	mtMap := map[string]struct {
		measurement    inputs.Measurement
		optionalFields []string
		optionalTags   []string
		extraTags      map[string]string
	}{
		"http_dial_testing": {
			measurement: &httpMeasurement{},
			optionalFields: []string{
				"fail_reason",
				"response_download",
				"response_connection",
				"response_ttfb",
				"response_ssl",
				"response_dns",
			},
		},
		"tcp_dial_testing": {
			measurement: &tcpMeasurement{},
			optionalFields: []string{
				"traceroute",
			},
		},
		"icmp_dial_testing": {
			measurement: &icmpMeasurement{},
			optionalFields: []string{
				"traceroute",
			},
		},
		"websocket_dial_testing": {
			measurement: &websocketMeasurement{},
		},
	}

	return func(pts []*point.Point) error {
		pointMap := map[string]bool{}
		for _, pt := range pts {
			name := pt.Name()
			m, ok := mtMap[name]
			if !ok || pointMap[name] {
				continue
			}

			extraTags := map[string]string{}
			for k, v := range m.extraTags {
				extraTags[k] = v
			}

			msgs := inputs.CheckPoint(pt,
				inputs.WithDoc(m.measurement),
				inputs.WithOptionalFields(m.optionalFields...),
				inputs.WithOptionalTags(m.optionalTags...),
				inputs.WithExtraTags(extraTags),
			)
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed with %d errors", name, len(msgs))
			}
			pointMap[name] = true
		}

		for _, item := range selected {
			if !pointMap[item] {
				return fmt.Errorf("measurement %s not found", item)
			}
		}

		return nil
	}
}

func withIntegrationEnv(t *testing.T, fn func(sender *mockSender)) {
	t.Helper()

	oldExit := datakit.Exit
	oldWorker := dialWorker
	sender := &mockSender{}
	datakit.Exit = cliutils.NewSem()
	dialWorker = &worker{
		sender:               sender,
		maxJobNumber:         DefaultWorkerMaxJobNumber,
		maxJobChanNumber:     DefaultWorkerChannelNumber,
		maxCachePointsNumber: DefaultWorkerCacheMaxPoints,
	}
	dialWorker.init()
	once = sync.Once{}
	collectPointsCache = nil

	defer func() {
		if datakit.Exit != nil {
			datakit.Exit.Close()
		}
		datakit.Exit = oldExit
		dialWorker = oldWorker
		once = sync.Once{}
	}()
	fn(sender)
}

func TestIntegrate(t *testing.T) {
	t.Run("mock server pull and dispatch produces HTTP point", func(t *testing.T) {
		withIntegrationEnv(t, func(sender *mockSender) {
			mockSrv := newMockDialtestingServer(t, nil)
			const taskName = "integrate-http-task"
			expectedPostURL := fmt.Sprintf("%s/v1/write/logging?token=test-token", mockSrv.URL())
			mockSrv.pullResponse = newPullResponse(t, map[string]interface{}{
				"region": map[string]interface{}{
					"name":    "region-cn",
					"name_en": "region-en",
					"isp":     "telecom",
				},
				"HTTP": []string{
					newHTTPTaskString(t, taskName, "http://127.0.0.1:1", fmt.Sprintf("%s?token=test-token", mockSrv.URL())),
				},
			})

			ipt := defaultInput()
			ipt.Server = mockSrv.URL()
			ipt.RegionID = RegionID
			ipt.AK = DialAK
			ipt.SK = DialSK
			ipt.cli = mockSrv.client(t)

			body, err := ipt.pullTask()
			require.NoError(t, err)
			require.NoError(t, ipt.dispatchTasks(body))

			assert.Eventually(t, func() bool {
				return len(sender.Points()) >= 1
			}, 5*time.Second, 50*time.Millisecond)

			points := sender.Points()
			require.NotEmpty(t, points)
			assert.Equal(t, 1, mockSrv.PullCount())
			assert.Equal(t, "region-cn", ipt.regionName)
			assert.Equal(t, "region-en", ipt.regionNameEn)
			assert.Equal(t, "telecom", ipt.RegionTags["isp"])

			var gotPoint *point.Point
			for _, pt := range points {
				lineProto := pt.LineProto()
				if pt.Name() == "http_dial_testing" && strings.Contains(lineProto, "name="+taskName) {
					gotPoint = pt
					break
				}
			}
			require.NotNil(t, gotPoint)

			lineProto := gotPoint.LineProto()
			assert.Contains(t, lineProto, "http_dial_testing,")
			assert.Contains(t, lineProto, "node_name=region-cn")
			assert.Contains(t, lineProto, "datakit_version=")
			assert.Contains(t, lineProto, "url=http://127.0.0.1:1")
			assert.Contains(t, lineProto, "seq_number=1i")
			assert.Contains(t, lineProto, "success=")
			assert.Contains(t, lineProto, "task=")

			urls := sender.URLs()
			require.NotEmpty(t, urls)
			assert.Contains(t, urls, expectedPostURL)

			ipt.Terminate()
		})
	})

	t.Run("mock server receives variable updates", func(t *testing.T) {
		withIntegrationEnv(t, func(sender *mockSender) {
			mockSrv := newMockDialtestingServer(t, newPullResponse(t, map[string]interface{}{
				"region": map[string]interface{}{
					"name": "region-cn",
				},
			}))

			ipt := defaultInput()
			ipt.Server = mockSrv.URL()
			ipt.RegionID = RegionID
			ipt.AK = DialAK
			ipt.SK = DialSK
			ipt.cli = mockSrv.client(t)
			ipt.variables.ipt = ipt
			ipt.variables.run()

			assert.Eventually(t, func() bool {
				return ipt.variables.reqURL != nil
			}, 3*time.Second, 20*time.Millisecond)

			ipt.variables.updateVariableValue(dt.Variable{
				UUID:            "var-1",
				OwnerExternalID: "owner-1",
			}, "value-1", 0)

			assert.Eventually(t, func() bool {
				return len(mockSrv.VariableUpdates()) == 1
			}, 5*time.Second, 20*time.Millisecond)

			updates := mockSrv.VariableUpdates()
			require.Len(t, updates, 1)

			var vars []dt.Variable
			require.NoError(t, json.Unmarshal(updates[0], &vars))
			require.Len(t, vars, 1)
			assert.Equal(t, "var-1", vars[0].UUID)
			assert.Equal(t, "owner-1", vars[0].OwnerExternalID)
			assert.Equal(t, "value-1", vars[0].Value)

			ipt.Terminate()
		})
	})

	t.Run("mock server updates existing task in place", func(t *testing.T) {
		withIntegrationEnv(t, func(sender *mockSender) {
			mockSrv := newMockDialtestingServer(t, nil)

			firstTask := newHTTPTaskString(t, "update-http-task", "http://127.0.0.1:1", fmt.Sprintf("%s?token=test-token", mockSrv.URL()))
			secondTask := newHTTPTaskString(t, "update-http-task", "http://127.0.0.1:1", fmt.Sprintf("%s?token=test-token", mockSrv.URL()))

			ipt := defaultInput()
			ipt.Server = mockSrv.URL()
			ipt.RegionID = RegionID
			ipt.AK = DialAK
			ipt.SK = DialSK
			ipt.cli = mockSrv.client(t)

			mockSrv.pullResponse = newPullResponse(t, map[string]interface{}{
				"HTTP": []string{firstTask},
			})
			body, err := ipt.pullTask()
			require.NoError(t, err)
			require.NoError(t, ipt.dispatchTasks(body))

			var created *dialer
			assert.Eventually(t, func() bool {
				value, ok := ipt.curTasks.Load("_update-http-task")
				if ok {
					created = value.(*dialer)
				}
				return ok
			}, 3*time.Second, 20*time.Millisecond)
			require.NotNil(t, created)

			mockSrv.pullResponse = newPullResponse(t, map[string]interface{}{
				"HTTP": []string{secondTask},
			})
			body, err = ipt.pullTask()
			require.NoError(t, err)
			require.NoError(t, ipt.dispatchTasks(body))

			value, ok := ipt.curTasks.Load("_update-http-task")
			require.True(t, ok)
			updated := value.(*dialer)
			assert.Same(t, created, updated)
			assert.Eventually(t, func() bool {
				return updated.task.GetExternalID() == "update-http-task"
			}, 2*time.Second, 20*time.Millisecond)

			ipt.Terminate()
		})
	})

	t.Run("mock server stop task removes existing dialer", func(t *testing.T) {
		withIntegrationEnv(t, func(sender *mockSender) {
			mockSrv := newMockDialtestingServer(t, nil)

			taskJSON := newHTTPTaskString(t, "stop-http-task", "http://127.0.0.1:1", fmt.Sprintf("%s?token=test-token", mockSrv.URL()))
			var taskPayload map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(taskJSON), &taskPayload))
			taskPayload["status"] = dt.StatusStop
			stopTaskJSONBytes, err := json.Marshal(taskPayload)
			require.NoError(t, err)

			ipt := defaultInput()
			ipt.Server = mockSrv.URL()
			ipt.RegionID = RegionID
			ipt.AK = DialAK
			ipt.SK = DialSK
			ipt.cli = mockSrv.client(t)

			mockSrv.pullResponse = newPullResponse(t, map[string]interface{}{
				"HTTP": []string{taskJSON},
			})
			body, err := ipt.pullTask()
			require.NoError(t, err)
			require.NoError(t, ipt.dispatchTasks(body))

			assert.Eventually(t, func() bool {
				_, ok := ipt.curTasks.Load("_stop-http-task")
				return ok
			}, 3*time.Second, 20*time.Millisecond)

			mockSrv.pullResponse = newPullResponse(t, map[string]interface{}{
				"HTTP": []string{string(stopTaskJSONBytes)},
			})
			body, err = ipt.pullTask()
			require.NoError(t, err)
			require.NoError(t, ipt.dispatchTasks(body))

			assert.Eventually(t, func() bool {
				_, ok := ipt.curTasks.Load("_stop-http-task")
				return !ok
			}, 3*time.Second, 20*time.Millisecond)

			ipt.Terminate()
		})
	})

	t.Run("mock server invalid region payload is tolerated", func(t *testing.T) {
		withIntegrationEnv(t, func(sender *mockSender) {
			mockSrv := newMockDialtestingServer(t, nil)
			mockSrv.pullResponse = newPullResponse(t, map[string]interface{}{
				"region": "invalid-region",
				"HTTP": []string{
					newHTTPTaskString(t, "invalid-region-task", "http://127.0.0.1:1", fmt.Sprintf("%s?token=test-token", mockSrv.URL())),
				},
			})

			ipt := defaultInput()
			ipt.Server = mockSrv.URL()
			ipt.RegionID = RegionID
			ipt.AK = DialAK
			ipt.SK = DialSK
			ipt.cli = mockSrv.client(t)

			body, err := ipt.pullTask()
			require.NoError(t, err)
			assert.NotPanics(t, func() {
				require.NoError(t, ipt.dispatchTasks(body))
			})

			assert.Eventually(t, func() bool {
				return len(sender.Points()) >= 1
			}, 5*time.Second, 50*time.Millisecond)

			ipt.Terminate()
		})
	})

	t.Run("mock server region disabled response stops all tasks", func(t *testing.T) {
		withIntegrationEnv(t, func(sender *mockSender) {
			mockSrv := newMockDialtestingServer(t, []byte(`kodo.RegionNotFoundOrDisabled`))
			mockSrv.pullStatusCode = http.StatusBadRequest

			ipt := defaultInput()
			ipt.Server = mockSrv.URL()
			ipt.RegionID = RegionID
			ipt.AK = DialAK
			ipt.SK = DialSK
			ipt.cli = mockSrv.client(t)

			task := newHTTPDialtestingTask(t, "region-disabled-task")
			d := newDialer(task, ipt)
			ipt.curTasks.Store(task.ID(), d)

			body, err := ipt.pullTask()
			assert.Error(t, err)
			assert.Nil(t, body)

			_, ok := ipt.curTasks.Load(task.ID())
			assert.False(t, ok)
		})
	})
}
