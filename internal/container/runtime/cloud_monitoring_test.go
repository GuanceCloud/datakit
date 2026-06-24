// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudMonitoringRuntime(t *testing.T) {
	t1 := time.Now().UTC().Truncate(time.Second).Add(-time.Minute)
	t2 := t1.Add(time.Minute)
	var requests atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		filter := r.URL.Query().Get("filter")
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(filter, metricContainerMemoryUsage) {
			assert.Contains(t, filter, `metric.labels.memory_type = "non-evictable"`)
		}

		switch {
		case strings.Contains(filter, metricContainerUptime):
			_, _ = fmt.Fprintf(w, `{"timeSeries":[%s]}`, containerSeries(
				metricContainerUptime, "DOUBLE", t2, "120", true))
		case strings.Contains(filter, metricContainerCPUUsage):
			_, _ = fmt.Fprintf(w, `{"timeSeries":[%s]}`, containerCPUTimeSeries(t1, t2))
		case strings.Contains(filter, metricContainerCPULimit):
			_, _ = fmt.Fprintf(w, `{"timeSeries":[%s]}`, containerSeries(
				metricContainerCPULimit, "DOUBLE", t2, "2", true))
		case strings.Contains(filter, metricContainerMemoryUsage):
			_, _ = fmt.Fprintf(w, `{"timeSeries":[%s]}`, containerSeries(
				metricContainerMemoryUsage, "INT64", t2, "1048576", false))
		case strings.Contains(filter, metricContainerMemoryLimit):
			_, _ = fmt.Fprintf(w, `{"timeSeries":[%s]}`, containerSeries(
				metricContainerMemoryLimit, "INT64", t2, "2097152", false))
		case strings.Contains(filter, metricPodNetworkReceived):
			_, _ = fmt.Fprintf(w, `{"timeSeries":[%s,%s]}`,
				podSeries(metricPodNetworkReceived, t2, "100"),
				podSeries(metricPodNetworkReceived, t2, "25"))
		case strings.Contains(filter, metricPodNetworkSent):
			_, _ = fmt.Fprintf(w, `{"timeSeries":[%s]}`,
				podSeries(metricPodNetworkSent, t2, "75"))
		default:
			http.Error(w, "unexpected filter", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	rt, err := NewCloudMonitoringRuntime(CloudMonitoringConfig{
		ProjectID:   "project-1",
		ClusterName: "cluster-1",
		Location:    "asia-southeast1",
		Endpoint:    server.URL,
	}, server.Client())
	require.NoError(t, err)

	containers, err := rt.ListContainers()
	require.NoError(t, err)
	require.Len(t, containers, 1)
	assert.Equal(t, "app", containers[0].Name)
	assert.Equal(t, CloudMonitoringRuntime, containers[0].RuntimeName)
	assert.Equal(t, "pod-1", containers[0].Labels["io.kubernetes.pod.name"])
	assert.Equal(t, t2.Add(-120*time.Second).Unix(), time.Unix(0, containers[0].CreatedAt).Unix())

	top, err := rt.ContainerTop(containers[0].ID)
	require.NoError(t, err)
	assert.InDelta(t, 10, top.CPUPercent, 0.001)
	assert.Equal(t, int64(100), top.CPUUsageMillicores)
	assert.Equal(t, int64(2000), top.CPULimitMillicores)
	assert.Equal(t, int64(1048576), top.MemoryWorkingSet)
	assert.Equal(t, int64(2097152), top.MemoryLimitInBytes)
	assert.Equal(t, int64(125), top.NetworkRcvd)
	assert.Equal(t, int64(75), top.NetworkSent)

	_, err = rt.ListContainers()
	require.NoError(t, err)
	assert.Equal(t, int64(7), requests.Load(), "the second list should use the short-lived cache")
}

func TestCloudMonitoringRuntimeIgnoresInactiveContainers(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		filter := r.URL.Query().Get("filter")
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(filter, metricContainerUptime):
			_, _ = fmt.Fprintf(w, `{"timeSeries":[%s]}`, containerSeries(
				metricContainerUptime, "DOUBLE", now.Add(-10*time.Minute), "120", true))
		case strings.Contains(filter, metricContainerMemoryUsage):
			_, _ = fmt.Fprintf(w, `{"timeSeries":[%s]}`, containerSeries(
				metricContainerMemoryUsage, "INT64", now, "1048576", false))
		default:
			_, _ = fmt.Fprint(w, `{"timeSeries":[]}`)
		}
	}))
	defer server.Close()

	rt, err := NewCloudMonitoringRuntime(CloudMonitoringConfig{
		ProjectID:    "project-1",
		ClusterName:  "cluster-1",
		Location:     "asia-southeast1",
		Endpoint:     server.URL,
		ActiveWithin: 5 * time.Minute,
	}, server.Client())
	require.NoError(t, err)

	containers, err := rt.ListContainers()
	require.NoError(t, err)
	assert.Empty(t, containers)

	containers, err = rt.ListContainers()
	require.NoError(t, err)
	assert.Empty(t, containers)
	assert.Equal(t, int64(7), requests.Load(), "an empty result should also use the short-lived cache")
}

func containerSeries(metricType, valueType string, end time.Time, value string, doubleValue bool) string {
	valueJSON := fmt.Sprintf(`"int64Value":%q`, value)
	if doubleValue {
		valueJSON = fmt.Sprintf(`"doubleValue":%s`, value)
	}
	return fmt.Sprintf(`{
		"metric":{"type":%q},
		"resource":{"type":"k8s_container","labels":{
			"namespace_name":"default","pod_name":"pod-1","container_name":"app"
		}},
		"valueType":%q,
		"points":[{"interval":{"endTime":%q},"value":{%s}}]
	}`, metricType, valueType, end.Format(time.RFC3339Nano), valueJSON)
}

func containerCPUTimeSeries(start, end time.Time) string {
	return fmt.Sprintf(`{
		"metric":{"type":%q},
		"resource":{"type":"k8s_container","labels":{
			"namespace_name":"default","pod_name":"pod-1","container_name":"app"
		}},
		"points":[
			{"interval":{"endTime":%q},"value":{"doubleValue":20}},
			{"interval":{"endTime":%q},"value":{"doubleValue":14}}
		]
	}`, metricContainerCPUUsage, end.Format(time.RFC3339Nano), start.Format(time.RFC3339Nano))
}

func podSeries(metricType string, end time.Time, value string) string {
	return fmt.Sprintf(`{
		"metric":{"type":%q},
		"resource":{"type":"k8s_pod","labels":{
			"namespace_name":"default","pod_name":"pod-1"
		}},
		"points":[{"interval":{"endTime":%q},"value":{"int64Value":%q}}]
	}`, metricType, end.Format(time.RFC3339Nano), value)
}
