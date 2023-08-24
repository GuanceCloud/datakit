// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kubernetes collect resources metric/object/event.
package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var klog = logger.DefaultSLogger("k8s")

type k8sClient client.Client

type Config struct {
	EnableK8sMetric             bool
	EnablePodMetric             bool
	EnableK8sEvent              bool
	EnableExtractK8sLabelAsTags bool
	ExtraTags                   map[string]string
	GlobalCustomerKeys          []string
}

type Kube struct {
	cfg    *Config
	client k8sClient

	canCollectPodMetrics bool
	onWatchingEvent      *atomic.Bool

	paused func() bool
	done   <-chan interface{}
}

var getGlobalCustomerKeys = func() []string { return nil }

func NewKubeCollector(client client.Client, cfg *Config, paused func() bool, done <-chan interface{}) (*Kube, error) {
	klog = logger.SLogger("k8s")

	if client == nil {
		return nil, fmt.Errorf("invalid kubernetes client, cannot be nil")
	}
	if cfg == nil {
		return nil, fmt.Errorf("invalid kubernetes collector config, cannot be nil")
	}

	getGlobalCustomerKeys = func() []string {
		return cfg.GlobalCustomerKeys
	}

	return &Kube{
		cfg:             cfg,
		client:          client,
		paused:          paused,
		done:            done,
		onWatchingEvent: &atomic.Bool{},
	}, nil
}

func (*Kube) Name() string {
	return globalName
}

func (k *Kube) Metric() ([]inputs.Measurement, error) {
	if !k.cfg.EnableK8sMetric {
		klog.Info("collect k8s metric: off")
		return nil, nil
	}

	counterWithName := make(map[string]map[string]int) // example: map["pod"]["kube-system"] = 10
	var res []inputs.Measurement
	var mu sync.Mutex

	g := goroutine.NewGroup(goroutine.Option{Name: "k8s-metrics"})

	for name, fn := range metricResourceList {
		func(name string, fn resourceHandle) {
			g.Go(func(_ context.Context) error {
				ctx := context.Background()
				if name == "pod" {
					ctx = context.WithValue(ctx, canCollectPodMetricsKey, k.canCollectPodMetrics)
					ctx = context.WithValue(ctx, setExtraK8sLabelAsTagsKey, k.cfg.EnableExtractK8sLabelAsTags)
				}

				meas, err := fn(ctx, k.client)
				if err != nil {
					klog.Warnf("failed to get %s resources, err: %s, ignored", name, err)
					return nil
				}

				mu.Lock()
				for _, mea := range meas {
					mea.addExtraTags(k.cfg.ExtraTags)
					res = append(res, mea)

					if counterWithName[name] == nil {
						counterWithName[name] = make(map[string]int)
					}
					counterWithName[name][mea.namespace()]++
				}
				mu.Unlock()
				return nil
			})
		}(name, fn)
	}

	_ = g.Wait()

	ns := transToNamespaceMeasurements(counterWithName)
	res = append(res, ns...)

	return res, nil
}

func (k *Kube) Object() ([]inputs.Measurement, error) {
	// update metrics-server state
	k.canCollectPodMetrics = k.verifyMetricsServerAccess()

	var res []inputs.Measurement
	var mu sync.Mutex

	g := goroutine.NewGroup(goroutine.Option{Name: "k8s-object"})

	for name, fn := range objectResourceList {
		func(name string, fn resourceHandle) {
			g.Go(func(_ context.Context) error {
				ctx := context.Background()
				if name == "pod" {
					ctx = context.WithValue(ctx, canCollectPodMetricsKey, k.canCollectPodMetrics)
					ctx = context.WithValue(ctx, setExtraK8sLabelAsTagsKey, k.cfg.EnableExtractK8sLabelAsTags)
				}

				meas, err := fn(ctx, k.client)
				if err != nil {
					klog.Warnf("failed to get %s resources, err: %s, ignored", name, err)
					return nil
				}

				mu.Lock()
				for _, mea := range meas {
					mea.addExtraTags(k.cfg.ExtraTags)
					res = append(res, mea)
				}
				mu.Unlock()
				return nil
			})
		}(name, fn)
	}

	_ = g.Wait()

	return res, nil
}

func (k *Kube) Logging() error {
	if !k.cfg.EnableK8sEvent {
		klog.Debug("collect k8s event: off")
		return nil
	}

	if k.onWatchingEvent.Load() {
		return nil
	}

	k.onWatchingEvent.Store(true)
	klog.Debug("collect k8s event starting")

	g := datakit.G("k8s-event")
	g.Go(func(ctx context.Context) error {
		k.watchingEvent()
		k.onWatchingEvent.Store(false)
		return nil
	})
	return nil
}

func (k *Kube) verifyMetricsServerAccess() bool {
	if !k.cfg.EnablePodMetric {
		return false
	}
	_, err := k.client.GetPodMetricsesForNamespace("datakit").List(context.TODO(), metaV1ListOption)
	if err != nil {
		klog.Warnf("unable to access metrics-server, err: %s, skip collecting pod metrics. retry in 5 minutes", err)
		return false
	}
	return true
}

func transToNamespaceMeasurements(counterWithName map[string]map[string]int) []inputs.Measurement {
	// counterWithName 的翻转，用来构建 point
	// example: map["kube-system"]["pod"] = 10
	var res []inputs.Measurement

	counterWithNamespace := make(map[string]map[string]int)
	for name, m := range counterWithName {
		for namespace, count := range m {
			if counterWithNamespace[namespace] == nil {
				counterWithNamespace[namespace] = make(map[string]int)
			}
			counterWithNamespace[namespace][name] = count
		}
	}
	for namespace, m := range counterWithNamespace {
		p := typed.NewPointKV()
		p.SetTag("namespace", namespace)
		for name, count := range m {
			p.SetField(name, count)
		}
		res = append(res, &count{p})
	}

	return res
}

func transLabelKey(s string) string {
	return strings.ReplaceAll(s, ".", "_")
}

type count struct{ typed.PointKV }

func (c *count) LineProto() (*point.Point, error) {
	return point.NewPoint("kubernetes", c.Tags(), c.Fields(), metricOpt)
}

//nolint:lll
func (*count) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes",
		Desc: "The count of the Kubernetes resource.",
		Type: "metric",
		Tags: map[string]interface{}{
			"namespace": &inputs.TagInfo{Desc: "namespace"},
		},
		Fields: map[string]interface{}{
			"cronjob":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "CronJob count"},
			"daemonset":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "Service count"},
			"deployment":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "Deployment count"},
			"job":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "Job count"},
			"node":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "Node count"},
			"pod":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "Pod count"},
			"replicaset":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "ReplicaSet count"},
			"statefulset": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "StatefulSet count"},
			"service":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "Service count"},
		},
	}
}
