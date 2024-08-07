// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeMeta(t *testing.T) {
	t.Run("node-name", func(t *testing.T) {
		obj := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-01",
			},
		}

		pr := newNodeParser(obj)

		matched, res := pr.matches("__kubernetes_node_name")
		assert.Equal(t, true, matched)
		assert.Equal(t, "node-01", res)
	})

	t.Run("node-label-and-annotation", func(t *testing.T) {
		obj := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"hostname": "node-master",
				},
				Annotations: map[string]string{
					"os": "linux",
				},
			},
		}

		pr := newNodeParser(obj)

		matched, res := pr.matches("__kubernetes_node_label_hostname")
		assert.Equal(t, true, matched)
		assert.Equal(t, "node-master", res)

		matched, res = pr.matches("__kubernetes_node_annotation_os")
		assert.Equal(t, true, matched)
		assert.Equal(t, "linux", res)

		matched, res = pr.matches("__kubernetes_pod_annotation_arch-nonexistent")
		assert.Equal(t, false, matched)
		assert.Equal(t, "", res)
	})

	t.Run("node-address", func(t *testing.T) {
		obj := &corev1.Node{
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{
						Type:    corev1.NodeHostName,
						Address: "node-hostname",
					},
					{
						Type:    corev1.NodeInternalIP,
						Address: "172.16.10.10",
					},
					{
						Type:    corev1.NodeExternalIP,
						Address: "172.16.10.11",
					},
				},
			},
		}

		pr := newNodeParser(obj)

		matched, res := pr.matches("__kubernetes_node_address_Hostname")
		assert.Equal(t, true, matched)
		assert.Equal(t, "node-hostname", res)

		matched, res = pr.matches("__kubernetes_node_address_InternalIP")
		assert.Equal(t, true, matched)
		assert.Equal(t, "172.16.10.10", res)

		matched, res = pr.matches("__kubernetes_node_address_ExternalIP")
		assert.Equal(t, true, matched)
		assert.Equal(t, "172.16.10.11", res)
	})

	t.Run("node-kubelet-port", func(t *testing.T) {
		obj := &corev1.Node{
			Status: corev1.NodeStatus{
				DaemonEndpoints: corev1.NodeDaemonEndpoints{
					KubeletEndpoint: corev1.DaemonEndpoint{
						Port: 10250,
					},
				},
			},
		}

		pr := newNodeParser(obj)

		matched, res := pr.matches("__kubernetes_node_kubelet_endpoint_port")
		assert.Equal(t, true, matched)
		assert.Equal(t, "10250", res)
	})

	t.Run("node-scrape", func(t *testing.T) {
		obj := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"prometheus.io/scrape": "true",
				},
			},
		}

		pr := newNodeParser(obj)

		should := pr.shouldScrape("true")
		assert.Equal(t, true, should)

		should = pr.shouldScrape("__kubernetes_node_annotation_prometheus.io/scrape")
		assert.Equal(t, true, should)

		should = pr.shouldScrape("__kubernetes_node_annotation_prometheus.io/scrape_nonexistent")
		assert.Equal(t, false, should)
	})
}

func TestParseNode(t *testing.T) {
	ins := &Instance{
		Target: Target{
			Scheme:  "https",
			Address: "__kubernetes_node_address_InternalIP",
			Port:    "__kubernetes_node_kubelet_endpoint_port",
			Path:    "__kubernetes_node_annotation_prometheus.io/path",
			Params:  "name=hello&name2=world",
		},
		Custom: Custom{
			Measurement: "__kubernetes_node_label_app",
			Tags: map[string]string{
				"instance": "__kubernetes_mate_instance",
				"host":     "__kubernetes_mate_host",
				"hostname": "__kubernetes_node_address_Hostname",
			},
		},
	}

	obj := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-01",
			Labels: map[string]string{
				"app": "node",
			},
			Annotations: map[string]string{
				"prometheus.io/path": "/metrics",
			},
		},
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{
				{
					Type:    corev1.NodeHostName,
					Address: "node-hostname",
				},
				{
					Type:    corev1.NodeInternalIP,
					Address: "172.16.10.10",
				},
			},
			DaemonEndpoints: corev1.NodeDaemonEndpoints{
				KubeletEndpoint: corev1.DaemonEndpoint{
					Port: 10250,
				},
			},
		},
	}

	out := &basePromConfig{
		urlstr:      "https://172.16.10.10:10250/metrics?name=hello&name2=world",
		measurement: "node",
		tags: map[string]string{
			"instance": "172.16.10.10:10250",
			"host":     "172.16.10.10",
			"hostname": "node-hostname",
		},
	}

	pr := newNodeParser(obj)

	res, err := pr.parsePromConfig(ins)
	assert.NoError(t, err)

	assert.Equal(t, out.urlstr, res.urlstr)
	assert.Equal(t, out.measurement, res.measurement)

	assert.Equal(t, len(out.tags), len(res.tags))
	for k := range out.tags {
		assert.Equal(t, out.tags[k], res.tags[k])
	}
}
