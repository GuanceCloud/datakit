package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
)

func TestAttachOperatorUsesEBPFPodView(t *testing.T) {
	var gotView string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotView = r.URL.Query().Get("view")

		pods := []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-a",
					Namespace: "default",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(pods); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer server.Close()

	k := &K8sClient{}
	if err := AttachOperator(k, server.URL); err != nil {
		t.Fatalf("attach operator: %v", err)
	}
	if err := k.ListAllPods(); err != nil {
		t.Fatalf("list all pods: %v", err)
	}
	if gotView != k8sclient.PodListViewEBPFV1 {
		t.Fatalf("view query = %q, want %q", gotView, k8sclient.PodListViewEBPFV1)
	}

	pods, err := k.ListPods("default")
	if err != nil {
		t.Fatalf("list pods from cache: %v", err)
	}
	if len(pods) != 1 {
		t.Fatalf("pods len = %d, want 1", len(pods))
	}
	if pods[0].Name != "pod-a" {
		t.Fatalf("pod name = %q, want pod-a", pods[0].Name)
	}
}
