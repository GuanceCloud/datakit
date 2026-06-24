// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/lang"
	gcplogging "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cloudprovider/gcp/logging"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/podutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/ansi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type cloudLoggingState = gcplogging.State
type cloudLoggingListRequest = gcplogging.ListRequest
type cloudLogEntry = gcplogging.Entry

type gcpCloudLoggingCollector struct {
	ipt       *Input
	client    *http.Client
	k8sClient k8sclient.Client
	endpoint  string

	defaults  *loggingDefaults
	logFilter filter.Filter
	state     cloudLoggingState
}

type cloudLogFeedGroup struct {
	source       string
	pipeline     string
	storageIndex string
	points       []*point.Point
}

func newGCPCloudLoggingCollector(ipt *Input, httpClient *http.Client, k8sClient k8sclient.Client) (Collector, error) {
	if ipt.GCPProjectID == "" {
		return nil, fmt.Errorf("gcp project id is required")
	}
	if ipt.GCPClusterName == "" {
		return nil, fmt.Errorf("gke cluster name is required")
	}
	if ipt.GCPClusterLocation == "" {
		return nil, fmt.Errorf("gke cluster location is required")
	}
	if httpClient == nil {
		return nil, fmt.Errorf("http client is required")
	}

	logFilter, err := createLogFilter(ipt)
	if err != nil {
		return nil, err
	}

	collector := &gcpCloudLoggingCollector{
		ipt:       ipt,
		client:    httpClient,
		k8sClient: k8sClient,
		endpoint:  gcplogging.DefaultEndpoint,
		defaults:  newLoggingDefaults(ipt),
		logFilter: logFilter,
		state: cloudLoggingState{
			Seen: map[string]int64{},
		},
	}
	collector.defaults.extraTags = inputs.MergeTags(ipt.Tagger.ElectionTags(), ipt.Tags, "")
	if err := collector.loadState(); err != nil {
		l.Warnf("load GCP Cloud Logging state failed: %s", err)
	}
	return collector, nil
}

func (c *gcpCloudLoggingCollector) StartCollect() {
	activationTicker := time.NewTicker(5 * time.Second)
	loggingTicker := time.NewTicker(c.ipt.LoggingSearchInterval)
	defer activationTicker.Stop()
	defer loggingTicker.Stop()

	active := false
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("cloud logging collector stopped")
			return

		case <-activationTicker.C:
			if !c.ipt.leaderGate().Allowed() {
				active = false
				continue
			}
			if !active {
				active = true
				c.collect()
			}

		case <-loggingTicker.C:
			if !c.ipt.leaderGate().Allowed() {
				continue
			}
			c.collect()
		}
	}
}

func (c *gcpCloudLoggingCollector) collect() {
	now := time.Now()
	start := now.Add(-c.ipt.GCPCloudLoggingLookback)
	if !c.state.Watermark.IsZero() {
		start = c.state.Watermark.Add(-c.ipt.GCPCloudLoggingOverlap)
	}

	entries, err := c.listEntries(context.Background(), start, now)
	if err != nil {
		l.Warnf("query GCP Cloud Logging failed: %s", err)
		return
	}
	if len(entries) == 0 {
		c.advanceWatermark(now)
		c.pruneSeen(now)
		if err := c.saveState(); err != nil {
			l.Warnf("save GCP Cloud Logging state failed: %s", err)
		}
		return
	}

	podCache := map[string]*apicorev1.Pod{}
	groups := map[string]*cloudLogFeedGroup{}
	stagedSeen := map[string]int64{}

	for i := range entries {
		entry := &entries[i]
		timestamp, ok := cloudLogTimestamp(entry)
		if !ok {
			continue
		}

		key := cloudLogEntryKey(entry)
		if _, ok := c.state.Seen[key]; ok {
			continue
		}
		if _, ok := stagedSeen[key]; ok {
			continue
		}

		pod := c.lookupPod(entry, podCache)
		cfg, info, ok := c.resolveLogConfig(entry, pod)
		if !ok {
			continue
		}

		message := cloudLogMessage(entry)
		if message == "" {
			continue
		}
		if cfg.RemoveAnsiEscapeCodes || c.defaults.removeAnsiEscapeCodes {
			message = string(ansi.Strip([]byte(message)))
		}

		pt := c.buildLogPoint(entry, timestamp, message, cfg, info)
		stagedSeen[key] = now.Unix()
		groupKey := strings.Join([]string{cfg.Source, cfg.Pipeline, cfg.StorageIndex}, "\x00")
		group := groups[groupKey]
		if group == nil {
			group = &cloudLogFeedGroup{
				source:       cfg.Source,
				pipeline:     cfg.Pipeline,
				storageIndex: cfg.StorageIndex,
			}
			groups[groupKey] = group
		}
		group.points = append(group.points, pt)
	}

	if err := c.feedGroups(groups); err != nil {
		l.Warnf("feed GCP Cloud Logging failed: %s", err)
		return
	}

	c.advanceWatermark(now)
	for key, seenAt := range stagedSeen {
		c.state.Seen[key] = seenAt
	}
	c.pruneSeen(now)
	if err := c.saveState(); err != nil {
		l.Warnf("save GCP Cloud Logging state failed: %s", err)
	}
}

func (c *gcpCloudLoggingCollector) advanceWatermark(now time.Time) {
	gcplogging.AdvanceWatermark(&c.state, now)
}

func (c *gcpCloudLoggingCollector) listEntries(ctx context.Context, start, end time.Time) ([]cloudLogEntry, error) {
	return gcplogging.ListEntries(ctx, c.client, gcplogging.Config{
		ProjectID:   c.ipt.GCPProjectID,
		ClusterName: c.ipt.GCPClusterName,
		Location:    c.ipt.GCPClusterLocation,
		Endpoint:    c.endpoint,
	}, start, end)
}

func (c *gcpCloudLoggingCollector) lookupPod(entry *cloudLogEntry, cache map[string]*apicorev1.Pod) *apicorev1.Pod {
	if c.k8sClient == nil {
		return nil
	}
	namespace := entry.Resource.Labels["namespace_name"]
	podName := entry.Resource.Labels["pod_name"]
	if namespace == "" || podName == "" {
		return nil
	}

	key := namespace + "/" + podName
	if pod, ok := cache[key]; ok {
		return pod
	}

	pod, err := c.k8sClient.GetPods(namespace).Get(
		context.Background(),
		podName,
		metav1.GetOptions{ResourceVersion: "0"},
	)
	if err != nil {
		l.Debugf("query pod %s for cloud logging failed: %s", key, err)
		cache[key] = nil
		return nil
	}
	cache[key] = pod
	return pod
}

func (c *gcpCloudLoggingCollector) resolveLogConfig(
	entry *cloudLogEntry,
	pod *apicorev1.Pod,
) (*logConfig, *containerLogInfo, bool) {
	containerName := entry.Resource.Labels["container_name"]
	namespace := entry.Resource.Labels["namespace_name"]
	podName := entry.Resource.Labels["pod_name"]

	info := &containerLogInfo{
		containerID:   cloudLogContainerID(c.ipt.GCPClusterName, namespace, podName, containerName),
		containerName: containerName,
		runtime:       runtime.CloudMonitoringRuntime,
		podName:       podName,
		podNamespace:  namespace,
	}
	if pod != nil {
		info.podUID = string(pod.UID)
		info.podIP = pod.Status.PodIP
		info.podLabels = pod.Labels
		info.ownerKind, info.ownerName = podutil.PodOwner(pod)
		info.image = podutil.ContainerImageFromPod(containerName, pod)
		if id := containerIDFromPod(containerName, pod); id != "" {
			info.containerID = id
		}
	}

	if !c.logFilter.Match(filter.FilterImage, info.image) ||
		!c.logFilter.Match(filter.FilterNamespace, namespace) {
		return nil, nil, false
	}

	cfg := &logConfig{Type: "cloud_logging", Source: containerName}
	if pod != nil {
		configText := pod.Annotations[fmt.Sprintf(logConfigAnnotationKeyFormat, "")]
		if specific := pod.Annotations[fmt.Sprintf(logConfigAnnotationKeyFormat, containerName+".")]; specific != "" {
			configText = specific
		}
		if configText != "" {
			parsed, ok := cloudStdoutLogConfig(configText)
			if !ok {
				return nil, nil, false
			}
			cfg = parsed
		}
	}

	if cfg.Disable {
		return nil, nil, false
	}
	cfg.fillDefaultSource(info)
	if cfg.Service == "" {
		cfg.Service = cfg.Source
	}
	cfg.addTags(info.buildTags())
	cfg.addTags(c.defaults.extraTags)
	cfg.addTags(c.defaults.setLabelAsTags(info.podLabels))
	if pod != nil && pod.Spec.NodeName != "" {
		cfg.addTags(map[string]string{"node_name": pod.Spec.NodeName, "host": pod.Spec.NodeName})
	}
	cfg.replacedTagsKey()
	cfg.setExtraSourceMap(c.defaults)
	return cfg, info, true
}

func cloudStdoutLogConfig(configText string) (*logConfig, bool) {
	var configs []*logConfig
	if err := json.Unmarshal([]byte(configText), &configs); err != nil {
		l.Warnf("parse cloud logging annotation failed: %s", err)
		return nil, false
	}
	for _, cfg := range configs {
		if cfg == nil {
			continue
		}
		switch cfg.Type {
		case "", "stdout", "cloud_logging", runtime.DockerRuntime, "containerd", "crio", "cri-o":
			return cfg, true
		}
	}
	return nil, false
}

func (c *gcpCloudLoggingCollector) buildLogPoint(
	entry *cloudLogEntry,
	timestamp time.Time,
	message string,
	cfg *logConfig,
	info *containerLogInfo,
) *point.Point {
	var kvs point.KVs
	kvs = kvs.Add("message", message)
	kvs = kvs.AddTag("status", cloudLogStatus(entry.Severity))
	kvs = kvs.AddTag("service", cfg.Service)
	kvs = kvs.AddTag("stream", cloudLogStream(entry.LogName))
	kvs = kvs.Add("filepath", entry.LogName)
	kvs = kvs.Add("gcp_insert_id", entry.InsertID)
	kvs = kvs.AddTag("gcp_project_id", c.ipt.GCPProjectID)
	kvs = kvs.AddTag("gcp_location", c.ipt.GCPClusterLocation)
	kvs = kvs.AddTag("cluster_name_k8s", c.ipt.GCPClusterName)
	if info.podName != "" && info.podIP != "" {
		kvs = kvs.AddTag("pod_ip", info.podIP)
	}
	for key, value := range cfg.Tags {
		kvs = kvs.AddTag(key, value)
	}

	return point.NewPoint(
		cfg.Source,
		kvs,
		append(point.DefaultLoggingOptions(), point.WithTime(timestamp))...,
	)
}

func (c *gcpCloudLoggingCollector) feedGroups(groups map[string]*cloudLogFeedGroup) error {
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		group := groups[key]
		if len(group.points) == 0 {
			continue
		}

		feedName := dkio.FeedSource("logging", group.source)
		opts := []dkio.FeedOption{
			dkio.WithElection(true),
			dkio.WithSource(feedName),
		}
		if group.storageIndex != "" {
			feedName = dkio.FeedSource(feedName, group.storageIndex)
			opts[1] = dkio.WithSource(feedName)
			opts = append(opts, dkio.WithStorageIndex(group.storageIndex))
		}
		if group.pipeline != "" {
			opts = append(opts, dkio.WithPipelineOption(&lang.LogOption{
				ScriptMap: map[string]string{group.source: group.pipeline},
			}))
		}

		if err := c.ipt.Feeder.Feed(point.Logging, group.points, opts...); err != nil {
			return err
		}
		collectPtsVec.WithLabelValues("cloud-logging").Add(float64(len(group.points)))
	}
	return nil
}

func (c *gcpCloudLoggingCollector) loadState() error {
	state, err := gcplogging.LoadState(c.ipt.GCPCloudLoggingStateFile)
	if err != nil {
		return err
	}
	c.state = state
	return nil
}

func (c *gcpCloudLoggingCollector) saveState() error {
	return gcplogging.SaveState(c.ipt.GCPCloudLoggingStateFile, c.state)
}

func (c *gcpCloudLoggingCollector) pruneSeen(now time.Time) {
	gcplogging.PruneSeen(&c.state, now, c.ipt.GCPCloudLoggingOverlap, gcplogging.MaxSeenEntries)
}

func cloudLogTimestamp(entry *cloudLogEntry) (time.Time, bool) {
	return gcplogging.Timestamp(entry)
}

func cloudLogEntryKey(entry *cloudLogEntry) string {
	return gcplogging.EntryKey(entry)
}

func cloudLogMessage(entry *cloudLogEntry) string {
	return gcplogging.Message(entry)
}

func cloudLogStatus(severity string) string {
	switch strings.ToUpper(severity) {
	case "DEBUG":
		return "debug"
	case "WARNING":
		return "warning"
	case "ERROR":
		return "error"
	case "CRITICAL":
		return "critical"
	case "ALERT":
		return "alert"
	case "EMERGENCY":
		return "emerg"
	case "NOTICE":
		return "notice"
	default:
		return pipeline.DefaultStatus
	}
}

func cloudLogStream(logName string) string {
	if strings.HasSuffix(logName, "/stderr") {
		return "stderr"
	}
	return "stdout"
}

func cloudLogContainerID(cluster, namespace, podName, containerName string) string {
	return strings.Join([]string{runtime.CloudMonitoringRuntime, cluster, namespace, podName, containerName}, "/")
}

func containerIDFromPod(containerName string, pod *apicorev1.Pod) string {
	statuses := append([]apicorev1.ContainerStatus{}, pod.Status.InitContainerStatuses...)
	statuses = append(statuses, pod.Status.ContainerStatuses...)
	for _, status := range statuses {
		if status.Name != containerName || status.ContainerID == "" {
			continue
		}
		if idx := strings.Index(status.ContainerID, "://"); idx >= 0 {
			return status.ContainerID[idx+3:]
		}
		return status.ContainerID
	}
	return ""
}
