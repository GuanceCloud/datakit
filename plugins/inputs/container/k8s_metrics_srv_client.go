// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"strconv"

	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type k8sMetricsClientX interface {
	getPodMetrics() metricsv1beta1.PodMetricsInterface
	getPodMetricsForNamespace(namespace string) metricsv1beta1.PodMetricsInterface
	getNodeMetrics() metricsv1beta1.NodeMetricsInterface
}

type k8sMetricsClient struct {
	*metricsv1beta1.MetricsV1beta1Client
}

func newK8sMetricsClient(restConfig *rest.Config) (*k8sMetricsClient, error) {
	client, err := metricsv1beta1.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &k8sMetricsClient{client}, nil
}

func (c *k8sMetricsClient) getPodMetrics() metricsv1beta1.PodMetricsInterface {
	return c.PodMetricses("")
}

func (c *k8sMetricsClient) getPodMetricsForNamespace(namespace string) metricsv1beta1.PodMetricsInterface {
	return c.PodMetricses(namespace)
}

func (c *k8sMetricsClient) getNodeMetrics() metricsv1beta1.NodeMetricsInterface {
	return c.NodeMetricses()
}

func gatherPodMetrics(client k8sMetricsClientX, namespace, name string) (*podSrvMetric, error) {
	met, err := client.getPodMetricsForNamespace(namespace).Get(context.Background(), name, metaV1GetOption)
	if err != nil {
		return nil, err
	}
	return parsePodMetrics(met), nil
}

type podSrvMetric struct {
	cpuUsage         float64
	memoryUsageBytes int64
}

func parsePodMetrics(item *v1beta1.PodMetrics) *podSrvMetric {
	if len(item.Containers) == 0 {
		return nil
	}

	cpu := item.Containers[0].Usage["cpu"]
	mem := item.Containers[0].Usage["memory"]

	for i := 1; i < len(item.Containers); i++ {
		if c, ok := item.Containers[i].Usage["cpu"]; ok {
			cpu.Add(c)
		}
		if m, ok := item.Containers[i].Usage["memory"]; ok {
			mem.Add(m)
		}
	}

	cpuUsage, err := strconv.ParseFloat(cpu.AsDec().String(), 64)
	if err != nil {
		l.Debugf("k8s pod metrics, parsed cpu err: %s", err)
	}
	memUsage, _ := mem.AsInt64()

	return &podSrvMetric{
		cpuUsage:         cpuUsage * 100, // percentage
		memoryUsageBytes: memUsage,
	}
}
