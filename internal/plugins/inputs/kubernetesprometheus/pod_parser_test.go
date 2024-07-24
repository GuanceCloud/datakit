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

func TestPodMeta(t *testing.T) {
	t.Run("pod-name", func(t *testing.T) {
		obj := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-01",
				Namespace: "pod-ns",
			},
		}

		pr := newPodParser(obj)

		matched, res := pr.matches("__kubernetes_pod_name")
		assert.Equal(t, true, matched)
		assert.Equal(t, "pod-01", res)

		matched, res = pr.matches("__kubernetes_pod_namespace")
		assert.Equal(t, true, matched)
		assert.Equal(t, "pod-ns", res)

		matched, res = pr.matches("nonexistent-key")
		assert.Equal(t, false, matched)
		assert.Equal(t, "", res)
	})

	t.Run("pod-spec-and-tatus", func(t *testing.T) {
		obj := &corev1.Pod{
			Spec: corev1.PodSpec{
				NodeName: "node-name",
			},
			Status: corev1.PodStatus{
				PodIP:  "172.16.10.10",
				HostIP: "172.16.10.11",
			},
		}

		pr := newPodParser(obj)

		matched, res := pr.matches("__kubernetes_pod_node_name")
		assert.Equal(t, true, matched)
		assert.Equal(t, "node-name", res)

		matched, res = pr.matches("__kubernetes_pod_ip")
		assert.Equal(t, true, matched)
		assert.Equal(t, "172.16.10.10", res)

		matched, res = pr.matches("__kubernetes_pod_host_ip")
		assert.Equal(t, true, matched)
		assert.Equal(t, "172.16.10.11", res)
	})

	t.Run("pod-label-and-annotation", func(t *testing.T) {
		obj := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "nginx",
				},
				Annotations: map[string]string{
					"prometheus.io/port": "9090",
				},
			},
		}

		pr := newPodParser(obj)

		matched, res := pr.matches("__kubernetes_pod_label_app")
		assert.Equal(t, true, matched)
		assert.Equal(t, "nginx", res)

		matched, res = pr.matches("__kubernetes_pod_annotation_prometheus.io/port")
		assert.Equal(t, true, matched)
		assert.Equal(t, "9090", res)

		matched, res = pr.matches("__kubernetes_pod_annotation_prometheus.io/port-nonexistent")
		assert.Equal(t, false, matched)
		assert.Equal(t, "", res)
	})

	t.Run("pod-container-port", func(t *testing.T) {
		obj := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "nginx",
						Ports: []corev1.ContainerPort{
							{
								Name:          "metrics",
								ContainerPort: 9090,
							},
							{
								Name:          "http-metrics",
								ContainerPort: 9091,
							},
						},
					},
					{
						Name: "redis",
						Ports: []corev1.ContainerPort{
							{
								Name:          "metrics-port",
								ContainerPort: 9092,
							},
						},
					},
				},
			},
		}

		pr := newPodParser(obj)

		matched, res := pr.matches("__kubernetes_pod_container_nginx_port_metrics_number")
		assert.Equal(t, true, matched)
		assert.Equal(t, "9090", res)

		matched, res = pr.matches("__kubernetes_pod_container_redis_port_metrics-port_number")
		assert.Equal(t, true, matched)
		assert.Equal(t, "9092", res)

		matched, res = pr.matches("__kubernetes_pod_container_nonexistent_port_metrics_number")
		assert.Equal(t, false, matched)
		assert.Equal(t, "", res)

		matched, res = pr.matches("__kubernetes_pod_container_nginx_port_metrics_number_nonexistent")
		assert.Equal(t, false, matched)
		assert.Equal(t, "", res)
	})

	t.Run("pod-scrape", func(t *testing.T) {
		obj := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"prometheus.io/scrape": "true",
				},
			},
		}

		pr := newPodParser(obj)

		should := pr.shouldScrape("true")
		assert.Equal(t, true, should)

		should = pr.shouldScrape("__kubernetes_pod_annotation_prometheus.io/scrape")
		assert.Equal(t, true, should)

		should = pr.shouldScrape("__kubernetes_pod_annotation_prometheus.io/scrape_nonexistent")
		assert.Equal(t, false, should)
	})
}

func TestParsePod(t *testing.T) {
	ins := &Instance{
		Target: Target{
			Scheme:  "https",
			Address: "__kubernetes_pod_ip",
			Port:    "__kubernetes_pod_container_nginx_port_metrics_number",
			Path:    "__kubernetes_pod_annotation_prometheus.io/path",
			Params:  "name=hello&name2=world",
		},
		Custom: Custom{
			Measurement: "__kubernetes_pod_label_app",
			Tags: map[string]string{
				"pod_name":      "__kubernetes_pod_name",
				"pod_namespace": "__kubernetes_pod_namespace",
			},
		},
	}

	obj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-123",
			Namespace: "nginx-ns",
			Labels: map[string]string{
				"app": "nginx",
			},
			Annotations: map[string]string{
				"prometheus.io/path": "/metrics",
			},
		},
		Status: corev1.PodStatus{
			PodIP: "172.16.10.10",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "nginx",
					Ports: []corev1.ContainerPort{
						{
							Name:          "metrics",
							ContainerPort: 9090,
						},
						{
							Name:          "nonexistent",
							ContainerPort: 9091,
						},
					},
				},
			},
		},
	}

	out := &basePromConfig{
		urlstr:      "https://172.16.10.10:9090/metrics?name=hello&name2=world",
		measurement: "nginx",
		tags: map[string]string{
			"instance":      "172.16.10.10:9090",
			"pod_name":      "nginx-123",
			"pod_namespace": "nginx-ns",
		},
	}

	pr := newPodParser(obj)

	res, err := pr.parsePromConfig(ins)
	assert.NoError(t, err)

	assert.Equal(t, out.urlstr, res.urlstr)
	assert.Equal(t, out.measurement, res.measurement)

	assert.Equal(t, len(out.tags), len(res.tags))
	for k := range out.tags {
		assert.Equal(t, out.tags[k], res.tags[k])
	}
}
