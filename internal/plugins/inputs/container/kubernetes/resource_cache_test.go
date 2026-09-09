// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	climetrics "github.com/GuanceCloud/cliutils/metrics"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	statsv1alpha1 "k8s.io/kubelet/pkg/apis/stats/v1alpha1"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

type cacheTestFeeder struct {
	dkio.Feeder
	mu     sync.Mutex
	points map[point.Category][]*point.Point
}

func (f *cacheTestFeeder) Feed(category point.Category, pts []*point.Point, _ ...dkio.FeedOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.points == nil {
		f.points = make(map[point.Category][]*point.Point)
	}
	f.points[category] = append(f.points[category], pts...)
	return nil
}

func (f *cacheTestFeeder) snapshot(category point.Category) []*point.Point {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]*point.Point(nil), f.points[category]...)
}

func cacheTestConfig() *Config {
	return &Config{NodeName: "node-a", EnableK8sMetric: true, EnableK8sObject: true, EnableCollectJob: true,
		MetricCollecInterval: time.Minute, ObjectCollecInterval: 5 * time.Minute, Feeder: &cacheTestFeeder{}}
}

func waitResourceCache(t *testing.T, c *resourceCache) {
	t.Helper()
	require.Eventually(t, func() bool {
		for name := range c.resources {
			if !c.synced(name) {
				return false
			}
		}
		return true
	}, 5*time.Second, time.Millisecond)
}

func TestResourceCollectionUsesOnlyInformerSnapshots(t *testing.T) {
	meta := metav1.ObjectMeta{Name: "test", Namespace: "default", UID: "test-uid"}
	nodeMeta := metav1.ObjectMeta{Name: "node-a", UID: "node-uid"}
	objects := []runtime.Object{
		&appsv1.Deployment{ObjectMeta: meta}, &appsv1.DaemonSet{ObjectMeta: meta},
		&appsv1.StatefulSet{ObjectMeta: meta}, &appsv1.ReplicaSet{ObjectMeta: meta},
		&batchv1.Job{ObjectMeta: meta}, &batchv1.CronJob{ObjectMeta: meta},
		&corev1.Service{ObjectMeta: meta}, &corev1.Endpoints{ObjectMeta: meta},
		&corev1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "pv", UID: "pv-uid"}},
		&corev1.PersistentVolumeClaim{ObjectMeta: meta},
		&corev1.Pod{ObjectMeta: meta, Spec: corev1.PodSpec{NodeName: "node-a"}, Status: corev1.PodStatus{Phase: corev1.PodPending}},
		&corev1.Node{ObjectMeta: nodeMeta, Status: corev1.NodeStatus{Capacity: corev1.ResourceList{
			corev1.ResourceCPU: apiresource.MustParse("4"), corev1.ResourceMemory: apiresource.MustParse("8Gi"),
		}}},
	}
	client := newCacheTestClientset(t, objects...)
	cfg := cacheTestConfig()
	c := newClusterResourceCache(context.Background(), client, cfg, 1)
	c.start()
	t.Cleanup(c.stop)
	waitResourceCache(t, c)
	require.Eventually(t, func() bool { return countAPIAction(client, "watch", "") == len(c.resources) }, time.Second, time.Millisecond)
	before := len(client.Actions())
	for range 3 {
		for _, constructor := range nonNodeLocalResources {
			constructor(nil, c.cfg).gatherMetric(c.ctx, 1234)
			constructor(nil, c.cfg).gatherObject(c.ctx)
		}
		capacity := getCapacityFromNode(c, "node-a")
		require.EqualValues(t, 4000, capacity.cpuCapacityMillicores)
		require.EqualValues(t, 8*1024*1024*1024, capacity.memCapacity)
	}
	require.Len(t, client.Actions(), before, "periodic collection must not query the API")
	require.Equal(t, 1, countAPIAction(client, "list", "deployments"), "changes and periodic collection share one informer")
	feeder := cfg.Feeder.(*cacheTestFeeder)
	metrics, objectsOut := map[string]bool{}, map[string]bool{}
	for _, pt := range feeder.snapshot(point.Metric) {
		metrics[pt.Name()] = true
	}
	for _, pt := range feeder.snapshot(point.Object) {
		objectsOut[pt.Name()] = true
	}
	for _, name := range []string{"kube_deployment", "kube_daemonset", "kube_statefulset", "kube_replicaset", "kube_job", "kube_cronjob", "kube_service", "kube_endpoint", "kube_pod", "kube_node", "kubernetes"} {
		require.True(t, metrics[name], "missing metric %s", name)
	}
	for _, name := range []string{deploymentObjectClass, daemonsetObjectClass, statefulsetObjectClass, replicasetObjectClass,
		jobObjectClass, cronjobObjectClass, serviceObjectClass, podObjectClass, nodeObjectClass, persistentvolumeObjectClass, persistentvolumeclaimObjectClass} {
		require.True(t, objectsOut[name], "missing object %s", name)
	}
}

func countAPIAction(client *cacheTestClientset, verb, resourceName string) int {
	n := 0
	for _, action := range client.Actions() {
		if action.GetVerb() == verb && (resourceName == "" || action.GetResource().Resource == resourceName) {
			n++
		}
	}
	return n
}

func TestResourceCacheSkipsUnsyncedAndCancelsInitialList(t *testing.T) {
	cfg := cacheTestConfig()
	c := newResourceCache(context.Background(), cfg, "cluster", 1)
	entered := make(chan struct{})
	var requests atomic.Int64
	addResourceInformer(c, "deployment", &appsv1.Deployment{}, "",
		func(ctx context.Context, _ metav1.ListOptions) (*appsv1.DeploymentList, error) {
			requests.Add(1)
			close(entered)
			<-ctx.Done()
			return nil, ctx.Err()
		}, func(context.Context, metav1.ListOptions) (watch.Interface, error) {
			return nil, fmt.Errorf("watch before initial list completed")
		})
	c.start()
	t.Cleanup(c.stop)
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("LIST did not start")
	}
	for range 3 {
		newDeployment(nil, c.cfg).gatherMetric(context.Background(), 1234)
		newDeployment(nil, c.cfg).gatherObject(context.Background())
	}
	require.EqualValues(t, 1, requests.Load())
	require.Empty(t, cfg.Feeder.(*cacheTestFeeder).snapshot(point.Metric))
	require.Empty(t, cfg.Feeder.(*cacheTestFeeder).snapshot(point.Object))
	stopped := make(chan struct{})
	go func() { c.stop(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("initial LIST prevented shutdown")
	}
}

func TestResourceCacheBatchesAndCopiesSnapshots(t *testing.T) {
	var objects []runtime.Object
	for i := range 251 {
		objects = append(objects, &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("deploy-%d", i), Namespace: "default", Labels: map[string]string{"original": "yes"},
		}, Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "app", Image: "before"}},
		}}}})
	}
	client := newCacheTestClientset(t, objects...)
	c := newResourceCache(context.Background(), cacheTestConfig(), "cluster", 1)
	deployments := client.AppsV1().Deployments("")
	addResourceInformer(c, "deployment", &appsv1.Deployment{}, "", deployments.List, deployments.Watch)
	c.start()
	t.Cleanup(c.stop)
	waitResourceCache(t, c)
	var batches []int
	require.True(t, cachedBatches[appsv1.Deployment](c.ctx, c.cfg, "deployment", func(items []appsv1.Deployment) {
		batches = append(batches, len(items))
		for idx := range items {
			items[idx].Labels["original"] = "changed"
			items[idx].Spec.Template.Spec.Containers[0].Image = "after"
		}
	}))
	require.Equal(t, []int{100, 100, 51}, batches)
	for _, item := range c.resources["deployment"].GetStore().List() {
		deployment := item.(*appsv1.Deployment)
		require.Equal(t, "yes", deployment.Labels["original"])
		require.Equal(t, "before", deployment.Spec.Template.Spec.Containers[0].Image)
	}
	batches = nil
	require.False(t, cachedBatches[appsv1.Deployment](c.ctx, c.cfg, "deployment", func(items []appsv1.Deployment) {
		batches = append(batches, len(items))
		c.invalidate()
	}))
	require.Equal(t, []int{100}, batches, "retiring a term stops remaining batches")
}

func TestLeaderTermsAreIdempotentAndDoNotOverlap(t *testing.T) {
	client := newCacheTestClientset(t)
	var mu sync.Mutex
	active := make(map[string]int)
	client.PrependWatchReactor("*", func(action ktesting.Action) (bool, watch.Interface, error) {
		key := action.GetResource().Resource
		inner, err := client.Tracker().Watch(action.GetResource(), action.GetNamespace())
		if err != nil {
			return true, nil, err
		}
		mu.Lock()
		active[key]++
		mu.Unlock()
		return true, &cacheTestWatch{Interface: inner, stop: func() { mu.Lock(); active[key]--; mu.Unlock() }}, nil
	})
	cfg := cacheTestConfig()
	k := &Kube{cfg: cfg, apiClient: client, wake: make(chan struct{}, 1)}
	t.Cleanup(func() { k.SetLeader(false); k.reconcileCluster(context.Background()) })
	k.reconcileCluster(context.Background())
	require.Empty(t, client.Actions(), "followers must not start cluster informers")
	k.SetLeader(true)
	k.reconcileCluster(context.Background())
	first := k.cluster
	waitResourceCache(t, first)
	require.Eventually(t, func() bool { return countAPIAction(client, "watch", "") == len(first.resources) }, time.Second, time.Millisecond)
	before := len(client.Actions())
	for range 5 {
		k.SetLeader(true)
		k.reconcileCluster(context.Background())
	}
	require.Same(t, first, k.cluster)
	require.Len(t, client.Actions(), before)
	for range 5 {
		old := k.cluster
		k.SetLeader(false)
		require.False(t, old.usable(), "Pause invalidates immediately, before the collection loop wakes")
		k.SetLeader(false)
		k.SetLeader(true)
		k.reconcileCluster(context.Background())
		require.NotSame(t, old, k.cluster)
		require.Greater(t, k.cluster.id, old.id)
		waitResourceCache(t, k.cluster)
		require.False(t, cachedBatches[corev1.Pod](context.Background(), old.cfg, "pod", func([]corev1.Pod) { t.Fatal("retired cache supplied pods") }))
		require.NoError(t, old.cfg.Feeder.Feed(point.Object, []*point.Point{point.NewPoint("retired", nil)}))
	}
	require.Empty(t, cfg.Feeder.(*cacheTestFeeder).snapshot(point.Object))
	k.SetLeader(false)
	k.reconcileCluster(context.Background())
	require.Nil(t, k.cluster)
	// HTTP cancellation reaches the test server asynchronously.
	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		for _, n := range active {
			if n != 0 {
				return false
			}
		}
		return true
	}, time.Second, time.Millisecond)
}

type cacheTestWatch struct {
	watch.Interface
	once sync.Once
	stop func()
}

func (w *cacheTestWatch) Stop() { w.once.Do(func() { w.Interface.Stop(); w.stop() }) }

func TestLocalCacheSurvivesElectionAndUsesNodeSelectors(t *testing.T) {
	client := newCacheTestClientset(t)
	cfg := cacheTestConfig()
	cfg.NodeLocal = true
	local := newLocalResourceCache(context.Background(), client, cfg)
	local.start()
	t.Cleanup(local.stop)
	waitResourceCache(t, local)
	k := &Kube{cfg: cfg, apiClient: client, local: local, wake: make(chan struct{}, 1)}
	t.Cleanup(func() { k.SetLeader(false); k.reconcileCluster(context.Background()) })
	for range 2 {
		k.SetLeader(true)
		k.reconcileCluster(context.Background())
		waitResourceCache(t, k.cluster)
		k.SetLeader(false)
		k.reconcileCluster(context.Background())
	}
	require.True(t, local.synced("pod"))
	require.True(t, local.synced("node"))
	podLocalLists, nodeLocalLists, pendingLists := 0, 0, 0
	for _, action := range client.Actions() {
		a, ok := action.(ktesting.ListAction)
		if !ok {
			continue
		}
		selector := a.GetListRestrictions().Fields.String()
		switch a.GetResource().Resource {
		case "pods":
			if selector == fields.OneTermEqualSelector("spec.nodeName", cfg.NodeName).String() {
				podLocalLists++
			}
			if selector == "status.phase=Pending" {
				pendingLists++
			}
		case "nodes":
			if selector == "metadata.name=node-a" {
				nodeLocalLists++
			}
		}
	}
	require.Equal(t, 1, podLocalLists)
	require.Equal(t, 1, nodeLocalLists)
	require.Equal(t, 2, pendingLists)
}

func TestLocalCacheBorrowsRuntimePodInformer(t *testing.T) {
	client := newCacheTestClientset(t)
	cfg := cacheTestConfig()
	cfg.NodeLocal = true
	runtimeCache := newLocalResourceCache(context.Background(), client, cfg)
	runtimeCache.start()
	t.Cleanup(runtimeCache.stop)
	waitResourceCache(t, runtimeCache)
	cfg.LocalPodInformer = runtimeCache.resources["pod"]
	local := newLocalResourceCache(context.Background(), client, cfg)
	local.start()
	t.Cleanup(local.stop)
	waitResourceCache(t, local)
	require.Same(t, cfg.LocalPodInformer, local.resources["pod"])
	local.stop()
	require.True(t, runtimeCache.synced("pod"))
	require.Equal(t, 1, countAPIAction(client, "list", "pods"))
}

func TestDisabledResourcesDoNotStartInformers(t *testing.T) {
	client := newCacheTestClientset(t)
	cfg := &Config{NodeLocal: true}
	c := newClusterResourceCache(context.Background(), client, cfg, 1)
	require.Len(t, c.resources, 3, "only existing change streams are needed with periodic collection disabled")
	c.stop()
	cfg.EnableK8sMetric = true
	c = newClusterResourceCache(context.Background(), client, cfg, 2)
	t.Cleanup(c.stop)
	require.NotContains(t, c.resources, "job")
	require.NotContains(t, c.resources, "persistentvolume")
	require.NotContains(t, c.resources, "persistentvolumeclaim")
	require.Contains(t, c.resources, "endpoint")
}

func TestInformerChangesReuseCacheAndIgnoreRetiredTerm(t *testing.T) {
	require.NoError(t, changes.LoadK8sManifest())
	dep := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default", UID: types.UID("app")},
		Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Image: "before"}}}}}}
	client := newCacheTestClientset(t, dep)
	cfg := cacheTestConfig()
	c := newClusterResourceCache(context.Background(), client, cfg, 1)
	c.start()
	t.Cleanup(c.stop)
	waitResourceCache(t, c)
	require.Eventually(t, func() bool { return countAPIAction(client, "watch", "deployments") == 1 }, time.Second, time.Millisecond)
	updated := dep.DeepCopy()
	updated.Spec.Template.Spec.Containers[0].Image = "after"
	_, err := client.AppsV1().Deployments("default").Update(c.ctx, updated, metav1.UpdateOptions{})
	require.NoError(t, err)
	feeder := cfg.Feeder.(*cacheTestFeeder)
	require.Eventually(t, func() bool { return len(feeder.snapshot(point.KeyEvent)) > 0 }, 5*time.Second, time.Millisecond)
	require.Contains(t, feeder.snapshot(point.KeyEvent)[0].Get("diff"), "after")
	require.Equal(t, 1, countAPIAction(client, "list", "deployments"))
	c.stop()
	before := len(feeder.snapshot(point.KeyEvent))
	handler := c.changeHandler("deployment", cache.ResourceEventHandlerFuncs{AddFunc: func(interface{}) { t.Fatal("retired handler invoked") },
		UpdateFunc: func(interface{}, interface{}) { t.Fatal("retired handler invoked") }, DeleteFunc: func(interface{}) { t.Fatal("retired handler invoked") }})
	handler.OnAdd(updated)
	handler.OnUpdate(dep, updated)
	handler.OnDelete(cache.DeletedFinalStateUnknown{Key: "default/app", Obj: updated})
	require.Equal(t, before, len(feeder.snapshot(point.KeyEvent)))
}

func TestInformerRelistsAfterResourceVersionExpires(t *testing.T) {
	client := newCacheTestClientset(t, &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "before", Namespace: "default"}})
	var mu sync.Mutex
	var watches []*watch.RaceFreeFakeWatcher
	client.PrependWatchReactor("pods", func(ktesting.Action) (bool, watch.Interface, error) {
		w := watch.NewRaceFreeFake()
		mu.Lock()
		watches = append(watches, w)
		mu.Unlock()
		return true, w, nil
	})
	c := newResourceCache(context.Background(), cacheTestConfig(), "cluster", 1)
	pods := client.CoreV1().Pods("")
	addResourceInformer(c, "pod", &corev1.Pod{}, "", pods.List, pods.Watch)
	c.start()
	t.Cleanup(c.stop)
	waitResourceCache(t, c)
	require.Eventually(t, func() bool { mu.Lock(); defer mu.Unlock(); return len(watches) == 1 }, time.Second, time.Millisecond)
	require.NoError(t, client.Tracker().Delete(corev1.SchemeGroupVersion.WithResource("pods"), "default", "before"))
	require.NoError(t, client.Tracker().Add(&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "after", Namespace: "default"}}))
	mu.Lock()
	first := watches[0]
	mu.Unlock()
	first.Error(&metav1.Status{Status: metav1.StatusFailure, Reason: metav1.StatusReasonExpired, Code: 410})
	require.Eventually(t, func() bool {
		_, exists, err := c.resources["pod"].GetStore().GetByKey("default/after")
		return err == nil && exists && countAPIAction(client, "list", "pods") >= 2
	}, 10*time.Second, 10*time.Millisecond)
	_, exists, err := c.resources["pod"].GetStore().GetByKey("default/before")
	require.NoError(t, err)
	require.False(t, exists)
}

type blockingCacheFeeder struct {
	dkio.Feeder
	entered chan struct{}
	release chan struct{}
	calls   atomic.Int64
}

func (f *blockingCacheFeeder) Feed(point.Category, []*point.Point, ...dkio.FeedOption) error {
	f.calls.Add(1)
	f.entered <- struct{}{}
	<-f.release
	return nil
}

func TestResumeJoinsAdmittedFeedBeforeStartingNextTerm(t *testing.T) {
	client := newCacheTestClientset(t)
	feeder := &blockingCacheFeeder{entered: make(chan struct{}, 2), release: make(chan struct{})}
	cfg := cacheTestConfig()
	cfg.Feeder = feeder
	k := &Kube{cfg: cfg, apiClient: client, wake: make(chan struct{}, 1)}
	t.Cleanup(func() { k.SetLeader(false); k.reconcileCluster(context.Background()) })
	k.SetLeader(true)
	k.reconcileCluster(context.Background())
	old := k.cluster
	waitResourceCache(t, old)
	feedDone := make(chan struct{})
	go func() { _ = old.cfg.Feeder.Feed(point.Object, nil); close(feedDone) }()
	<-feeder.entered
	k.SetLeader(false)
	require.False(t, old.usable(), "Pause must invalidate without waiting for a blocked feeder")
	k.SetLeader(true)
	resumed := make(chan struct{})
	go func() { k.reconcileCluster(context.Background()); close(resumed) }()
	select {
	case <-resumed:
		t.Error("new term started before the old feed drained")
	case <-time.After(20 * time.Millisecond):
	}
	close(feeder.release)
	select {
	case <-resumed:
	case <-time.After(5 * time.Second):
		t.Fatal("Resume did not finish after feed drained")
	}
	<-feedDone
	require.NotSame(t, old, k.cluster)
	require.NoError(t, old.cfg.Feeder.Feed(point.Object, nil))
	require.EqualValues(t, 1, feeder.calls.Load(), "retired term admitted a new feed")
}

type cacheTestKubelet struct {
	k8sClient
	calls atomic.Int64
}

func (k *cacheTestKubelet) GetStatsSummary() (*statsv1alpha1.Summary, error) {
	k.calls.Add(1)
	return &statsv1alpha1.Summary{Pods: []statsv1alpha1.PodStats{{PodRef: statsv1alpha1.PodReference{Name: "app", Namespace: "default"}}}}, nil
}

func TestPodCollectionWaitsForNodeCapacitySync(t *testing.T) {
	client := newCacheTestClientset(t, &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default"},
		Spec: corev1.PodSpec{NodeName: "node-a"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}})
	cfg := cacheTestConfig()
	cfg.NodeLocal, cfg.EnablePodMetric = true, true
	c := newResourceCache(context.Background(), cfg, "node-local", 0)
	pods := client.CoreV1().Pods("")
	addResourceInformer(c, "pod", &corev1.Pod{}, "", pods.List, pods.Watch)
	ready := make(chan struct{})
	addResourceInformer(c, "node", &corev1.Node{}, "", func(ctx context.Context, _ metav1.ListOptions) (*corev1.NodeList, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ready:
		}
		return &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "node-a"}, Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{corev1.ResourceCPU: apiresource.MustParse("4"), corev1.ResourceMemory: apiresource.MustParse("8Gi")},
		}}}}, nil
	}, func(context.Context, metav1.ListOptions) (watch.Interface, error) {
		return watch.NewRaceFreeFake(), nil
	})
	c.start()
	t.Cleanup(c.stop)
	require.Eventually(t, c.resources["pod"].HasSynced, 5*time.Second, time.Millisecond)
	require.False(t, c.resources["node"].HasSynced())
	kubelet := &cacheTestKubelet{}
	newPodLocal(kubelet, c.cfg).gatherMetric(c.ctx, 1234)
	newPodLocal(kubelet, c.cfg).gatherObject(c.ctx)
	feeder := cfg.Feeder.(*cacheTestFeeder)
	require.Empty(t, feeder.snapshot(point.Metric))
	require.Empty(t, feeder.snapshot(point.Object))
	require.Zero(t, kubelet.calls.Load())
	stateOnlyCfg := *c.cfg
	stateOnlyCfg.EnablePodMetric = false
	stateOnlyFeeder := &cacheTestFeeder{}
	stateOnlyCfg.Feeder = stateOnlyFeeder
	newPodLocal(kubelet, &stateOnlyCfg).gatherMetric(c.ctx, 1234)
	require.NotEmpty(t, stateOnlyFeeder.snapshot(point.Metric), "state-only Pod metrics do not require Node permissions")
	require.Zero(t, kubelet.calls.Load())
	close(ready)
	waitResourceCache(t, c)
	newPodLocal(kubelet, c.cfg).gatherMetric(c.ctx, 1234)
	newPodLocal(kubelet, c.cfg).gatherObject(c.ctx)
	require.EqualValues(t, 2, kubelet.calls.Load())
	objects := feeder.snapshot(point.Object)
	require.Len(t, objects, 1)
	require.EqualValues(t, int64(8*1024*1024*1024), objects[0].Get("mem_capacity"))
}

type cacheTestSyncInformer struct {
	cache.SharedIndexInformer
	synced  atomic.Bool
	entered chan struct{}
}

func (i *cacheTestSyncInformer) HasSynced() bool {
	if i.synced.Load() {
		return true
	}
	select {
	case i.entered <- struct{}{}:
	default:
	}
	return false
}

func TestLocalObjectsStartAfterCacheSync(t *testing.T) {
	for _, synced := range []bool{true, false} {
		t.Run(fmt.Sprintf("synced=%t", synced), func(t *testing.T) {
			client := newCacheTestClientset(t, &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-a"}})
			pods := &cacheTestSyncInformer{
				SharedIndexInformer: cache.NewSharedIndexInformer(nil, &corev1.Pod{}, 0, nil),
				entered:             make(chan struct{}, 1),
			}
			require.NoError(t, pods.GetStore().Add(&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default", UID: "app", Annotations: map[string]string{
					annotationPromExport: "[[inputs.prom]]\nurl='http://127.0.0.1:1/metrics'\ninterval='30s'\n",
				}},
				Spec:   corev1.PodSpec{NodeName: "node-a"},
				Status: corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "127.0.0.1"},
			}))
			cfg := cacheTestConfig()
			cfg.NodeLocal, cfg.LocalPodInformer = true, pods
			cfg.MetricCollecInterval, cfg.ObjectCollecInterval = time.Hour, time.Hour
			k := &Kube{cfg: cfg, client: &cacheTestKubelet{}, apiClient: client, wake: make(chan struct{}, 1)}
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			go func() { k.run(ctx); close(done) }()
			t.Cleanup(func() {
				cancel()
				select {
				case <-done:
				case <-time.After(5 * time.Second):
					t.Error("collector did not stop while waiting for cache sync")
				}
			})
			require.Eventually(t, func() bool { return countAPIAction(client, "watch", "nodes") == 1 }, time.Second, time.Millisecond)
			feeder := cfg.Feeder.(*cacheTestFeeder)
			require.Empty(t, feeder.snapshot(point.Object))
			require.Zero(t, promValue(t, podAnnotationPromActiveTasks))
			requests := len(client.Actions())
			if synced {
				pods.synced.Store(true)
				require.Eventually(t, func() bool {
					return len(feeder.snapshot(point.Object)) == 1 && promValue(t, podAnnotationPromActiveTasks) == 1
				}, time.Second, time.Millisecond, "first objects and Prom tasks must not wait for the hourly ticker")
				require.Equal(t, podObjectClass, feeder.snapshot(point.Object)[0].Name())
			}
			cancel()
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("collector did not stop")
			}
			require.Len(t, client.Actions(), requests, "initial collection must not query the API")
			require.Zero(t, promValue(t, podAnnotationPromActiveTasks))
		})
	}
}

func TestChangeNotificationWaitsForInitialSync(t *testing.T) {
	c := newResourceCache(context.Background(), cacheTestConfig(), "cluster", 1)
	t.Cleanup(c.stop)
	informer := &cacheTestSyncInformer{entered: make(chan struct{}, 1)}
	c.resources["deployment"] = informer
	var received atomic.Int64
	handler := c.changeHandler("deployment", cache.ResourceEventHandlerFuncs{AddFunc: func(interface{}) { received.Add(1) }})
	done := make(chan struct{})
	go func() {
		handler.OnAdd(&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(time.Now())}})
		close(done)
	}()
	select {
	case <-informer.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not wait for sync")
	}
	require.Zero(t, received.Load())
	informer.synced.Store(true)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not resume after sync")
	}
	require.EqualValues(t, 1, received.Load())
}

func TestResourceCacheRejectsPartialInitialList(t *testing.T) {
	cfg := cacheTestConfig()
	c := newResourceCache(context.Background(), cfg, "cluster", 1)
	var calls atomic.Int64
	var allowSecondPage atomic.Bool
	addResourceInformer(c, "deployment", &appsv1.Deployment{}, "", func(_ context.Context, options metav1.ListOptions) (*appsv1.DeploymentList, error) {
		calls.Add(1)
		if options.Continue == "" {
			return &appsv1.DeploymentList{ListMeta: metav1.ListMeta{ResourceVersion: "1", Continue: "next"},
				Items: []appsv1.Deployment{{ObjectMeta: metav1.ObjectMeta{Name: "first", Namespace: "default"}}}}, nil
		}
		if !allowSecondPage.Load() {
			return nil, fmt.Errorf("second page unavailable")
		}
		return &appsv1.DeploymentList{ListMeta: metav1.ListMeta{ResourceVersion: "1"},
			Items: []appsv1.Deployment{{ObjectMeta: metav1.ObjectMeta{Name: "second", Namespace: "default"}}}}, nil
	}, func(context.Context, metav1.ListOptions) (watch.Interface, error) {
		return watch.NewRaceFreeFake(), nil
	})
	c.start()
	t.Cleanup(c.stop)
	require.Eventually(t, func() bool { return calls.Load() >= 2 }, 5*time.Second, time.Millisecond)
	require.False(t, c.synced("deployment"))
	newDeployment(nil, c.cfg).gatherMetric(c.ctx, 1234)
	newDeployment(nil, c.cfg).gatherObject(c.ctx)
	feeder := cfg.Feeder.(*cacheTestFeeder)
	require.Empty(t, feeder.snapshot(point.Metric))
	require.Empty(t, feeder.snapshot(point.Object))
	allowSecondPage.Store(true)
	waitResourceCache(t, c)
	newDeployment(nil, c.cfg).gatherObject(c.ctx)
	require.Len(t, feeder.snapshot(point.Object), 2)
}

func TestResourceCacheMetricsTrackSyncAndRetirement(t *testing.T) {
	client := newCacheTestClientset(t)
	c := newResourceCache(context.Background(), &Config{}, "cluster", 42)
	deployments := client.AppsV1().Deployments("")
	addResourceInformer(c, "deployment", &appsv1.Deployment{}, "", deployments.List, deployments.Watch)
	pods := &cacheTestSyncInformer{entered: make(chan struct{}, 1)}
	c.resources["pod"], c.borrowed["pod"] = pods, true
	starts := promValue(t, informerStartsVec.WithLabelValues("cluster"))
	c.start()
	t.Cleanup(c.stop)
	require.Eventually(t, func() bool {
		return countAPIAction(client, "watch", "deployments") == 1 &&
			promValue(t, informerUnsyncedVec.WithLabelValues("cluster")) == 1
	}, time.Second, time.Millisecond)
	require.Equal(t, starts+1, promValue(t, informerStartsVec.WithLabelValues("cluster")), "borrowed informers must not count as new starts")
	pods.synced.Store(true)
	require.Eventually(t, func() bool {
		return promValue(t, informerUnsyncedVec.WithLabelValues("cluster")) == 0
	}, time.Second, time.Millisecond)
	c.invalidate()
	require.Zero(t, promValue(t, informerUnsyncedVec.WithLabelValues("cluster")))

	families, err := climetrics.Gather()
	require.NoError(t, err)
	var names []string
	for _, family := range families {
		if !strings.HasPrefix(family.GetName(), "datakit_input_container_kubernetes_informer_") {
			continue
		}
		names = append(names, family.GetName())
		require.LessOrEqual(t, len(family.GetMetric()), 2)
		for _, metric := range family.GetMetric() {
			require.Len(t, metric.GetLabel(), 1)
			require.Equal(t, "scope", metric.GetLabel()[0].GetName())
			require.Contains(t, []string{"node-local", "cluster"}, metric.GetLabel()[0].GetValue())
		}
	}
	require.ElementsMatch(t, []string{
		"datakit_input_container_kubernetes_informer_starts_total",
		"datakit_input_container_kubernetes_informer_unsynced",
	}, names)
}
