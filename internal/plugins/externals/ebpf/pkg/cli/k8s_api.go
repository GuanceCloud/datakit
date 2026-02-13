// Package cli used to create k8s client and get some k8s info
package cli

import (
	"context"
	"fmt"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/tools/cache"

	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
)

type K8sClient struct {
	informer       map[string]cache.SharedIndexInformer
	operatorClient *k8sclient.OperatorClient

	workloadLabels      []string
	workloadLabelPrefix string
	timeout             time.Duration

	Pods     map[string][]*corev1.Pod
	Services map[string][]*corev1.Service

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
	// 从已拉取的 Pod 推导 namespace
	cli.mu.RLock()
	defer cli.mu.RUnlock()
	nsSet := make(map[string]struct{})
	for ns := range cli.Pods {
		nsSet[ns] = struct{}{}
	}
	items := make([]corev1.Namespace, 0, len(nsSet))
	for ns := range nsSet {
		items = append(items, corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	}
	return &corev1.NamespaceList{Items: items}, nil
}

func (cli *K8sClient) ListAllPods() error {
	if cli.operatorClient != nil {
		ctx, cancel := newTimeoutContext(cli.timeout)
		defer cancel()
		podList, err := cli.operatorClient.GetPods("").List(ctx, metav1.ListOptions{})
		if err == nil {
			pods := make(map[string][]*corev1.Pod)
			for i := range podList.Items {
				pod := &podList.Items[i]
				ns := pod.ObjectMeta.Namespace
				pods[ns] = append(pods[ns], pod)
			}
			if len(pods) == 0 {
				log.Warnf("operator list pods is empty")
			}
			cli.mu.Lock()
			defer cli.mu.Unlock()
			cli.Pods = pods
			return nil
		}

		log.Warnf("operator list pods failed: %s", err)
	} else {
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
	}
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
	if cli.informer == nil {
		return nil
	}
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
