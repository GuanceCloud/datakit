// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"

	apicorev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type podSrvMetric struct {
	cpuUsage         int64 // Milli
	memoryUsageBytes int64
}

func queryPodMetrics(ctx context.Context, client k8sClient, name string, namespace string) (*podSrvMetric, error) {
	item, err := client.GetPodMetricsesForNamespace(namespace).Get(ctx, name, metaV1GetOption)
	if err != nil {
		return nil, fmt.Errorf("falied of query metrics-server for pod %s, err: %w", name, err)
	}
	return parsePodMetrics(item)
}

func parsePodMetrics(item *v1beta1.PodMetrics) (*podSrvMetric, error) {
	if len(item.Containers) == 0 {
		return nil, fmt.Errorf("unreachable, not found container in pod")
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

	cpuUsage, ok := cpu.AsInt64()
	if !ok {
		cpuUsage, ok = cpu.AsDec().Unscaled()
		if !ok {
			cpuUsage = 0
		}
	}
	memUsage, _ := mem.AsInt64()

	// rounded up to Milli
	if cpuUsage < 1000000 {
		cpuUsage = 1000000
	}

	return &podSrvMetric{
		cpuUsage:         cpuUsage / 1e6,
		memoryUsageBytes: memUsage,
	}, nil
}

func getMemoryCapacityFromResourceLimit(containers []apicorev1.Container) int64 {
	var limit int64
	for _, c := range containers {
		qu := c.Resources.Limits["memory"]
		memLimit, _ := qu.AsInt64()
		limit += memLimit
	}
	return limit
}

// getMemoryCapacityFromNode return memory capacity for node.
func getCapacityFromNode(ctx context.Context, client k8sClient, nodeName string) (cpuCapacity int64, memCapacity int64) {
	node, err := client.GetNodes().Get(ctx, nodeName, metaV1GetOption)
	if err != nil {
		return
	}

	c := node.Status.Capacity["cpu"]
	cpuCapacity, _ = c.AsDec().Unscaled()

	m := node.Status.Capacity["memory"]
	memCapacity, _ = m.AsInt64()
	return
}
