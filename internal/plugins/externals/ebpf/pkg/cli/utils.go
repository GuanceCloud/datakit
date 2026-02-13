// Package cli used to create k8s client and get some k8s info
package cli

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func newSelectorFromMap(m map[string]string) (labels.Selector, []error) {
	var errs []error
	rSet := make(labels.Requirements, 0, len(m))
	for k, v := range m {
		if r, err := labels.NewRequirement(k, selection.Equals, []string{v}); err != nil {
			errs = append(errs, err)
		} else {
			rSet = append(rSet, *r)
		}
	}
	return labels.NewSelector().Add(rSet...), errs
}

type ServicePort corev1.ServicePort

type ContainerInfo struct {
	ID   string
	Name string

	PodUID    string
	PodName   string
	Namespace string

	Pid int
}

type PodInfo struct {
	UID       string
	Name      string
	Namespace string

	Labels      map[string]string
	Annotations map[string]string

	HostIPs []string
	PodIPs  []string
	Ports   []corev1.ContainerPort

	HostIPC     bool
	HostNetwork bool
	HostPID     bool

	OwnerReferences []metav1.OwnerReference

	StartTime time.Time
}

type Port struct {
	Port     uint32
	Protocol string
}

type ServiceInfo struct {
	UID       string
	Name      string
	Namespace string

	Labels      map[string]string
	Annotations map[string]string

	Type string

	ClusterIPs  []string
	ExternalIPs []string
	Port        []ServicePort

	OwnerReferences []metav1.OwnerReference

	Selector labels.Selector
}

func GetContainersInfo(cliLi []*CRIClient) (infs []*ContainerInfo, errs []error) {
	for _, cli := range cliLi {
		if cli == nil {
			continue
		}
		containers, err := cli.ListContainers()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		for _, c := range containers {
			inf := &ContainerInfo{
				ID: c.Id,
			}
			if uid, ok := c.Labels["io.kubernetes.pod.uid"]; ok {
				inf.PodUID = uid
			}
			if name, ok := c.Labels["io.kubernetes.pod.name"]; ok {
				inf.PodName = name
			}
			if namespace, ok := c.Labels["io.kubernetes.pod.namespace"]; ok {
				inf.Namespace = namespace
			}
			if ctrName, ok := c.Labels["io.kubernetes.container.name"]; ok {
				inf.Name = ctrName
			}

			if pid, err := cli.GetContainerPID(c.Id); err != nil {
				errs = append(errs, err)
			} else {
				inf.Pid = pid
			}
			infs = append(infs, inf)
		}
	}
	return infs, errs
}

// PodContainerMapping returns a map of pods to containers,
// keyed by k8s namespace and pod UID.
func PodContainerMapping(containers []*ContainerInfo) map[string]map[string]*ContainerInfo {
	result := map[string]map[string]*ContainerInfo{}
	for _, c := range containers {
		if c.PodUID == "" || c.Namespace == "" {
			continue
		}
		if _, ok := result[c.Namespace]; !ok {
			result[c.Namespace] = map[string]*ContainerInfo{}
		}
		result[c.Namespace][c.PodUID] = c
	}
	return result
}

func GetPodInfo(cli *K8sClient, ns string) ([]*PodInfo, error) {
	list, err := cli.ListPods(ns)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	result := []*PodInfo{}
	for _, elem := range list {
		podIPs := []string{}
		for _, ip := range elem.Status.PodIPs {
			podIPs = append(podIPs, ip.IP)
		}
		if len(podIPs) == 0 {
			podIPs = append(podIPs, elem.Status.PodIP)
		}
		ports := []corev1.ContainerPort{}
		for _, c := range elem.Spec.Containers {
			ports = append(ports, c.Ports...)
		}
		hostIPs := []string{elem.Status.HostIP}
		pod := &PodInfo{
			UID:       string(elem.GetUID()),
			Name:      elem.GetName(),
			Namespace: elem.GetNamespace(),

			Labels:      elem.Labels,
			Annotations: elem.Annotations,

			HostIPs: hostIPs,
			PodIPs:  podIPs,
			Ports:   ports,

			HostIPC:     elem.Spec.HostIPC,
			HostNetwork: elem.Spec.HostNetwork,
			HostPID:     elem.Spec.HostPID,

			OwnerReferences: elem.OwnerReferences,
		}
		if elem.Status.StartTime != nil {
			pod.StartTime = elem.Status.StartTime.Time
		}
		result = append(result, pod)
	}
	return result, nil
}

func GetServiceInfo(cli *K8sClient, ns string) ([]*ServiceInfo, error) {
	list, err := cli.ListServices(ns)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	result := []*ServiceInfo{}

	ports := func(p []corev1.ServicePort) []ServicePort {
		result := []ServicePort{}
		for _, port := range p {
			result = append(result, ServicePort(port))
		}
		return result
	}

	for _, elem := range list {
		service := &ServiceInfo{
			UID:             string(elem.GetUID()),
			Name:            elem.GetName(),
			Namespace:       elem.GetNamespace(),
			Labels:          elem.Labels,
			Annotations:     elem.Annotations,
			Type:            string(elem.Spec.Type),
			ClusterIPs:      elem.Spec.ClusterIPs,
			ExternalIPs:     elem.Spec.ExternalIPs,
			Port:            ports(elem.Spec.Ports),
			OwnerReferences: elem.OwnerReferences,
		}

		if len(elem.Spec.Selector) > 0 {
			service.Selector, _ = newSelectorFromMap(elem.Spec.Selector)
		}

		result = append(result, service)
	}
	return result, nil
}
