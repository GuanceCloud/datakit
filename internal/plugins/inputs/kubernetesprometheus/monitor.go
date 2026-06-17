// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
)

const (
	monitorResyncPeriod = 20 * time.Second
	monitorRetryPeriod  = 5 * time.Second
)

type monitorReconcileFunc func(context.Context, string, interface{}) ([]string, error)
type monitorListFunc func(context.Context, metav1.ListOptions) (runtime.Object, error)
type monitorWatchFunc func(context.Context, metav1.ListOptions) (watch.Interface, error)

type monitorWatcher struct {
	role      Role
	informer  cache.SharedIndexInformer
	queue     workqueue.DelayingInterface
	store     cache.Store
	scrape    scrapeManagerInterface
	reconcile monitorReconcileFunc

	list             monitorListFunc
	watchUnavailable *atomic.Bool
	resourceVersions map[string]string
	taskKeys         map[string][]string
}

func newMonitorWatcher(
	role Role,
	informer cache.SharedIndexInformer,
	scrape scrapeManagerInterface,
	reconcile monitorReconcileFunc,
) *monitorWatcher {
	return &monitorWatcher{
		role:      role,
		informer:  informer,
		queue:     workqueue.NewNamedDelayingQueue(string(role)),
		store:     informer.GetStore(),
		scrape:    scrape,
		reconcile: reconcile,

		resourceVersions: make(map[string]string),
		taskKeys:         make(map[string][]string),
	}
}

func (w *monitorWatcher) Run(ctx context.Context) {
	defer w.queue.ShutDown()

	w.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			w.enqueue(obj)
		},
		UpdateFunc: func(_, obj interface{}) {
			w.enqueue(obj)
		},
		DeleteFunc: func(obj interface{}) {
			w.enqueue(obj)
		},
	})

	managerGo.Go(func(_ context.Context) error {
		for w.process(ctx) {
		}
		return nil
	})

	managerGo.Go(func(_ context.Context) error {
		w.informer.Run(ctx.Done())
		return nil
	})

	if w.list != nil {
		managerGo.Go(func(_ context.Context) error {
			w.resync(ctx)
			return nil
		})
	}

	if !cache.WaitForCacheSync(ctx.Done(), w.informer.HasSynced) {
		klog.Warnf("failed to sync %s informer cache", w.role)
		return
	}

	<-ctx.Done()
}

func (w *monitorWatcher) resync(ctx context.Context) {
	ticker := time.NewTicker(monitorResyncPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if w.watchUnavailable != nil && w.watchUnavailable.Load() {
				w.syncList(ctx)
				continue
			}
			for _, key := range w.store.ListKeys() {
				w.queue.Add(key)
			}
		}
	}
}

func (w *monitorWatcher) syncList(ctx context.Context) {
	list, err := w.list(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		klog.Warnf("failed to list %s: %s", w.role, err)
		return
	}

	items, err := meta.ExtractList(list)
	if err != nil {
		klog.Warnf("failed to extract %s list: %s", w.role, err)
		return
	}

	current := make(map[string]struct{}, len(items))
	for _, item := range items {
		key, err := cache.MetaNamespaceKeyFunc(item)
		if err != nil {
			klog.Warnf("cannot get %s key: %s", w.role, err)
			continue
		}
		current[key] = struct{}{}
		if err := w.store.Update(item); err != nil {
			klog.Warnf("failed to update %s %s store: %s", w.role, key, err)
			continue
		}
		w.queue.Add(key)
	}

	for _, key := range w.store.ListKeys() {
		if _, found := current[key]; found {
			continue
		}
		obj, exists, err := w.store.GetByKey(key)
		if err != nil || !exists {
			continue
		}
		if err := w.store.Delete(obj); err != nil {
			klog.Warnf("failed to delete %s %s from store: %s", w.role, key, err)
			continue
		}
		w.queue.Add(key)
	}
}

func (w *monitorWatcher) enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Warnf("cannot get %s key: %s", w.role, err)
		return
	}
	w.queue.Add(key)
}

func (w *monitorWatcher) process(ctx context.Context) bool {
	keyObj, quit := w.queue.Get()
	if quit {
		return false
	}
	defer w.queue.Done(keyObj)

	key, ok := keyObj.(string)
	if !ok {
		klog.Warnf("unexpected %s queue key %v", w.role, keyObj)
		return true
	}

	obj, exists, err := w.store.GetByKey(key)
	if err != nil {
		klog.Warnf("failed to get %s %s from informer store: %s", w.role, key, err)
		w.queue.AddAfter(key, monitorRetryPeriod)
		return true
	}

	if !exists {
		klog.Infof("deleted %s %s", w.role, key)
		w.terminateTasks(key)
		delete(w.resourceVersions, key)
		return true
	}

	accessor, err := meta.Accessor(obj)
	if err != nil {
		klog.Warnf("cannot access %s %s metadata: %s", w.role, key, err)
		return true
	}

	resourceVersion := accessor.GetResourceVersion()
	if oldVersion, found := w.resourceVersions[key]; found && oldVersion != resourceVersion {
		klog.Infof("updated %s %s", w.role, key)
		w.terminateTasks(key)
	} else if !found {
		klog.Infof("discovered %s %s", w.role, key)
	}

	taskKeys, err := w.reconcile(ctx, key, obj)
	if err != nil {
		klog.Warnf("failed to reconcile %s %s: %s", w.role, key, err)
		w.queue.AddAfter(key, monitorRetryPeriod)
		return true
	}

	w.cleanStaleTasks(key, taskKeys)
	w.taskKeys[key] = unique(taskKeys)
	w.resourceVersions[key] = resourceVersion
	return true
}

func (w *monitorWatcher) cleanStaleTasks(key string, taskKeys []string) {
	keep := make(map[string]struct{}, len(taskKeys))
	for _, taskKey := range taskKeys {
		keep[taskKey] = struct{}{}
	}

	for _, taskKey := range w.taskKeys[key] {
		if _, found := keep[taskKey]; !found {
			w.scrape.removeScrape(w.role, taskKey)
		}
	}
}

func (w *monitorWatcher) terminateTasks(key string) {
	for _, taskKey := range w.taskKeys[key] {
		w.scrape.removeScrape(w.role, taskKey)
	}
	delete(w.taskKeys, key)
}

func watchMonitor(
	role Role,
	watchUnavailable *atomic.Bool,
	warnOnce *sync.Once,
	watchFn monitorWatchFunc,
	options metav1.ListOptions,
) (watch.Interface, error) {
	watcher, err := watchFn(context.Background(), options)
	if apierrors.IsForbidden(err) || apierrors.IsUnauthorized(err) {
		watchUnavailable.Store(true)
		warnOnce.Do(func() {
			klog.Warnf("no watch permission for %s, falling back to periodic list", role)
		})
		return watch.NewRaceFreeFake(), nil
	}
	return watcher, err
}

func newPodMonitorInformer(client k8sclient.Client) (cache.SharedIndexInformer, monitorListFunc, *atomic.Bool) {
	monitors := client.GetPrmetheusPodMonitors(metav1.NamespaceAll)
	watchUnavailable := &atomic.Bool{}
	var warnOnce sync.Once
	list := func(ctx context.Context, options metav1.ListOptions) (runtime.Object, error) {
		return monitors.List(ctx, options)
	}
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return list(context.Background(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return watchMonitor(RolePodMonitor, watchUnavailable, &warnOnce, monitors.Watch, options)
			},
		},
		&monitoringv1.PodMonitor{},
		0,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	return informer, list, watchUnavailable
}

func newServiceMonitorInformer(client k8sclient.Client) (cache.SharedIndexInformer, monitorListFunc, *atomic.Bool) {
	monitors := client.GetPrmetheusServiceMonitors(metav1.NamespaceAll)
	watchUnavailable := &atomic.Bool{}
	var warnOnce sync.Once
	list := func(ctx context.Context, options metav1.ListOptions) (runtime.Object, error) {
		return monitors.List(ctx, options)
	}
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return list(context.Background(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return watchMonitor(RoleServiceMonitor, watchUnavailable, &warnOnce, monitors.Watch, options)
			},
		},
		&monitoringv1.ServiceMonitor{},
		0,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	return informer, list, watchUnavailable
}

func monitorObjectError(role Role, obj interface{}) error {
	return fmt.Errorf("converting to %s object failed, %T", role, obj)
}
