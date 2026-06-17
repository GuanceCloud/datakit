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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

type recordingScrapeManager struct {
	removed []string
}

func (*recordingScrapeManager) runWorker(context.Context, int, time.Duration) {}
func (*recordingScrapeManager) registerScrape(Role, string, string, scraper)  {}
func (*recordingScrapeManager) refreshTraits(Role, string, string)            {}
func (*recordingScrapeManager) isTraitsExists(Role, string, string) bool      { return false }
func (*recordingScrapeManager) isScrapeExists(Role, string, string) bool      { return false }
func (*recordingScrapeManager) tryCleanScrapes(Role, string, []string) bool   { return false }

func (m *recordingScrapeManager) removeScrape(_ Role, key string) {
	m.removed = append(m.removed, key)
}

func TestMonitorWatcherProcess(t *testing.T) {
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{},
		&monitoringv1.PodMonitor{},
		0,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	scrapeManager := &recordingScrapeManager{}
	taskKeys := []string{"task-a"}
	reconcileCount := 0
	watcher := newMonitorWatcher(
		RolePodMonitor,
		informer,
		scrapeManager,
		func(context.Context, string, interface{}) ([]string, error) {
			reconcileCount++
			return taskKeys, nil
		},
	)
	t.Cleanup(watcher.queue.ShutDown)

	ctx := context.Background()
	key := "default/example"
	monitor := &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       "default",
			Name:            "example",
			ResourceVersion: "1",
		},
	}

	require.NoError(t, watcher.store.Add(monitor))
	watcher.queue.Add(key)
	require.True(t, watcher.process(ctx))
	require.Equal(t, 1, reconcileCount)
	require.Equal(t, []string{"task-a"}, watcher.taskKeys[key])
	require.Equal(t, "1", watcher.resourceVersions[key])
	require.Empty(t, scrapeManager.removed)

	taskKeys = []string{"task-b"}
	watcher.queue.Add(key)
	require.True(t, watcher.process(ctx))
	require.Equal(t, 2, reconcileCount)
	require.Equal(t, []string{"task-b"}, watcher.taskKeys[key])
	require.Equal(t, []string{"task-a"}, scrapeManager.removed)

	updated := monitor.DeepCopy()
	updated.ResourceVersion = "2"
	require.NoError(t, watcher.store.Update(updated))
	taskKeys = []string{"task-c"}
	watcher.queue.Add(key)
	require.True(t, watcher.process(ctx))
	require.Equal(t, 3, reconcileCount)
	require.Equal(t, []string{"task-c"}, watcher.taskKeys[key])
	require.Equal(t, "2", watcher.resourceVersions[key])
	require.Equal(t, []string{"task-a", "task-b"}, scrapeManager.removed)

	require.NoError(t, watcher.store.Delete(updated))
	watcher.queue.Add(key)
	require.True(t, watcher.process(ctx))
	require.Equal(t, 3, reconcileCount)
	require.NotContains(t, watcher.taskKeys, key)
	require.NotContains(t, watcher.resourceVersions, key)
	require.Equal(t, []string{"task-a", "task-b", "task-c"}, scrapeManager.removed)
}

func TestMonitorWatcherSyncList(t *testing.T) {
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{},
		&monitoringv1.PodMonitor{},
		0,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	watcher := newMonitorWatcher(RolePodMonitor, informer, &recordingScrapeManager{}, nil)
	t.Cleanup(watcher.queue.ShutDown)

	old := &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "old", ResourceVersion: "1"},
	}
	stale := &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "stale", ResourceVersion: "1"},
	}
	require.NoError(t, watcher.store.Add(old))
	require.NoError(t, watcher.store.Add(stale))

	updated := old.DeepCopy()
	updated.ResourceVersion = "2"
	added := &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "added", ResourceVersion: "1"},
	}
	watcher.list = func(context.Context, metav1.ListOptions) (runtime.Object, error) {
		return &monitoringv1.PodMonitorList{Items: []*monitoringv1.PodMonitor{updated, added}}, nil
	}

	watcher.syncList(context.Background())

	obj, exists, err := watcher.store.GetByKey("default/old")
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, "2", obj.(*monitoringv1.PodMonitor).ResourceVersion)

	_, exists, err = watcher.store.GetByKey("default/added")
	require.NoError(t, err)
	require.True(t, exists)

	_, exists, err = watcher.store.GetByKey("default/stale")
	require.NoError(t, err)
	require.False(t, exists)
	require.Equal(t, 3, watcher.queue.Len())
}

func TestWatchMonitorFallsBackOnForbidden(t *testing.T) {
	unavailable := &atomic.Bool{}
	var warnOnce sync.Once
	forbidden := apierrors.NewForbidden(
		schema.GroupResource{Group: "monitoring.coreos.com", Resource: "podmonitors"},
		"",
		fmt.Errorf("watch denied"),
	)

	watcher, err := watchMonitor(
		RolePodMonitor,
		unavailable,
		&warnOnce,
		func(context.Context, metav1.ListOptions) (watch.Interface, error) {
			return nil, forbidden
		},
		metav1.ListOptions{},
	)
	require.NoError(t, err)
	require.NotNil(t, watcher)
	require.True(t, unavailable.Load())
	watcher.Stop()
}
