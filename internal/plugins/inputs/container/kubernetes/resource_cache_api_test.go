// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ktesting "k8s.io/client-go/testing"
)

// Use the vendored object tracker behind a real REST client. This exercises
// HTTP watch cancellation without a Kubernetes cluster or extra dependencies.
type cacheTestClientset struct {
	*kubernetes.Clientset
	ktesting.Fake
	tracker ktesting.ObjectTracker
}

func (c *cacheTestClientset) Tracker() ktesting.ObjectTracker { return c.tracker }

func newCacheTestClientset(t *testing.T, objects ...runtime.Object) *cacheTestClientset {
	t.Helper()
	c := &cacheTestClientset{tracker: ktesting.NewObjectTracker(scheme.Scheme, scheme.Codecs.UniversalDecoder())}
	for _, object := range objects {
		require.NoError(t, c.tracker.Add(object))
	}
	c.AddReactor("*", "*", ktesting.ObjectReaction(c.tracker))
	// The default watcher must be created per request, not shared between terms.
	c.AddWatchReactor("*", func(action ktesting.Action) (bool, watch.Interface, error) {
		w, err := c.tracker.Watch(action.GetResource(), action.GetNamespace())
		return true, w, err
	})
	server := httptest.NewServer(http.HandlerFunc(c.serveHTTP))
	t.Cleanup(server.Close)
	var err error
	c.Clientset, err = kubernetes.NewForConfig(&rest.Config{Host: server.URL, QPS: -1})
	require.NoError(t, err)
	return c
}

func (c *cacheTestClientset) serveHTTP(w http.ResponseWriter, req *http.Request) {
	parts := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	var gv schema.GroupVersion
	if parts[0] == "api" {
		gv.Version = parts[1]
		parts = parts[2:]
	} else {
		gv.Group, gv.Version = parts[1], parts[2]
		parts = parts[3:]
	}
	namespace := ""
	if parts[0] == "namespaces" {
		namespace = parts[1]
		parts = parts[2:]
	}
	resourceName := parts[0]
	gvk := gv.WithKind(map[string]string{
		"pods": "Pod", "nodes": "Node", "deployments": "Deployment", "daemonsets": "DaemonSet", "replicasets": "ReplicaSet",
		"statefulsets": "StatefulSet", "jobs": "Job", "cronjobs": "CronJob", "endpoints": "Endpoints", "services": "Service",
		"persistentvolumes": "PersistentVolume", "persistentvolumeclaims": "PersistentVolumeClaim",
	}[resourceName])
	gvr := gv.WithResource(resourceName)
	options := metav1.ListOptions{FieldSelector: req.URL.Query().Get("fieldSelector"), LabelSelector: req.URL.Query().Get("labelSelector"), ResourceVersion: req.URL.Query().Get("resourceVersion")}
	w.Header().Set("Content-Type", "application/json")
	if req.URL.Query().Get("watch") == "true" {
		stream, err := c.InvokesWatch(ktesting.NewWatchAction(gvr, namespace, options))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer stream.Stop()
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		for {
			select {
			case <-req.Context().Done():
				return
			case event, ok := <-stream.ResultChan():
				if !ok {
					return
				}
				obj := event.Object.DeepCopyObject()
				kinds, _, err := scheme.Scheme.ObjectKinds(obj)
				if err != nil || len(kinds) == 0 {
					return
				}
				obj.GetObjectKind().SetGroupVersionKind(kinds[0])
				if err := json.NewEncoder(w).Encode(map[string]interface{}{"type": event.Type, "object": obj}); err != nil {
					return
				}
				w.(http.Flusher).Flush()
			}
		}
	}
	var obj runtime.Object
	var err error
	switch req.Method {
	case http.MethodGet:
		obj, err = c.Invokes(ktesting.NewListAction(gvr, gvk, namespace, options), nil)
		if err == nil {
			obj.GetObjectKind().SetGroupVersionKind(gv.WithKind(gvk.Kind + "List"))
			if accessor, accessErr := meta.ListAccessor(obj); accessErr == nil {
				accessor.SetResourceVersion("1")
			}
		}
	case http.MethodPut:
		obj, err = scheme.Scheme.New(gvk)
		if err == nil {
			err = json.NewDecoder(req.Body).Decode(obj)
		}
		if err == nil {
			obj, err = c.Invokes(ktesting.NewUpdateAction(gvr, namespace, obj), nil)
		}
		if err == nil {
			obj.GetObjectKind().SetGroupVersionKind(gvk)
		}
	default:
		http.Error(w, "unsupported test request", http.StatusMethodNotAllowed)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		return
	}
}
