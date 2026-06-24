// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gcpauth "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cloudprovider/gcp/auth"
	gcpmetadata "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cloudprovider/gcp/metadata"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

type cloudLoggingTestFeeder struct {
	feeds  atomic.Int64
	points []*point.Point
	err    error
}

func (f *cloudLoggingTestFeeder) Feed(_ point.Category, points []*point.Point, _ ...dkio.FeedOption) error {
	if f.err != nil {
		return f.err
	}
	f.feeds.Add(1)
	f.points = append(f.points, points...)
	return nil
}

func (f *cloudLoggingTestFeeder) FeedLastError(string, ...metrics.LastErrorOption) {}

func TestGCPCloudLoggingDeduplication(t *testing.T) {
	timestamp := time.Date(2026, 6, 9, 1, 2, 3, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request cloudLoggingListRequest
		if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&request)) {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		assert.Equal(t, "timestamp desc", request.OrderBy)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{
			"entries":[{
				"logName":"projects/project-1/logs/stdout",
				"timestamp":%q,
				"insertId":"insert-1",
				"severity":"INFO",
				"textPayload":"hello from autopilot",
				"resource":{"type":"k8s_container","labels":{
					"cluster_name":"cluster-1",
					"location":"asia-southeast1",
					"namespace_name":"default",
					"pod_name":"pod-1",
					"container_name":"app"
				}}
			}]
		}`, timestamp.Format(time.RFC3339Nano))
	}))
	defer server.Close()

	stateFile := filepath.Join(t.TempDir(), "state.json")
	feeder := &cloudLoggingTestFeeder{}
	ipt := newInput()
	ipt.GCPProjectID = "project-1"
	ipt.GCPClusterName = "cluster-1"
	ipt.GCPClusterLocation = "asia-southeast1"
	ipt.GCPCloudLoggingStateFile = stateFile
	ipt.Feeder = feeder

	collector, err := newGCPCloudLoggingCollector(ipt, server.Client(), nil)
	require.NoError(t, err)
	cloudCollector := collector.(*gcpCloudLoggingCollector)
	cloudCollector.endpoint = server.URL

	cloudCollector.collect()
	cloudCollector.collect()

	assert.Equal(t, int64(1), feeder.feeds.Load())
	require.Len(t, feeder.points, 1)
	assert.Equal(t, "hello from autopilot", feeder.points[0].KVs().Get("message").GetS())
	assert.Equal(t, "app", feeder.points[0].Name())
	assert.Equal(t, timestamp.UnixNano(), feeder.points[0].Time().UnixNano())
	assert.FileExists(t, stateFile)

	restartedFeeder := &cloudLoggingTestFeeder{}
	ipt.Feeder = restartedFeeder
	restarted, err := newGCPCloudLoggingCollector(ipt, server.Client(), nil)
	require.NoError(t, err)
	restarted.(*gcpCloudLoggingCollector).endpoint = server.URL
	restarted.(*gcpCloudLoggingCollector).collect()
	assert.Zero(t, restartedFeeder.feeds.Load())
}

func TestGCPCloudLoggingWatermarkUsesCompletedQueryTime(t *testing.T) {
	now := time.Date(2026, 6, 9, 2, 0, 0, 0, time.UTC)
	collector := &gcpCloudLoggingCollector{}

	collector.advanceWatermark(now)
	assert.Equal(t, now, collector.state.Watermark)

	collector.advanceWatermark(now.Add(-time.Minute))
	assert.Equal(t, now, collector.state.Watermark)
}

func TestGCPMetadataTokenTransportCachesToken(t *testing.T) {
	var tokenRequests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			tokenRequests.Add(1)
			assert.Equal(t, "Google", r.Header.Get("Metadata-Flavor"))
			_, _ = io.WriteString(w, `{"access_token":"token-1","expires_in":3600,"token_type":"Bearer"}`)
		case "/api":
			assert.Equal(t, "Bearer token-1", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	transport := gcpauth.NewMetadataTokenTransport(server.Client().Transport, server.URL+"/token", time.Now)
	client := &http.Client{Transport: transport}
	for i := 0; i < 2; i++ {
		resp, err := client.Get(server.URL + "/api")
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	}
	assert.Equal(t, int64(1), tokenRequests.Load())
}

func TestResolveGCPCloudConfigFromMetadata(t *testing.T) {
	t.Setenv("ENV_CLUSTER_NAME_K8S", "default")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Google", r.Header.Get("Metadata-Flavor"))
		switch r.URL.Path {
		case "/project/project-id":
			_, _ = io.WriteString(w, "project-1")
		case "/instance/attributes/cluster-name":
			_, _ = io.WriteString(w, "cluster-1")
		case "/instance/attributes/cluster-location":
			_, _ = io.WriteString(w, "asia-southeast1")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	ipt := newInput()
	err := resolveGCPCloudConfig(context.Background(), ipt, gcpmetadata.NewClientWithHTTP(server.URL, server.Client()))
	require.NoError(t, err)
	assert.Equal(t, "project-1", ipt.GCPProjectID)
	assert.Equal(t, "cluster-1", ipt.GCPClusterName)
	assert.Equal(t, "asia-southeast1", ipt.GCPClusterLocation)
	assert.Equal(t, "cluster-1", getClusterNameK8s())
}

func TestResolveGCPCloudConfigPreservesOverrides(t *testing.T) {
	t.Setenv("ENV_CLUSTER_NAME_K8S", "configured-cluster")

	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		http.Error(w, "metadata should not be requested", http.StatusInternalServerError)
	}))
	defer server.Close()

	ipt := newInput()
	ipt.GCPProjectID = "configured-project"
	ipt.GCPClusterName = "configured-cluster"
	ipt.GCPClusterLocation = "configured-location"

	err := resolveGCPCloudConfig(context.Background(), ipt, gcpmetadata.NewClientWithHTTP(server.URL, server.Client()))
	require.NoError(t, err)
	assert.Zero(t, requests.Load())
	assert.Equal(t, "configured-cluster", getClusterNameK8s())
}

func TestResolveGCPCloudConfigMetadataError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not available", http.StatusNotFound)
	}))
	defer server.Close()

	err := resolveGCPCloudConfig(context.Background(), newInput(), gcpmetadata.NewClientWithHTTP(server.URL, server.Client()))
	require.ErrorContains(t, err, "discover gcp project id")
	require.ErrorContains(t, err, "404 Not Found")
}
