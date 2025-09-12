// Package cli used to create k8s client and get some k8s info
package cli

import (
	"context"
	"fmt"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type K8sClient struct {
	*kubernetes.Clientset

	informer map[string]cache.SharedIndexInformer

	workloadLabels      []string
	workloadLablePrefix string
	timeout             time.Duration

	Pods         map[string][]*corev1.Pod
	Services     map[string][]*corev1.Service
	Deployments  map[string][]*appsv1.Deployment
	Jobs         map[string][]*batchv1.Job
	Nodes        map[string][]*corev1.Node
	DaemonSets   map[string][]*appsv1.DaemonSet
	StatefulSets map[string][]*appsv1.StatefulSet
	ReplicaSets  map[string][]*appsv1.ReplicaSet
	CronJobs     map[string][]*batchv1.CronJob

	mu sync.RWMutex
}

func newTimeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx := context.Background()
	if timeout > 0 {
		return context.WithTimeout(ctx, timeout)
	}
	return ctx, func() {}
}

func (cli *K8sClient) ListNamespaces() (*corev1.NamespaceList, error) {
	ctx, cancel := newTimeoutContext(cli.timeout)
	defer cancel()
	return cli.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
}

func (cli *K8sClient) ListAllPods() error {
	v, ok := cli.informer[ResourceTypePod]
	if !ok {
		return nil
	}
	pods := map[string][]*corev1.Pod{}
	for _, v := range v.GetIndexer().List() {
		pod, ok := v.(*corev1.Pod)
		if !ok {
			return fmt.Errorf("pod is not corev1.Pod")
		}

		ns := pod.ObjectMeta.Namespace
		if v, ok := pods[ns]; ok {
			pods[ns] = append(v, pod)
		} else {
			pods[ns] = []*corev1.Pod{pod}
		}
	}

	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.Pods = pods

	return nil
}

func (cli *K8sClient) ListPods(ns string) ([]*corev1.Pod, error) {
	cli.mu.RLock()
	defer cli.mu.RUnlock()

	if ns == "" {
		var res []*corev1.Pod
		for _, v := range cli.Pods {
			res = append(res, v...)
		}
		return res, nil
	}

	if v, ok := cli.Pods[ns]; ok {
		return v, nil
	}
	return nil, nil
}

func (cli *K8sClient) ListAllServices() error {
	v, ok := cli.informer[ResourceTypeService]
	if !ok {
		return nil
	}
	svcs := map[string][]*corev1.Service{}
	for _, v := range v.GetIndexer().List() {
		svc, ok := v.(*corev1.Service)
		if !ok {
			return fmt.Errorf("service is not corev1.Service")
		}

		ns := svc.ObjectMeta.Namespace
		if v, ok := svcs[ns]; ok {
			svcs[ns] = append(v, svc)
		} else {
			svcs[ns] = []*corev1.Service{svc}
		}
	}

	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.Services = svcs

	return nil
}

func (cli *K8sClient) ListServices(ns string) ([]*corev1.Service, error) {
	cli.mu.RLock()
	defer cli.mu.RUnlock()

	if ns == "" {
		var res []*corev1.Service
		for _, v := range cli.Services {
			res = append(res, v...)
		}
		return res, nil
	}

	if v, ok := cli.Services[ns]; ok {
		return v, nil
	}
	return nil, nil
}

func (cli *K8sClient) ListAllReplicaSets() error {
	v, ok := cli.informer[ResourceTypeReplicaSet]
	if !ok {
		return nil
	}
	rps := map[string][]*appsv1.ReplicaSet{}
	for _, v := range v.GetIndexer().List() {
		svc, ok := v.(*appsv1.ReplicaSet)
		if !ok {
			return fmt.Errorf("replicaset is not appsv1.ReplicaSet")
		}

		ns := svc.ObjectMeta.Namespace
		if v, ok := rps[ns]; ok {
			rps[ns] = append(v, svc)
		} else {
			rps[ns] = []*appsv1.ReplicaSet{svc}
		}
	}

	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.ReplicaSets = rps

	return nil
}

func (cli *K8sClient) ListReplicaSets(ns string) ([]*appsv1.ReplicaSet, error) {
	cli.mu.RLock()
	defer cli.mu.RUnlock()

	if ns == "" {
		var res []*appsv1.ReplicaSet
		for _, v := range cli.ReplicaSets {
			res = append(res, v...)
		}
		return res, nil
	}

	if v, ok := cli.ReplicaSets[ns]; ok {
		return v, nil
	}
	return nil, nil
}

func (cli *K8sClient) ListAllDeployments() error {
	v, ok := cli.informer[ResourceTypeDeployment]
	if !ok {
		return nil
	}
	dps := map[string][]*appsv1.Deployment{}
	for _, v := range v.GetIndexer().List() {
		svc, ok := v.(*appsv1.Deployment)
		if !ok {
			return fmt.Errorf("deployment is not appsv1.Deployment")
		}

		ns := svc.ObjectMeta.Namespace
		if v, ok := dps[ns]; ok {
			dps[ns] = append(v, svc)
		} else {
			dps[ns] = []*appsv1.Deployment{svc}
		}
	}

	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.Deployments = dps

	return nil
}

func (cli *K8sClient) ListDeployments(ns string) (*appsv1.DeploymentList, error) {
	ctx, cancel := newTimeoutContext(cli.timeout)
	defer cancel()
	return cli.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
}

func (cli *K8sClient) ListAllStatefulSets() error {
	v, ok := cli.informer[ResourceTypeStatefulSet]
	if !ok {
		return nil
	}
	ss := map[string][]*appsv1.StatefulSet{}
	for _, v := range v.GetIndexer().List() {
		svc, ok := v.(*appsv1.StatefulSet)
		if !ok {
			return fmt.Errorf("statefulset is not appsv1.StatefulSet")
		}

		ns := svc.ObjectMeta.Namespace
		if v, ok := ss[ns]; ok {
			ss[ns] = append(v, svc)
		} else {
			ss[ns] = []*appsv1.StatefulSet{svc}
		}
	}

	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.StatefulSets = ss

	return nil
}

func (cli *K8sClient) ListStatefulSets(ns string) ([]*appsv1.StatefulSet, error) {
	cli.mu.RLock()
	defer cli.mu.RUnlock()

	if ns == "" {
		var res []*appsv1.StatefulSet
		for _, v := range cli.StatefulSets {
			res = append(res, v...)
		}
		return res, nil
	}

	if v, ok := cli.StatefulSets[ns]; ok {
		return v, nil
	}
	return nil, nil
}

func (cli *K8sClient) ListAllDaemonSets() error {
	v, ok := cli.informer[ResourceTypeDaemonSet]
	if !ok {
		return nil
	}
	ds := map[string][]*appsv1.DaemonSet{}
	for _, v := range v.GetIndexer().List() {
		svc, ok := v.(*appsv1.DaemonSet)
		if !ok {
			return fmt.Errorf("daemonset is not appsv1.DaemonSet")
		}

		ns := svc.ObjectMeta.Namespace
		if v, ok := ds[ns]; ok {
			ds[ns] = append(v, svc)
		} else {
			ds[ns] = []*appsv1.DaemonSet{svc}
		}
	}

	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.DaemonSets = ds

	return nil
}

func (cli *K8sClient) ListDaemonSet(ns string) ([]*appsv1.DaemonSet, error) {
	cli.mu.RLock()
	defer cli.mu.RUnlock()

	if ns == "" {
		var res []*appsv1.DaemonSet
		for _, v := range cli.DaemonSets {
			res = append(res, v...)
		}
		return res, nil
	}

	if v, ok := cli.DaemonSets[ns]; ok {
		return v, nil
	}
	return nil, nil
}

func (cli *K8sClient) ListAllCronJobs() error {
	v, ok := cli.informer[ResourceTypeCronJob]
	if !ok {
		return nil
	}
	cj := map[string][]*batchv1.CronJob{}
	for _, v := range v.GetIndexer().List() {
		svc, ok := v.(*batchv1.CronJob)
		if !ok {
			return fmt.Errorf("cronjob is not batchv1.CronJob")
		}

		ns := svc.ObjectMeta.Namespace
		if v, ok := cj[ns]; ok {
			cj[ns] = append(v, svc)
		} else {
			cj[ns] = []*batchv1.CronJob{svc}
		}
	}

	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.CronJobs = cj

	return nil
}

func (cli *K8sClient) ListCronJobs(ns string) ([]*batchv1.CronJob, error) {
	cli.mu.RLock()
	defer cli.mu.RUnlock()

	if ns == "" {
		var res []*batchv1.CronJob
		for _, v := range cli.CronJobs {
			res = append(res, v...)
		}
		return res, nil
	}

	if v, ok := cli.CronJobs[ns]; ok {
		return v, nil
	}
	return nil, nil
}

func (cli *K8sClient) ListAllJobs() error {
	v, ok := cli.informer[ResourceTypeJob]
	if !ok {
		return nil
	}
	jobs := map[string][]*batchv1.Job{}
	for _, v := range v.GetIndexer().List() {
		svc, ok := v.(*batchv1.Job)
		if !ok {
			return fmt.Errorf("job is not batchv1.Job")
		}

		ns := svc.ObjectMeta.Namespace
		if v, ok := jobs[ns]; ok {
			jobs[ns] = append(v, svc)
		} else {
			jobs[ns] = []*batchv1.Job{svc}
		}
	}

	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.Jobs = jobs

	return nil
}

func (cli *K8sClient) ListJobs(ns string) ([]*batchv1.Job, error) {
	cli.mu.RLock()
	defer cli.mu.RUnlock()

	if ns == "" {
		var res []*batchv1.Job
		for _, v := range cli.Jobs {
			res = append(res, v...)
		}
		return res, nil
	}

	if v, ok := cli.Jobs[ns]; ok {
		return v, nil
	}
	return nil, nil
}
