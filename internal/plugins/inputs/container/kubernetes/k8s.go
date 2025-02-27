// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kubernetes collect resources metric/object/event.
package kubernetes

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/container/option"

	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var klog = logger.DefaultSLogger("k8s")

type k8sClient k8sclient.Client

type Config struct {
	NodeName                      string
	NodeLocal                     bool
	EnableK8sMetric               bool
	EnableK8sObject               bool
	EnableK8sEvent                bool
	EnablePodMetric               bool
	EnableK8sSelfMetricByProm     bool
	EnableExtractK8sLabelAsTagsV1 bool
	ExtraTags                     map[string]string
	DisableCollectJob             bool
	Feeder                        dkio.Feeder

	PodFilterForMetric      filter.Filter
	LabelAsTagsForMetric    LabelsOption
	LabelAsTagsForNonMetric LabelsOption
}

type Kube struct {
	cfg    *Config
	client k8sClient

	nodeName                 string
	onWatchingEvent          *atomic.Bool
	lastEventResourceVersion string
	paused                   func() bool
	done                     <-chan interface{}
}

func NewKubeCollector(client k8sclient.Client, cfg *Config, paused func() bool, done <-chan interface{}) (*Kube, error) {
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
		cfg:             cfg,
		client:          client,
		nodeName:        nodeName,
		paused:          paused,
		done:            done,
		onWatchingEvent: &atomic.Bool{},
	}, nil
}

func (*Kube) Name() string {
	return name
}

func (k *Kube) Metric(feed func([]*point.Point) error, opts ...option.CollectOption) {
	if !k.cfg.EnableK8sMetric {
		return
	}

	c := option.DefaultOption()
	for _, opt := range opts {
		opt(c)
	}
	if c.Paused && !k.cfg.NodeLocal {
		return
	}

	k.gather("metric", feed, c.Paused)
}

func (k *Kube) Object(feed func([]*point.Point) error, opts ...option.CollectOption) {
	if !k.cfg.EnableK8sObject {
		return
	}

	c := option.DefaultOption()
	for _, opt := range opts {
		opt(c)
	}
	if c.Paused && !k.cfg.NodeLocal {
		return
	}

	k.gather("object", feed, c.Paused)
}

func (k *Kube) Logging(feed func([]*point.Point) error) {
	if !k.cfg.EnableK8sEvent || k.paused() || k.onWatchingEvent.Load() {
		return
	}

	k.onWatchingEvent.Store(true)
	klog.Debug("collect k8s event starting")

	g := datakit.G("k8s-event")

	g.Go(func(ctx context.Context) error {
		k.gatherEvent(feed)
		k.onWatchingEvent.Store(false)
		return nil
	})
}

func (k *Kube) getActiveNamespaces(ctx context.Context) ([]string, error) {
	list, err := k.client.GetNamespaces().List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		return nil, err
	}
	var ns []string
	for _, item := range list.Items {
		if item.Status.Phase == apicorev1.NamespaceActive {
			ns = append(ns, item.Name)
		}
	}
	return ns, nil
}

// Kubernetes collection (Deployment/Pod/other..) uses election because it needs access the api-server.
const kubeElection = true

func (k *Kube) Election() bool { return kubeElection }

func getLocalNodeName() (string, error) {
	var e string
	if os.Getenv("NODE_NAME") != "" {
		e = os.Getenv("NODE_NAME")
	}
	if os.Getenv("ENV_K8S_NODE_NAME") != "" {
		e = os.Getenv("ENV_K8S_NODE_NAME")
	}
	if e != "" {
		return e, nil
	}
	return "", fmt.Errorf("invalid ENV_K8S_NODE_NAME environment, cannot be empty")
}

type count struct{}

//nolint:lll
func (*count) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes",
		Desc: "The count of the Kubernetes resource.",
		Type: "metric",
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

//nolint:gochecknoinits
func init() {
	setupMetrics()
	registerMeasurements(&count{})
}
