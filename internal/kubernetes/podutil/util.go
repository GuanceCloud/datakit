// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package podutil implements some basic functions for handling Pods.
package podutil

import (
	"strings"

	apicorev1 "k8s.io/api/core/v1"
)

func PodOwner(pod *apicorev1.Pod) (kind string, name string) {
	if len(pod.OwnerReferences) == 0 {
		return
	}
	switch pod.OwnerReferences[0].Kind {
	case "ReplicaSet":
		if hash, ok := pod.Labels["pod-template-hash"]; ok {
			kind = "deployment"
			name = strings.TrimSuffix(pod.OwnerReferences[0].Name, "-"+hash)
		}
	case "DaemonSet", "StatefulSet", "Job", "CronJob":
		kind = pod.OwnerReferences[0].Kind
		name = pod.OwnerReferences[0].Name
	default:
		// skip
	}

	kind = strings.ToLower(kind)
	name = strings.ToLower(name)
	return
}

func SumCPULimits(pod *apicorev1.Pod) int64 {
	var sum int64
	for _, c := range pod.Spec.Containers {
		limit := c.Resources.Limits["cpu"]
		sum += limit.MilliValue()
	}
	return sum
}

func SumMemoryLimits(pod *apicorev1.Pod) int64 {
	var sum int64
	for _, c := range pod.Spec.Containers {
		limit := c.Resources.Limits["memory"]
		sum += limit.Value()
	}
	return sum
}

func SumCPURequests(pod *apicorev1.Pod) int64 {
	var sum int64
	for _, c := range pod.Spec.Containers {
		req := c.Resources.Requests["cpu"]
		sum += req.MilliValue()
	}
	return sum
}

func SumMemoryRequests(pod *apicorev1.Pod) int64 {
	var sum int64
	for _, c := range pod.Spec.Containers {
		req := c.Resources.Requests["memory"]
		sum += req.Value()
	}
	return sum
}

func ContainerImageFromPod(containerName string, pod *apicorev1.Pod) string {
	if containerName != "" {
		for _, container := range pod.Spec.Containers {
			if container.Name == containerName {
				return container.Image
			}
		}
	}
	return ""
}

func ContainerLimitInPod(containerName string, pod *apicorev1.Pod) (cpuLimit int64, memLimit int64) {
	if containerName != "" {
		for _, c := range pod.Spec.Containers {
			if c.Name != containerName {
				continue
			}
			cpu := c.Resources.Limits["cpu"]
			mem := c.Resources.Limits["memory"]
			return cpu.MilliValue(), mem.Value()
		}
	}
	return
}

func ContainerRequestInPod(containerName string, pod *apicorev1.Pod) (cpuRequest int64, memRequest int64) {
	if containerName != "" {
		for _, c := range pod.Spec.Containers {
			if c.Name != containerName {
				continue
			}
			cpu := c.Resources.Requests["cpu"]
			mem := c.Resources.Requests["memory"]
			return cpu.MilliValue(), mem.Value()
		}
	}
	return
}
