// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	CloudMonitoringRuntime = "cloud_monitoring"

	metricContainerUptime      = "kubernetes.io/container/uptime"
	metricContainerCPUUsage    = "kubernetes.io/container/cpu/core_usage_time"
	metricContainerCPULimit    = "kubernetes.io/container/cpu/limit_cores"
	metricContainerMemoryUsage = "kubernetes.io/container/memory/used_bytes"
	metricContainerMemoryLimit = "kubernetes.io/container/memory/limit_bytes"
	metricPodNetworkReceived   = "kubernetes.io/pod/network/received_bytes_count"
	metricPodNetworkSent       = "kubernetes.io/pod/network/sent_bytes_count"

	defaultMonitoringEndpoint = "https://monitoring.googleapis.com"
)

type CloudMonitoringConfig struct {
	ProjectID    string
	ClusterName  string
	Location     string
	Endpoint     string
	Window       time.Duration
	ActiveWithin time.Duration
}

type cloudMonitoringRuntime struct {
	cfg    CloudMonitoringConfig
	client *http.Client

	mu          sync.RWMutex
	containers  map[string]*Container
	tops        map[string]*ContainerTop
	refreshedAt time.Time
}

func NewCloudMonitoringRuntime(cfg CloudMonitoringConfig, client *http.Client) (ContainerRuntime, error) {
	if cfg.ProjectID == "" {
		return nil, fmt.Errorf("gcp project id is required")
	}
	if cfg.ClusterName == "" {
		return nil, fmt.Errorf("gke cluster name is required")
	}
	if cfg.Location == "" {
		return nil, fmt.Errorf("gke cluster location is required")
	}
	if client == nil {
		return nil, fmt.Errorf("http client is required")
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultMonitoringEndpoint
	}
	if cfg.Window <= 0 {
		cfg.Window = 10 * time.Minute
	}
	if cfg.ActiveWithin <= 0 {
		cfg.ActiveWithin = 5 * time.Minute
	}

	return &cloudMonitoringRuntime{
		cfg:        cfg,
		client:     client,
		containers: map[string]*Container{},
		tops:       map[string]*ContainerTop{},
	}, nil
}

func (c *cloudMonitoringRuntime) Version() (*VersionInfo, error) {
	return &VersionInfo{PlatformName: CloudMonitoringRuntime, APIVersion: "v3"}, nil
}

func (c *cloudMonitoringRuntime) ListContainers() ([]*Container, error) {
	c.mu.RLock()
	if !c.refreshedAt.IsZero() && time.Since(c.refreshedAt) < 30*time.Second {
		result := copyCloudContainers(c.containers)
		c.mu.RUnlock()
		return result, nil
	}
	c.mu.RUnlock()

	containers, tops, err := c.collect(context.Background(), time.Now())
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.containers = containers
	c.tops = tops
	c.refreshedAt = time.Now()
	c.mu.Unlock()

	return copyCloudContainers(containers), nil
}

func copyCloudContainers(containers map[string]*Container) []*Container {
	keys := make([]string, 0, len(containers))
	for key := range containers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]*Container, 0, len(keys))
	for _, key := range keys {
		item := *containers[key]
		item.Labels = copyStringMap(item.Labels)
		result = append(result, &item)
	}
	return result
}

func (c *cloudMonitoringRuntime) ContainerStatus(id string) (*ContainerStatus, error) {
	c.mu.RLock()
	item, ok := c.containers[id]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("container %q not found", id)
	}

	return &ContainerStatus{
		ID:    item.ID,
		Name:  item.Name,
		Image: item.Image,
	}, nil
}

func (c *cloudMonitoringRuntime) ContainerTop(id string) (*ContainerTop, error) {
	c.mu.RLock()
	top, ok := c.tops[id]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("container %q metrics not found", id)
	}

	result := *top
	return &result, nil
}

type cloudMetricSpec struct {
	metricType   string
	resourceType string
	metricFilter string
}

func (c *cloudMonitoringRuntime) collect(ctx context.Context, now time.Time) (map[string]*Container, map[string]*ContainerTop, error) {
	containers := map[string]*Container{}
	tops := map[string]*ContainerTop{}
	podNetwork := map[string]*ContainerTop{}

	specs := []cloudMetricSpec{
		{metricType: metricContainerUptime, resourceType: "k8s_container"},
		{metricType: metricContainerCPUUsage, resourceType: "k8s_container"},
		{metricType: metricContainerCPULimit, resourceType: "k8s_container"},
		{
			metricType:   metricContainerMemoryUsage,
			resourceType: "k8s_container",
			metricFilter: `metric.labels.memory_type = "non-evictable"`,
		},
		{metricType: metricContainerMemoryLimit, resourceType: "k8s_container"},
		{metricType: metricPodNetworkReceived, resourceType: "k8s_pod"},
		{metricType: metricPodNetworkSent, resourceType: "k8s_pod"},
	}

	var errs []string
	for _, spec := range specs {
		series, err := c.listTimeSeries(ctx, spec, now.Add(-c.cfg.Window), now)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", spec.metricType, err))
			continue
		}
		c.applySeries(
			spec.metricType,
			series,
			containers,
			tops,
			podNetwork,
			now.Add(-c.cfg.ActiveWithin),
		)
	}

	for id, item := range containers {
		top := ensureTop(tops, id)
		if item.CreatedAt == 0 {
			item.CreatedAt = now.UnixNano()
		}
		podKey := cloudPodKey(
			item.Labels["io.kubernetes.pod.namespace"],
			item.Labels["io.kubernetes.pod.name"],
		)
		if network, ok := podNetwork[podKey]; ok {
			top.NetworkRcvd = network.NetworkRcvd
			top.NetworkSent = network.NetworkSent
		}
	}

	if len(containers) == 0 && len(errs) != 0 {
		return nil, nil, fmt.Errorf("query cloud monitoring: %s", strings.Join(errs, "; "))
	}
	return containers, tops, nil
}

func (c *cloudMonitoringRuntime) applySeries(
	metricType string,
	series []cloudTimeSeries,
	containers map[string]*Container,
	tops map[string]*ContainerTop,
	podNetwork map[string]*ContainerTop,
	activeCutoff time.Time,
) {
	for i := range series {
		item := &series[i]
		namespace := item.Resource.Labels["namespace_name"]
		podName := item.Resource.Labels["pod_name"]

		if item.Resource.Type == "k8s_pod" {
			key := cloudPodKey(namespace, podName)
			top := ensureTop(podNetwork, key)
			value, _, ok := latestCloudPoint(item.Points)
			if !ok {
				continue
			}
			switch metricType {
			case metricPodNetworkReceived:
				top.NetworkRcvd += int64(value)
			case metricPodNetworkSent:
				top.NetworkSent += int64(value)
			}
			continue
		}

		containerName := item.Resource.Labels["container_name"]
		if namespace == "" || podName == "" || containerName == "" {
			continue
		}

		id := cloudContainerID(c.cfg.ClusterName, namespace, podName, containerName)
		if metricType == metricContainerUptime {
			value, end, ok := latestCloudPoint(item.Points)
			if !ok || value < 0 || end.Before(activeCutoff) {
				continue
			}
			container := ensureCloudContainer(containers, id, namespace, podName, containerName)
			container.CreatedAt = end.Add(-time.Duration(value * float64(time.Second))).UnixNano()
			ensureTop(tops, id)
			continue
		}

		if _, ok := containers[id]; !ok {
			continue
		}
		top := ensureTop(tops, id)

		switch metricType {
		case metricContainerCPUUsage:
			if cores, ok := cloudPointRate(item.Points); ok {
				top.CPUPercent = cores * 100
				top.CPUUsageMillicores = int64(cores * 1000)
			}
		case metricContainerCPULimit:
			if value, _, ok := latestCloudPoint(item.Points); ok {
				top.CPULimitMillicores = int64(value * 1000)
			}
		case metricContainerMemoryUsage:
			if value, _, ok := latestCloudPoint(item.Points); ok {
				top.MemoryWorkingSet = int64(value)
			}
		case metricContainerMemoryLimit:
			if value, _, ok := latestCloudPoint(item.Points); ok {
				top.MemoryLimitInBytes = int64(value)
			}
		}
	}
}

func ensureCloudContainer(containers map[string]*Container, id, namespace, podName, containerName string) *Container {
	if item, ok := containers[id]; ok {
		return item
	}

	item := &Container{
		ID:             id,
		Name:           containerName,
		RuntimeName:    CloudMonitoringRuntime,
		RuntimeVersion: "v3",
		State:          "Running",
		Status:         "Running",
		Labels: map[string]string{
			"io.kubernetes.container.name": containerName,
			"io.kubernetes.pod.name":       podName,
			"io.kubernetes.pod.namespace":  namespace,
		},
	}
	containers[id] = item
	return item
}

func ensureTop(tops map[string]*ContainerTop, id string) *ContainerTop {
	if top, ok := tops[id]; ok {
		return top
	}
	top := &ContainerTop{ID: id}
	tops[id] = top
	return top
}

func cloudContainerID(cluster, namespace, podName, containerName string) string {
	return strings.Join([]string{CloudMonitoringRuntime, cluster, namespace, podName, containerName}, "/")
}

func cloudPodKey(namespace, podName string) string {
	return namespace + "/" + podName
}

func copyStringMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

type cloudTimeSeriesListResponse struct {
	TimeSeries    []cloudTimeSeries `json:"timeSeries"`
	NextPageToken string            `json:"nextPageToken"`
}

type cloudTimeSeries struct {
	Metric struct {
		Type   string            `json:"type"`
		Labels map[string]string `json:"labels"`
	} `json:"metric"`
	Resource struct {
		Type   string            `json:"type"`
		Labels map[string]string `json:"labels"`
	} `json:"resource"`
	Points []cloudPoint `json:"points"`
}

type cloudPoint struct {
	Interval struct {
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	} `json:"interval"`
	Value struct {
		DoubleValue *float64        `json:"doubleValue"`
		Int64Value  json.RawMessage `json:"int64Value"`
	} `json:"value"`
}

func (c *cloudMonitoringRuntime) listTimeSeries(
	ctx context.Context,
	spec cloudMetricSpec,
	start, end time.Time,
) ([]cloudTimeSeries, error) {
	var result []cloudTimeSeries
	pageToken := ""

	for {
		u, err := url.Parse(strings.TrimRight(c.cfg.Endpoint, "/") +
			"/v3/projects/" + url.PathEscape(c.cfg.ProjectID) + "/timeSeries")
		if err != nil {
			return nil, fmt.Errorf("build monitoring url: %w", err)
		}

		filter := fmt.Sprintf(
			`metric.type = %q AND resource.type = %q AND resource.labels.cluster_name = %q AND resource.labels.location = %q`,
			spec.metricType, spec.resourceType, c.cfg.ClusterName, c.cfg.Location,
		)
		if spec.metricFilter != "" {
			filter += " AND " + spec.metricFilter
		}
		query := u.Query()
		query.Set("filter", filter)
		query.Set("interval.startTime", start.UTC().Format(time.RFC3339Nano))
		query.Set("interval.endTime", end.UTC().Format(time.RFC3339Nano))
		query.Set("view", "FULL")
		query.Set("pageSize", "1000")
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		u.RawQuery = query.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("create monitoring request: %w", err)
		}
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("send monitoring request: %w", err)
		}

		var body cloudTimeSeriesListResponse
		decodeErr := decodeCloudResponse(resp, &body)
		if decodeErr != nil {
			return nil, decodeErr
		}
		result = append(result, body.TimeSeries...)

		if body.NextPageToken == "" {
			return result, nil
		}
		pageToken = body.NextPageToken
	}
}

func decodeCloudResponse(resp *http.Response, dst any) error {
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		return fmt.Errorf("cloud api returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decode cloud api response: %w", err)
	}
	return nil
}

func latestCloudPoint(points []cloudPoint) (float64, time.Time, bool) {
	var (
		found bool
		value float64
		end   time.Time
	)
	for i := range points {
		pointEnd, err := time.Parse(time.RFC3339Nano, points[i].Interval.EndTime)
		if err != nil {
			continue
		}
		pointValue, ok := cloudPointValue(points[i])
		if !ok {
			continue
		}
		if !found || pointEnd.After(end) {
			found = true
			value = pointValue
			end = pointEnd
		}
	}
	return value, end, found
}

func cloudPointRate(points []cloudPoint) (float64, bool) {
	type valueAt struct {
		value float64
		end   time.Time
		start string
	}
	values := make([]valueAt, 0, len(points))
	for i := range points {
		end, err := time.Parse(time.RFC3339Nano, points[i].Interval.EndTime)
		if err != nil {
			continue
		}
		value, ok := cloudPointValue(points[i])
		if ok {
			values = append(values, valueAt{
				value: value,
				end:   end,
				start: points[i].Interval.StartTime,
			})
		}
	}
	if len(values) < 2 {
		return 0, false
	}
	sort.Slice(values, func(i, j int) bool { return values[i].end.Before(values[j].end) })

	latest := values[len(values)-1]
	previous := values[len(values)-2]
	seconds := latest.end.Sub(previous.end).Seconds()
	if seconds <= 0 || latest.value < previous.value ||
		(latest.start != "" && previous.start != "" && latest.start != previous.start) {
		return 0, false
	}
	return (latest.value - previous.value) / seconds, true
}

func cloudPointValue(point cloudPoint) (float64, bool) {
	if point.Value.DoubleValue != nil {
		return *point.Value.DoubleValue, true
	}
	if len(point.Value.Int64Value) == 0 {
		return 0, false
	}

	raw := strings.Trim(string(point.Value.Int64Value), `"`)
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false
	}
	return value, true
}
