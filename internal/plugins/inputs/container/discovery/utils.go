// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package discovery

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	apicorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func queryPodOwner(pod *apicorev1.Pod) string {
	if len(pod.OwnerReferences) != 0 {
		switch pod.OwnerReferences[0].Kind {
		case "ReplicaSet":
			if hash, ok := pod.Labels["pod-template-hash"]; ok {
				return strings.TrimRight(pod.OwnerReferences[0].Name, "-"+hash)
			}
		case "DaemonSet", "StatefulSet":
			return pod.OwnerReferences[0].Name
		default:
			// skip
		}
	}
	return ""
}

func completePromConfig(config string, item *apicorev1.Pod) string {
	podIP := item.Status.PodIP

	// 从 ip 列表中使用 index 获取 ip
	func() {
		indexStr, ok := item.Annotations[annotationPromIPIndex]
		if !ok {
			return
		}
		idx, err := strconv.Atoi(indexStr)
		if err != nil {
			klog.Warnf("parsing 'prom.instances.ip_index' failed for source %s, err: %s", item.Name, err)
			return
		}
		if !(0 <= idx && idx < len(item.Status.PodIPs)) {
			klog.Warnf("parsing 'prom.instances.ip_index' failed for source %s, excepting less len(%d) got %d", item.Name, len(item.Status.PodIPs), idx)
			return
		}
		podIP = item.Status.PodIPs[idx].IP
	}()

	ownerName := queryPodOwner(item)
	config = strings.ReplaceAll(config, "$OWNER", ownerName)
	config = strings.ReplaceAll(config, "$IP", podIP)
	config = strings.ReplaceAll(config, "$NAMESPACE", item.Namespace)
	config = strings.ReplaceAll(config, "$PODNAME", item.Name)
	config = strings.ReplaceAll(config, "$NODENAME", item.Spec.NodeName)

	return config
}

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

func parseScrapeFromProm(scrape string) bool {
	if scrape == "" {
		return false
	}
	b, _ := strconv.ParseBool(scrape)
	return b
}

func findContainerPortForPod(pod *apicorev1.Pod, portName string) int {
	if portName == "" {
		return -1
	}
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			if port.Name == portName {
				return int(port.ContainerPort)
			}
		}
	}
	return -1
}

func findContainerPortForService(svc *apicorev1.Service, pod *apicorev1.Pod, portName string) int {
	if portName == "" {
		return -1
	}

	for _, s := range svc.Spec.Ports {
		if s.Name != portName {
			continue
		}

		switch s.TargetPort.Type {
		case intstr.Int:
			return s.TargetPort.IntValue()
		case intstr.String:
			return findContainerPortForPod(pod, s.TargetPort.String())
		}
	}

	return -1
}
