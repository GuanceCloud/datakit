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
	loggingv1alpha1 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/pkg/apis/datakits/v1alpha1"
	externalversions "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/pkg/client/informers/externalversions"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var crdWatcherG = datakit.G("crd-watcher")

type loggingConfigWatcher struct {
	client      k8sclient.Client
	coordinator *containerLogCoordinator

	queue    workqueue.DelayingInterface
	informer cache.SharedIndexInformer
	store    cache.Store

	stopCh chan struct{}
}

func newLoggingConfigWatcher(client k8sclient.Client, coordinator *containerLogCoordinator) *loggingConfigWatcher {
	return &loggingConfigWatcher{
		client:      client,
		coordinator: coordinator,
		queue:       workqueue.NewDelayingQueue(),
		stopCh:      make(chan struct{}),
	}
}

func (w *loggingConfigWatcher) start(ctx context.Context) {
	l.Info("starting logging config watcher")

	// RBAC 预检：尝试进行一次最小化的 List 调用，若无权限则退出
	if clientset := w.client.LoggingClient(); clientset != nil {
		_, err := clientset.LoggingV1alpha1().ClusterLoggingConfigs().List(ctx, metav1.ListOptions{Limit: 1})
		if apierrors.IsForbidden(err) || apierrors.IsUnauthorized(err) {
			l.Errorf("missing RBAC to access ClusterLoggingConfig: %v; exit logging config watcher", err)
			return
		}
	}

	w.setupInformer()

	crdWatcherG.Go(func(_ context.Context) error {
		w.processQueue(ctx)
		return nil
	})

	crdWatcherG.Go(func(_ context.Context) error {
		w.informer.Run(w.stopCh)
		return nil
	})

	if !cache.WaitForCacheSync(w.stopCh, w.informer.HasSynced) {
		l.Error("failed to sync informer cache")
		return
	}

	l.Info("logging config watcher started successfully")

	<-ctx.Done()
	w.stop()
}

func (w *loggingConfigWatcher) stop() {
	close(w.stopCh)
	w.queue.ShutDown()
	l.Info("logging config watcher stopped")
}

func (w *loggingConfigWatcher) setupInformer() {
	clientset := w.client.LoggingClient()
	informerFactory := externalversions.NewSharedInformerFactoryWithOptions(
		clientset, 0,
		externalversions.WithTweakListOptions(func(v *metav1.ListOptions) { v.Limit = 50 }),
	)

	w.informer = informerFactory.Logging().V1alpha1().ClusterLoggingConfigs().Informer()
	w.store = w.informer.GetStore()

	w.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			w.enqueue(obj, "add")
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			w.enqueue(newObj, "update")
		},
		DeleteFunc: func(obj interface{}) {
			w.enqueue(obj, "delete")
		},
	})
}

func (w *loggingConfigWatcher) enqueue(obj interface{}, action string) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		l.Errorf("failed to get key for object: %v", err)
		return
	}

	w.queue.AddAfter(key, time.Second)
	l.Debugf("enqueued %s event for key: %s", action, key)
}

func (w *loggingConfigWatcher) processQueue(ctx context.Context) {
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

func (w *loggingConfigWatcher) processNextItem() bool {
	keyObj, quit := w.queue.Get()
	if quit {
		return false
	}
	defer w.queue.Done(keyObj)

	key := keyObj.(string)

	obj, exists, err := w.store.GetByKey(key)
	if err != nil {
		l.Errorf("failed to get object by key %s: %v", key, err)
		return true
	}

	if !exists {
		w.coordinator.deleteCRDLoggingConfig(key)
		l.Infof("logging config deleted: %s", key)
		return true
	}

	logging, ok := obj.(*loggingv1alpha1.ClusterLoggingConfig)
	if !ok {
		l.Warnf("failed to convert object to ClusterLoggingConfig: %v", obj)
		return true
	}

	l.Debugf("processNextItem: key=%s, object=%+v", key, logging)

	crdConfig, err := newCRDLoggingConfig(key, logging)
	if err != nil {
		l.Errorf("failed to parse CRD config for key %s: %v", key, err)
		return true
	}

	w.coordinator.updateCRDLoggingConfig(key, crdConfig)
	return true
}

func startLoggingConfigWatcher(client k8sclient.Client, coordinator *containerLogCoordinator) {
	ctx, cancel := context.WithCancel(context.Background())
	watcher := newLoggingConfigWatcher(client, coordinator)

	crdWatcherG.Go(func(_ context.Context) error {
		watcher.start(ctx)
		return nil
	})

	<-datakit.Exit.Wait()
	cancel()
	l.Info("logging config watcher exiting...")
}
