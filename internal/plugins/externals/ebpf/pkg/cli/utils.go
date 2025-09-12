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

func newSelector(labelSelector *metav1.LabelSelector) (labels.Selector, []error) {
	if labelSelector == nil {
		return labels.NewSelector(), nil
	}

	rSet := make(labels.Requirements, 0,
		len(labelSelector.MatchExpressions)+
			len(labelSelector.MatchLabels))

	var errs []error
	for _, v := range labelSelector.MatchExpressions {
		if r, err := labels.NewRequirement(v.Key, selection.Operator(v.Operator), v.Values); err != nil {
			errs = append(errs, err)
		} else {
			rSet = append(rSet, *r)
		}
	}

	for k, v := range labelSelector.MatchLabels {
		r, err := labels.NewRequirement(k, selection.Equals, []string{v})
		if err != nil {
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

type ReplicaSetInfo struct {
	UID             string
	Name            string
	Namespace       string
	Labels          map[string]string
	Annotations     map[string]string
	OwnerReferences []metav1.OwnerReference
	Selector        labels.Selector
}

type DeploymentInfo struct {
	UID       string
	Name      string
	Namespace string

	Labels      map[string]string
	Annotations map[string]string

	OwnerReferences []metav1.OwnerReference

	Selector labels.Selector
}

type StatefulSetInfo struct {
	UID       string
	Name      string
	Namespace string

	Labels      map[string]string
	Annotations map[string]string

	OwnerReferences []metav1.OwnerReference

	Selector labels.Selector
}

type DaemonSetInfo struct {
	UID       string
	Name      string
	Namespace string

	Labels      map[string]string
	Annotations map[string]string

	OwnerReferences []metav1.OwnerReference

	Selector labels.Selector
}

type CronJobInfo struct {
	UID       string
	Name      string
	Namespace string

	Labels      map[string]string
	Annotations map[string]string

	OwnerReferences []metav1.OwnerReference

	Selector labels.Selector
}

type JobInfo struct {
	UID       string
	Name      string
	Namespace string

	Labels      map[string]string
	Annotations map[string]string

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

func GetReplicaSetInfo(cli *K8sClient, ns string) ([]*ReplicaSetInfo, error) {
	list, err := cli.ListReplicaSets(ns)
	if err != nil {
		return nil, fmt.Errorf("failed to list replicasets: %w", err)
	}

	result := []*ReplicaSetInfo{}

	for _, elem := range list {
		selector, _ := newSelector(elem.Spec.Selector)
		replicaSet := &ReplicaSetInfo{
			UID:             string(elem.GetUID()),
			Name:            elem.GetName(),
			Namespace:       elem.GetNamespace(),
			Labels:          elem.Labels,
			Annotations:     elem.Annotations,
			Selector:        selector,
			OwnerReferences: elem.OwnerReferences,
		}

		result = append(result, replicaSet)
	}
	return result, nil
}

func GetDeploymentInfo(cli *K8sClient, ns string) ([]*DeploymentInfo, error) {
	list, err := cli.ListDeployments(ns)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	result := []*DeploymentInfo{}

	for _, elem := range list.Items {
		selector, _ := newSelector(elem.Spec.Selector)
		deployment := &DeploymentInfo{
			UID:             string(elem.GetUID()),
			Name:            elem.GetName(),
			Namespace:       elem.GetNamespace(),
			Labels:          elem.Labels,
			Annotations:     elem.Annotations,
			Selector:        selector,
			OwnerReferences: elem.OwnerReferences,
		}
		result = append(result, deployment)
	}

	return result, nil
}

func GetStatefulSetInfo(cli *K8sClient, ns string) ([]*StatefulSetInfo, error) {
	list, err := cli.ListStatefulSets(ns)
	if err != nil {
		return nil, fmt.Errorf("failed to list statefulsets: %w", err)
	}

	result := []*StatefulSetInfo{}

	for _, elem := range list {
		selector, _ := newSelector(elem.Spec.Selector)
		statefulset := &StatefulSetInfo{
			UID:             string(elem.GetUID()),
			Name:            elem.GetName(),
			Namespace:       elem.GetNamespace(),
			Labels:          elem.Labels,
			Annotations:     elem.Annotations,
			Selector:        selector,
			OwnerReferences: elem.OwnerReferences,
		}
		result = append(result, statefulset)
	}
	return result, nil
}

func GetDaemonSetInfo(cli *K8sClient, ns string) ([]*DaemonSetInfo, error) {
	list, err := cli.ListDaemonSet(ns)
	if err != nil {
		return nil, fmt.Errorf("failed to list daemonsets: %w", err)
	}

	result := []*DaemonSetInfo{}

	for _, elem := range list {
		selector, _ := newSelector(elem.Spec.Selector)
		daemonset := &DaemonSetInfo{
			UID:             string(elem.GetUID()),
			Name:            elem.GetName(),
			Namespace:       elem.GetNamespace(),
			Labels:          elem.Labels,
			Annotations:     elem.Annotations,
			Selector:        selector,
			OwnerReferences: elem.OwnerReferences,
		}
		result = append(result, daemonset)
	}
	return result, nil
}

func GetCronJobInfo(cli *K8sClient, ns string) ([]*CronJobInfo, error) {
	list, err := cli.ListCronJobs(ns)
	if err != nil {
		return nil, fmt.Errorf("failed to list cronjobs: %w", err)
	}

	result := []*CronJobInfo{}

	for _, elem := range list {
		selector, _ := newSelector(elem.Spec.JobTemplate.Spec.Selector)
		cronjob := &CronJobInfo{
			UID:             string(elem.GetUID()),
			Name:            elem.GetName(),
			Namespace:       elem.GetNamespace(),
			Labels:          elem.Labels,
			Annotations:     elem.Annotations,
			Selector:        selector,
			OwnerReferences: elem.OwnerReferences,
		}
		result = append(result, cronjob)
	}
	return result, nil
}

func GetJobInfo(cli *K8sClient, ns string) ([]*JobInfo, error) {
	list, err := cli.ListJobs(ns)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	result := []*JobInfo{}

	for _, elem := range list {
		selector, _ := newSelector(elem.Spec.Selector)
		job := &JobInfo{
			UID:             string(elem.GetUID()),
			Name:            elem.GetName(),
			Namespace:       elem.GetNamespace(),
			Labels:          elem.Labels,
			Annotations:     elem.Annotations,
			Selector:        selector,
			OwnerReferences: elem.OwnerReferences,
		}
		result = append(result, job)
	}
	return result, nil
}
