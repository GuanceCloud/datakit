// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"strings"

	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type podInfo struct {
	pod       *apicorev1.Pod
	ownerKind string
	ownerName string
}

func (p *podInfo) containerImage(containerName string) string {
	for _, container := range p.pod.Spec.Containers {
		if container.Name == containerName {
			return container.Image
		}
	}
	return ""
}

func (p *podInfo) owner() (string, string) {
	return strings.ToLower(p.ownerKind), p.ownerName
}

func (p *podInfo) cpuLimit(containerName string) int64 {
	if containerName == "" {
		return 0
	}

	for _, c := range p.pod.Spec.Containers {
		if c.Name != containerName {
			continue
		}

		cpu := c.Resources.Limits["cpu"]

		limit, ok := cpu.AsInt64()
		if !ok {
			limit, ok = cpu.AsDec().Unscaled()
			if !ok {
				limit = 0
			}
		}
		return limit
	}

	return 0
}

func (c *container) queryPodInfo(ctx context.Context, podName, podNamespace string) (*podInfo, error) {
	pod, err := c.k8sClient.GetPods(podNamespace).Get(ctx, podName, metav1.GetOptions{ResourceVersion: "0"})
	if err != nil {
		return nil, fmt.Errorf("unable query pod %s, err: %w", podName, err)
	}

	info := podInfo{pod: pod}

	if len(pod.OwnerReferences) != 0 {
		switch pod.OwnerReferences[0].Kind {
		case "ReplicaSet":
			if hash, ok := pod.Labels["pod-template-hash"]; ok {
				info.ownerKind = "Deployment"
				info.ownerName = strings.TrimRight(pod.OwnerReferences[0].Name, "-"+hash)
			}
		case "DaemonSet", "StatefulSet":
			info.ownerKind = pod.OwnerReferences[0].Kind
			info.ownerName = pod.OwnerReferences[0].Name
		default:
			// skip
		}
	}

	return &info, nil
}
