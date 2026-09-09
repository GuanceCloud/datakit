// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestPodWatcherRecoversAfterRBACGranted(t *testing.T) {
	var allowed atomic.Bool
	var denied atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !allowed.Load() {
			denied.Add(1)
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(&metav1.Status{TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Status"},
				Status: metav1.StatusFailure, Reason: metav1.StatusReasonForbidden, Code: 403})
			return
		}
		if req.URL.Query().Get("watch") == "true" {
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
			<-req.Context().Done()
			return
		}
		_ = json.NewEncoder(w).Encode(&corev1.PodList{TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "PodList"},
			ListMeta: metav1.ListMeta{ResourceVersion: "1"}, Items: []corev1.Pod{{
				ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default", UID: "app"},
				Spec:       corev1.PodSpec{NodeName: "node-a"},
			}}})
	}))
	t.Cleanup(server.Close)
	client, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
	require.NoError(t, err)
	watcher := newPodWatcher(client, newContainerLogCoordinator(nil), "node-a")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { watcher.start(ctx); close(done) }()
	t.Cleanup(func() { cancel(); <-done })
	require.Eventually(t, func() bool { return denied.Load() > 0 }, time.Second, time.Millisecond)
	require.False(t, watcher.informer.HasSynced())
	allowed.Store(true)
	require.Eventually(t, watcher.informer.HasSynced, 10*time.Second, time.Millisecond)
	pod, err := watcher.podMetadata.Get(context.Background(), "default", "app")
	require.NoError(t, err)
	require.Equal(t, "node-a", pod.Spec.NodeName)
}

func TestPodWatcherCancelsBlockedInitialList(t *testing.T) {
	entered := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		entered <- struct{}{}
		<-req.Context().Done()
	}))
	t.Cleanup(server.Close)
	client, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
	require.NoError(t, err)
	watcher := newPodWatcher(client, newContainerLogCoordinator(nil), "node-a")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { watcher.start(ctx); close(done) }()
	t.Cleanup(func() { cancel(); <-done })
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("LIST did not start")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Pod watcher did not cancel its HTTP request")
	}
}
