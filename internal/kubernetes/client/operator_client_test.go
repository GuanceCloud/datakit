// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestOperatorPodsListViewQuery(t *testing.T) {
	tests := []struct {
		name     string
		view     string
		wantView string
	}{
		{
			name: "default view",
		},
		{
			name:     "ebpf view",
			view:     PodListViewEBPFV1,
			wantView: PodListViewEBPFV1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotPath string
			var gotView string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
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

			c, err := NewKubernetesClientForOperator(server.URL)
			if err != nil {
				t.Fatalf("new operator client: %v", err)
			}
			c.SetPodListView(tc.view)

			podList, err := c.GetPods("").List(context.Background(), metav1.ListOptions{})
			if err != nil {
				t.Fatalf("list pods: %v", err)
			}
			if gotPath != "/v1/cluster/api/v1/pods" {
				t.Fatalf("path = %q, want /v1/cluster/api/v1/pods", gotPath)
			}
			if gotView != tc.wantView {
				t.Fatalf("view query = %q, want %q", gotView, tc.wantView)
			}
			if len(podList.Items) != 1 {
				t.Fatalf("pod list len = %d, want 1", len(podList.Items))
			}
			if podList.Items[0].Name != "pod-a" {
				t.Fatalf("pod name = %q, want pod-a", podList.Items[0].Name)
			}
		})
	}
}
