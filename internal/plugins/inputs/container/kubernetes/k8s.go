// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kubernetes collect resources metric/object/event.
package kubernetes

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	objectInterval = time.Minute * 5
	metricInterval = time.Second * 60

	klog = logger.DefaultSLogger("k8s")
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

	PodFilterForMetric            filter.Filter
	EnableExtractK8sLabelAsTagsV1 bool
	LabelAsTagsForMetric          LabelsOption
	LabelAsTagsForNonMetric       LabelsOption

	ExtraTags map[string]string
	Feeder    dkio.Feeder
}

type Kube struct {
	cfg    *Config
	client k8sClient

	nodeName                 string
	onWatchingEvent          *atomic.Bool
	onWatchingChange         *atomic.Bool
	lastEventResourceVersion string

	paused    bool
	chanPause chan bool
}

func NewKubeCollector(client k8sclient.Client, cfg *Config, chanPause chan bool) (*Kube, error) {
	klog = logger.SLogger("k8s")

	if client == nil {
		return nil, fmt.Errorf("invalid kubernetes client, cannot be nil")
	}
	if cfg == nil {
		return nil, fmt.Errorf("invalid kubernetes collector config, cannot be nil")
	}

	nodeName, err := getLocalNodeName()
	if err != nil {
		return nil, err
	}

	return &Kube{
		cfg:              cfg,
		client:           client,
		nodeName:         nodeName,
		onWatchingEvent:  &atomic.Bool{},
		onWatchingChange: &atomic.Bool{},
		paused:           true,
		chanPause:        chanPause,
	}, nil
}

func (k *Kube) StartCollect() {
	tickers := []*time.Ticker{
		time.NewTicker(metricInterval),
		time.NewTicker(objectInterval),
	}
	for _, t := range tickers {
		defer t.Stop()
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "k8s-pod-prom-worker"})
	g.Go(func(_ context.Context) error {
		startPromWorker()
		return nil
	})

	if k.cfg.EnableK8sObject {
		k.gatherObject()
	}

	ctx, cancel := context.WithCancel(context.Background()) // nolint
	start := time.Now()

	for {
		select {
		case <-datakit.Exit.Wait():
			cancel()
			klog.Info("k8s collect exit")
			return

		case k.paused = <-k.chanPause:
			if k.paused {
				cancel()
				klog.Info("not leader for election")
			} else {
				ctx, cancel = context.WithCancel(context.Background())
				k.tryWatchEventAndChange(ctx)
			}

		case tt := <-tickers[0].C:
			if k.cfg.EnableK8sMetric {
				nextts := inputs.AlignTimeMillSec(tt, start.UnixMilli(), metricInterval.Milliseconds())
				start = time.UnixMilli(nextts)

				k.gatherMetric(start.UnixNano())
			}

		case <-tickers[1].C:
			if k.cfg.EnableK8sObject {
				k.gatherObject()
			}
		}
	}
}

func (k *Kube) gatherMetric(timestamp int64) {
	var (
		start = time.Now()
		g     = goroutine.NewGroup(goroutine.Option{Name: "k8s-metric"})
		ctx   = context.Background()
	)

	if !k.paused {
		for idx, newResourceFn := range nonNodeLocalResources {
			func(name string, newResource resourceConstructor) {
				g.Go(func(_ context.Context) error {
					st := time.Now()
					rc := newResource(k.client, k.cfg)
					rc.gatherMetric(ctx, timestamp)
					collectResourceCostVec.WithLabelValues("metric", name).Observe(time.Since(st).Seconds())
					return nil
				})
			}(nonNodeLocalResourcesNames[idx], newResourceFn)
		}
	}

	if k.cfg.NodeLocal {
		for idx, newResourceFn := range nodeLocalResources {
			func(name string, newResource resourceConstructor) {
				g.Go(func(_ context.Context) error {
					st := time.Now()
					rc := newResource(k.client, k.cfg)
					rc.gatherMetric(ctx, timestamp)
					collectResourceCostVec.WithLabelValues("metric", name).Observe(time.Since(st).Seconds())
					return nil
				})
			}(nodeLocalResourcesNames[idx], newResourceFn)
		}
	}

	collectCostVec.WithLabelValues("metric").Observe(time.Since(start).Seconds())
}

func (k *Kube) gatherObject() {
	var (
		start = time.Now()
		g     = goroutine.NewGroup(goroutine.Option{Name: "k8s-object"})
		ctx   = context.Background()
	)

	if !k.paused {
		for idx, newResourceFn := range nonNodeLocalResources {
			func(name string, newResource resourceConstructor) {
				g.Go(func(_ context.Context) error {
					st := time.Now()
					rc := newResource(k.client, k.cfg)
					rc.gatherObject(ctx)
					collectResourceCostVec.WithLabelValues("object", name).Observe(time.Since(st).Seconds())
					return nil
				})
			}(nonNodeLocalResourcesNames[idx], newResourceFn)
		}
	}

	if k.cfg.NodeLocal {
		for idx, newResourceFn := range nodeLocalResources {
			func(name string, newResource resourceConstructor) {
				g.Go(func(_ context.Context) error {
					st := time.Now()
					rc := newResource(k.client, k.cfg)
					rc.gatherObject(ctx)
					collectResourceCostVec.WithLabelValues("object", name).Observe(time.Since(st).Seconds())
					return nil
				})
			}(nodeLocalResourcesNames[idx], newResourceFn)
		}
	}

	collectCostVec.WithLabelValues("object").Observe(time.Since(start).Seconds())
}

func (k *Kube) tryWatchEventAndChange(ctx context.Context) {
	if k.cfg.EnableK8sEvent && !k.onWatchingEvent.Load() {
		klog.Info("collect k8s event starting")
		g := datakit.G("k8s-event")

		k.onWatchingEvent.Store(true)
		g.Go(func(_ context.Context) error {
			k.gatherEvent(ctx)
			k.onWatchingEvent.Store(false)
			return nil
		})
	}

	if !k.onWatchingChange.Load() {
		klog.Info("collect k8s object-change starting")
		g := datakit.G("k8s-object-change")

		k.onWatchingChange.Store(true)
		g.Go(func(_ context.Context) error {
			k.gatherChange(ctx)
			k.onWatchingChange.Store(false)
			return nil
		})
	}
}

func (k *Kube) gatherChange(ctx context.Context) {
	apiClient, err := k8sclient.GetAPIClient()
	if err != nil {
		klog.Warnf("failed of apiclient: %s", err)
		return
	}

	for _, newResource := range nonNodeLocalResources {
		rc := newResource(k.client, k.cfg)
		rc.addChangeInformer(apiClient.InformerFactory)
	}

	apiClient.InformerFactory.Start(ctx.Done())
	apiClient.InformerFactory.WaitForCacheSync(ctx.Done())
	<-ctx.Done()
}

type count struct{}

//nolint:lll
func (*count) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes",
		Desc: "The count of the Kubernetes resource.",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"namespace": &inputs.TagInfo{Desc: "namespace"},
			"node_name": &inputs.TagInfo{Desc: "NodeName is a request to schedule this pod onto a specific node (only supported Pod and Container)."},
		},
		Fields: map[string]interface{}{
			"cronjob":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "CronJob count"},
			"daemonset":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "Service count"},
			"deployment":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "Deployment count"},
			"job":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "Job count"},
			"node":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "Node count"},
			"endpoint":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "Endpoint count"},
			"pod":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "Pod count"},
			"replicaset":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "ReplicaSet count"},
			"statefulset": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "StatefulSet count"},
			"service":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "Service count"},
			"container":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "Container count"},
		},
	}
}

type objectChangeEvent struct{}

//nolint:lll
func (*objectChangeEvent) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "event",
		Desc: "The change in the Specification of Kubernetes resources, where only the Specification is compared, excluding Meta and Status.",
		Cat:  point.KeyEvent,
		Tags: map[string]interface{}{
			"df_source":        inputs.NewTagInfo("The event source is always `change`."),
			"df_event_id":      inputs.NewTagInfo("The event ID is generated by UUIDv4, e.g. `event-<lowercase UUIDv4>`."),
			"df_status":        inputs.NewTagInfo("The event source is always `info`."),
			"df_sub_status":    inputs.NewTagInfo("Always `info`."),
			"df_resource":      inputs.NewTagInfo("The name of Kubernetes resource, e.g. `deployment-abc-123`"),
			"df_resource_type": inputs.NewTagInfo("The type of Kubernetes resource, e.g. `Deployment/DamonSet`"),
			"df_title":         inputs.NewTagInfo("This is a template field, concatenated from other values: `[{{df_resource_type}}] {{df_resource}} configuration changed`."),
			"df_uid":           inputs.NewTagInfo("The UID of Kubernetes resource."),
			"df_namespace":     inputs.NewTagInfo("The namespace of Kubernetes resource."),
		},
		Fields: map[string]interface{}{
			"df_check_range_start": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.UnknownType, Unit: inputs.TimestampSec, Desc: "Current system time"},
			"df_check_range_end":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.UnknownType, Unit: inputs.TimestampSec, Desc: "Current system time"},
			"df_date_range":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.UnknownType, Unit: inputs.UnknownUnit, Desc: "Always `0`."},
			"df_message":           &inputs.FieldInfo{DataType: inputs.String, Type: inputs.UnknownType, Unit: inputs.UnknownUnit, Desc: "Diff text of resource changes."},
		},
	}
}

//nolint:gochecknoinits
func init() {
	setupMetrics()
	registerMeasurements(&count{})
	registerMeasurements(&objectChangeEvent{})
}
