// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kubernetes collect resources metric/object/event.
package kubernetes

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	defaultChangeLanguage = changes.LangEn
	watchRetryInterval    = time.Second * 10
	klog                  = logger.DefaultSLogger("k8s")
)

type k8sClient k8sclient.Client

type Config struct {
	NodeName         string
	NodeLocal        bool
	EnableK8sMetric  bool
	EnableK8sObject  bool
	EnableK8sEvent   bool
	EnablePodMetric  bool
	EnableCollectJob bool

	MetricCollecInterval time.Duration
	ObjectCollecInterval time.Duration

	PodFilterForMetric            filter.Filter
	EnableExtractK8sLabelAsTagsV1 bool
	LabelAsTagsForMetric          LabelsOption
	LabelAsTagsForNonMetric       LabelsOption

	ExtraTags map[string]string
	Feeder    dkio.Feeder

	// The runtime watcher owns this process-lifetime, node-filtered informer.
	LocalPodInformer cache.SharedIndexInformer

	resourceCache    *resourceCache
	promPodPublisher promPodPublisher
}

type Kube struct {
	cfg       *Config
	client    k8sClient
	apiClient kubeclient.Interface

	mu         sync.Mutex
	leader     bool
	generation uint64
	stopped    bool
	wake       chan struct{}
	cluster    *resourceCache

	// Owned by the collection loop; election callbacks never change these.
	local                    *resourceCache
	lastEventResourceVersion string
}

func NewKubeCollector(client k8sclient.Client, cfg *Config) (*Kube, error) {
	klog = logger.SLogger("k8s", logger.WithRateLimiter(promTaskLimitLogRate, "legacy-pod-prom-task-limit"))
	if client == nil {
		return nil, fmt.Errorf("invalid kubernetes client, cannot be nil")
	}
	if cfg == nil {
		return nil, fmt.Errorf("invalid kubernetes collector config, cannot be nil")
	}
	if cfg.NodeName == "" {
		nodeName, err := config.GetLocalNodeName()
		if err != nil {
			return nil, err
		}
		cfg.NodeName = nodeName
	}
	return &Kube{cfg: cfg, client: client, apiClient: client.KubernetesClientset(), wake: make(chan struct{}, 1)}, nil
}

// SetLeader records every transition even while collection is busy. Invalidate
// the old term synchronously; a coalesced wakeup only schedules its teardown.
func (k *Kube) SetLeader(leader bool) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.stopped || k.leader == leader {
		return
	}
	k.leader = leader
	k.generation++
	if k.cluster != nil {
		k.cluster.invalidate()
	}
	select {
	case k.wake <- struct{}{}:
	default:
	}
}

func (k *Kube) StartCollect() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g := goroutine.NewGroup(goroutine.Option{Name: "k8s-lifecycle"})
	g.Go(func(_ context.Context) error {
		select {
		case <-datakit.Exit.Wait():
			cancel()
		case <-ctx.Done():
		}
		return nil
	})
	k.run(ctx)
	cancel()
	if err := g.Wait(); err != nil {
		klog.Warnf("wait Kubernetes lifecycle failed: %s", err)
	}
}

func (k *Kube) run(ctx context.Context) {
	var localSynced <-chan struct{}
	if k.cfg.NodeLocal && (k.cfg.EnableK8sMetric || k.cfg.EnableK8sObject) {
		k.local = newLocalResourceCache(ctx, k.apiClient, k.cfg)
		k.local.start()
		if k.cfg.EnableK8sObject {
			ready := make(chan struct{})
			localSynced = ready
			local := k.local
			local.wg.Go(func() {
				if cache.WaitForCacheSync(local.ctx.Done(), local.resources["pod"].HasSynced, local.resources["node"].HasSynced) {
					close(ready)
				}
			})
		}
	}
	defer func() {
		k.mu.Lock()
		k.stopped = true
		cluster := k.cluster
		k.cluster = nil
		if cluster != nil {
			cluster.invalidate()
		}
		k.mu.Unlock()
		if cluster != nil {
			cluster.stop()
		}
		if k.local != nil {
			k.local.stop()
			k.local = nil
		}
		klog.Info("k8s collect exit")
	}()

	metricTicker := time.NewTicker(k.cfg.MetricCollecInterval)
	objectTicker := time.NewTicker(k.cfg.ObjectCollecInterval)
	defer metricTicker.Stop()
	defer objectTicker.Stop()

	k.reconcileCluster(ctx)
	if k.cfg.EnableK8sObject && localSynced == nil {
		k.gatherObject()
	}
	start := ntp.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case <-k.wake:
			k.reconcileCluster(ctx)
		case <-localSynced:
			// Start local objects and annotation scrapes without waiting for the
			// next object tick. Cache synchronization must not block the loop.
			localSynced = nil
			k.gatherObject()
		case tt := <-metricTicker.C:
			if k.cfg.EnableK8sMetric {
				start = inputs.AlignTime(tt, start, k.cfg.MetricCollecInterval)
				k.gatherMetric(start.UnixNano())
			}
		case <-objectTicker.C:
			if k.cfg.EnableK8sObject {
				k.gatherObject()
			}
		}
	}
}

// Only the collection loop creates or joins terms. Resume cannot overlap the
// previous term, including its event handlers and annotation scrape workers.
func (k *Kube) reconcileCluster(ctx context.Context) {
	k.mu.Lock()
	old := k.cluster
	if old != nil && old.id == k.generation && old.usable() {
		k.mu.Unlock()
		return
	}
	k.cluster = nil
	k.mu.Unlock()
	if old != nil {
		old.stop()
	}

	k.mu.Lock()
	defer k.mu.Unlock()
	if !k.leader || k.stopped || ctx.Err() != nil {
		return
	}
	c := newClusterResourceCache(ctx, k.apiClient, k.cfg, k.generation)
	k.cluster = c
	c.start()
	if k.cfg.EnableK8sEvent {
		c.wg.Go(func() { k.runEvents(c) })
	}
	klog.Infof("started Kubernetes cluster cache term %d", c.id)
}

func (k *Kube) runEvents(c *resourceCache) {
	for c.usable() {
		k.gatherEvent(c.ctx, c.cfg)
		timer := time.NewTimer(watchRetryInterval)
		select {
		case <-c.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (k *Kube) gatherMetric(timestamp int64) {
	k.gatherResources("metric", timestamp)
}

func (k *Kube) gatherObject() {
	k.gatherResources("object", 0)
}

func (k *Kube) gatherResources(category string, timestamp int64) {
	start := time.Now()
	g := goroutine.NewGroup(goroutine.Option{Name: "k8s-" + category})
	k.mu.Lock()
	cluster := k.cluster
	k.mu.Unlock()
	for _, scope := range []struct {
		cache        *resourceCache
		constructors []resourceConstructor
		names        []string
	}{
		{cluster, nonNodeLocalResources, nonNodeLocalResourcesNames},
		{k.local, nodeLocalResources, nodeLocalResourcesNames},
	} {
		if !scope.cache.usable() {
			continue
		}
		for idx, constructor := range scope.constructors {
			name := scope.names[idx]
			g.Go(func(_ context.Context) error {
				st := time.Now()
				rc := constructor(k.client, scope.cache.cfg)
				if category == "metric" {
					rc.gatherMetric(scope.cache.ctx, timestamp)
				} else {
					rc.gatherObject(scope.cache.ctx)
				}
				collectResourceCostVec.WithLabelValues(category, name).Observe(time.Since(st).Seconds())
				return nil
			})
		}
	}
	if err := g.Wait(); err != nil {
		klog.Warnf("wait Kubernetes %s collection failed: %s", category, err)
	}
	collectCostVec.WithLabelValues(category).Observe(time.Since(start).Seconds())
}

type K8sResourceCount struct{}

//nolint:lll
func (*K8sResourceCount) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kubernetes",
		Desc:   "The count of the Kubernetes resource.",
		DescZh: "Kubernetes 中的资源计数。",
		Cat:    point.Metric,
		Tags: map[string]interface{}{
			"namespace": &inputs.TagInfo{Desc: "namespace"},
			"node_name": &inputs.TagInfo{Desc: "NodeName is a request to schedule this pod onto a specific node (only supported Pod and Container)."},
		},
		Fields: map[string]interface{}{
			"cronjob":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of Kubernetes CronJobs in the selected scope."},
			"daemonset":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of Kubernetes DaemonSets in the selected scope."},
			"deployment":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of Kubernetes Deployments in the selected scope."},
			"job":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of Kubernetes Jobs in the selected scope."},
			"node":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of Kubernetes Nodes in the selected scope."},
			"endpoint":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of Kubernetes Endpoints in the selected scope."},
			"pod":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of Kubernetes Pods in the selected scope."},
			"replicaset":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of Kubernetes ReplicaSets in the selected scope."},
			"statefulset": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of Kubernetes StatefulSets in the selected scope."},
			"service":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of Kubernetes Services in the selected scope."},
			"container":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of containers in the selected Kubernetes scope."},
		},
	}
}

type ObjectChangeEvent struct{}

//nolint:lll
func (*ObjectChangeEvent) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "event",
		Desc:   "Changes to major Kubernetes resources (Pods/Deployments/Services) will trigger change events with the following tag and fields. For the complete list of change events, see [here](change-event.md#k8s).",
		DescZh: "Kubernetes 中主要资源（Pod/Deployment/Service 等）变更将触发如下形式的变更事件。完整的变更列表，参见[这里](change-event.md#k8s)。",
		Cat:    point.KeyEvent,
		Tags: map[string]interface{}{
			"df_source":                    inputs.NewTagInfo("The event source is always `change`."),
			"df_event_id":                  inputs.NewTagInfo("The event ID is generated by UUIDv4, e.g. `event-<lowercase UUIDv4>`."),
			"df_status":                    inputs.NewTagInfo("The event source is always `info`."),
			"df_sub_status":                inputs.NewTagInfo("Always `info`."),
			"class":                        inputs.NewTagInfo("The type of Kubernetes resource, e.g. `kubernetes_deployments/kubernetes_nodes/..`"),
			"deployment_name/node_name/..": inputs.NewTagInfo("The name of Kubernetes resource, e.g. `deployment-abc-123`"),
			"uid":                          inputs.NewTagInfo("The UID of Kubernetes resource."),
			"namespace":                    inputs.NewTagInfo("The namespace of Kubernetes resource."),
		},
		Fields: map[string]interface{}{
			"df_title":   &inputs.FieldInfo{DataType: inputs.String, Type: inputs.UnknownType, Unit: inputs.NoUnit, Desc: "Title text summarizing the Kubernetes resource change."},
			"df_message": &inputs.FieldInfo{DataType: inputs.String, Type: inputs.UnknownType, Unit: inputs.NoUnit, Desc: "This is a template field, concatenated from other values: `[{{df_resource_type}}] {{df_resource}} configuration changed`."},
			"diff":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.UnknownType, Unit: inputs.NoUnit, Desc: "Unified diff text showing the Kubernetes resource changes."},
		},
	}
}

//nolint:gochecknoinits
func init() {
	setupMetrics()
}
