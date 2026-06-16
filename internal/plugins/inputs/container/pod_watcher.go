// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"os"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var podWatcherG = goroutine.G("pod-watcher")

var podLoggingScanRetryDelays = []time.Duration{
	0,
	time.Second,
	3 * time.Second,
}

type podLoggingScanAttempt int

type podWatcher struct {
	client      k8sclient.Client
	coordinator *containerLogCoordinator

	queue    workqueue.DelayingInterface
	informer cache.SharedIndexInformer

	stopCh chan struct{}
}

func newPodWatcher(client k8sclient.Client, coordinator *containerLogCoordinator) *podWatcher {
	return &podWatcher{
		client:      client,
		coordinator: coordinator,
		queue:       workqueue.NewDelayingQueue(),
		stopCh:      make(chan struct{}),
	}
}

func (w *podWatcher) start(ctx context.Context) {
	l.Info("starting pod watcher")

	// RBAC 预检：尝试进行一次最小化的 List 调用，若无权限则退出
	if clientset := w.client.KubernetesClientset(); clientset != nil {
		_, err := clientset.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{Limit: 1})
		if apierrors.IsForbidden(err) || apierrors.IsUnauthorized(err) {
			l.Errorf("missing RBAC to access Pod: %v; exit pod watcher", err)
			return
		}
	}

	w.setupInformer()

	podWatcherG.Go(func(_ context.Context) error {
		w.processQueue(ctx)
		return nil
	})

	podWatcherG.Go(func(_ context.Context) error {
		w.informer.Run(w.stopCh)
		return nil
	})

	if !cache.WaitForCacheSync(w.stopCh, w.informer.HasSynced) {
		l.Error("failed to sync informer cache")
		return
	}

	l.Info("pod watcher started successfully")

	<-ctx.Done()
	w.stop()
}

func (w *podWatcher) stop() {
	close(w.stopCh)
	w.queue.ShutDown()
	l.Info("pod watcher stopped")
}

func (w *podWatcher) setupInformer() {
	clientset := w.client.KubernetesClientset()
	informerFactory := informers.NewSharedInformerFactory(clientset, 0)

	w.informer = informerFactory.Core().V1().Pods().Informer()

	w.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				l.Warnf("failed to convert add object to Pod: %T", obj)
				return
			}
			w.enqueueLoggingScans(pod)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod, oldOK := oldObj.(*corev1.Pod)
			newPod, newOK := newObj.(*corev1.Pod)
			if !oldOK || !newOK {
				l.Warnf("failed to convert update objects to Pod: old=%T, new=%T", oldObj, newObj)
				return
			}

			// 检查 Pod 是否进入 Terminating 状态
			if newPod.DeletionTimestamp != nil {
				w.enqueue(newPod, "update")
				return
			}

			if oldPod.Spec.NodeName != newPod.Spec.NodeName ||
				(oldPod.Status.Phase != corev1.PodRunning && newPod.Status.Phase == corev1.PodRunning) {
				w.enqueueLoggingScans(newPod)
			}
		},
		// DeleteFunc 通常是在 Pod 彻底消失后触发
		DeleteFunc: func(obj interface{}) {
			w.enqueue(obj, "delete")
		},
	})
}

func (w *podWatcher) enqueueLoggingScans(pod *corev1.Pod) {
	nodeName := os.Getenv("ENV_K8S_NODE_NAME")
	if nodeName == "" {
		nodeName = datakit.DKHost
	}
	if pod.Spec.NodeName == "" || pod.Spec.NodeName != nodeName {
		return
	}

	for attempt, delay := range podLoggingScanRetryDelays {
		w.queue.AddAfter(podLoggingScanAttempt(attempt), delay)
	}
	l.Debugf("enqueued logging scans for podUID: %s", pod.UID)
}

func (w *podWatcher) enqueue(obj interface{}, action string) {
	// 处理删除事件，可能被包装在 DeletedFinalStateUnknown 中
	if deletedObj, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		obj = deletedObj.Obj
	}

	pod, ok := obj.(*corev1.Pod)
	if !ok || pod == nil {
		l.Warnf("failed to convert object to Pod: %v", obj)
		return
	}

	podUID := string(pod.UID)
	if podUID == "" {
		l.Warnf("pod UID is empty in %s event", action)
		return
	}

	w.queue.AddAfter(podUID, time.Second)
	l.Debugf("enqueued %s event for podUID: %s", action, podUID)
}

func (w *podWatcher) processQueue(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if !w.processNextItem() {
				return
			}
		}
	}
}

func (w *podWatcher) processNextItem() bool {
	queueObj, quit := w.queue.Get()
	if quit {
		return false
	}
	defer w.queue.Done(queueObj)

	var podUID string
	switch item := queueObj.(type) {
	case podLoggingScanAttempt:
		w.coordinator.requestLoggingScan()
		l.Debugf("requested logging scan, attempt=%d", item)
		return true
	case string:
		if item == "" {
			l.Warn("pod UID is empty")
			return true
		}
		podUID = item
	default:
		l.Errorf("unexpected pod queue item: %v", queueObj)
		return true
	}

	l.Debugf("processing pod terminating/deletion event: podUID=%s", podUID)
	w.coordinator.requestLoggingScan()
	return true
}

func startPodWatcher(client k8sclient.Client, coordinator *containerLogCoordinator) {
	ctx, cancel := context.WithCancel(context.Background())
	watcher := newPodWatcher(client, coordinator)

	podWatcherG.Go(func(_ context.Context) error {
		watcher.start(ctx)
		return nil
	})

	<-datakit.Exit.Wait()
	cancel()
	l.Info("pod watcher exiting...")
}
