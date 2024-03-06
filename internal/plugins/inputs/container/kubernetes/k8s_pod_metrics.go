// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"

	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	statsv1alpha1 "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func getMemoryLimitFromResource(containers []apicorev1.Container) int64 {
	var limit int64
	for _, c := range containers {
		qu := c.Resources.Limits["memory"]
		memLimit, _ := qu.AsInt64()
		limit += memLimit
	}
	return limit
}

func getMaxCPULimitFromResource(containers []apicorev1.Container) int64 {
	var limit int64
	for _, c := range containers {
		qu := c.Resources.Limits["cpu"]
		cpuLimit := qu.MilliValue()
		if cpuLimit > limit {
			limit = cpuLimit
		}
	}
	return limit
}

type nodeCapacity struct {
	nodeName              string
	cpuCapacityMillicores int64 // unit Milli
	memCapacity           int64 // unit Bytes
}

// getMemoryCapacityFromNode return memory capacity for node.
func getCapacityFromNode(ctx context.Context, client k8sClient, nodeName string) (capacity nodeCapacity) {
	node, err := client.GetNodes().Get(ctx, nodeName, metav1.GetOptions{ResourceVersion: "0"})
	if err != nil {
		return
	}

	capacity.nodeName = nodeName

	c := node.Status.Capacity["cpu"]
	capacity.cpuCapacityMillicores = c.MilliValue()

	m := node.Status.Capacity["memory"]
	capacity.memCapacity, _ = m.AsInt64()

	return
}

type podSrvMetric struct {
	cpuUsage           float64
	cpuUsageMilliCores int64
	memoryUsageBytes   int64

	networkRcvdBytes               int64
	networkSentBytes               int64
	ephemeralStorageUsedBytes      int64
	ephemeralStorageAvailableBytes int64
	ephemeralStorageCapacityBytes  int64
}

type PodMetricsCollect interface {
	GetPodMetrics(ctx context.Context, namespace, name string) (*podSrvMetric, error)
}

type podMetricsFromAPIServer struct {
	client k8sClient
}

func newPodMetricsFromAPIServer(client k8sClient) *podMetricsFromAPIServer {
	return &podMetricsFromAPIServer{client: client}
}

func (p *podMetricsFromAPIServer) GetPodMetrics(ctx context.Context, namespace, name string) (*podSrvMetric, error) {
	item, err := p.client.GetPodMetricses(namespace).Get(ctx, name, metav1.GetOptions{ResourceVersion: "0"})
	if err != nil {
		return nil, fmt.Errorf("falied of query metrics-server for pod %s, err: %w", name, err)
	}
	return parsePodMetrics(item)
}

func parsePodMetrics(item *v1beta1.PodMetrics) (*podSrvMetric, error) {
	if len(item.Containers) == 0 {
		return nil, fmt.Errorf("not found container in pod")
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

	cpuMilliCores := cpu.MilliValue()
	memUsage, _ := mem.AsInt64()

	podMetricsQueryCountVec.WithLabelValues("api-server").Inc()

	return &podSrvMetric{
		cpuUsage:           float64(cpuMilliCores) / 1e3 * 100.0,
		cpuUsageMilliCores: cpuMilliCores,
		memoryUsageBytes:   memUsage,
	}, nil
}

type podMetricsFromKubelet struct {
	client       k8sClient
	metricsCache *statsv1alpha1.Summary
}

func newPodMetricsFromKubelet(client k8sClient) *podMetricsFromKubelet {
	return &podMetricsFromKubelet{client: client}
}

func (p *podMetricsFromKubelet) GetPodMetrics(ctx context.Context, namespace, name string) (*podSrvMetric, error) {
	if p.metricsCache == nil {
		m, err := p.client.GetMetricsFromKubelet()
		if err != nil {
			return nil, fmt.Errorf("falied of query kubelet stats, err: %w", err)
		}
		if m != nil {
			p.metricsCache = m
		}
		podMetricsQueryCountVec.WithLabelValues("kubelet").Inc()
	}
	return hitPodMetrics(p.metricsCache, namespace, name)
}

func (p *podMetricsFromKubelet) GetPodsVolumeInfo(ctx context.Context) ([]*podVolumeInfo, error) {
	if p.metricsCache == nil {
		m, err := p.client.GetMetricsFromKubelet()
		if err != nil {
			return nil, fmt.Errorf("falied of query kubelet stats/summary, err: %w", err)
		}
		if m != nil {
			p.metricsCache = m
		}
		podMetricsQueryCountVec.WithLabelValues("kubelet").Inc()
	}
	return composePodsVolumeInfo(p.metricsCache)
}

func hitPodMetrics(item *statsv1alpha1.Summary, namespace, name string) (*podSrvMetric, error) {
	if item == nil {
		return nil, fmt.Errorf("invalid kubelet stats")
	}
	if len(item.Pods) == 0 {
		return nil, fmt.Errorf("not found pod in kubelet")
	}

	metrics := &podSrvMetric{}

	for _, stats := range item.Pods {
		if stats.PodRef.Name == name && stats.PodRef.Namespace == namespace {
			if stats.CPU != nil && stats.CPU.UsageNanoCores != nil {
				cpuUsageMilliCores := int64(*stats.CPU.UsageNanoCores) / 1e6
				// minimum 1 Milli
				if cpuUsageMilliCores < 1 {
					cpuUsageMilliCores = 1
				}
				metrics.cpuUsageMilliCores = cpuUsageMilliCores
				metrics.cpuUsage = float64(cpuUsageMilliCores) / 1e3 * 100.0
			}

			if stats.Memory != nil && stats.Memory.WorkingSetBytes != nil {
				metrics.memoryUsageBytes = int64(*stats.Memory.WorkingSetBytes)
			}

			if stats.Network != nil {
				if stats.Network.RxBytes != nil {
					metrics.networkRcvdBytes = int64(*stats.Network.RxBytes)
				}
				if stats.Network.TxBytes != nil {
					metrics.networkSentBytes = int64(*stats.Network.TxBytes)
				}
			}

			if stats.EphemeralStorage != nil {
				if stats.EphemeralStorage.UsedBytes != nil {
					metrics.ephemeralStorageUsedBytes = int64(*stats.EphemeralStorage.UsedBytes)
				}
				if stats.EphemeralStorage.AvailableBytes != nil {
					metrics.ephemeralStorageAvailableBytes = int64(*stats.EphemeralStorage.AvailableBytes)
				}
				if stats.EphemeralStorage.CapacityBytes != nil {
					metrics.ephemeralStorageCapacityBytes = int64(*stats.EphemeralStorage.CapacityBytes)
				}
			}

			return metrics, nil
		}
	}
	return nil, fmt.Errorf("not found %s:%s pod metrics from kubelet", namespace, name)
}

type podVolumeInfo struct {
	nodeName, podName, namespace string
	pvcName, volumeMountName     string
	available                    int64
	capacity                     int64
	used                         int64
	inodes                       int64
	inodesUsed                   int64
	inodesFree                   int64
}

func composePodsVolumeInfo(item *statsv1alpha1.Summary) ([]*podVolumeInfo, error) {
	var res []*podVolumeInfo

	for _, pod := range item.Pods {
		for _, volume := range pod.VolumeStats {
			if volume.PVCRef == nil {
				continue
			}
			// Must be equal.
			if volume.PVCRef.Namespace != pod.PodRef.Namespace {
				continue
			}

			info := &podVolumeInfo{
				nodeName:        item.Node.NodeName,
				podName:         pod.PodRef.Name,
				namespace:       pod.PodRef.Namespace,
				pvcName:         volume.PVCRef.Name,
				volumeMountName: volume.Name,
			}

			if volume.AvailableBytes != nil {
				info.available = int64(*volume.AvailableBytes)
			}
			if volume.CapacityBytes != nil {
				info.capacity = int64(*volume.CapacityBytes)
			}
			if volume.UsedBytes != nil {
				info.used = int64(*volume.UsedBytes)
			}
			if volume.Inodes != nil {
				info.inodes = int64(*volume.Inodes)
			}
			if volume.InodesUsed != nil {
				info.inodesUsed = int64(*volume.InodesUsed)
			}
			if volume.InodesFree != nil {
				info.inodesFree = int64(*volume.InodesFree)
			}

			res = append(res, info)
		}
	}

	return res, nil
}
