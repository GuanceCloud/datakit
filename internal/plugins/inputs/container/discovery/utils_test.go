// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCompletePromConfig(t *testing.T) {
	cfg := `
[inputs.prom]
  urls = ["$IP:8080/metrics"]
  [inputs.prom.tags]
    "pod_name" = "$PODNAME"
    "pod_namespace" = "$NAMESPACE"
    "node_name" = "$NODENAME"
`

	cases := []struct {
		name string
		in   *apicorev1.Pod
		out  string
	}{
		{
			name: "use PodIP",
			in: &apicorev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake_name",
					Namespace: "fake_namespace",
				},
				Spec: apicorev1.PodSpec{
					NodeName: "fake_node_name",
				},
				Status: apicorev1.PodStatus{
					PodIP: "172.16.0.2",
				},
			},

			out: `
[inputs.prom]
  urls = ["172.16.0.2:8080/metrics"]
  [inputs.prom.tags]
    "pod_name" = "fake_name"
    "pod_namespace" = "fake_namespace"
    "node_name" = "fake_node_name"
`,
		},

		{
			name: "use PodIPs[1]",
			in: &apicorev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake_name",
					Namespace: "fake_namespace",
					Annotations: map[string]string{
						"datakit/prom.instances.ip_index": "1",
					},
				},
				Spec: apicorev1.PodSpec{
					NodeName: "fake_node_name",
				},
				Status: apicorev1.PodStatus{
					PodIP: "172.16.0.2",
					PodIPs: []apicorev1.PodIP{
						{IP: "172.16.0.3"},
						{IP: "172.16.0.4"},
						{IP: "172.16.0.5"},
					},
				},
			},

			out: `
[inputs.prom]
  urls = ["172.16.0.4:8080/metrics"]
  [inputs.prom.tags]
    "pod_name" = "fake_name"
    "pod_namespace" = "fake_namespace"
    "node_name" = "fake_node_name"
`,
		},

		{
			name: "invalid PodIPs idx for 'X'",
			in: &apicorev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake_name",
					Namespace: "fake_namespace",
					Annotations: map[string]string{
						"datakit/prom.instances.ip_index": "X",
					},
				},
				Spec: apicorev1.PodSpec{
					NodeName: "fake_node_name",
				},
				Status: apicorev1.PodStatus{
					PodIP: "172.16.0.2",
					PodIPs: []apicorev1.PodIP{
						{IP: "172.16.0.3"},
						{IP: "172.16.0.4"},
						{IP: "172.16.0.5"},
					},
				},
			},

			out: `
[inputs.prom]
  urls = ["172.16.0.2:8080/metrics"]
  [inputs.prom.tags]
    "pod_name" = "fake_name"
    "pod_namespace" = "fake_namespace"
    "node_name" = "fake_node_name"
`,
		},

		{
			name: "invalid PodIPs idx for '100'",
			in: &apicorev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake_name",
					Namespace: "fake_namespace",
					Annotations: map[string]string{
						"datakit/prom.instances.ip_index": "100",
					},
				},
				Spec: apicorev1.PodSpec{
					NodeName: "fake_node_name",
				},
				Status: apicorev1.PodStatus{
					PodIP: "172.16.0.2",
					PodIPs: []apicorev1.PodIP{
						{IP: "172.16.0.3"},
						{IP: "172.16.0.4"},
						{IP: "172.16.0.5"},
					},
				},
			},

			out: `
[inputs.prom]
  urls = ["172.16.0.2:8080/metrics"]
  [inputs.prom.tags]
    "pod_name" = "fake_name"
    "pod_namespace" = "fake_namespace"
    "node_name" = "fake_node_name"
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := completePromConfig(cfg, tc.in)
			assert.Equal(t, tc.out, res)
		})
	}
}
