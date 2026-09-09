// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

// A resourceCache belongs either to the process (node-local) or to one leader
// term. Its informer set is immutable once started. Retired terms never serve
// data, even if a caller still holds a snapshot or an informer has not exited.
type resourceCache struct {
	id          uint64
	scope       string
	startedAt   time.Time
	ctx         context.Context
	cancel      context.CancelFunc
	valid       atomic.Bool
	admissionMu sync.Mutex
	feeding     sync.WaitGroup
	wg          sync.WaitGroup
	resources   map[string]cache.SharedIndexInformer
	borrowed    map[string]bool
	cfg         *Config
}

func newResourceCache(parent context.Context, cfg *Config, scope string, id uint64) *resourceCache {
	ctx, cancel := context.WithCancel(parent)
	copyConfig := *cfg
	c := &resourceCache{
		id: id, scope: scope, startedAt: time.Now().UTC(), ctx: ctx, cancel: cancel,
		resources: make(map[string]cache.SharedIndexInformer), borrowed: make(map[string]bool), cfg: &copyConfig,
	}
	c.valid.Store(true)
	copyConfig.resourceCache = c
	copyConfig.Feeder = &cacheFeeder{Feeder: cfg.Feeder, cache: c}
	return c
}

func newLocalResourceCache(parent context.Context, client kubeclient.Interface, cfg *Config) *resourceCache {
	c := newResourceCache(parent, cfg, "node-local", 0)
	if cfg.LocalPodInformer != nil {
		c.resources["pod"] = cfg.LocalPodInformer
		c.borrowed["pod"] = true // The runtime/logging watcher owns Run and shutdown.
	} else {
		pods := client.CoreV1().Pods(metav1.NamespaceAll)
		addResourceInformer(c, "pod", &corev1.Pod{},
			fields.OneTermEqualSelector("spec.nodeName", cfg.NodeName).String(), pods.List, pods.Watch)
	}
	nodes := client.CoreV1().Nodes()
	addResourceInformer(c, "node", &corev1.Node{},
		fields.OneTermEqualSelector("metadata.name", cfg.NodeName).String(), nodes.List, nodes.Watch)
	return c
}

func newClusterResourceCache(parent context.Context, client kubeclient.Interface, cfg *Config, id uint64) *resourceCache {
	c := newResourceCache(parent, cfg, "cluster", id)
	if cfg.EnableK8sMetric || cfg.EnableK8sObject {
		selector := ""
		if cfg.NodeLocal {
			selector = fields.OneTermEqualSelector("status.phase", string(corev1.PodPending)).String()
		}
		pods := client.CoreV1().Pods(metav1.NamespaceAll)
		addResourceInformer(c, "pod", &corev1.Pod{}, selector, pods.List, pods.Watch)
		nodes := client.CoreV1().Nodes()
		addResourceInformer(c, "node", &corev1.Node{}, "", nodes.List, nodes.Watch)
		replicasets := client.AppsV1().ReplicaSets(metav1.NamespaceAll)
		addResourceInformer(c, "replicaset", &appsv1.ReplicaSet{}, "", replicasets.List, replicasets.Watch)
		cronjobs := client.BatchV1().CronJobs(metav1.NamespaceAll)
		addResourceInformer(c, "cronjob", &batchv1.CronJob{}, "", cronjobs.List, cronjobs.Watch)
		services := client.CoreV1().Services(metav1.NamespaceAll)
		addResourceInformer(c, "service", &corev1.Service{}, "", services.List, services.Watch)
		if cfg.EnableCollectJob {
			jobs := client.BatchV1().Jobs(metav1.NamespaceAll)
			addResourceInformer(c, "job", &batchv1.Job{}, "", jobs.List, jobs.Watch)
		}
	}
	if cfg.EnableK8sMetric {
		endpoints := client.CoreV1().Endpoints(metav1.NamespaceAll)
		addResourceInformer(c, "endpoint", &corev1.Endpoints{}, "", endpoints.List, endpoints.Watch)
	}
	if cfg.EnableK8sObject {
		volumes := client.CoreV1().PersistentVolumes()
		addResourceInformer(c, "persistentvolume", &corev1.PersistentVolume{}, "", volumes.List, volumes.Watch)
		claims := client.CoreV1().PersistentVolumeClaims(metav1.NamespaceAll)
		addResourceInformer(c, "persistentvolumeclaim", &corev1.PersistentVolumeClaim{}, "", claims.List, claims.Watch)
	}
	// These informers also drive object changes when periodic collection is off.
	deployments := client.AppsV1().Deployments(metav1.NamespaceAll)
	addResourceInformer(c, "deployment", &appsv1.Deployment{}, "", deployments.List, deployments.Watch)
	daemonsets := client.AppsV1().DaemonSets(metav1.NamespaceAll)
	addResourceInformer(c, "daemonset", &appsv1.DaemonSet{}, "", daemonsets.List, daemonsets.Watch)
	statefulsets := client.AppsV1().StatefulSets(metav1.NamespaceAll)
	addResourceInformer(c, "statefulset", &appsv1.StatefulSet{}, "", statefulsets.List, statefulsets.Watch)
	newDeployment(nil, c.cfg).(*deployment).addChangeHandler(c.resources["deployment"])
	newDaemonset(nil, c.cfg).(*daemonset).addChangeHandler(c.resources["daemonset"])
	newStatefulset(nil, c.cfg).(*statefulset).addChangeHandler(c.resources["statefulset"])
	return c
}

func addResourceInformer[T runtime.Object](c *resourceCache, name string, object runtime.Object, selector string,
	list func(context.Context, metav1.ListOptions) (T, error),
	watchFn func(context.Context, metav1.ListOptions) (watch.Interface, error),
) {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = selector
			// Bind requests to the term, including the initial LIST. Closing Run's
			// stop channel alone does not cancel an in-flight HTTP request in v0.25.
			return list(c.ctx, options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = selector
			return watchFn(c.ctx, options)
		},
	}
	c.resources[name] = cache.NewSharedIndexInformer(lw, object, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
}

func (c *resourceCache) start() {
	informerUnsyncedVec.WithLabelValues(c.scope).Set(float64(len(c.resources)))
	for name, informer := range c.resources {
		if !c.borrowed[name] {
			informerStartsVec.WithLabelValues(c.scope).Inc()
			c.wg.Go(func() { informer.Run(c.ctx.Done()) })
		}
		c.wg.Go(func() {
			if cache.WaitForCacheSync(c.ctx.Done(), informer.HasSynced) {
				c.admissionMu.Lock()
				if c.usable() {
					informerUnsyncedVec.WithLabelValues(c.scope).Dec()
				}
				c.admissionMu.Unlock()
			}
		})
	}
	if c.cfg.EnableK8sObject && (c.scope == "node-local" || !c.cfg.NodeLocal) {
		manager := newPromTaskManager(c.ctx, c.cfg)
		c.cfg.promPodPublisher = manager
		c.wg.Go(manager.run)
	}
}

func (c *resourceCache) usable() bool {
	return c != nil && c.valid.Load() && c.ctx.Err() == nil
}

func (c *resourceCache) invalidate() {
	c.admissionMu.Lock()
	c.valid.Store(false)
	informerUnsyncedVec.WithLabelValues(c.scope).Set(0)
	c.admissionMu.Unlock()
	c.cancel()
}

func (c *resourceCache) stop() {
	c.invalidate()
	c.wg.Wait() // Includes informer event handlers and Prometheus scrape workers.
	c.feeding.Wait()
}

func (c *resourceCache) synced(name string) bool {
	return c.usable() && c.resources[name] != nil && c.resources[name].HasSynced()
}

// Process a single snapshot in bounded batches. Deep-copy only the current
// batch: builders and downstream feeders must never own informer objects.
func cachedBatches[T any, P interface {
	*T
	runtime.Object
}](ctx context.Context, cfg *Config, name string, consume func([]T)) bool {
	c := cfg.resourceCache
	if !c.synced(name) {
		return false
	}
	items := c.resources[name].GetStore().List()
	for offset := 0; ; {
		if ctx.Err() != nil || !c.usable() {
			return false
		}
		end := min(offset+int(queryLimit), len(items))
		batch := make([]T, 0, end-offset)
		for _, item := range items[offset:end] {
			obj, ok := item.(P)
			if !ok {
				klog.Warnf("unexpected cached %s type: %T", name, item)
				return false
			}
			batch = append(batch, *obj.DeepCopyObject().(P))
		}
		consume(batch)
		if end == len(items) {
			return c.usable() && ctx.Err() == nil
		}
		offset = end
	}
}

type cacheFeeder struct {
	dkio.Feeder
	cache *resourceCache
}

func (c *resourceCache) changeHandler(name string, handler cache.ResourceEventHandlerFuncs) cache.ResourceEventHandlerFuncs {
	ready := func() bool {
		// Initial Add notifications may include resources created while the LIST
		// was in flight. Wait for sync instead of dropping these notifications.
		return c.usable() && cache.WaitForCacheSync(c.ctx.Done(), c.resources[name].HasSynced) && c.usable()
	}
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if ready() {
				handler.OnAdd(obj.(runtime.Object).DeepCopyObject())
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if ready() {
				handler.OnUpdate(oldObj.(runtime.Object).DeepCopyObject(), newObj.(runtime.Object).DeepCopyObject())
			}
		},
		DeleteFunc: func(obj interface{}) {
			if deleted, ok := obj.(cache.DeletedFinalStateUnknown); ok {
				obj = deleted.Obj
			}
			if ready() {
				handler.OnDelete(obj.(runtime.Object).DeepCopyObject())
			}
		},
	}
}

func (f *cacheFeeder) Feed(category point.Category, pts []*point.Point, opts ...dkio.FeedOption) error {
	f.cache.admissionMu.Lock()
	if !f.cache.usable() || f.Feeder == nil {
		f.cache.admissionMu.Unlock()
		// These points have not been handed to IO.
		if pool := point.GetPointPool(); pool != nil {
			for _, pt := range pts {
				pool.Put(pt)
			}
		}
		return nil
	}
	// Admission is linearized with invalidation. Already admitted IO cannot be
	// recalled, so teardown joins it before starting the next term. Never hold
	// the admission lock across IO: Pause must invalidate even under backpressure.
	f.cache.feeding.Add(1)
	f.cache.admissionMu.Unlock()
	defer f.cache.feeding.Done()
	return f.Feeder.Feed(category, pts, opts...)
}
