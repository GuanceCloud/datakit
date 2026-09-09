// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
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
	client      kubernetes.Interface
	coordinator *containerLogCoordinator
	nodeName    string

	podMetadata podMetadataProvider

	queue    workqueue.DelayingInterface
	informer cache.SharedIndexInformer

	stopCh         chan struct{}
	stopOnce       sync.Once
	requestCtx     context.Context
	cancelRequests context.CancelFunc
}

func newPodWatcher(client kubernetes.Interface, coordinator *containerLogCoordinator, nodeName string) *podWatcher {
	ctx, cancel := context.WithCancel(context.Background())
	watcher := &podWatcher{
		client:         client,
		coordinator:    coordinator,
		nodeName:       nodeName,
		queue:          workqueue.NewDelayingQueue(),
		stopCh:         make(chan struct{}),
		requestCtx:     ctx,
		cancelRequests: cancel,
	}
	watcher.setupInformer()
	return watcher
}

func (w *podWatcher) start(ctx context.Context) {
	l.Info("starting pod watcher")

	workers := goroutine.NewGroup(goroutine.Option{Name: "pod-watcher-workers"})
	defer func() {
		w.stop()
		if err := workers.Wait(); err != nil {
			l.Errorf("pod watcher worker stopped with error: %s", err)
		}
	}()

	workers.Go(func(_ context.Context) error {
		w.processQueue(ctx)
		return nil
	})

	workers.Go(func(_ context.Context) error {
		w.informer.Run(w.stopCh)
		return nil
	})

	if !cache.WaitForCacheSync(ctx.Done(), w.informer.HasSynced) {
		if ctx.Err() == nil {
			l.Error("failed to sync informer cache")
		}
		return
	}

	l.Info("pod watcher started successfully")
	w.coordinator.requestLoggingScan()

	<-ctx.Done()
}

func (w *podWatcher) stop() {
	w.stopOnce.Do(func() {
		w.cancelRequests()
		close(w.stopCh)
		w.queue.ShutDown()
		l.Info("pod watcher stopped")
	})
}

func (w *podWatcher) setupInformer() {
	// The reflector owns retry/backoff, including RBAC errors. A one-shot
	// preflight must not permanently strand all consumers of this shared cache.
	w.informer = cache.NewSharedIndexInformer(&cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = w.podListOptions().FieldSelector
			return w.client.CoreV1().Pods(metav1.NamespaceAll).List(w.requestCtx, options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = w.podListOptions().FieldSelector
			return w.client.CoreV1().Pods(metav1.NamespaceAll).Watch(w.requestCtx, options)
		},
	}, &corev1.Pod{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	w.podMetadata = &informerPodMetadataProvider{
		lister:    corelisters.NewPodLister(w.informer.GetIndexer()),
		hasSynced: w.informer.HasSynced,
	}

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
			if !w.isLocalPod(newPod) {
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
	if !w.isLocalPod(pod) {
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
	if !w.isLocalPod(pod) {
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

func (w *podWatcher) podListOptions() metav1.ListOptions {
	return metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.nodeName", w.nodeName).String(),
	}
}

func (w *podWatcher) isLocalPod(pod *corev1.Pod) bool {
	return pod != nil && pod.Spec.NodeName != "" && pod.Spec.NodeName == w.nodeName
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

func startPodWatcher(watcher *podWatcher) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	podWatcherG.Go(func(_ context.Context) error {
		select {
		case <-datakit.Exit.Wait():
			cancel()
		case <-ctx.Done():
		}
		return nil
	})

	watcher.start(ctx)
	l.Info("pod watcher exiting...")
}
