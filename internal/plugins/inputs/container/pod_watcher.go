// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var podWatcherG = datakit.G("pod-watcher")

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
		UpdateFunc: func(oldObj, newObj interface{}) {
			newPod := newObj.(*corev1.Pod)
			oldPod := oldObj.(*corev1.Pod)

			// 检查 Pod 是否进入 Terminating 状态
			// 1. 检查 DeletionTimestamp 是否刚刚被设置（从 nil 变为非 nil）
			// 2. 或者检查 DeletionTimestamp 是否存在（Pod 处于 Terminating 状态）
			oldDeleting := oldPod.DeletionTimestamp != nil
			newDeleting := newPod.DeletionTimestamp != nil

			// Pod 刚刚进入 Terminating 状态（DeletionTimestamp 刚刚被设置）
			// 或者已经处于 Terminating 状态（DeletionTimestamp 已存在）
			if newDeleting {
				if !oldDeleting {
					// 刚刚进入 Terminating 状态
					l.Infof("Pod %s/%s is entering terminating state (DeletionTimestamp just set), removing log tasks, podUID=%s",
						newPod.Namespace, newPod.Name, string(newPod.UID))
				} else {
					// 已经处于 Terminating 状态
					l.Debugf("Pod %s/%s is already terminating (DeletionTimestamp exists), removing log tasks, podUID=%s",
						newPod.Namespace, newPod.Name, string(newPod.UID))
				}
				w.enqueue(newPod, "update")
			}
		},
		// DeleteFunc 通常是在 Pod 彻底消失后触发，
		// 作为兜底，确保即使 UpdateFunc 没有捕获到，也能执行清理
		DeleteFunc: func(obj interface{}) {
			w.enqueue(obj, "delete")
		},
	})
}

func (w *podWatcher) enqueue(obj interface{}, action string) {
	var pod *corev1.Pod

	// 处理删除事件，可能被包装在 DeletedFinalStateUnknown 中
	if deletedObj, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		obj = deletedObj.Obj
	}

	pod, ok := obj.(*corev1.Pod)
	if !ok {
		l.Warnf("failed to convert object to Pod: %v", obj)
		return
	}

	if pod == nil {
		l.Warnf("pod is nil in %s event", action)
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
	uidObj, quit := w.queue.Get()
	if quit {
		return false
	}
	defer w.queue.Done(uidObj)

	podUID, ok := uidObj.(string)
	if !ok {
		l.Errorf("failed to convert queue item to string: %v", uidObj)
		return true
	}

	if podUID == "" {
		l.Warnf("pod UID is empty")
		return true
	}

	l.Debugf("processing pod terminating/deletion event: podUID=%s", podUID)

	// 遍历 containerTasks，找到匹配的 podUID，执行 removeTask
	w.coordinator.taskMutex.RLock()
	toRemove := make([]string, 0)
	for containerID, task := range w.coordinator.containerTasks {
		if task.podUID == podUID {
			toRemove = append(toRemove, containerID)
		}
	}
	w.coordinator.taskMutex.RUnlock()

	// 在不持锁的情况下逐个删除（removeTask 内部自行加锁）
	for _, containerID := range toRemove {
		w.coordinator.removeTask(containerID)
		l.Infof("removed task for container %s due to pod terminating/deletion, podUID=%s", containerID, podUID)
	}

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
